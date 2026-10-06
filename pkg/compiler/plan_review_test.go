package compiler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
)

func reviewFixture(t *testing.T) Plan {
	t.Helper()
	plan, err := Compile(newTestCluster("small"))
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func decodeFixture(t *testing.T, plan Plan) Plan {
	t.Helper()
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodePlan(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

func TestReviewedPlanDetectsUnrenderedInputChanges(t *testing.T) {
	cluster := newTestCluster("small")
	before, err := Compile(cluster)
	if err != nil {
		t.Fatal(err)
	}
	// Exposure is not rendered into current artifacts, but is still desired input.
	cluster.Spec.Networking.Exposure = "private"
	after, err := Compile(cluster)
	if err != nil {
		t.Fatal(err)
	}
	diff, err := ComparePlans(before, after)
	if err != nil || !diff.Changed || !diff.InputChanged || len(diff.Artifacts) != 0 {
		t.Fatalf("unrendered input change lost: %+v, %v", diff, err)
	}
	if err := MatchPlan(before, after); err == nil {
		t.Fatal("changed spec accepted as reviewed")
	}
	if err := MatchPlan(decodeFixture(t, before), before); err != nil {
		t.Fatal(err)
	}
}

func TestComparisonIncludesFeatureAndTenantRemoval(t *testing.T) {
	cluster := newTestCluster("small")
	cluster.Spec.Features.Security = true
	cluster.Spec.GitOps.TenantPath = "tenants"
	cluster.Spec.Tenants = []v1alpha1.TenantSpec{{Name: "payments"}}
	before, err := Compile(cluster)
	if err != nil {
		t.Fatal(err)
	}
	cluster.Spec.Features.Security = false
	cluster.Spec.Features.Compliance = true
	cluster.Spec.Tenants = nil
	after, err := Compile(cluster)
	if err != nil {
		t.Fatal(err)
	}
	diff, err := ComparePlans(before, after)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, artifact := range diff.Artifacts {
		got[artifact.Path] = artifact.Action
	}
	want := map[string]string{"gitops/addon-security.yaml": "removed", "gitops/addon-compliance.yaml": "added", "tenants/payments.yaml": "removed", "tenants/kustomization.yaml": "changed"}
	if !reflect.DeepEqual(got, want) || !diff.Changed || !diff.InputChanged {
		t.Fatalf("unexpected diff: %+v", diff)
	}
	unchanged, err := ComparePlans(after, after)
	if err != nil || unchanged.Changed || unchanged.Artifacts == nil {
		t.Fatalf("unchanged inventory: %+v, %v", unchanged, err)
	}
}

func TestDecodePlanRejectsMalformedMetadata(t *testing.T) {
	plan := reviewFixture(t)
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	inputs := []string{"null", "{}", string(data) + " {}", strings.Replace(string(data), `"platform":`, `"unknown":1,"platform":`, 1), strings.Repeat(" ", MaxPlanBytes+1)}
	for _, input := range inputs {
		if _, err := DecodePlan(strings.NewReader(input)); err == nil {
			t.Fatal("invalid metadata accepted")
		}
	}
	for _, invalid := range invalidReviewPlans(plan) {
		if err := ValidatePlan(invalid); err == nil {
			t.Fatalf("invalid inventory accepted: %+v", invalid)
		}
	}
}

func invalidReviewPlans(valid Plan) []Plan {
	plans := []Plan{}
	for _, path := range []string{"../escape", "/absolute", ".", "a/../b", `a\b`, ""} {
		changed := valid
		changed.Artifacts = []Artifact{{Path: path, Size: 1, SHA256: valid.Artifacts[0].SHA256}}
		plans = append(plans, changed)
	}
	oldSchema := valid
	oldSchema.SchemaVersion = "mantl.io/plan/v1alpha1"
	noHash := valid
	noHash.SpecSHA256 = ""
	duplicate := valid
	duplicate.Artifacts = append([]Artifact{valid.Artifacts[0]}, valid.Artifacts[0])
	oversized := valid
	oversized.Artifacts = []Artifact{{Path: "large", Size: MaxArtifactBytes + 1, SHA256: valid.Artifacts[0].SHA256}}
	badHash := valid
	badHash.Artifacts = []Artifact{{Path: "bad", Size: 1, SHA256: "bad"}}
	return append(plans, oldSchema, noHash, duplicate, oversized, badHash)
}

func writeReviewArtifacts(t *testing.T, plan Plan, directory string) {
	t.Helper()
	for _, artifact := range plan.Artifacts {
		filename := filepath.Join(directory, artifact.Path)
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, artifact.Content, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReadReviewedArtifactsOwnsVerifiedBytes(t *testing.T) {
	plan := reviewFixture(t)
	directory := t.TempDir()
	writeReviewArtifacts(t, plan, directory)
	verified, err := ReadPlanArtifacts(decodeFixture(t, plan), directory)
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(directory, verified.Artifacts[0].Path)
	if err := os.WriteFile(filename, []byte("changed later"), 0600); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(verified.Artifacts[0].Content, plan.Artifacts[0].Content) {
		t.Fatal("source mutation altered verified bytes")
	}
	if _, err := ReadPlanArtifacts(plan, directory); err == nil {
		t.Fatal("tampering accepted")
	}
}

func TestReviewedArtifactsRejectMissingTamperedAndSymlinkedFiles(t *testing.T) {
	for _, mode := range []string{"missing", "same-size-tampering", "symlink", "directory", "escaping-directory"} {
		t.Run(mode, func(t *testing.T) {
			plan := reviewFixture(t)
			directory := t.TempDir()
			writeReviewArtifacts(t, plan, directory)
			artifact := plan.Artifacts[0]
			filename := filepath.Join(directory, artifact.Path)
			corruptReviewArtifact(t, mode, directory, filename, artifact)
			if _, err := ReadPlanArtifacts(plan, directory); err == nil {
				t.Fatalf("%s accepted", mode)
			}
		})
	}
}

func corruptReviewArtifact(t *testing.T, mode, directory, filename string, artifact Artifact) {
	t.Helper()
	if err := os.Remove(filename); err != nil {
		t.Fatal(err)
	}
	var err error
	switch mode {
	case "missing":
		return
	case "same-size-tampering":
		err = os.WriteFile(filename, bytes.Repeat([]byte("x"), artifact.Size), 0600)
	case "symlink":
		err = os.Symlink(filepath.Join(t.TempDir(), "outside"), filename)
	case "directory":
		err = os.Mkdir(filename, 0700)
	case "escaping-directory":
		if err := os.RemoveAll(filepath.Join(directory, "gitops")); err != nil {
			t.Fatal(err)
		}
		outside := t.TempDir()
		if err := os.WriteFile(filepath.Join(outside, filepath.Base(filename)), artifact.Content, 0600); err != nil {
			t.Fatal(err)
		}
		err = os.Symlink(outside, filepath.Join(directory, "gitops"))
	default:
		t.Fatalf("unknown corruption mode %s", mode)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadPlanRequiresBoundedRegularMetadata(t *testing.T) {
	plan := reviewFixture(t)
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	read, err := ReadPlan(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := MatchPlan(read, plan); err != nil {
		t.Fatal(err)
	}
	for _, filename := range []string{filepath.Join(t.TempDir(), "missing.json"), t.TempDir()} {
		if _, err := ReadPlan(filename); err == nil {
			t.Fatal("invalid metadata file accepted")
		}
	}
	invalid := plan
	invalid.SpecSHA256 = "invalid"
	if err := MatchPlan(invalid, plan); err == nil {
		t.Fatal("invalid reviewed hash accepted")
	}
	if err := MatchPlan(plan, invalid); err == nil {
		t.Fatal("invalid current hash accepted")
	}
	if _, err := ComparePlans(invalid, plan); err == nil {
		t.Fatal("invalid baseline compared")
	}
}
