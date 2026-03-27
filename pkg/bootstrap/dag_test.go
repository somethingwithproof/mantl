package bootstrap

import (
	"strings"
	"testing"
)

func TestProvisionInfra_MissingDistributionReturnsError(t *testing.T) {
	dag := &ExecutionDAG{
		BuildDir:     t.TempDir(),
		Distribution: "", // empty
	}
	err := dag.ProvisionInfra("aws", "test-cluster")
	if err == nil {
		t.Fatal("expected error for missing distribution, got nil")
	}
	if !strings.Contains(err.Error(), "distribution is required") {
		t.Errorf("error = %q, want it to contain 'distribution is required'", err.Error())
	}
}

func TestProvisionInfra_MissingBlueprintReturnsError(t *testing.T) {
	dag := &ExecutionDAG{
		BuildDir:     t.TempDir(),
		Distribution: "unknown-distro",
	}
	err := dag.ProvisionInfra("nonexistent-cloud", "test-cluster")
	if err == nil {
		t.Fatal("expected error for unknown blueprint, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to contain 'not found'", err.Error())
	}
}

func TestCreateLocalCluster_InvalidNameRejected(t *testing.T) {
	dag := &ExecutionDAG{}
	tests := []struct {
		name        string
		clusterName string
	}{
		{"name with uppercase", "MyCluster"},
		{"name starting with digit", "1cluster"},
		{"name with underscore", "my_cluster"},
		{"empty name", ""},
		{"name too long", strings.Repeat("a", 254)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dag.CreateLocalCluster(tt.clusterName)
			if err == nil {
				t.Fatalf("expected error for invalid cluster name %q, got nil", tt.clusterName)
			}
			if !strings.Contains(err.Error(), "invalid cluster name") {
				t.Errorf("error = %q, want it to contain 'invalid cluster name'", err.Error())
			}
		})
	}
}
