// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/bootstrap"
	"github.com/thomasvincent/mantl/pkg/compiler"
)

type recordingApply struct {
	dag     *bootstrap.ExecutionDAG
	calls   []string
	failAt  string
	failure error
	inspect func(context.Context, string) error
}

func (executor *recordingApply) invoke(ctx context.Context, stage string) error {
	executor.calls = append(executor.calls, stage)
	if executor.inspect != nil {
		if err := executor.inspect(ctx, stage); err != nil {
			return err
		}
	}
	if executor.failAt == stage {
		return executor.failure
	}
	return nil
}
func (executor *recordingApply) Preflight(ctx context.Context, _, _ string) error {
	return executor.invoke(ctx, "preflight")
}
func (executor *recordingApply) ProvisionInfra(ctx context.Context, _, _ string) error {
	return executor.invoke(ctx, "provision")
}
func (executor *recordingApply) InstallArgoCD(ctx context.Context) error {
	return executor.invoke(ctx, "install")
}
func (executor *recordingApply) BootstrapGitOps(ctx context.Context) error {
	return executor.invoke(ctx, "bootstrap")
}
func (executor *recordingApply) VerifyConvergence(ctx context.Context) error {
	return executor.invoke(ctx, "convergence")
}

func applyTestCommand(t *testing.T, executor *recordingApply) *cobra.Command {
	t.Helper()
	previous := kubeContext
	kubeContext = "fixture-context"
	t.Cleanup(func() { kubeContext = previous })
	command := newApplyCommand(func(dag *bootstrap.ExecutionDAG) applyExecutor {
		executor.dag = dag
		return executor
	})
	command.SetOut(&bytes.Buffer{})
	return command
}

func assertPrivateApplyBuild(t *testing.T, executor *recordingApply, outputDir string) {
	t.Helper()
	if executor.dag.BuildDir == outputDir || executor.dag.BuildDir == "" {
		t.Fatal("execution is not using a private artifact snapshot")
	}
	info, err := os.Stat(executor.dag.BuildDir)
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatalf("private directory permissions: %v", err)
	}
	for _, application := range executor.dag.Applications {
		file := filepath.Join(executor.dag.BuildDir, "gitops", application.Name+".yaml")
		info, err := os.Stat(file)
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("private application permissions: %v", err)
		}
	}
}

func TestApplyPreflightPrecedesWritesAndUsesPrivateInputs(t *testing.T) {
	executor := &recordingApply{}
	command := applyTestCommand(t, executor)
	outputDir := filepath.Join(t.TempDir(), "generated")
	executor.inspect = func(ctx context.Context, stage string) error {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("total deadline missing")
		}
		if stage == "preflight" {
			if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
				t.Fatal("output created before preflight")
			}
		} else {
			assertPrivateApplyBuild(t, executor, outputDir)
		}
		return nil
	}
	command.SetArgs([]string{planSpec(t), "--output-dir", outputDir})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(executor.calls, []string{"preflight", "provision", "install", "bootstrap", "convergence"}) {
		t.Fatalf("unexpected execution order: %v", executor.calls)
	}
	if kubeContext != "fixture-context" || executor.dag.KubeContext != kubeContext {
		t.Fatal("target context changed")
	}
	if _, err := os.Stat(executor.dag.BuildDir); !os.IsNotExist(err) {
		t.Fatal("private snapshot leaked after execution")
	}
}

func savedApplyFixture(t *testing.T) (string, string, string, compiler.Plan) {
	t.Helper()
	spec := planSpec(t)
	parent := t.TempDir()
	outputDir := filepath.Join(parent, "build")
	filename := filepath.Join(parent, "reviewed.json")
	command := newPlanCommand()
	command.SetOut(&bytes.Buffer{})
	command.SetArgs([]string{spec, "--output-dir", outputDir, "--save-plan", filename})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	plan, err := compiler.ReadPlan(filename)
	if err != nil {
		t.Fatal(err)
	}
	return spec, outputDir, filename, plan
}

func TestReviewedApplyKeepsVerifiedBytesAfterSourceChanges(t *testing.T) {
	spec, outputDir, filename, plan := savedApplyFixture(t)
	artifact := plan.Artifacts[0]
	expected, err := os.ReadFile(filepath.Join(outputDir, artifact.Path))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "gitops/unreviewed.yaml"), []byte("unreviewed"), 0600); err != nil {
		t.Fatal(err)
	}
	executor := &recordingApply{}
	command := applyTestCommand(t, executor)
	executor.inspect = func(_ context.Context, stage string) error {
		if stage == "preflight" {
			return os.WriteFile(filepath.Join(outputDir, artifact.Path), []byte("changed after verification"), 0600)
		}
		assertPrivateApplyBuild(t, executor, outputDir)
		data, err := os.ReadFile(filepath.Join(executor.dag.BuildDir, artifact.Path))
		if err != nil || !bytes.Equal(data, expected) {
			t.Fatalf("execution bytes changed after verification: %v", err)
		}
		if _, err := os.Stat(filepath.Join(executor.dag.BuildDir, "gitops/unreviewed.yaml")); !os.IsNotExist(err) {
			t.Fatal("unreviewed input entered snapshot")
		}
		return nil
	}
	command.SetArgs([]string{spec, "--plan", filename, "--output-dir", outputDir})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestReviewedApplyRejectsChangesBeforePreflight(t *testing.T) {
	for _, kind := range []string{"artifact", "spec"} {
		t.Run(kind, func(t *testing.T) {
			spec, outputDir, filename, plan := savedApplyFixture(t)
			changed := filepath.Join(outputDir, plan.Artifacts[0].Path)
			data := []byte("tampered")
			if kind == "spec" {
				changed = spec
				original, err := os.ReadFile(spec)
				if err != nil {
					t.Fatal(err)
				}
				data = bytes.Replace(original, []byte("exposure: private"), []byte("exposure: hybrid"), 1)
			}
			if err := os.WriteFile(changed, data, 0600); err != nil {
				t.Fatal(err)
			}
			executor := &recordingApply{}
			command := applyTestCommand(t, executor)
			command.SetArgs([]string{spec, "--plan", filename, "--output-dir", outputDir})
			if err := command.Execute(); err == nil || len(executor.calls) != 0 {
				t.Fatalf("unreviewed input reached execution: %v, %v", err, executor.calls)
			}
		})
	}
}

