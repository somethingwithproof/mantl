package compliance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

// Framework mirrors the on-disk compliance framework YAML
// (compliance/frameworks/<name>/framework.yaml). The file is a Kubernetes-style
// object whose controls live under spec, so FrameworkSpec is embedded with the
// "spec" tag: the document nests under spec while Go access stays flat
// (framework.Controls).
type Framework struct {
	FrameworkSpec `json:"spec"`
}

// FrameworkSpec holds the framework body: controls and their mappings.
type FrameworkSpec struct {
	Name     string    `json:"displayName"`
	Version  string    `json:"version"`
	Controls []Control `json:"controls"`
}

// Control is a single framework control with its policy and evidence mappings.
type Control struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Severity string    `json:"severity"`
	Mappings []Mapping `json:"mappings"`
	// EvidenceResources is retained for the audit controller, which still reads
	// it. The framework YAML does not populate it (so it stays nil and that loop
	// is a no-op), but keeping the field avoids breaking audit_controller.go.
	// The audit controller migrates to Mappings/EvidenceCollector in Phase 2b.
	EvidenceResources []EvidenceResource `json:"evidenceResources,omitempty"`
}

// EvidenceResource is a Kubernetes resource captured as evidence. Retained for
// audit_controller.go compatibility; superseded by EvidenceCollector in Phase 2b.
type EvidenceResource struct {
	Kind      string `json:"kind"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Group     string `json:"group,omitempty"`
}

// Mapping is one entry under a control: either a policy reference or an
// evidence collector. Type is "policy" or "evidence".
type Mapping struct {
	Type              string             `json:"type"`
	PolicyRef         *PolicyRef         `json:"policyRef,omitempty"`
	EvidenceCollector *EvidenceCollector `json:"evidenceCollector,omitempty"`
}

// PolicyRef names a Kyverno policy template and its enforcement mode.
type PolicyRef struct {
	Template string `json:"template"`
	Mode     string `json:"mode"`
}

// EvidenceCollector describes periodic evidence capture for a control.
// Parsed now; consumed in Phase 2b.
type EvidenceCollector struct {
	Type          string   `json:"type"`
	Schedule      string   `json:"schedule"`
	Query         string   `json:"query,omitempty"`
	Resources     []string `json:"resources,omitempty"`
	RetentionDays int      `json:"retentionDays,omitempty"`
}

// ComplianceProfileReconciler reconciles a ComplianceProfile object
type ComplianceProfileReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	FrameworkDir string
}

// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kyverno.io,resources=clusterpolicies,verbs=get;list;watch;create;update;patch

func (r *ComplianceProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// 1. Fetch the ComplianceProfile
	var profile v1alpha1.ComplianceProfile
	if err := r.Get(ctx, req.NamespacedName, &profile); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("Reconciling ComplianceProfile", "Name", profile.Name, "Framework", profile.Spec.Framework)

	// 2. Load the Framework mapping file
	frameworkFile := filepath.Join(r.FrameworkDir, fmt.Sprintf("%s.yaml", profile.Spec.Framework))
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

	var referenced []string
	for _, control := range framework.Controls {
		for _, m := range control.Mappings {
			if m.PolicyRef != nil && m.PolicyRef.Template != "" {
				referenced = append(referenced, m.PolicyRef.Template)
			}
		}
	}

	index, err := buildPolicyIndex([]string{
		"platform/security/kyverno/policies",
		"policies/kyverno",
		filepath.Join(r.FrameworkDir, profile.Spec.Framework, "policies"),
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	resolved, unresolved := resolveTemplates(referenced, index, templateAliases)
	l.Info("Resolved framework policies", "framework", framework.Name,
		"resolved", len(resolved), "unresolved", len(unresolved))

	profile.Status.State = "Active"
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
		Complete(r)
}
