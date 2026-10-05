package compliance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thomasvincent/mantl/pkg/framework"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sort"
	"strings"

	complianceapi "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// FindingReconciler reconciles Kyverno PolicyReports into Mantl Findings
type FindingReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	FrameworkDir   string
	RequireReports bool
	Namespace      string
}

// policyReportGVK is the GroupVersionKind for Kyverno PolicyReports.
var policyReportGVK = schema.GroupVersionKind{
	Group:   "wgpolicyk8s.io",
	Version: "v1alpha2",
	Kind:    "PolicyReport",
}

// +kubebuilder:rbac:groups=compliance.mantl.io,resources=findings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wgpolicyk8s.io,resources=policyreports;clusterpolicyreports,verbs=get;list;watch

func (r *FindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	report := &unstructured.Unstructured{}
	gvk := policyReportGVK
	if req.Namespace == "" {
		gvk.Kind = "ClusterPolicyReport"
	}
	report.SetGroupVersionKind(gvk)
	destination := req.Namespace
	if destination == "" {
		destination = r.Namespace
		if destination == "" {
			destination = "mantl-compliance-system"
		}
	}
	source := digest(gvk.Kind + "/" + req.Namespace + "/" + req.Name)[:32]
	var previous complianceapi.FindingList
	if err := r.List(ctx, &previous, client.InNamespace(destination), client.MatchingLabels{"mantl.io/report": source}); err != nil {
		return ctrl.Result{}, fmt.Errorf("list report findings: %w", err)
	}
	err := r.Get(ctx, req.NamespacedName, report)
	if err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		for i := range previous.Items {
			f := &previous.Items[i]
			f.Spec.Status = "unknown"
			f.Spec.Message = "Source PolicyReport deleted; current evaluation unavailable"
			if err := r.Update(ctx, f); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	var profiles complianceapi.ComplianceProfileList
	if r.FrameworkDir != "" {
		if err := r.List(ctx, &profiles); err != nil {
			return ctrl.Result{}, err
		}
	}
	seen := map[string]bool{}
	results, found, err := unstructured.NestedSlice(report.Object, "results")
	if err != nil || !found {
		return ctrl.Result{}, fmt.Errorf("PolicyReport has invalid or missing results")
	}
	for _, item := range results {
		result, ok := item.(map[string]interface{})
		if !ok {
			return ctrl.Result{}, fmt.Errorf("PolicyReport contains malformed result")
		}
		state, _, _ := unstructured.NestedString(result, "result")
		policy, _, _ := unstructured.NestedString(result, "policy")
		rule, _, _ := unstructured.NestedString(result, "rule")
		if policy == "" || rule == "" || state == "" {
			return ctrl.Result{}, fmt.Errorf("PolicyReport result lacks identity or evaluation")
		}
		resourceItems, _, resourceErr := unstructured.NestedSlice(result, "resources")
		if resourceErr != nil {
			return ctrl.Result{}, resourceErr
		}
		message, _, _ := unstructured.NestedString(result, "message")
		severity, _, _ := unstructured.NestedString(result, "severity")
		controls := []string{}
		frameworks := []string{}
		for _, profile := range profiles.Items {
			if req.Namespace != "" {
				namespaces := profile.Spec.Namespaces
				if len(namespaces) == 0 {
					namespaces = []string{profile.Namespace}
				}
				allowed := false
				for _, ns := range namespaces {
					if ns == req.Namespace {
						allowed = true
					}
				}
				if !allowed {
					continue
				}
			}
			fw, loadErr := framework.Load(r.FrameworkDir, profile.Spec.Framework, profile.Spec.Version)
			if loadErr != nil {
				return ctrl.Result{}, loadErr
			}
			for _, c := range fw.Controls {
				if !framework.Selected(c.ID, profile.Spec.IncludeControls) {
					continue
				}
				for _, m := range c.Mappings {
					if m.PolicyRef == nil {
						continue
					}
					template := m.PolicyRef.Template
					alias := templateAliases[template]
					if policy == template || policy == alias || policy == profile.Spec.Framework+"-"+template {
						controls = append(controls, c.ID)
						frameworks = append(frameworks, profile.Spec.Framework)
					}
				}
			}
		}
		sort.Strings(controls)
		sort.Strings(frameworks)
		controlID := policy
		frameworkID := "kyverno"
		if len(controls) > 0 {
			controlID = strings.Join(controls, ",")
			frameworkID = strings.Join(frameworks, ",")
		}
		for _, resourceItem := range resourceItems {
			res, ok := resourceItem.(map[string]interface{})
			if !ok {
				return ctrl.Result{}, fmt.Errorf("invalid PolicyReport resource")
			}
			kind, _, _ := unstructured.NestedString(res, "kind")
			ns, _, _ := unstructured.NestedString(res, "namespace")
			name, _, _ := unstructured.NestedString(res, "name")
			uid, _, _ := unstructured.NestedString(res, "uid")
			if name == "" || kind == "" {
				return ctrl.Result{}, fmt.Errorf("PolicyReport resource lacks identity")
			}
			findingName := buildFindingName(policy, rule, ns, name) + "-" + digest(source + kind + uid)[:12]
			if len(findingName) > 253 {
				findingName = strings.Trim(findingName[:220], "-") + "-" + digest(findingName)[:24]
			}
			seen[findingName] = true
			var f complianceapi.Finding
			getErr := r.Get(ctx, types.NamespacedName{Name: findingName, Namespace: destination}, &f)
			if getErr != nil && !errors.IsNotFound(getErr) {
				return ctrl.Result{}, getErr
			}
			if errors.IsNotFound(getErr) && state == "pass" {
				continue
			}
			f.Name = findingName
			f.Namespace = destination
			if f.Labels == nil {
				f.Labels = map[string]string{}
			}
			f.Labels["mantl.io/report"] = source
			status := state
			if state == "pass" {
				status = "resolved"
			} else if state != "fail" && state != "warn" {
				status = "unknown"
			}
			f.Spec = complianceapi.FindingSpec{ID: findingName, ControlID: controlID, Framework: frameworkID, Severity: normalizeSeverity(severity), Resource: fmt.Sprintf("%s/%s/%s", kind, ns, name), Message: message, Status: status}
			if errors.IsNotFound(getErr) {
				err = r.Create(ctx, &f)
			} else {
				err = r.Update(ctx, &f)
			}
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	for i := range previous.Items {
		f := &previous.Items[i]
		if !seen[f.Name] {
			f.Spec.Status = "resolved"
			f.Spec.Message = "Violation absent from current PolicyReport"
			if err := r.Update(ctx, f); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}
func digest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

// buildFindingName produces a deterministic, DNS-safe Finding name from its key fields.
func buildFindingName(policy, rule, namespace, resourceName string) string {
	name := fmt.Sprintf("%s-%s-%s-%s", policy, rule, namespace, resourceName)
	// Trim to stay within the 253-char Kubernetes name limit and sanitize.
	if len(name) > 253 {
		name = name[:220] + "-" + digest(name)[:24]
	}
	// Replace characters that are not allowed in DNS names with dashes.
	safe := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			safe = append(safe, c)
		} else if c >= 'A' && c <= 'Z' {
			safe = append(safe, c+32) // toLower
		} else {
			safe = append(safe, '-')
		}
	}
	// Strip leading/trailing dashes
	result := string(safe)
	for len(result) > 0 && result[0] == '-' {
		result = result[1:]
	}
	for len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}
	return result
}

// normalizeSeverity maps Kyverno severity levels to Mantl severity levels.
func normalizeSeverity(severity string) string {
	switch severity {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	default:
		return "medium"
	}
}

// policyReportAvailable reports whether the PolicyReport CRD is registered. It
// distinguishes a genuinely absent type (no-match) from a discovery failure:
// only a no-match is a clean "absent" (false, nil). Any other error is returned
// so the caller fails closed rather than silently disabling finding aggregation,
// which would make a non-compliant cluster look clean.
func policyReportAvailable(mapper meta.RESTMapper) (bool, error) {
	_, err := mapper.RESTMapping(policyReportGVK.GroupKind(), policyReportGVK.Version)
	switch {
	case err == nil:
		return true, nil
	case meta.IsNoMatchError(err):
		return false, nil
	default:
		return false, err
	}
}

func (r *FindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// PolicyReport (Kyverno) is an optional dependency. If its CRD is absent at
	// startup, skip the watch rather than crash the manager; the controller is a
	// no-op until Kyverno is installed and the operator is restarted. A discovery
	// failure (not a clean absence) is propagated so the manager fails to start
	// instead of silently disabling violation aggregation.
	available, err := policyReportAvailable(mgr.GetRESTMapper())
	if err != nil {
		return fmt.Errorf("checking PolicyReport CRD availability: %w", err)
	}
	reportWatch := prometheus.NewGauge(prometheus.GaugeOpts{Name: "mantl_compliance_report_watch_ready", Help: "1 when PolicyReport CRD was available at operator startup"})
	metrics.Registry.MustRegister(reportWatch)
	if available {
		reportWatch.Set(1)
	}
	if !available {
		if r.RequireReports {
			return fmt.Errorf("PolicyReport CRD is required; retry startup after Kyverno is installed")
		}
		mgr.GetLogger().Info("PolicyReport CRD not registered; FindingReconciler disabled until Kyverno is installed and the operator restarts")
		return nil
	}
	report := &unstructured.Unstructured{}
	report.SetGroupVersionKind(policyReportGVK)
	builder := ctrl.NewControllerManagedBy(mgr).Named("finding-aggregator").Watches(report, &handler.EnqueueRequestForObject{})
	cluster := &unstructured.Unstructured{}
	clusterGVK := policyReportGVK
	clusterGVK.Kind = "ClusterPolicyReport"
	cluster.SetGroupVersionKind(clusterGVK)
	if _, err := mgr.GetRESTMapper().RESTMapping(clusterGVK.GroupKind(), clusterGVK.Version); err == nil {
		builder = builder.Watches(cluster, &handler.EnqueueRequestForObject{})
	} else if !meta.IsNoMatchError(err) {
		return err
	}
	return builder.Complete(r)
}
