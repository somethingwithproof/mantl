package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
	rbac "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

// featureAppMap defines the mapping between the MantlCluster Features struct field names
// and their corresponding manifest paths in the repository.
const rootPlatformApp = "root-platform"

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
	gitopsDir := filepath.Join(outputDir, "gitops")
	if err := os.MkdirAll(gitopsDir, 0755); err != nil {
		return fmt.Errorf("failed to create gitops directory: %w", err)
	}

	// Remove only known generated application files before writing the selected topology.
	for _, name := range []string{rootPlatformApp, "addon-observability", "addon-security", "addon-secrets", "addon-compliance", "addon-progressivedelivery", "platform-tenants"} {
		if err := os.Remove(filepath.Join(gitopsDir, name+".yaml")); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("clear generated application: %w", err)
		}
	}

	for _, app := range GitOpsApplications(cluster) {
		if err := renderAppSource(gitopsDir, app.Name, app.Path, app.Repository, app.Revision); err != nil {
			return err
		}
	}
	return renderTenantTopology(cluster, outputDir)
}

// GitOpsApplication describes one application generated directly from a Mantl spec.
// Nested applications and tenant workloads have their own lifecycle and are not
// included in this top-level topology.
type GitOpsApplication struct {
	Name       string
	Path       string
	Repository string
	Revision   string
}

// GitOpsApplications is the shared desired topology for rendering and observation.
func GitOpsApplications(cluster *v1alpha1.MantlCluster) []GitOpsApplication {
	if cluster == nil {
		return nil
	}
	repository, revision := cluster.Spec.GitOps.Repository, cluster.Spec.GitOps.Revision
	if repository == "" {
		repository = "https://github.com/somethingwithproof/mantl.git"
	}
	if revision == "" {
		revision = "main"
	}
	apps := []GitOpsApplication{{Name: rootPlatformApp, Path: "deploy/gitops/core", Repository: repository, Revision: revision}}
	add := func(name, path string, enabled bool) {
		if enabled {
			apps = append(apps, GitOpsApplication{Name: name, Path: path, Repository: repository, Revision: revision})
		}
	}
	features := cluster.Spec.Features
	add("addon-observability", featureAppMap["Observability"], features.Observability)
	add("addon-security", featureAppMap["Security"], features.Security)
	add("addon-secrets", featureAppMap["Secrets"], features.Secrets)
	operatorPath := cluster.Spec.GitOps.OperatorPath
	if operatorPath == "" {
		operatorPath = featureAppMap["Compliance"]
	}
	add("addon-compliance", operatorPath, features.Compliance)
	add("addon-progressivedelivery", featureAppMap["ProgressiveDelivery"], features.ProgressiveDelivery)
	add("platform-tenants", cluster.Spec.GitOps.TenantPath, len(cluster.Spec.Tenants) > 0 || cluster.Spec.GitOps.TenantPath != "")
	return apps
}

func renderTenantTopology(cluster *v1alpha1.MantlCluster, outputDir string) error {
	// Keep an empty tenant application when its source is configured so ArgoCD can
	// reconcile removals. Never discover desired tenants from stale output files.
	if len(cluster.Spec.Tenants) == 0 && cluster.Spec.GitOps.TenantPath == "" {
		return nil
	}
	tenantsDir := filepath.Join(outputDir, "tenants")
	if err := os.MkdirAll(tenantsDir, 0755); err != nil {
		return err
	}
	names := []string{}
	for _, tenant := range cluster.Spec.Tenants {
		if err := renderTenant(tenantsDir, tenant, tenantEnvironment(cluster)); err != nil {
			return err
		}
		names = append(names, tenant.Name+".yaml")
	}
	kustomization, err := yaml.Marshal(map[string]interface{}{"apiVersion": "kustomize.config.k8s.io/v1beta1", "kind": "Kustomization", "resources": names})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(tenantsDir, "kustomization.yaml"), kustomization, 0600)
}

// renderTenant generates Namespace and RBAC RoleBindings for a tenant.
func renderTenant(outputDir string, tenant v1alpha1.TenantSpec, environment string) error {
	data, err := tenantManifest(tenant, environment)
	if err != nil {
		return err
	}
	if !dnsName.MatchString(tenant.Name) || len(tenant.Name) > 63 {
		return fmt.Errorf("invalid tenant name")
	}
	return os.WriteFile(filepath.Join(outputDir, tenant.Name+".yaml"), data, 0600)
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
	data, err := applicationManifest(name, path, repository, revision)
	if err != nil {
		return err
	}
	if !dnsName.MatchString(name) || len(name) > 63 {
		return fmt.Errorf("invalid application name")
	}
	return os.WriteFile(filepath.Join(outputDir, name+".yaml"), data, 0600)
}

func tenantEnvironment(cluster *v1alpha1.MantlCluster) string {
	if cluster.Spec.Environment != "" {
		return cluster.Spec.Environment
	}
	return map[string]string{"small": "dev", "medium": "staging", "full": "production"}[cluster.Spec.Profile.Size]
}

func tenantManifest(tenant v1alpha1.TenantSpec, environment string) ([]byte, error) {
	if environment == "" {
		environment = "dev"
	}
	if environment != "dev" && environment != "staging" && environment != "production" {
		return nil, fmt.Errorf("invalid tenant environment")
	}
	if tenant.Name == "kustomization" {
		return nil, fmt.Errorf("tenant name kustomization conflicts with generated manifest index")
	}
	ns := tenant.Namespace
	if ns == "" {
		ns = tenant.Name
	}

	// Create Namespace
	namespace := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": ns,
			"labels": map[string]interface{}{
				"mantl.io/tenant":     tenant.Name,
				"opencost.io/team":    tenant.Name,
				"opencost.io/project": "mantl",
				"environment":         environment,
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
				"apiGroup": rbac.GroupName,
				"kind":     "ClusterRole",
				"name":     "admin",
			},
			"subjects": make([]map[string]interface{}, 0),
		}

		for _, admin := range tenant.Admins {
			rb["subjects"] = append(rb["subjects"].([]map[string]interface{}), map[string]interface{}{
				"kind":     "User",
				"name":     admin,
				"apiGroup": rbac.GroupName,
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
			return nil, err
		}
		content += "---\n" + string(data)
	}
	return []byte(content), nil
}

func applicationManifest(name, path, repository, revision string) ([]byte, error) {
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

	if name == "root-platform" {
		patch, err := json.Marshal([]map[string]interface{}{{"op": "replace", "path": "/spec/source/repoURL", "value": repository}, {"op": "replace", "path": "/spec/source/targetRevision", "value": revision}})
		if err != nil {
			return nil, err
		}
		app["spec"].(map[string]interface{})["source"].(map[string]interface{})["kustomize"] = map[string]interface{}{"patches": []interface{}{map[string]interface{}{"target": map[string]string{"group": "argoproj.io", "version": "v1alpha1", "kind": "Application", "name": "kyverno-policies"}, "patch": string(patch)}}}
	}
	data, err := yaml.Marshal(app)
	if err != nil {
		return nil, fmt.Errorf("failed to render application %s: %w", name, err)
	}

	return data, nil
}
