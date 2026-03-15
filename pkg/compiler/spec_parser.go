package compiler

import (
	"fmt"
	"os"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
	"sigs.k8s.io/yaml"
)

// ParseSpec reads the MantlCluster YAML spec from a file and unmarshals it.
func ParseSpec(filePath string) (*v1alpha1.MantlCluster, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var cluster v1alpha1.MantlCluster
	if err := yaml.Unmarshal(data, &cluster); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}

	// Basic validation
	if cluster.Spec.Provider.Kind == "" {
		return nil, fmt.Errorf("spec.provider.kind is required")
	}

	return &cluster, nil
}
