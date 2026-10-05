package cloudcontract

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Result struct {
	Provider    string          `json:"provider"`
	Checks      map[string]bool `json:"checks"`
	Limitations []string        `json:"limitations"`
}
type document struct {
	blocks []*hclsyntax.Block
	source map[*hclsyntax.Block][]byte
}

func read(dir string) (document, error) {
	d := document{source: map[*hclsyntax.Block][]byte{}}
	files, err := filepath.Glob(filepath.Join(dir, "*.tf"))
	if err != nil {
		return d, err
	}
	if len(files) == 0 {
		return d, fmt.Errorf("no Terraform files at %s", dir)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return d, err
		}
		parsed, diags := hclsyntax.ParseConfig(data, file, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return d, fmt.Errorf("parse Terraform: %s", diags.Error())
		}
		body := parsed.Body.(*hclsyntax.Body)
		for _, b := range body.Blocks {
			d.blocks = append(d.blocks, b)
			d.source[b] = data
		}
	}
	return d, nil
}
func (d document) block(kind string, labels ...string) *hclsyntax.Block {
	for _, b := range d.blocks {
		if b.Type == kind && strings.Join(b.Labels, "/") == strings.Join(labels, "/") {
			return b
		}
	}
	return nil
}
func (d document) boolean(b *hclsyntax.Block, key string, want bool) bool {
	if b == nil {
		return false
	}
	a := b.Body.Attributes[key]
	if a == nil {
		return false
	}
	vars := map[string]cty.Value{}
	for _, variable := range d.blocks {
		if variable.Type == "variable" && len(variable.Labels) == 1 {
			def := variable.Body.Attributes["default"]
			if def != nil {
				value, diags := def.Expr.Value(nil)
				if !diags.HasErrors() {
					vars[variable.Labels[0]] = value
				}
			}
		}
	}
	value, diags := a.Expr.Value(&hcl.EvalContext{Variables: map[string]cty.Value{"var": cty.ObjectVal(vars)}})
	return !diags.HasErrors() && value.IsKnown() && !value.IsNull() && value.Type() == cty.Bool && value.True() == want
}
func (d document) contains(b *hclsyntax.Block, key, needle string) bool {
	if b == nil {
		return false
	}
	a := b.Body.Attributes[key]
	if a == nil {
		return false
	}
	data := d.source[b]
	return strings.Contains(string(a.Expr.Range().SliceBytes(data)), needle)
}
func Validate(root string) ([]Result, error) {
	results := []Result{}
	for _, name := range []string{"aws-eks", "gcp-gke", "azure-aks"} {
		d, err := read(filepath.Join(root, "infra/terraform/blueprints", name))
		if err != nil {
			return nil, err
		}
		checks := map[string]bool{}
		contract := d.block("output", "platform_contract")
		checks["shared_contract"] = contract != nil
		checks["version_input"] = d.block("variable", "kubernetes_version") != nil
		switch name {
		case "aws-eks":
			b := d.block("module", "eks")
			checks["private_api_default"] = d.boolean(b, "cluster_endpoint_private_access", true) && d.boolean(b, "cluster_endpoint_public_access", false)
			checks["audit_logging"] = d.contains(b, "cluster_enabled_log_types", "audit")
			checks["secrets_encryption"] = d.contains(b, "cluster_encryption_config", "provider_key_arn") // pragma: allowlist secret (attribute name, not a credential)
			checks["workload_identity"] = d.boolean(b, "enable_irsa", true)
		case "gcp-gke":
			b := d.block("module", "gke")
			checks["private_api_default"] = d.boolean(b, "enable_private_endpoint", true) && d.boolean(b, "enable_private_nodes", true)
			checks["audit_logging"] = d.contains(b, "logging_service", "logging.googleapis.com") && d.block("resource", "google_project_iam_audit_config", "kubernetes") != nil
			checks["secrets_encryption"] = d.contains(b, "database_encryption", "ENCRYPTED")
			checks["workload_identity"] = d.contains(b, "identity_namespace", "svc.id.goog")
		case "azure-aks":
			b := d.block("resource", "azurerm_kubernetes_cluster", "main")
			checks["private_api_default"] = d.boolean(b, "private_cluster_enabled", true)
			checks["workload_identity"] = d.boolean(b, "oidc_issuer_enabled", true) && d.boolean(b, "workload_identity_enabled", true)
			checks["audit_logging"] = d.block("resource", "azurerm_monitor_diagnostic_setting", "aks_audit") != nil
			checks["secrets_encryption"] = false // pragma: allowlist secret (boolean validation result)
			if b != nil {
				for _, child := range b.Body.Blocks {
					if child.Type == "key_management_service" {
						checks["secrets_encryption"] = child.Body.Attributes["key_vault_key_id"] != nil // pragma: allowlist secret (schema field name)
					}
				}
			}
			checks["platform_federation"] = d.block("resource", "azurerm_federated_identity_credential", "external_secrets") != nil
		}
		results = append(results, Result{name, checks, []string{"Static configuration evidence only; no deployment or recovery evidence", "Evidence storage requires an externally provisioned S3 Object Lock bucket and scoped federation", "Least-privilege identity bindings and observed audit delivery require deployment review"}})
	}
	return results, nil
}
func Failures(results []Result) []string {
	failures := []string{}
	for _, r := range results {
		for key, pass := range r.Checks {
			if !pass {
				failures = append(failures, r.Provider+":"+key)
			}
		}
	}
	sort.Strings(failures)
	return failures
}
