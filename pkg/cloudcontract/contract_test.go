package cloudcontract

import (
	"os"
	"path/filepath"
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
	os.WriteFile(filepath.Join(dir, "main.tf"), []byte("variable \"private\" { default = false }\nmodule \"cluster\" { private_api = var.private }\n"), 0600)
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
		os.MkdirAll(dir, 0700)
		os.WriteFile(filepath.Join(dir, "main.tf"), []byte("variable \"kubernetes_version\" { default = \"1.31\" }\n"), 0600)
	}
	results, err := Validate(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(Failures(results)) == 0 {
		t.Fatal("missing modules accepted")
	}
}