func TestApplyStopsOnEveryFailedStage(t *testing.T) {
	stages := []string{"preflight", "provision", "install", "bootstrap", "convergence"}
	failure := errors.New("fixture stage failed")
	for i, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			executor := &recordingApply{failAt: stage, failure: failure}
			command := applyTestCommand(t, executor)
			outputDir := filepath.Join(t.TempDir(), "build")
			command.SetArgs([]string{planSpec(t), "--output-dir", outputDir})
			if err := command.Execute(); !errors.Is(err, failure) {
				t.Fatalf("underlying failure lost: %v", err)
			}
			if !reflect.DeepEqual(executor.calls, stages[:i+1]) {
				t.Fatalf("continued after failed stage: %v", executor.calls)
			}
			if stage == "preflight" {
				if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
					t.Fatal("failed preflight wrote generated output")
				}
			} else if _, err := os.Stat(executor.dag.BuildDir); !os.IsNotExist(err) {
				t.Fatal("failed apply leaked its private snapshot")
			}
		})
	}
}

func TestApplyCancellationStopsNextStage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	executor := &recordingApply{}
	command := applyTestCommand(t, executor)
	executor.inspect = func(_ context.Context, stage string) error {
		if stage == "provision" {
			cancel()
		}
		return nil
	}
	command.SetArgs([]string{planSpec(t), "--output-dir", t.TempDir()})
	if err := command.ExecuteContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	if !reflect.DeepEqual(executor.calls, []string{"preflight", "provision"}) {
		t.Fatalf("continued after cancellation: %v", executor.calls)
	}
}

func TestApplyRejectsInvalidOptionsAndMissingCloudContext(t *testing.T) {
	for _, args := range [][]string{{"--timeout", "0"}, {"--timeout", "3h"}, {"--output-dir", ""}, {"--plan", "missing.json"}} {
		executor := &recordingApply{}
		command := applyTestCommand(t, executor)
		command.SetArgs(append([]string{planSpec(t)}, args...))
		if err := command.Execute(); err == nil || len(executor.calls) != 0 {
			t.Fatalf("invalid options reached execution: %v", err)
		}
	}
	executor := &recordingApply{}
	command := applyTestCommand(t, executor)
	kubeContext = ""
	command.SetArgs([]string{planSpec(t)})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "--context") || len(executor.calls) != 0 {
		t.Fatalf("implicit cloud context accepted: %v", err)
	}
}

func TestLocalApplyDoesNotChangeGlobalContext(t *testing.T) {
	spec := planSpec(t)
	data, err := os.ReadFile(spec)
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.Replace(data, []byte("kind: aws"), []byte("kind: local"), 1)
	data = bytes.Replace(data, []byte("distribution: eks"), []byte("distribution: kind"), 1)
	if err := os.WriteFile(spec, data, 0600); err != nil {
		t.Fatal(err)
	}
	executor := &recordingApply{}
	command := applyTestCommand(t, executor)
	kubeContext = ""
	command.SetArgs([]string{spec, "--output-dir", t.TempDir()})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if kubeContext != "" || executor.dag.KubeContext != "kind-prod-us-east" {
		t.Fatal("local context leaked into global state")
	}
}

func TestApplyTotalDeadlineAndOutputFailure(t *testing.T) {
	executor := &recordingApply{}
	command := applyTestCommand(t, executor)
	executor.inspect = func(ctx context.Context, _ string) error {
		<-ctx.Done()
		return ctx.Err()
	}
	outputDir := filepath.Join(t.TempDir(), "build")
	command.SetArgs([]string{planSpec(t), "--output-dir", outputDir, "--timeout", "1ms"})
	if err := command.Execute(); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("total deadline lost: %v", err)
	}
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		t.Fatal("expired preflight wrote artifacts")
	}
	executor = &recordingApply{}
	command = applyTestCommand(t, executor)
	command.SetOut(failingPlanWriter{})
	command.SetArgs([]string{planSpec(t), "--output-dir", t.TempDir()})
	if err := command.Execute(); err == nil || !reflect.DeepEqual(executor.calls, []string{"preflight"}) {
		t.Fatalf("progress output failure reached provisioning: %v, %v", err, executor.calls)
	}
}
