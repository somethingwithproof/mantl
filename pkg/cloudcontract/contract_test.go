package cloudcontract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSupportedCloudContracts(t *testing.T) {
	results, err := Validate("../..")
	if err != nil {
		t.Fatal(err)
	}
	if failures := Failures(results); len(failures) > 0 {
		t.Fatal(failures)
	}
}
func TestPrivateDefaultChecksValueNotComment(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte("variable \"private\" { default = false }\nmodule \"cluster\" { private_api = var.private }\n"), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := read(dir)
	if err != nil {
		t.Fatal(err)
	}
	if d.boolean(d.block("module", "cluster"), "private_api", true) {
		t.Fatal("insecure private default accepted")
	}
}

func TestMissingCloudModulesReturnFailures(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"aws-eks", "gcp-gke", "azure-aks"} {
		dir := filepath.Join(root, "infra/terraform/blueprints", name)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte("variable \"kubernetes_version\" { default = \"1.31\" }\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	results, err := Validate(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(Failures(results)) == 0 {
		t.Fatal("missing modules accepted")
	}
}

func TestAWSContractRejectsSecurityRegressions(t *testing.T) {
	for _, tc := range []struct{ name, before, after, check string }{
		{"public API", "endpoint_public_access  = var.cluster_endpoint_public_access", "endpoint_public_access  = true", "private_api_default"},
		{"audit logs", `"api", "audit", "authenticator", "controllerManager", "scheduler"`, `"api", "authenticator", "controllerManager", "scheduler"`, "audit_logging"},
		{"encryption key", "provider_key_arn = module.kms.key_arn", "provider_key_id = module.kms.key_arn", "secrets_encryption"},
		{"workload identity", "enable_irsa             = true", "enable_irsa             = false", "workload_identity"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for _, provider := range []string{"aws-eks", "gcp-gke", "azure-aks"} {
				relative := filepath.Join("infra/terraform/blueprints", provider)
				files, err := filepath.Glob(filepath.Join("../..", relative, "*.tf"))
				if err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(root, relative)
				if err := os.MkdirAll(target, 0700); err != nil {
					t.Fatal(err)
				}
				for _, file := range files {
					data, err := os.ReadFile(file)
					if err != nil {
						t.Fatal(err)
					}
					if provider == "aws-eks" && filepath.Base(file) == "main.tf" {
						if !strings.Contains(string(data), tc.before) {
							t.Fatalf("fixture missing %q", tc.before)
						}
						data = []byte(strings.ReplaceAll(string(data), tc.before, tc.after))
					}
					if err := os.WriteFile(filepath.Join(target, filepath.Base(file)), data, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			results, err := Validate(root)
			if err != nil {
				t.Fatal(err)
			}
			for _, result := range results {
				if result.Provider == "aws-eks" && result.Checks[tc.check] {
					t.Fatalf("security regression accepted: %s", tc.check)
				}
			}
		})
	}
}
