package compliance

import (
	"os"
	"testing"

	"sigs.k8s.io/yaml"
)

// TestSOC2NewPoliciesResolve guards that the policies authored for the
// previously-unresolved SOC2 templates are picked up by the resolver, so the
// coverage gap genuinely shrinks rather than the policy files drifting unused.
func TestSOC2NewPoliciesResolve(t *testing.T) {
	data, err := os.ReadFile("../../compliance/frameworks/soc2/framework.yaml")
	if err != nil {
		t.Fatalf("read framework: %v", err)
	}
	var fw Framework
	if err := yaml.Unmarshal(data, &fw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var templates []string
	for _, c := range fw.Controls {
		for _, m := range c.Mappings {
			if m.PolicyRef != nil && m.PolicyRef.Template != "" {
				templates = append(templates, m.PolicyRef.Template)
			}
		}
	}

	idx, err := buildPolicyIndex([]string{
		"../../platform/security/kyverno/policies",
		"../../policies/kyverno",
	})
	if err != nil {
		t.Fatalf("buildPolicyIndex: %v", err)
	}

	_, unresolved := resolveTemplates(templates, idx, templateAliases)
	stillUnresolved := map[string]bool{}
	for _, u := range unresolved {
		stillUnresolved[u] = true
	}

	for _, want := range []string{
		"restrict-external-ips",
		"require-replicas",
		"require-change-annotations",
		"disallow-default-serviceaccount",
	} {
		if stillUnresolved[want] {
			t.Errorf("template %q should resolve to a policy now but is still unresolved", want)
		}
	}
}
