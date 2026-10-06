// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"
	"errors"
	"github.com/thomasvincent/mantl/pkg/compiler"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyConvergenceHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	dag := &ExecutionDAG{}
	if err := dag.VerifyConvergence(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled convergence: %v", err)
	}
}

func TestCommandPropagatesCancellationWithoutExecuting(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	dag := &ExecutionDAG{}
	if err := dag.command(ctx, executable).Run(); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled command: %v", err)
	}
}

func TestCommandKeepsExplicitKubernetesContext(t *testing.T) {
	dag := &ExecutionDAG{KubeContext: "fixture-context"}
	command := dag.command(context.Background(), "kubectl", "get", "applications")
	if strings.Join(command.Args[1:], " ") != "--context fixture-context get applications" {
		t.Fatalf("unexpected context arguments: %v", command.Args)
	}
}

func fakeKubectl(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	binary := filepath.Join(directory, "kubectl")
	logfile := filepath.Join(directory, "calls")
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$KUBECTL_FIXTURE_LOG"
case "$*" in
  *"apply -f -") cat >/dev/null ;;
esac
case "$KUBECTL_FIXTURE_FAIL" in
  namespace) exit 1 ;;
  install) case "$*" in *"https://"*) exit 1 ;; esac ;;
  wait) case "$*" in *" wait "*) exit 1 ;; esac ;;
esac
`
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MANTL_KUBECTL_PATH", binary)
	t.Setenv("KUBECTL_FIXTURE_LOG", logfile)
	t.Setenv("KUBECTL_FIXTURE_FAIL", "")
	return logfile
}

func TestArgoCDInstallationKeepsContextAndStopsAtFailedPhase(t *testing.T) {
	for _, phase := range []string{"", "namespace", "install", "wait"} {
		t.Run(phase, func(t *testing.T) {
			logfile := fakeKubectl(t)
			t.Setenv("KUBECTL_FIXTURE_FAIL", phase)
			dag := &ExecutionDAG{KubeContext: "fixture"}
			err := dag.InstallArgoCD(context.Background())
			if (phase == "") != (err == nil) {
				t.Fatalf("installation phase %q: %v", phase, err)
			}
			data, readErr := os.ReadFile(logfile)
			if readErr != nil {
				t.Fatal(readErr)
			}
			calls := strings.Split(strings.TrimSpace(string(data)), "\n")
			expected := map[string]int{"": 3, "namespace": 1, "install": 2, "wait": 3}[phase]
			if len(calls) != expected {
				t.Fatalf("installation continued after failed phase: %v", calls)
			}
			for _, call := range calls {
				if !strings.HasPrefix(call, "--context fixture ") {
					t.Fatalf("implicit cluster context: %s", call)
				}
			}
		})
	}
}

func TestBootstrapGitOpsAppliesOnlyGeneratedManifests(t *testing.T) {
	logfile := fakeKubectl(t)
	build := t.TempDir()
	directory := filepath.Join(build, "gitops")
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"root.yaml", "notes.txt", "unreviewed.yaml"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	dag := &ExecutionDAG{BuildDir: build, KubeContext: "fixture", Applications: []compiler.GitOpsApplication{{Name: "root"}}}
	if err := dag.BootstrapGitOps(context.Background()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(logfile)
	if err != nil || strings.Contains(string(data), "notes.txt") || strings.Contains(string(data), "unreviewed.yaml") || !strings.Contains(string(data), "root.yaml") {
		t.Fatalf("unexpected bootstrap inputs: %s, %v", data, err)
	}
	t.Setenv("KUBECTL_FIXTURE_FAIL", "namespace")
	if err := dag.BootstrapGitOps(context.Background()); err == nil {
		t.Fatal("bootstrap ignored failed manifest application")
	}
}

func TestProvisionInfra_MissingDistributionReturnsError(t *testing.T) {
	dag := &ExecutionDAG{
		BuildDir:     t.TempDir(),
		Distribution: "", // empty
	}
	err := dag.ProvisionInfra(context.Background(), "aws", "test-cluster")
	if err == nil {
		t.Fatal("expected error for missing distribution, got nil")
	}
	if !strings.Contains(err.Error(), "distribution is required") {
		t.Errorf("error = %q, want it to contain 'distribution is required'", err.Error())
	}
}

func TestProvisionInfra_MissingBlueprintReturnsError(t *testing.T) {
	dag := &ExecutionDAG{
		BuildDir:     t.TempDir(),
		Distribution: "unknown-distro",
	}
	err := dag.ProvisionInfra(context.Background(), "nonexistent-cloud", "test-cluster")
	if err == nil {
		t.Fatal("expected error for unknown blueprint, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to contain 'not found'", err.Error())
	}
}

func TestCreateLocalCluster_InvalidNameRejected(t *testing.T) {
	dag := &ExecutionDAG{}
	tests := []struct {
		name        string
		clusterName string
	}{
		{"name with uppercase", "MyCluster"},
		{"name starting with digit", "1cluster"},
		{"name with underscore", "my_cluster"},
		{"empty name", ""},
		{"name too long", strings.Repeat("a", 254)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dag.CreateLocalCluster(context.Background(), tt.clusterName)
			if err == nil {
				t.Fatalf("expected error for invalid cluster name %q, got nil", tt.clusterName)
			}
			if !strings.Contains(err.Error(), "invalid cluster name") {
				t.Errorf("error = %q, want it to contain 'invalid cluster name'", err.Error())
			}
		})
	}
}

func TestBootstrapValidatesAllInputsBeforeApplying(t *testing.T) {
	logfile := fakeKubectl(t)
	build := t.TempDir()
	if err := os.Mkdir(filepath.Join(build, "gitops"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(build, "gitops/root.yaml"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, applications := range [][]compiler.GitOpsApplication{nil, {{Name: "../escape"}}, {{Name: "root"}, {Name: "missing"}}} {
		dag := &ExecutionDAG{BuildDir: build, Applications: applications}
		if err := dag.BootstrapGitOps(context.Background()); err == nil {
			t.Fatal("invalid bootstrap inputs accepted")
		}
	}
	if _, err := os.Stat(logfile); !os.IsNotExist(err) {
		t.Fatal("invalid topology reached kubectl apply")
	}
}
