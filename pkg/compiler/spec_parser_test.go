package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSpec_ValidSpec(t *testing.T) {
	specYAML := `
apiVersion: platform.mantl.io/v1alpha1
kind: MantlCluster
metadata:
  name: test-cluster
spec:
  provider:
    kind: aws
    region: us-east-1
    accountId: "123456789012"
  kubernetes:
    distribution: eks
    version: "1.29"
  profile:
    size: small
  networking:
    domain: example.com
    exposure: public
`
	dir := t.TempDir()
	specFile := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(specFile, []byte(specYAML), 0600); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	cluster, err := ParseSpec(specFile)
	if err != nil {
		t.Fatalf("ParseSpec returned error: %v", err)
	}
	if cluster.Name != "test-cluster" {
		t.Errorf("cluster.Name = %q, want %q", cluster.Name, "test-cluster")
	}
	if cluster.Spec.Provider.Kind != "aws" {
		t.Errorf("provider.kind = %q, want %q", cluster.Spec.Provider.Kind, "aws")
	}
}

func TestParseSpec_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "missing provider.kind",
			yaml: `
spec:
  provider:
    region: us-east-1
  kubernetes:
    distribution: eks
    version: "1.29"
  profile:
    size: small
  networking:
    domain: example.com
`,
			wantErr: "spec.provider.kind",
		},
		{
			name: "missing provider.region",
			yaml: `
spec:
  provider:
    kind: aws
  kubernetes:
    distribution: eks
    version: "1.29"
  profile:
    size: small
  networking:
    domain: example.com
`,
			wantErr: "spec.provider.region",
		},
		{
			name: "missing kubernetes.distribution",
			yaml: `
spec:
  provider:
    kind: aws
    region: us-east-1
  kubernetes:
    version: "1.29"
  profile:
    size: small
  networking:
    domain: example.com
`,
			wantErr: "spec.kubernetes.distribution",
		},
		{
			name: "missing kubernetes.version",
			yaml: `
spec:
  provider:
    kind: aws
    region: us-east-1
  kubernetes:
    distribution: eks
  profile:
    size: small
  networking:
    domain: example.com
`,
			wantErr: "spec.kubernetes.version",
		},
		{
			name: "missing profile.size",
			yaml: `
spec:
  provider:
    kind: aws
    region: us-east-1
  kubernetes:
    distribution: eks
    version: "1.29"
  profile: {}
  networking:
    domain: example.com
`,
			wantErr: "spec.profile.size",
		},
		{
			name: "missing networking.domain",
			yaml: `
spec:
  provider:
    kind: aws
    region: us-east-1
  kubernetes:
    distribution: eks
    version: "1.29"
  profile:
    size: small
  networking:
    exposure: public
`,
			wantErr: "spec.networking.domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			specFile := filepath.Join(dir, "spec.yaml")
			if err := os.WriteFile(specFile, []byte(tt.yaml), 0600); err != nil {
				t.Fatalf("failed to write spec file: %v", err)
			}

			_, err := ParseSpec(specFile)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseSpec_NonExistentFile(t *testing.T) {
	_, err := ParseSpec("/nonexistent/spec.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read spec file") {
		t.Errorf("error = %q, want it to contain 'failed to read spec file'", err.Error())
	}
}

func TestParseSpec_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(specFile, []byte("not: valid: yaml: {{{"), 0600); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	_, err := ParseSpec(specFile)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal spec") {
		t.Errorf("error = %q, want it to contain 'failed to unmarshal spec'", err.Error())
	}
}
