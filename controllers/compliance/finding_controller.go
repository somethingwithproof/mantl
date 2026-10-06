// SPDX-License-Identifier: Apache-2.0

package compliance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/framework"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sort"
	"strings"
	"time"

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
	Store     evidence.Store
	ClusterID string
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
	destination := reportDestination(req.Namespace, r.Namespace)
	source := digest(gvk.Kind + "/" + req.Namespace + "/" + req.Name)[:32]
	var previous complianceapi.FindingList
	if err := r.List(ctx, &previous, client.InNamespace(destination), client.MatchingLabels{"mantl.io/report": source}); err != nil {
		return ctrl.Result{}, fmt.Errorf("list report findings: %w", err)
	}
	err := r.Get(ctx, req.NamespacedName, report)
	if err != nil {
		return ctrl.Result{}, r.unavailableReport(ctx, previous.Items, err)
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
		if err := r.reconcileReportResult(ctx, req.Namespace, destination, source, item, profiles.Items, seen); err != nil {
			return ctrl.Result{}, err
		}
	}
	if err := r.markPreviousFindings(ctx, previous.Items, seen, "resolved", "Violation absent from current PolicyReport"); err != nil {
		return ctrl.Result{}, err
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

func (r *FindingReconciler) controlsForPolicy(namespace, policy string, profiles []complianceapi.ComplianceProfile) ([]string, []string, error) {
	controls, frameworks := []string{}, []string{}
	for _, profile := range profiles {
		if !framework.NamespaceSelected(profile.Spec.Namespaces, profile.Namespace, namespace) {
			continue
		}
		fw, loadErr := framework.LoadSelected(r.FrameworkDir, profile.Spec.Framework, profile.Spec.Version, profile.Spec.BundleDigest, profile.Spec.IncludeControls)
		if loadErr != nil {
			return nil, nil, loadErr
		}
		matched := framework.PolicyControls(fw, profile.Spec.IncludeControls, profile.Spec.Framework, policy, templateAliases)
		controls = append(controls, matched...)
		for range matched {
			frameworks = append(frameworks, profile.Spec.Framework)
		}
	}
	return controls, frameworks, nil
}

type findingReportResult struct {
	policy, rule, state, message, severity, control, framework string
}

func (r *FindingReconciler) persistReportResource(ctx context.Context, destination, source string, details findingReportResult, resourceItem interface{}) (string, error) {
	policy, rule, state, message, severity := details.policy, details.rule, details.state, details.message, details.severity
	controlID, frameworkID := details.control, details.framework
	res, ok := resourceItem.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid PolicyReport resource")
	}
	kind, _, _ := unstructured.NestedString(res, "kind")
	ns, _, _ := unstructured.NestedString(res, "namespace")
	name, _, _ := unstructured.NestedString(res, "name")
	uid, _, _ := unstructured.NestedString(res, "uid")
	if name == "" || kind == "" {
		return "", fmt.Errorf("PolicyReport resource lacks identity")
	}
	findingName := buildFindingName(policy, rule, ns, name) + "-" + digest(source + kind + uid)[:12]
	if len(findingName) > 253 {
		findingName = strings.Trim(findingName[:220], "-") + "-" + digest(findingName)[:24]
	}
	var f complianceapi.Finding
	getErr := r.Get(ctx, types.NamespacedName{Name: findingName, Namespace: destination}, &f)
	if getErr != nil && !errors.IsNotFound(getErr) {
		return "", getErr
	}
	if errors.IsNotFound(getErr) && state == "pass" {
		return findingName, nil
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
	if f.Annotations == nil {
		f.Annotations = map[string]string{}
	}
	f.Annotations[observedAtAnnotation] = time.Now().UTC().Format(time.RFC3339Nano)
	f.Spec = complianceapi.FindingSpec{ID: findingName, ControlID: controlID, Framework: frameworkID, Severity: normalizeSeverity(severity), Resource: fmt.Sprintf("%s/%s/%s", kind, ns, name), Message: message, Status: status}
	err := r.persistFinding(ctx, &f, errors.IsNotFound(getErr))
	if err != nil {
		return "", err
	}
	return findingName, nil
}

func (r *FindingReconciler) markPreviousFindings(ctx context.Context, previous []complianceapi.Finding, seen map[string]bool, state, message string) error {
	for i := range previous {
		f := &previous[i]
		if seen != nil && seen[f.Name] {
			continue
		}
		if f.Annotations == nil {
			f.Annotations = map[string]string{}
		}
		f.Annotations[observedAtAnnotation] = time.Now().UTC().Format(time.RFC3339Nano)
		f.Spec.Status, f.Spec.Message = state, message
		if err := r.persistFinding(ctx, f, false); err != nil {
			return err
		}
	}
	return nil
}

func (r *FindingReconciler) reconcileReportResult(ctx context.Context, namespace, destination, source string, item interface{}, profiles []complianceapi.ComplianceProfile, seen map[string]bool) error {

	details, resourceItems, err := parseFindingResult(item)
	if err != nil {
		return err
	}
	policy := details.policy
	controls, frameworks, mappingErr := r.controlsForPolicy(namespace, policy, profiles)
	if mappingErr != nil {
		return mappingErr
	}
	sort.Strings(controls)
	sort.Strings(frameworks)
	controlID := policy
	frameworkID := "kyverno"
	if len(controls) > 0 {
		controlID = strings.Join(controls, ",")
		frameworkID = strings.Join(frameworks, ",")
	}
	details.control, details.framework = controlID, frameworkID
	for _, resourceItem := range resourceItems {
		name, err := r.persistReportResource(ctx, destination, source, details, resourceItem)
		if err != nil {
			return err
		}
		seen[name] = true
	}
	return nil
}

func parseFindingResult(item interface{}) (findingReportResult, []interface{}, error) {
	result, ok := item.(map[string]interface{})
	if !ok {
		return findingReportResult{}, nil, fmt.Errorf("PolicyReport contains malformed result")
	}
	state, _, _ := unstructured.NestedString(result, "result")
	policy, _, _ := unstructured.NestedString(result, "policy")
	rule, _, _ := unstructured.NestedString(result, "rule")
	if policy == "" || rule == "" || state == "" {
		return findingReportResult{}, nil, fmt.Errorf("PolicyReport result lacks identity or evaluation")
	}
	resourceItems, _, resourceErr := unstructured.NestedSlice(result, "resources")
	if resourceErr != nil {
		return findingReportResult{}, nil, resourceErr
	}
	message, _, _ := unstructured.NestedString(result, "message")
	severity, _, _ := unstructured.NestedString(result, "severity")
	return findingReportResult{policy: policy, rule: rule, state: state, message: message, severity: severity}, resourceItems, nil
}

func reportDestination(namespace, configured string) string {
	if namespace != "" {
		return namespace
	}
	if configured != "" {
		return configured
	}
	return "mantl-compliance-system"
}

func (r *FindingReconciler) unavailableReport(ctx context.Context, previous []complianceapi.Finding, err error) error {
	if !errors.IsNotFound(err) {
		return err
	}
	return r.markPreviousFindings(ctx, previous, nil, "unknown", "Source PolicyReport deleted; current evaluation unavailable")
}
