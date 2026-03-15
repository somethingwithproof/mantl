package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ComplianceAuditReconciler reconciles a ComplianceAudit object
type ComplianceAuditReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	EvidenceBucket string
}

// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceaudits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceaudits/status,verbs=get;update;patch

func (r *ComplianceAuditReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// 1. Fetch the Audit
	var audit v1alpha1.ComplianceAudit
	if err := r.Get(ctx, req.NamespacedName, &audit); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if audit.Status.Phase == "Completed" || audit.Status.Phase == "Failed" {
		return ctrl.Result{}, nil
	}

	l.Info("Starting Compliance Audit", "Profile", audit.Spec.Profile)

	// 2. Update status to Running
	audit.Status.Phase = "Running"
	now := metav1.Now()
	audit.Status.StartTime = &now
	if err := r.Status().Update(ctx, &audit); err != nil {
		return ctrl.Result{}, err
	}

	// 3. Trigger Evidence Collection (Simulated for one resource)
	// In reality, we'd loop over all resources mapped in the profile
	l.Info("Collecting evidence for audit", "Resource", "NetworkPolicies")
	
	snap, err := evidence.CaptureResource("networkpolicies", "", "")
	if err != nil {
		l.Error(err, "Failed to capture evidence")
		audit.Status.Phase = "Failed"
		r.Status().Update(ctx, &audit)
		return ctrl.Result{}, err
	}

	// 4. Upload to S3
	uri, err := evidence.UploadToS3(ctx, snap, r.EvidenceBucket)
	if err != nil {
		l.Error(err, "Failed to upload evidence to S3")
		audit.Status.Phase = "Failed"
		r.Status().Update(ctx, &audit)
		return ctrl.Result{}, err
	}

	// 5. Finalize Audit
	audit.Status.Phase = "Completed"
	end := metav1.Now()
	audit.Status.EndTime = &end
	audit.Status.EvidenceURI = uri
	audit.Status.FindingCount = 1 // Simplified
	
	if err := r.Status().Update(ctx, &audit); err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Compliance Audit completed successfully", "Evidence", uri)

	return ctrl.Result{}, nil
}

func (r *ComplianceAuditReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ComplianceAudit{}).
		Complete(r)
}
