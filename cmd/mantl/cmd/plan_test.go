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

func planSpec(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../../../examples/mantl-spec.yaml")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "spec.yaml")
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestPlanPreviewAndGeneration(t *testing.T) {
	spec := planSpec(t)
	outputDir := filepath.Join(t.TempDir(), "generated")
	run := func(dryRun bool) compiler.Plan {
		t.Helper()
		command := newPlanCommand()
		var output bytes.Buffer
		command.SetOut(&output)
		args := []string{spec, "--format", "json", "--output-dir", outputDir}
		if dryRun {
			args = append(args, "--dry-run")
		}
		command.SetArgs(args)
		if err := command.Execute(); err != nil {
			t.Fatal(err)
		}
		var plan compiler.Plan
		if err := json.Unmarshal(output.Bytes(), &plan); err != nil {
			t.Fatalf("invalid JSON: %s: %v", output.String(), err)
		}
		return plan
	}
	preview := run(true)
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		t.Fatal("dry-run touched output directory")
	}
	generated := run(false)
	before, _ := json.Marshal(preview)
	after, _ := json.Marshal(generated)
	if !bytes.Equal(before, after) {
		t.Fatal("preview differs from generated inventory")
	}
	for _, artifact := range generated.Artifacts {
		if _, err := os.Stat(filepath.Join(outputDir, artifact.Path)); err != nil {
			t.Fatal(err)
		}
	}
	// A preview must preserve existing files, including unrelated user files.
	marker := filepath.Join(outputDir, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	run(true)
	if data, err := os.ReadFile(marker); err != nil || string(data) != "keep" {
		t.Fatal("preview modified existing output")
	}
}

func TestPlanErrorsDoNotGenerate(t *testing.T) {
	spec := planSpec(t)
	for _, args := range [][]string{{spec, "--format", "invalid"}, {spec, "--output-dir", ""}, {"missing.yaml"}} {
		outputDir := filepath.Join(t.TempDir(), "generated")
		command := newPlanCommand()
		command.SetArgs(append([]string{"--output-dir", outputDir}, args...))
		var output bytes.Buffer
		command.SetOut(&output)
		command.SetErr(&output)
		if err := command.Execute(); err == nil {
			t.Fatal("expected error")
		}
		if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
			t.Fatal("invalid request wrote output")
		}
	}
	command := newPlanCommand()
	command.SetArgs([]string{spec, "--dry-run"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := command.ExecuteContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}

type failingPlanWriter struct{}

func (failingPlanWriter) Write([]byte) (int, error) { return 0, errors.New("output unavailable") }

func TestPlanOutputFailureAndText(t *testing.T) {
	for _, format := range []string{"json", "text"} {
		command := newPlanCommand()
		command.SetArgs([]string{planSpec(t), "--dry-run", "--format", format})
		command.SetOut(failingPlanWriter{})
		if err := command.Execute(); err == nil {
			t.Fatal("output error discarded")
		}
	}
	var output bytes.Buffer
	if err := writePlan(&output, "text", compiler.Plan{Platform: "fixture", Artifacts: []compiler.Artifact{{Path: "file", SHA256: "hash", Size: 4}}, Notices: []string{"offline"}}, "build", true); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte("Would generate 1 artifacts")) || !bytes.Contains(output.Bytes(), []byte("hash  file")) || !bytes.Contains(output.Bytes(), []byte("offline")) {
		t.Fatalf("incomplete text preview: %s", output.String())
	}
}

func TestPlanGenerationFailurePreservesError(t *testing.T) {
	for _, artifactType := range []string{"terraform", "gitops"} {
		t.Run(artifactType, func(t *testing.T) {
			outputDir := t.TempDir()
			blocked := filepath.Join(outputDir, "gitops")
			if artifactType == "terraform" {
				blocked = outputDir
				if err := os.Remove(outputDir); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(blocked, []byte("preserve"), 0600); err != nil {
				t.Fatal(err)
			}
			command := newPlanCommand()
			command.SetArgs([]string{planSpec(t), "--output-dir", outputDir})
			var output bytes.Buffer
			command.SetOut(&output)
			err := command.Execute()
			var pathError *os.PathError
			if !errors.As(err, &pathError) || !strings.Contains(err.Error(), "write "+artifactType+" artifacts to "+outputDir) {
				t.Fatalf("write error lacks context or underlying cause: %v", err)
			}
			if output.Len() != 0 {
				t.Fatal("failed generation emitted a success inventory")
			}
			if data, err := os.ReadFile(blocked); err != nil || string(data) != "preserve" {
				t.Fatal("write failure modified blocking file")
			}
		})
	}
}
