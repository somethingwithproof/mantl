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
	"Security":            "platform/security",
	"Secrets":             "platform/secrets",
	"Compliance":          "deploy/operator/base",
	"ProgressiveDelivery": "platform/progressive-delivery",
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
	// This makes the code DRY: adding a new feature field to the API automatically
	// picks up the mapping if it exists in featureAppMap.
	features := cluster.Spec.Features
	v := reflect.ValueOf(features)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		isEnabled := v.Field(i).Bool()

		if isEnabled {
			if path, exists := featureAppMap[fieldName]; exists {
				appName := fmt.Sprintf("addon-%s", strings.ToLower(fieldName))
				if err := renderApp(gitopsDir, appName, path); err != nil {
					return err
				}
			}
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
