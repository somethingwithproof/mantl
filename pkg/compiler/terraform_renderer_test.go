package compiler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
)

func newTestCluster(size string) *v1alpha1.MantlCluster {
	return &v1alpha1.MantlCluster{
		ObjectMeta: metav1.ObjectMeta{Name: "test-cluster"},
		Spec: v1alpha1.MantlClusterSpec{
			Provider: v1alpha1.ProviderSpec{
				Kind:      "aws",
				Region:    "us-east-1",
				AccountID: "123456789012",
			},
			Kubernetes: v1alpha1.KubernetesSpec{
				Distribution: "eks",
				Version:      "1.29",
			},
			Profile: v1alpha1.ProfileSpec{
				Size: size,
			},
			Networking: v1alpha1.NetworkingSpec{
				Domain: "example.com",
				VpcID:  "vpc-abc123",
			},
		},
	}
}

func TestRenderTerraform_MappedSizes(t *testing.T) {
	tests := []struct {
		size    string
		wantEnv string
	}{
		{"small", "dev"},
		{"medium", "staging"},
		{"full", "prod"},
	}

	for _, tt := range tests {
		t.Run(tt.size, func(t *testing.T) {
			dir := t.TempDir()
			cluster := newTestCluster(tt.size)

			if err := RenderTerraform(cluster, dir); err != nil {
				t.Fatalf("RenderTerraform(%q) returned error: %v", tt.size, err)
			}

			data, err := os.ReadFile(filepath.Join(dir, "terraform.tfvars.json"))
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}

			var vars map[string]interface{}
			if err := json.Unmarshal(data, &vars); err != nil {
				t.Fatalf("failed to parse output JSON: %v", err)
			}

			if got := vars["environment"]; got != tt.wantEnv {
				t.Errorf("environment = %q, want %q", got, tt.wantEnv)
			}
		})
	}
}

func TestRenderTerraform_UnmappedSize(t *testing.T) {
	dir := t.TempDir()
	cluster := newTestCluster("xlarge")

	err := RenderTerraform(cluster, dir)
	if err == nil {
		t.Fatal("expected error for unmapped size \"xlarge\", got nil")
	}
	if got := err.Error(); !containsStr(got, "unsupported profile size") {
		t.Errorf("error = %q, want it to contain \"unsupported profile size\"", got)
	}
}

func TestRenderTerraform_EmptySize(t *testing.T) {
	dir := t.TempDir()
	cluster := newTestCluster("")

	err := RenderTerraform(cluster, dir)
	if err == nil {
		t.Fatal("expected error for empty size, got nil")
	}
	if got := err.Error(); !containsStr(got, "unsupported profile size") {
		t.Errorf("error = %q, want it to contain \"unsupported profile size\"", got)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
