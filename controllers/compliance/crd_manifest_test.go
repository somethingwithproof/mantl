// SPDX-License-Identifier: Apache-2.0

package compliance

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/yaml"
)

// TestCRDManifestsCoverRegisteredKinds guards that every compliance kind has a
// generated CRD manifest under config/crd/bases. Without these CRDs the
// controllers' Get calls fail on a real cluster, so a missing or renamed CRD is
// a deployment break this catches in CI rather than at install time.
func TestCRDManifestsCoverRegisteredKinds(t *testing.T) {
	want := map[string]bool{
		"ComplianceProfile": false,
		"Finding":           false,
		"ComplianceAudit":   false,
	}

	dir := "../../config/crd/bases"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read CRD dir: %v", err)
	}

	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var crd struct {
			Kind string `json:"kind"`
			Spec struct {
				Group string `json:"group"`
				Names struct {
					Kind string `json:"kind"`
				} `json:"names"`
			} `json:"spec"`
		}
		if err := yaml.Unmarshal(data, &crd); err != nil {
			t.Fatalf("%s: %v", e.Name(), err)
		}
		if crd.Kind != "CustomResourceDefinition" {
			t.Errorf("%s: kind=%q, want CustomResourceDefinition", e.Name(), crd.Kind)
		}
		if crd.Spec.Group != "compliance.mantl.io" {
			t.Errorf("%s: group=%q, want compliance.mantl.io", e.Name(), crd.Spec.Group)
		}
		if _, ok := want[crd.Spec.Names.Kind]; ok {
			want[crd.Spec.Names.Kind] = true
		}
	}

	for kind, found := range want {
		if !found {
			t.Errorf("no CRD manifest found for kind %q", kind)
		}
	}
}
