package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
)

// TFVars represents the JSON format for Terraform variables.
type TFVars map[string]interface{}

// RenderTerraform generates the terraform.tfvars.json file for the given spec.
func RenderTerraform(cluster *v1alpha1.MantlCluster, outputDir string) error {
	if cluster == nil {
		return fmt.Errorf("cluster must not be nil")
	}

	// 1. Prepare variables based on the spec
	// Map profile size to environment name
	envMap := map[string]string{
		"small":  "dev",
		"medium": "staging",
		"full":   "prod",
	}
	environment := envMap[cluster.Spec.Profile.Size]
	if environment == "" {
		return fmt.Errorf("unsupported profile size %q: expected one of small, medium, full", cluster.Spec.Profile.Size)
	}

	if cluster.Spec.Environment != "" {
		switch cluster.Spec.Environment {
		case "dev", "staging":
			environment = cluster.Spec.Environment
		case "production":
			environment = "prod"
		default:
			return fmt.Errorf("unsupported environment %q", cluster.Spec.Environment)
		}
	}
	vars := TFVars{"cluster_name": cluster.Name, "kubernetes_version": cluster.Spec.Kubernetes.Version, "environment": environment}
	switch cluster.Spec.Provider.Kind {
	case "gcp":
		vars["region"] = cluster.Spec.Provider.Region
		vars["project"] = cluster.Spec.Provider.AccountID
	case "azure":
		vars["location"] = cluster.Spec.Provider.Region
		vars["resource_group_name"] = cluster.Spec.Provider.AccountID
	default:
		vars["region"] = cluster.Spec.Provider.Region
	}

	// 3. Serialize to JSON
	data, err := json.MarshalIndent(vars, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal terraform variables: %w", err)
	}

	// 4. Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 5. Write to file
	outputPath := filepath.Join(outputDir, "terraform.tfvars.json")
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write terraform variables file: %w", err)
	}

	return nil
}
