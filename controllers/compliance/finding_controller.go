package compliance

import (
	"context"
	"fmt"

	complianceapi "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// FindingReconciler reconciles Kyverno PolicyReports into Mantl Findings
type FindingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// policyReportGVK is the GroupVersionKind for Kyverno PolicyReports.
var policyReportGVK = schema.GroupVersionKind{
	Group:   "wgpolicyk8s.io",
	Version: "v1alpha2",
	Kind:    "PolicyReport",
}

// +kubebuilder:rbac:groups=compliance.mantl.io,resources=findings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wgpolicyk8s.io,resources=policyreports,verbs=get;list;watch

func (r *FindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling PolicyReport signals into Findings", "Request", req.NamespacedName)

	// Fetch the PolicyReport as unstructured to avoid requiring the wgpolicyk8s.io
	// CRDs to be registered in the scheme at operator startup.
	report := &unstructured.Unstructured{}
	report.SetGroupVersionKind(policyReportGVK)
	if err := r.Get(ctx, req.NamespacedName, report); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Extract results from the unstructured PolicyReport
	results, _, _ := unstructured.NestedSlice(report.Object, "results")
	for _, item := range results {
		result, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		resultStatus, _, _ := unstructured.NestedString(result, "result")
		if resultStatus != "fail" {
			continue
		}

		policy, _, _ := unstructured.NestedString(result, "policy")
		rule, _, _ := unstructured.NestedString(result, "rule")
		message, _, _ := unstructured.NestedString(result, "message")
		severity, _, _ := unstructured.NestedString(result, "severity")

		resources, _, _ := unstructured.NestedSlice(result, "resources")
		for _, resItem := range resources {
			resMap, ok := resItem.(map[string]interface{})
			if !ok {
				continue
			}
			resKind, _, _ := unstructured.NestedString(resMap, "kind")
			resNamespace, _, _ := unstructured.NestedString(resMap, "namespace")
			resName, _, _ := unstructured.NestedString(resMap, "name")

			findingName := buildFindingName(policy, rule, resNamespace, resName)
			finding := &complianceapi.Finding{}
			err := r.Get(ctx, types.NamespacedName{Name: findingName, Namespace: req.Namespace}, finding)
			if err != nil && !errors.IsNotFound(err) {
				l.Error(err, "failed to fetch Finding", "name", findingName)
				continue
			}

			if errors.IsNotFound(err) {
				// Create a new Finding CRD for this policy violation
				newFinding := &complianceapi.Finding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      findingName,
						Namespace: req.Namespace,
						Labels: map[string]string{
							"mantl.io/policy": policy,
							"mantl.io/rule":   rule,
						},
					},
					Spec: complianceapi.FindingSpec{
						ControlID: policy,
						Framework: "kyverno",
						Severity:  normalizeSeverity(severity),
						Resource:  fmt.Sprintf("%s/%s/%s", resKind, resNamespace, resName),
						Message:   message,
						Status:    "fail",
					},
				}
				if createErr := r.Create(ctx, newFinding); createErr != nil {
					l.Error(createErr, "failed to create Finding", "name", findingName)
				} else {
					l.Info("Created Finding for policy violation", "finding", findingName, "policy", policy)
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// buildFindingName produces a deterministic, DNS-safe Finding name from its key fields.
func buildFindingName(policy, rule, namespace, resourceName string) string {
	name := fmt.Sprintf("%s-%s-%s-%s", policy, rule, namespace, resourceName)
	// Trim to stay within the 253-char Kubernetes name limit and sanitize.
	if len(name) > 253 {
		name = name[:253]
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
	if !available {
		mgr.GetLogger().Info("PolicyReport CRD not registered; FindingReconciler disabled until Kyverno is installed and the operator restarts")
		return nil
	}
	report := &unstructured.Unstructured{}
	report.SetGroupVersionKind(policyReportGVK)
	return ctrl.NewControllerManagedBy(mgr).
		Named("finding-aggregator").
		Watches(report, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
