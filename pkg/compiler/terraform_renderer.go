// SPDX-License-Identifier: Apache-2.0

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
	data, err := terraformManifest(cluster)
	if err != nil {
		return err
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

func terraformManifest(cluster *v1alpha1.MantlCluster) ([]byte, error) {
	if cluster == nil {
		return nil, fmt.Errorf("cluster must not be nil")
	}

	environment, err := environmentName(cluster)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("failed to marshal terraform variables: %w", err)
	}

	return data, nil
}

// environmentName is the one environment vocabulary for Terraform variables and
// tenant labels: dev, staging, production. Profile size supplies it when the
// spec does not.
func environmentName(cluster *v1alpha1.MantlCluster) (string, error) {
	if cluster == nil {
		return "", fmt.Errorf("cluster must not be nil")
	}
	if cluster.Spec.Environment != "" {
		switch cluster.Spec.Environment {
		case "dev", "staging", "production":
			return cluster.Spec.Environment, nil
		default:
			return "", fmt.Errorf("unsupported environment %q", cluster.Spec.Environment)
		}
	}
	switch cluster.Spec.Profile.Size {
	case "small":
		return "dev", nil
	case "medium":
		return "staging", nil
	case "full":
		return "production", nil
	default:
		return "", fmt.Errorf("unsupported profile size %q: expected one of small, medium, full", cluster.Spec.Profile.Size)
	}
}
