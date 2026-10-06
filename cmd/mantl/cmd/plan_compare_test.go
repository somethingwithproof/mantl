package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPlanComparisonGatePrintsJSONAndPreservesFiles(t *testing.T) {
	spec, outputDir, filename, _ := savedApplyFixture(t)
	data, err := os.ReadFile(spec)
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.Replace(data, []byte("security: true"), []byte("security: false"), 1)
	if err := os.WriteFile(spec, data, 0600); err != nil {
		t.Fatal(err)
	}
	baseline, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(outputDir, "gitops/addon-security.yaml")
	original, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	command := newPlanCommand()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{spec, "--dry-run", "--compare", filename, "--fail-on-change", "--format", "json", "--output-dir", outputDir})
	if err := command.Execute(); err == nil {
		t.Fatal("change gate accepted a removed feature")
	}
	var report planReport
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatalf("change gate did not produce valid JSON: %s, %v", output.String(), err)
	}
	if report.Comparison == nil || !report.Comparison.Changed || len(report.Comparison.Artifacts) != 1 || report.Comparison.Artifacts[0].Action != "removed" {
		t.Fatalf("feature removal missing: %+v", report.Comparison)
	}
	checkUnchangedFile(t, filename, baseline)
	checkUnchangedFile(t, artifact, original)
}

func checkUnchangedFile(t *testing.T, filename string, original []byte) {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil || !bytes.Equal(data, original) {
		t.Fatalf("comparison modified %s: %v", filename, err)
	}
}

func TestPlanUnchangedComparisonGateSucceeds(t *testing.T) {
	spec, _, filename, _ := savedApplyFixture(t)
	command := newPlanCommand()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{spec, "--dry-run", "--compare", filename, "--fail-on-change"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte("changed=false inputChanged=false")) {
		t.Fatal("unchanged comparison absent from text")
	}
}

func TestPlanInvalidReviewOptionsDoNotWrite(t *testing.T) {
	spec := planSpec(t)
	for _, args := range [][]string{{"--dry-run", "--save-plan", filepath.Join(t.TempDir(), "review.json")}, {"--fail-on-change"}, {"--compare", "missing.json"}, {"--save-plan", spec}} {
		outputDir := filepath.Join(t.TempDir(), "build")
		command := newPlanCommand()
		command.SetOut(&bytes.Buffer{})
		command.SetArgs(append([]string{spec, "--output-dir", outputDir}, args...))
		if err := command.Execute(); err == nil {
			t.Fatal("invalid review options accepted")
		}
		if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
			t.Fatal("invalid review options wrote files")
		}
	}
}

func TestPlanCannotSaveInsideGeneratedArtifacts(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "build")
	command := newPlanCommand()
	command.SetArgs([]string{planSpec(t), "--output-dir", outputDir, "--save-plan", filepath.Join(outputDir, "review.json")})
	if err := command.Execute(); err == nil {
		t.Fatal("saved inventory inside artifact directory accepted")
	}
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		t.Fatal("invalid saved path wrote files")
	}
}
