package compliance

import (
	"context"
	"fmt"

	"github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ComplianceProfileReconciler reconciles a ComplianceProfile object
type ComplianceProfileReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kyverno.io,resources=clusterpolicies,verbs=get;list;watch;create;update;patch;delete

func (r *ComplianceProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. Fetch the ComplianceProfile
	var profile v1alpha1.ComplianceProfile
	if err := r.Get(ctx, req.NamespacedName, &profile); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Reconciling ComplianceProfile", "Name", profile.Name)

	// 2. Logic to map Framework -> Controls -> Kyverno Policies
	// In a real implementation, this would look up the Framework spec
	// and generate/apply Kyverno ClusterPolicies mapped to its controls.
	// We'll simulate the mapping phase here.

	mappedPolicies := []string{"disallow-default-namespace", "require-labels", "restrict-image-registries"}
	log.Info("Mapped profile to Kyverno policies", "policies", mappedPolicies)

	// Here we would apply those policies to the cluster via the client.
	// For this implementation, we will log the expansion and proceed.
	for _, p := range mappedPolicies {
		log.Info("Would enforce policy", "policyName", p)
	}

	// 3. Update the Status to active
	// This represents that the controller has successfully compiled
	// the abstract profile into concrete technical enforcement.
	// profile.Status.State = "Active"
	// profile.Status.ActivePolicies = int32(len(mappedPolicies))
	// if err := r.Status().Update(ctx, &profile); err != nil {
	// 	return ctrl.Result{}, err
	// }

	log.Info("ComplianceProfile reconciled successfully")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComplianceProfileReconciler) SetupWithManager(mgr ctrl.Session) error {
	// To actually watch for changes to the generated Kyverno policies, we would add:
	// .Owns(&kyverno.ClusterPolicy{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ComplianceProfile{}).
		Complete(r)
}
