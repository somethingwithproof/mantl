package compiler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

			outPath := filepath.Join(dir, "terraform.tfvars.json")
			data, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("failed to read output file %q: %v", outPath, err)
			}

			var vars map[string]interface{}
			if err := json.Unmarshal(data, &vars); err != nil {
				t.Fatalf("failed to parse output JSON: %v\ndata: %s", err, data)
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
	if got := err.Error(); !strings.Contains(got, "unsupported profile size") {
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
	if got := err.Error(); !strings.Contains(got, "unsupported profile size") {
		t.Errorf("error = %q, want it to contain \"unsupported profile size\"", got)
	}
}

func TestRenderTerraform_InvalidDirectory(t *testing.T) {
	cluster := newTestCluster("small")

	err := RenderTerraform(cluster, "/dev/null/impossible")
	if err == nil {
		t.Fatal("expected error for invalid directory path, got nil")
	}
}

func TestRenderTerraform_AllTFVarsFields(t *testing.T) {
	dir := t.TempDir()
	cluster := newTestCluster("small")

	if err := RenderTerraform(cluster, dir); err != nil {
		t.Fatalf("RenderTerraform returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "terraform.tfvars.json"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	var vars map[string]interface{}
	if err := json.Unmarshal(data, &vars); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	expected := map[string]string{
		"cluster_name":       "test-cluster",
		"region":             "us-east-1",
		"kubernetes_version": "1.29",
		"environment":        "dev",
	}

	for key, want := range expected {
		got, ok := vars[key]
		if !ok {
			t.Errorf("missing key %q in tfvars output", key)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestRenderTerraform_NilCluster(t *testing.T) {
	dir := t.TempDir()

	err := RenderTerraform(nil, dir)
	if err == nil {
		t.Fatal("expected error for nil cluster, got nil")
	}
	if got := err.Error(); !strings.Contains(got, "cluster must not be nil") {
		t.Errorf("error = %q, want it to contain \"cluster must not be nil\"", got)
	}
}
