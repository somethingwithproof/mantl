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

	// Validate required fields
	if cluster.Spec.Provider.Kind == "" {
		return nil, fmt.Errorf("spec.provider.kind is required")
	}
	if cluster.Spec.Provider.Region == "" {
		return nil, fmt.Errorf("spec.provider.region is required")
	}
	if cluster.Spec.Kubernetes.Distribution == "" {
		return nil, fmt.Errorf("spec.kubernetes.distribution is required")
	}
	if cluster.Spec.Kubernetes.Version == "" {
		return nil, fmt.Errorf("spec.kubernetes.version is required")
	}
	if cluster.Spec.Profile.Size == "" {
		return nil, fmt.Errorf("spec.profile.size is required")
	}
	if cluster.Spec.Networking.Domain == "" {
		return nil, fmt.Errorf("spec.networking.domain is required")
	}

	return &cluster, nil
}
