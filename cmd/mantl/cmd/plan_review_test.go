// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasvincent/mantl/pkg/compiler"
)

func TestSavedPlanVerificationWithoutSpec(t *testing.T) {
	for _, format := range []string{"text", "json"} {
		t.Run(format, func(t *testing.T) {
			spec, directory, filename, plan := savedApplyFixture(t)
			if err := os.Remove(spec); err != nil {
				t.Fatal(err)
			}
			original, err := os.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}
			// Extra files are outside the inventory and are not trusted or executed.
			marker := filepath.Join(directory, "unreviewed.yaml")
			if err := os.WriteFile(marker, []byte("preserve"), 0600); err != nil {
				t.Fatal(err)
			}
			command := newPlanCommand()
			var output bytes.Buffer
			command.SetOut(&output)
			command.SetArgs([]string{"verify", filename, "--output-dir", directory, "--format", format})
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if format == "json" {
				var report planVerificationReport
				if err := json.Unmarshal(output.Bytes(), &report); err != nil {
					t.Fatal(err)
				}
				if report.SchemaVersion != "mantl.io/plan-verification/v1alpha1" || report.Scope != "artifact-integrity" || !report.Verified {
					t.Fatalf("unexpected verification report: %+v", report)
				}
				if err := compiler.MatchPlan(plan, report.Inventory); err != nil {
					t.Fatal(err)
				}
				if bytes.Contains(output.Bytes(), []byte(`"Content"`)) {
					t.Fatal("verification exposed artifact contents")
				}
			} else if !strings.Contains(output.String(), "Verified") || !strings.Contains(output.String(), "Artifact integrity only") {
				t.Fatalf("verification scope absent: %s", output.String())
			}
			checkUnchangedFile(t, filename, original)
			checkUnchangedFile(t, marker, []byte("preserve"))
		})
	}
}

func TestSavedPlanVerificationRejectsInvalidArtifacts(t *testing.T) {
	for _, failure := range []string{"missing", "tampered", "symlink", "escape", "schema"} {
		t.Run(failure, func(t *testing.T) {
			_, directory, filename, plan := savedApplyFixture(t)
			artifact := filepath.Join(directory, plan.Artifacts[0].Path)
			data, err := os.ReadFile(artifact)
			if err != nil {
				t.Fatal(err)
			}
			switch failure {
			case "missing", "symlink":
				if err := os.Remove(artifact); err != nil {
					t.Fatal(err)
				}
				if failure == "symlink" {
					target := filepath.Join(t.TempDir(), "artifact")
					if err := os.WriteFile(target, data, 0600); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(target, artifact); err != nil {
						t.Fatal(err)
					}
				}
			case "tampered":
				data[0] ^= 1
				if err := os.WriteFile(artifact, data, 0600); err != nil {
					t.Fatal(err)
				}
			case "escape":
				plan.Artifacts[0].Path = "../outside"
			case "schema":
				plan.SchemaVersion = "mantl.io/plan/v1alpha1"
			}
			if failure == "escape" || failure == "schema" {
				if err := savePlan(plan, filename); err != nil {
					t.Fatal(err)
				}
			}
			command := newPlanCommand()
			var output bytes.Buffer
			command.SetOut(&output)
			command.SetArgs([]string{"verify", filename, "--output-dir", directory, "--format", "json"})
			if err := command.Execute(); err == nil || output.Len() != 0 {
				t.Fatalf("invalid artifacts emitted success: %s, %v", output.String(), err)
			}
		})
	}
}

