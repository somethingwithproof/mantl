package compiler

import (
	"fmt"
	"io/ioutil"
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
	"Observability":       "platform/observability",
	"Security":            "platform/security/base",
	"Secrets":             "platform/secrets/base",
	"Compliance":          "deploy/operator/base",
	"ProgressiveDelivery": "platform/progressive-delivery/base",
}

// RenderGitOps generates the ArgoCD application manifests for the platform.
func RenderGitOps(cluster *v1alpha1.MantlCluster, outputDir string) error {
	gitopsDir := filepath.Join(outputDir, "gitops")
	if err := os.MkdirAll(gitopsDir, 0755); err != nil {
		return fmt.Errorf("failed to create gitops directory: %w", err)
	}

	// 1. Generate the Root Application (App-of-Apps)
	if err := renderApp(gitopsDir, "root-platform", "platform/base"); err != nil {
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
				appName := fmt.Sprintf("addon-%s", strings.ToLower(fieldName))
				if err := renderApp(gitopsDir, appName, path); err != nil {
					return err
				}
			}
		}
	}

	// 3. Generate Tenant manifests
	if len(cluster.Spec.Tenants) > 0 {
		tenantsDir := filepath.Join(outputDir, "tenants")
		if err := os.MkdirAll(tenantsDir, 0755); err != nil {
			return err
		}

		for _, tenant := range cluster.Spec.Tenants {
			if err := renderTenant(tenantsDir, tenant); err != nil {
				return err
			}
		}

		// Create an ArgoCD application to manage all tenants
		if err := renderApp(gitopsDir, "platform-tenants", "tenants"); err != nil {
			return err
		}
	}

	return nil
}

// renderTenant generates Namespace and RBAC RoleBindings for a tenant.
func renderTenant(outputDir string, tenant v1alpha1.TenantSpec) error {
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

	return ioutil.WriteFile(tenantFile, []byte(content), 0600)
}

// syncWavePriority assigns ArgoCD sync-wave numbers to known infrastructure
// components. Lower waves deploy first; anything not listed defaults to "3".
var syncWavePriority = map[string]string{
	"cert-manager":     "1",
	"external-secrets": "1",
	"kyverno":          "2",
	"monitoring":       "2",
}

// renderApp is a helper to generate a standard ArgoCD Application manifest.
func renderApp(outputDir, name, path string) error {
	wave := "3"
	for component, w := range syncWavePriority {
		if strings.Contains(name, component) {
			wave = w
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
				"repoURL":        "https://github.com/thomasvincent/mantl",
				"targetRevision": "main",
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
	return ioutil.WriteFile(outPath, data, 0600)
}
