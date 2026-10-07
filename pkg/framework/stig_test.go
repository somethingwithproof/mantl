// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/auditplan"
	"github.com/thomasvincent/mantl/pkg/controlbundle"
	"github.com/thomasvincent/mantl/pkg/evaluation"
	"github.com/thomasvincent/mantl/pkg/framework"
)

func TestSTIGCatalogSurvivesVerifiedBundle(t *testing.T) {
	var archive bytes.Buffer
	manifest, err := controlbundle.Build("../..", "0.5.0-stig.1", &archive)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "content")
	if _, err = controlbundle.Unpack(&archive, dest, manifest.Digest()); err != nil {
		t.Fatal(err)
	}
	fw, err := framework.LoadPinned(filepath.Join(dest, "frameworks"), "stig-kubernetes", "2.3", manifest.Digest())
	if err != nil {
		t.Fatal(err)
	}
	if len(fw.Controls) != 91 || fw.ContentDigest != manifest.Digest() {
		t.Fatalf("catalog count or content identity changed: %d %s", len(fw.Controls), fw.ContentDigest)
	}
	now := time.Now().UTC()
	assertSTIGCollectionScope(t, fw, now)
	index := map[string]string{}
	for _, name := range []string{"stig-disallow-secret-env", "stig-restrict-privileged-host-ports"} {
		index[name] = filepath.Join(dest, "policies", "stig-kubernetes", name+".yaml")
	}
	mapped := 0
	for _, control := range fw.Controls {
		if assertSTIGControlCoverage(t, control, index, now) {
			mapped++
		}
	}
	if mapped != 2 {
		t.Fatalf("expected two partial workload mappings, got %d", mapped)
	}
	if _, err = framework.LoadSelected(filepath.Join(dest, "frameworks"), "stig-kubernetes", "2.3", manifest.Digest(), []string{"V-unknown"}); err == nil {
		t.Fatal("accepted an unknown STIG requirement")
	}
	if _, err = framework.Load(filepath.Join(dest, "frameworks"), "stig-kubernetes", "2.4"); err == nil {
		t.Fatal("accepted an unshipped STIG revision")
	}
}

func assertSTIGCollectionScope(t *testing.T, fw *framework.Framework, now time.Time) {
	t.Helper()
	profile := &api.ComplianceProfile{Spec: api.ComplianceProfileSpec{Namespaces: []string{"team-a"}}}
	audit := &api.ComplianceAudit{Spec: api.ComplianceAuditSpec{Frequency: "daily"}}
	plan := auditplan.Build(fw, profile, audit, now)
	if len(plan.Tasks) != 2 || len(plan.Gaps) != 0 {
		t.Fatalf("unexpected workload collection scope: %+v", plan)
	}
	for _, task := range plan.Tasks {
		if task.Namespace != "team-a" || task.Resource != "pods" {
			t.Fatalf("unexpected resource scope: %+v", task)
		}
	}
}

func assertSTIGControlCoverage(t *testing.T, control framework.Control, index map[string]string, now time.Time) bool {
	t.Helper()
	expected, unsupported, evidenceRequired, err := evaluation.MappingExpectations(control, []string{"team-a"}, index, nil)
	if err != nil || !unsupported {
		t.Fatalf("%s lost its manual-review gap: %v %v", control.ID, unsupported, err)
	}
	in := evaluation.Input{Expected: expected, Unsupported: unsupported, EvidenceRequired: evidenceRequired, EvidenceAt: &now, EvidenceURI: "s3://example/version", Now: now, MaxAge: time.Hour}
	for _, key := range expected {
		in.Observations = append(in.Observations, evaluation.Observation{Key: key, Result: "pass", At: now, Resource: "Pod/team-a/app"})
	}
	result := evaluation.Assess(in)
	if result.Result != "unknown" || result.Coverage != "partial" {
		t.Fatalf("%s incorrectly passed incomplete STIG evidence: %+v", control.ID, result)
	}
	if len(expected) == 0 {
		return false
	}
	in.Observations[0].Result = "fail"
	if evaluation.Assess(in).Result != "fail" {
		t.Fatalf("%s hid a workload violation", control.ID)
	}
	return true
}
