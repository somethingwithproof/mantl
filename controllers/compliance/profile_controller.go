package compliance

import (
	"context"

	"path/filepath"

	"github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/framework"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type Framework = framework.Framework
type Control = framework.Control
type PolicyRef = framework.PolicyRef
type Mapping = framework.Mapping
type EvidenceCollector = framework.EvidenceCollector

// ComplianceProfileReconciler reconciles a ComplianceProfile object
type ComplianceProfileReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	FrameworkDir string
}

// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceprofiles/status,verbs=get;update;patch
// The operator only reads ClusterPolicies; GitOps owns their application (ADR 006).
// +kubebuilder:rbac:groups=kyverno.io,resources=clusterpolicies,verbs=get;list;watch

func (r *ComplianceProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// 1. Fetch the ComplianceProfile
	var profile v1alpha1.ComplianceProfile
	if err := r.Get(ctx, req.NamespacedName, &profile); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("Reconciling ComplianceProfile", "Name", profile.Name, "Framework", profile.Spec.Framework)

	// 2. Load the Framework mapping file
	loaded, err := framework.LoadSelected(r.FrameworkDir, profile.Spec.Framework, profile.Spec.Version, profile.Spec.BundleDigest, profile.Spec.IncludeControls)
	if err != nil {
		profile.Status.State = "Invalid"
		profile.Status.ActivePolicies = 0
		profile.Status.UnresolvedTemplates = []string{"framework configuration is invalid or unavailable"}
		if statusErr := r.Status().Update(ctx, &profile); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}
	catalog := *loaded

	var referenced []string
	for _, control := range catalog.Controls {
		if !framework.Selected(control.ID, profile.Spec.IncludeControls) {
			continue
		}
		for _, m := range control.Mappings {
			if m.PolicyRef != nil && m.PolicyRef.Template != "" {
				referenced = append(referenced, m.PolicyRef.Template)
			}
		}
	}

	index, err := buildPolicyIndex([]string{
		filepath.Join(r.FrameworkDir, "..", "policies", "runtime"),
		filepath.Join(r.FrameworkDir, profile.Spec.Framework, "policies"),
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	resolved, unresolved := resolveTemplates(referenced, index, templateAliases)
	l.Info("Resolved framework policies", "framework", catalog.Name,
		"resolved", len(resolved), "unresolved", len(unresolved))

	profile.Status.State = "Active"
	if len(unresolved) > 0 {
		profile.Status.State = "CoverageGap"
	}
	profile.Status.ActivePolicies = int32(len(resolved))
	profile.Status.UnresolvedTemplates = unresolved
	if err := r.Status().Update(ctx, &profile); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComplianceProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ComplianceProfile{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
