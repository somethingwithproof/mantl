package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
	"sigs.k8s.io/yaml"
)

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

	// 2. Generate Addon Applications based on features
	if cluster.Spec.Features.Observability {
		if err := renderApp(gitopsDir, "observability-stack", "platform/observability"); err != nil {
			return err
		}
	}

	if cluster.Spec.Features.Security {
		if err := renderApp(gitopsDir, "security-policies", "platform/security"); err != nil {
			return err
		}
	}

	if cluster.Spec.Features.Secrets {
		if err := renderApp(gitopsDir, "external-secrets", "platform/secrets"); err != nil {
			return err
		}
	}

	if cluster.Spec.Features.Compliance {
		if err := renderApp(gitopsDir, "compliance-operator", "compliance/operator"); err != nil {
			return err
		}
	}

	return nil
}

// renderApp is a helper to generate a standard ArgoCD Application manifest.
func renderApp(outputDir, name, path string) error {
	app := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": "argocd",
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
		return err
	}

	return ioutil.WriteFile(filepath.Join(outputDir, fmt.Sprintf("%s.yaml", name)), data, 0644)
}
