package compliance

import "testing"

func TestBuildPolicyIndex(t *testing.T) {
	// The repo's platform policy dir must yield a known policy by base name.
	idx, err := buildPolicyIndex([]string{"../../platform/security/kyverno/policies"})
	if err != nil {
		t.Fatalf("buildPolicyIndex: %v", err)
	}
	if _, ok := idx["require-network-policy"]; !ok {
		t.Fatalf("index missing require-network-policy; got keys %v", keys(idx))
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
