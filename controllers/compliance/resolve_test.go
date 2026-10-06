// SPDX-License-Identifier: Apache-2.0

package compliance

import (
	"reflect"
	"testing"
)

func TestResolveTemplates(t *testing.T) {
	index := map[string]string{
		"require-network-policy":    "platform/security/kyverno/policies/require-network-policy.yaml",
		"restrict-image-registries": "platform/security/kyverno/policies/restrict-image-registries.yaml",
	}
	aliases := map[string]string{
		"restrict-registries": "restrict-image-registries",
	}
	resolved, unresolved := resolveTemplates(
		[]string{"require-network-policy", "restrict-registries", "require-rbac-least-privilege"},
		index, aliases,
	)
	wantResolved := []string{
		"platform/security/kyverno/policies/require-network-policy.yaml",
		"platform/security/kyverno/policies/restrict-image-registries.yaml",
	}
	if !reflect.DeepEqual(resolved, wantResolved) {
		t.Errorf("resolved = %v, want %v", resolved, wantResolved)
	}
	if !reflect.DeepEqual(unresolved, []string{"require-rbac-least-privilege"}) {
		t.Errorf("unresolved = %v, want [require-rbac-least-privilege]", unresolved)
	}
}