func TestSavedPlanDiffWithoutSpecOrArtifacts(t *testing.T) {
	for _, change := range []string{"none", "input", "artifacts"} {
		t.Run(change, func(t *testing.T) {
			spec, directory, baseline, plan := savedApplyFixture(t)
			if err := os.Remove(spec); err != nil {
				t.Fatal(err)
			}
			// Move artifacts out of reach; diff should only read inventory metadata.
			if err := os.Rename(directory, directory+"-unavailable"); err != nil {
				t.Fatal(err)
			}
			if change == "input" {
				plan.SpecSHA256 = strings.Repeat("a", 64)
			}
			if change == "artifacts" {
				plan.Artifacts[0].SHA256 = strings.Repeat("b", 64)
				plan.Artifacts = plan.Artifacts[:len(plan.Artifacts)-1]
				added := plan.Artifacts[0]
				added.Path = "added.yaml"
				plan.Artifacts = append(plan.Artifacts, added)
			}
			candidate := filepath.Join(filepath.Dir(baseline), "candidate.json")
			if err := savePlan(plan, candidate); err != nil {
				t.Fatal(err)
			}
			original, err := os.ReadFile(candidate)
			if err != nil {
				t.Fatal(err)
			}
			for _, gate := range []bool{false, true} {
				command := newPlanCommand()
				var output bytes.Buffer
				command.SetOut(&output)
				args := []string{"diff", baseline, candidate, "--format", "json"}
				if gate {
					args = append(args, "--fail-on-change")
				}
				command.SetArgs(args)
				err := command.Execute()
				if (err != nil) != (gate && change != "none") {
					t.Fatalf("unexpected change gate result: %v", err)
				}
				var report savedPlanDiffReport
				if err := json.Unmarshal(output.Bytes(), &report); err != nil {
					t.Fatal(err)
				}
				if report.SchemaVersion != "mantl.io/plan-comparison/v1alpha1" || report.Comparison.Changed != (change != "none") || report.Comparison.InputChanged != (change == "input") {
					t.Fatalf("unexpected comparison: %+v", report)
				}
				if change == "artifacts" {
					actions := map[string]bool{}
					for _, artifact := range report.Comparison.Artifacts {
						actions[artifact.Action] = true
					}
					if len(report.Comparison.Artifacts) != 3 || !actions["added"] || !actions["changed"] || !actions["removed"] {
						t.Fatalf("artifact changes missing: %+v", report.Comparison)
					}
				}
				checkUnchangedFile(t, candidate, original)
			}
		})
	}
}

func TestSavedPlanReviewErrorsAndCancellation(t *testing.T) {
	_, directory, filename, _ := savedApplyFixture(t)
	for _, args := range [][]string{
		{"verify", filename, "--output-dir", directory}, {"diff", filename, filename},
	} {
		for _, format := range []string{"json", "text"} {
			command := newPlanCommand()
			command.SetArgs(append(append([]string{}, args...), "--format", format))
			command.SetOut(failingPlanWriter{})
			if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "output unavailable") {
				t.Fatalf("report error discarded: %v", err)
			}
		}
		command := newPlanCommand()
		command.SetArgs(args)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := command.ExecuteContext(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation lost: %v", err)
		}
		command = newPlanCommand()
		command.SetArgs(append(append([]string{}, args...), "--format", "invalid"))
		if err := command.Execute(); err == nil {
			t.Fatal("invalid format accepted")
		}
	}
	for _, args := range [][]string{{"verify", "missing.json"}, {"verify", filename, "--output-dir", ""}, {"diff", filename, "missing.json"}, {"diff", "missing.json", filename}} {
		command := newPlanCommand()
		var output bytes.Buffer
		command.SetOut(&output)
		command.SetArgs(args)
		if err := command.Execute(); err == nil || output.Len() != 0 {
			t.Fatalf("invalid input reported success: %s, %v", output.String(), err)
		}
	}
}

func TestSavedPlanDiffTextAndInvalidInventory(t *testing.T) {
	_, _, baseline, plan := savedApplyFixture(t)
	candidate := filepath.Join(t.TempDir(), "candidate.json")
	plan.SpecSHA256 = strings.Repeat("c", 64)
	if err := savePlan(plan, candidate); err != nil {
		t.Fatal(err)
	}
	command := newPlanCommand()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{"diff", baseline, candidate, "--fail-on-change"})
	if err := command.Execute(); err == nil || !strings.Contains(output.String(), "changed=true inputChanged=true") {
		t.Fatalf("text gate did not report input-only change: %s, %v", output.String(), err)
	}
	if err := os.WriteFile(candidate, []byte(`{"schemaVersion":"unsupported"}`), 0600); err != nil {
		t.Fatal(err)
	}
	command = newPlanCommand()
	output.Reset()
	command.SetOut(&output)
	command.SetArgs([]string{"diff", baseline, candidate, "--format", "json"})
	if err := command.Execute(); err == nil || output.Len() != 0 {
		t.Fatalf("invalid inventory emitted a comparison: %s, %v", output.String(), err)
	}
}
