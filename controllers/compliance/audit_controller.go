package compliance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

// ComplianceAuditReconciler reconciles a ComplianceAudit object
type ComplianceAuditReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	EvidenceBucket string
	FrameworkDir   string
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

	// 2. Fetch the associated Profile to find the Framework
	var profile v1alpha1.ComplianceProfile
	if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: audit.Spec.Profile}, &profile); err != nil {
		l.Error(err, "Failed to find ComplianceProfile", "profile", audit.Spec.Profile)
		return ctrl.Result{}, err
	}

	l.Info("Starting Compliance Audit", "Profile", audit.Spec.Profile, "Framework", profile.Spec.Framework)

	// 3. Update status to Running
	audit.Status.Phase = "Running"
	now := metav1.Now()
	audit.Status.StartTime = &now
	if err := r.Status().Update(ctx, &audit); err != nil {
		return ctrl.Result{}, err
	}

	// 4. Load the Framework mapping file
	if !validFrameworkName.MatchString(profile.Spec.Framework) {
		audit.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, &audit); statusErr != nil {
			l.Error(statusErr, "Failed to update audit status to Failed")
		}
		return ctrl.Result{}, fmt.Errorf("invalid framework name: %q", profile.Spec.Framework)
	}
	frameworkFile := filepath.Join(r.FrameworkDir, fmt.Sprintf("%s.yaml", profile.Spec.Framework))
	frameworkFile = filepath.Clean(frameworkFile)
	if !strings.HasPrefix(frameworkFile, filepath.Clean(r.FrameworkDir)+string(os.PathSeparator)) {
		audit.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, &audit); statusErr != nil {
			l.Error(statusErr, "Failed to update audit status to Failed")
		}
		return ctrl.Result{}, fmt.Errorf("invalid framework name: path traversal detected")
	}
	data, err := os.ReadFile(frameworkFile)
	if err != nil {
		l.Error(err, "Failed to read framework file", "file", frameworkFile)
		audit.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, &audit); statusErr != nil {
			l.Error(statusErr, "Failed to update audit status to Failed")
		}
		return ctrl.Result{}, err
	}

	var framework Framework
	if err := yaml.Unmarshal(data, &framework); err != nil {
		l.Error(err, "Failed to unmarshal framework yaml")
		audit.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, &audit); statusErr != nil {
			l.Error(statusErr, "Failed to update audit status to Failed")
		}
		return ctrl.Result{}, err
	}

	// 5. Iterate through all controls and capture evidence
	evidenceCount := 0
	for _, control := range framework.Controls {
		for _, res := range control.EvidenceResources {
			l.Info("Capturing evidence", "Control", control.ID, "Kind", res.Kind)
			
			snap, err := evidence.CaptureResource(res.Kind, res.Name, res.Namespace)
			if err != nil {
				l.Error(err, "Failed to capture evidence", "kind", res.Kind)
				continue
			}

			// 6. Upload to S3
			_, err = evidence.UploadToS3(ctx, snap, r.EvidenceBucket)
			if err != nil {
				l.Error(err, "Failed to upload evidence to S3", "kind", res.Kind)
				continue
			}
			evidenceCount++
		}
	}

	// 7. Finalize Audit
	audit.Status.Phase = "Completed"
	end := metav1.Now()
	audit.Status.EndTime = &end
	audit.Status.EvidenceCount = int32(evidenceCount)

	if err := r.Status().Update(ctx, &audit); err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Compliance Audit completed successfully", "EvidenceCount", evidenceCount)

	return ctrl.Result{}, nil
}

func (r *ComplianceAuditReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ComplianceAudit{}).
		Complete(r)
}
