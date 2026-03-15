package compliance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

// validFrameworkName restricts framework names to safe characters.
var validFrameworkName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Framework defines the structure of our compliance mapping files.
type Framework struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Controls []struct {
		ID                string             `json:"id"`
		Policies          []string           `json:"policies"`
		EvidenceResources []EvidenceResource `json:"evidenceResources"`
	} `json:"controls"`
}

// EvidenceResource defines a Kubernetes resource to be captured as evidence.
type EvidenceResource struct {
	Kind      string `json:"kind"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Group     string `json:"group,omitempty"`
}

// ComplianceProfileReconciler reconciles a ComplianceProfile object
type ComplianceProfileReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	FrameworkDir string
}

// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kyverno.io,resources=clusterpolicies,verbs=get;list;watch;create;update;patch;delete

func (r *ComplianceProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// 1. Fetch the ComplianceProfile
	var profile v1alpha1.ComplianceProfile
	if err := r.Get(ctx, req.NamespacedName, &profile); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("Reconciling ComplianceProfile", "Name", profile.Name, "Framework", profile.Spec.Framework)

	// 2. Load the Framework mapping file
	if !validFrameworkName.MatchString(profile.Spec.Framework) {
		return ctrl.Result{}, fmt.Errorf("invalid framework name: %q", profile.Spec.Framework)
	}
	frameworkFile := filepath.Join(r.FrameworkDir, fmt.Sprintf("%s.yaml", profile.Spec.Framework))
	frameworkFile = filepath.Clean(frameworkFile)
	if !strings.HasPrefix(frameworkFile, filepath.Clean(r.FrameworkDir)+string(os.PathSeparator)) {
		return ctrl.Result{}, fmt.Errorf("invalid framework name: path traversal detected")
	}
	data, err := os.ReadFile(frameworkFile)
	if err != nil {
		l.Error(err, "Failed to read framework file", "file", frameworkFile)
		return ctrl.Result{}, err
	}

	var framework Framework
	if err := yaml.Unmarshal(data, &framework); err != nil {
		l.Error(err, "Failed to unmarshal framework yaml")
		return ctrl.Result{}, err
	}

	// 3. Map controls to concrete policies
	var policiesToEnforce []string
	for _, control := range framework.Controls {
		policiesToEnforce = append(policiesToEnforce, control.Policies...)
	}

	l.Info("Dynamic mapping complete", "framework", framework.Name, "policyCount", len(policiesToEnforce))

	// 4. Update status
	profile.Status.State = "Active"
	profile.Status.ActivePolicies = int32(len(policiesToEnforce))
	if err := r.Status().Update(ctx, &profile); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComplianceProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ComplianceProfile{}).
		Complete(r)
}
