package framework

import (
	"reflect"
	"testing"
)

func TestPolicyControlsPreserveSelectionAliasesAndDuplicates(t *testing.T) {
	fw := &Framework{FrameworkSpec: FrameworkSpec{Controls: []Control{
		{ID: "C1", Mappings: []Mapping{{}, {PolicyRef: &PolicyRef{Template: "template"}}, {PolicyRef: &PolicyRef{Template: "template"}}}},
		{ID: "C2", Mappings: []Mapping{{PolicyRef: &PolicyRef{Template: "template"}}}},
	}}}
	for _, policy := range []string{"template", "actual", "soc2-template"} {
		got := PolicyControls(fw, []string{"C1"}, "soc2", policy, map[string]string{"template": "actual"})
		if !reflect.DeepEqual(got, []string{"C1", "C1"}) {
			t.Fatalf("mapping scope lost for %s: %v", policy, got)
		}
	}
	if got := PolicyControls(fw, nil, "soc2", "other", nil); len(got) != 0 {
		t.Fatalf("unrelated policy matched: %v", got)
	}
}

func TestNamespaceSelectionKeepsClusterAndTenantScopesDistinct(t *testing.T) {
	for _, test := range []struct {
		namespaces          []string
		fallback, namespace string
		want                bool
	}{
		{nil, "team", "", true}, {nil, "team", "team", true}, {nil, "team", "foreign", false},
		{[]string{"team", "other"}, "fallback", "other", true},
		{[]string{"team"}, "fallback", "fallback", false},
	} {
		if got := NamespaceSelected(test.namespaces, test.fallback, test.namespace); got != test.want {
			t.Fatalf("namespace %q selected=%v, want %v", test.namespace, got, test.want)
		}
	}
}
