package compliance

import (
	"os"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestFrameworkParsesRealSOC2(t *testing.T) {
	data, err := os.ReadFile("../../compliance/frameworks/soc2/framework.yaml")
	if err != nil {
		t.Fatalf("read framework: %v", err)
	}
	var fw Framework
	if err := yaml.Unmarshal(data, &fw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(fw.Controls) == 0 {
		t.Fatal("no controls parsed")
	}
	// CC6.1 must expose its two policy templates via mappings.
	var cc61 *Control
	for i := range fw.Controls {
		if fw.Controls[i].ID == "CC6.1" {
			cc61 = &fw.Controls[i]
		}
	}
	if cc61 == nil {
		t.Fatal("CC6.1 not found")
	}
	got := templatesOf(cc61)
	if len(got) != 2 || got[0] != "require-network-policy" || got[1] != "require-pod-security-context" {
		t.Fatalf("CC6.1 templates = %v, want [require-network-policy require-pod-security-context]", got)
	}
}

func templatesOf(c *Control) []string {
	var out []string
	for _, m := range c.Mappings {
		if m.PolicyRef != nil && m.PolicyRef.Template != "" {
			out = append(out, m.PolicyRef.Template)
		}
	}
	return out
}
