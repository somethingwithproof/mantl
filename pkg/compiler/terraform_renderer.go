package compiler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
)

// TFVars represents the JSON format for Terraform variables.
type TFVars map[string]interface{}

// RenderTerraform generates the terraform.tfvars.json file for the given spec.
func RenderTerraform(cluster *v1alpha1.MantlCluster, outputDir string) error {
	// 1. Prepare variables based on the spec
	// Map profile size to environment name
	envMap := map[string]string{
		"small":  "dev",
		"medium": "staging",
		"full":   "prod",
	}
	environment := envMap[cluster.Spec.Profile.Size]
	if environment == "" {
		environment = cluster.Spec.Profile.Size
	}

	vars := TFVars{
		"cluster_name":       cluster.Name,
		"region":             cluster.Spec.Provider.Region,
		"kubernetes_version": cluster.Spec.Kubernetes.Version,
		"vpc_id":             cluster.Spec.Networking.VpcID,
		"domain":             cluster.Spec.Networking.Domain,
		"environment":        environment,
	}

	// 2. Map provider-specific variables
	if cluster.Spec.Provider.Kind == "aws" {
		vars["account_id"] = cluster.Spec.Provider.AccountID
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
	if err := ioutil.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write terraform variables file: %w", err)
	}

	return nil
}
