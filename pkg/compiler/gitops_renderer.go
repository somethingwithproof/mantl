package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
	"sigs.k8s.io/yaml"
)

// featureAppMap defines the mapping between the MantlCluster Features struct field names
// and their corresponding manifest paths in the repository.
var featureAppMap = map[string]string{
	"Observability":       "platform/observability/prometheus/overlays/dev",
	"Security":            "platform/security/base",
	"Secrets":             "platform/secrets/base",
	"Compliance":          "deploy/operator/base",
	"ProgressiveDelivery": "platform/progressive-delivery/base",
}

// RenderGitOps generates the ArgoCD application manifests for the platform.
func RenderGitOps(cluster *v1alpha1.MantlCluster, outputDir string) error {
	if cluster == nil {
		return fmt.Errorf("cluster must not be nil")
	}
	repository := cluster.Spec.GitOps.Repository
	if repository == "" {
		repository = "https://github.com/somethingwithproof/mantl.git"
	}
	revision := cluster.Spec.GitOps.Revision
	if revision == "" {
		revision = "main"
	}
	render := func(dir, name, path string) error { return renderAppSource(dir, name, path, repository, revision) }
	gitopsDir := filepath.Join(outputDir, "gitops")
	if err := os.MkdirAll(gitopsDir, 0755); err != nil {
		return fmt.Errorf("failed to create gitops directory: %w", err)
	}

	// Remove only known generated application files before writing the selected topology.
	for _, name := range []string{"root-platform", "addon-observability", "addon-security", "addon-secrets", "addon-compliance", "addon-progressivedelivery", "platform-tenants"} {
		if err := os.Remove(filepath.Join(gitopsDir, name+".yaml")); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("clear generated application: %w", err)
		}
	}

	// 1. Generate the Root Application (App-of-Apps)
	if err := render(gitopsDir, "root-platform", "deploy/gitops/core"); err != nil {
		return err
	}

	// 2. Dynamically generate Addon Applications based on features using reflection
	features := cluster.Spec.Features
	featuresValue := reflect.ValueOf(features)
	featuresType := featuresValue.Type()

	for i := 0; i < featuresValue.NumField(); i++ {
		fieldName := featuresType.Field(i).Name
		isEnabled := featuresValue.Field(i).Bool()

		if isEnabled {
			if path, exists := featureAppMap[fieldName]; exists {
				if fieldName == "Compliance" && cluster.Spec.GitOps.OperatorPath != "" {
					path = cluster.Spec.GitOps.OperatorPath
				}
				appName := fmt.Sprintf("addon-%s", strings.ToLower(fieldName))
				if err := render(gitopsDir, appName, path); err != nil {
					return err
				}
			}
		}
	}

	// Keep an empty tenant application when its source is configured so ArgoCD can
	// reconcile removals. Never discover desired tenants from stale output files.
	if len(cluster.Spec.Tenants) > 0 || cluster.Spec.GitOps.TenantPath != "" {
		tenantsDir := filepath.Join(outputDir, "tenants")
		if err := os.MkdirAll(tenantsDir, 0755); err != nil {
			return err
		}

		names := []string{}
		for _, tenant := range cluster.Spec.Tenants {
			if err := renderTenant(tenantsDir, tenant); err != nil {
				return err
			}
			names = append(names, tenant.Name+".yaml")
		}
		kustomization, err := yaml.Marshal(map[string]interface{}{"apiVersion": "kustomize.config.k8s.io/v1beta1", "kind": "Kustomization", "resources": names})
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(tenantsDir, "kustomization.yaml"), kustomization, 0600); err != nil {
			return err
		}

		// Create an ArgoCD application to manage all tenants
		if err := render(gitopsDir, "platform-tenants", cluster.Spec.GitOps.TenantPath); err != nil {
			return err
		}
	}

	return nil
}

// renderTenant generates Namespace and RBAC RoleBindings for a tenant.
func renderTenant(outputDir string, tenant v1alpha1.TenantSpec) error {
	if tenant.Name == "kustomization" {
		return fmt.Errorf("tenant name kustomization conflicts with generated manifest index")
	}
	ns := tenant.Namespace
	if ns == "" {
		ns = tenant.Name
	}

	tenantFile := filepath.Join(outputDir, fmt.Sprintf("%s.yaml", tenant.Name))

	absOut, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output directory: %w", err)
	}
	absFile, err := filepath.Abs(tenantFile)
	if err != nil {
		return fmt.Errorf("failed to resolve tenant file path: %w", err)
	}
	if !strings.HasPrefix(absFile, absOut+string(filepath.Separator)) {
		return fmt.Errorf("tenant name %q results in path outside output directory", tenant.Name)
	}

	// Create Namespace
	namespace := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": ns,
			"labels": map[string]interface{}{
				"mantl.io/tenant": tenant.Name,
			},
		},
	}

	nsData, _ := yaml.Marshal(namespace)
	content := string(nsData) + "---\n"

	// Create RoleBinding for admins
	if len(tenant.Admins) > 0 {
		rb := map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "RoleBinding",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-admin", tenant.Name),
				"namespace": ns,
			},
			"roleRef": map[string]interface{}{
				"apiGroup": "rbac.authorization.k8s.io",
				"kind":     "ClusterRole",
				"name":     "admin",
			},
			"subjects": make([]map[string]interface{}, 0),
		}

		for _, admin := range tenant.Admins {
			rb["subjects"] = append(rb["subjects"].([]map[string]interface{}), map[string]interface{}{
				"kind":     "User",
				"name":     admin,
				"apiGroup": "rbac.authorization.k8s.io",
			})
		}

		rbData, _ := yaml.Marshal(rb)
		content += string(rbData)
	}

	// Tenant workloads start with default-deny ingress/egress and bounded resources.
	// Operators add reviewed application-specific policies through GitOps.
	isolation := []map[string]interface{}{
		{"apiVersion": "networking.k8s.io/v1", "kind": "NetworkPolicy", "metadata": map[string]interface{}{"name": "default-deny", "namespace": ns}, "spec": map[string]interface{}{"podSelector": map[string]interface{}{}, "policyTypes": []string{"Ingress", "Egress"}}},
		{"apiVersion": "v1", "kind": "ResourceQuota", "metadata": map[string]interface{}{"name": "tenant-quota", "namespace": ns}, "spec": map[string]interface{}{"hard": map[string]string{"requests.cpu": "2", "requests.memory": "4Gi", "limits.cpu": "4", "limits.memory": "8Gi", "pods": "20"}}},
		{"apiVersion": "v1", "kind": "LimitRange", "metadata": map[string]interface{}{"name": "tenant-defaults", "namespace": ns}, "spec": map[string]interface{}{"limits": []interface{}{map[string]interface{}{"type": "Container", "default": map[string]string{"cpu": "500m", "memory": "512Mi"}, "defaultRequest": map[string]string{"cpu": "100m", "memory": "128Mi"}}}}},
	}
	for _, obj := range isolation {
		data, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}
		content += "---\n" + string(data)
	}
	return os.WriteFile(tenantFile, []byte(content), 0600)
}

// syncWaveEntry pairs a component prefix with its ArgoCD sync-wave number.
type syncWaveEntry struct {
	prefix string
	wave   string
}

// syncWavePriority assigns ArgoCD sync-wave numbers to known infrastructure
// components. Lower waves deploy first; anything not listed defaults to "3".
// Ordered by prefix length descending so longer (more specific) prefixes match
// first, eliminating ambiguity when one prefix is a substring of another.
var syncWavePriority = []syncWaveEntry{
	{"external-secrets", "1"},
	{"cert-manager", "1"},
	{"monitoring", "2"},
	{"kyverno", "2"},
}

// renderApp is a helper to generate a standard ArgoCD Application manifest.
func renderApp(outputDir, name, path string) error {
	return renderAppSource(outputDir, name, path, "https://github.com/somethingwithproof/mantl.git", "main")
}
func renderAppSource(outputDir, name, path, repository, revision string) error {
	wave := "3"
	for _, entry := range syncWavePriority {
		if strings.HasPrefix(name, entry.prefix) {
			wave = entry.wave
			break
		}
	}

	app := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": "argocd",
			"annotations": map[string]interface{}{
				"argocd.argoproj.io/sync-wave": wave,
			},
		},
		"spec": map[string]interface{}{
			"project": "default",
			"source": map[string]interface{}{
				"repoURL":        repository,
				"targetRevision": revision,
				"path":           path,
			},
			"destination": map[string]interface{}{
				"server":    "https://kubernetes.default.svc",
				"namespace": "argocd",
			},
			"syncPolicy": map[string]interface{}{
				"automated": map[string]interface{}{
					"prune":    true,
					"selfHeal": true,
				},
				"syncOptions": []string{"CreateNamespace=true"},
			},
		},
	}

	data, err := yaml.Marshal(app)
	if err != nil {
		return fmt.Errorf("failed to render application %s: %w", name, err)
	}

	outPath := filepath.Join(outputDir, fmt.Sprintf("%s.yaml", name))
	absOut, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output directory: %w", err)
	}
	absFile, err := filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("failed to resolve output file path: %w", err)
	}
	if !strings.HasPrefix(absFile, absOut+string(filepath.Separator)) {
		return fmt.Errorf("app name %q results in path outside output directory", name)
	}
	return os.WriteFile(outPath, data, 0600)
}
