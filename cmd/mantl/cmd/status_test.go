package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/thomasvincent/mantl/pkg/compiler"
	"github.com/thomasvincent/mantl/pkg/platformstatus"
)

func statusSpec(t *testing.T) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "spec.yaml")
	const spec = `apiVersion: platform.mantl.io/v1alpha1
kind: MantlCluster
metadata:
  name: test-platform
spec:
  provider: {kind: local, region: local}
  kubernetes: {distribution: kind, version: "1.35"}
  profile: {size: small}
  networking: {domain: test.example}
  features: {compliance: true}
`
	if err := os.WriteFile(file, []byte(spec), 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

func statusFixture(t *testing.T, file string) []byte {
	t.Helper()
	cluster, err := compiler.ParseSpec(file)
	if err != nil {
		t.Fatal(err)
	}
	var items []any
	for _, app := range compiler.GitOpsApplications(cluster) {
		source := map[string]string{"repoURL": app.Repository, "targetRevision": app.Revision, "path": app.Path}
		dest := map[string]string{"server": "https://kubernetes.default.svc", "namespace": "argocd"}
		items = append(items, map[string]any{"metadata": map[string]string{"name": app.Name, "namespace": "argocd"}, "spec": map[string]any{"source": source, "destination": dest}, "status": map[string]any{"health": map[string]string{"status": "Healthy"}, "sync": map[string]any{"status": "Synced", "comparedTo": map[string]any{"source": source, "destination": dest}}}})
	}
	data, err := json.Marshal(map[string]any{"items": items})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestStatusJSONAndGate(t *testing.T) {
	file := statusSpec(t)
	fixture := statusFixture(t, file)
	origContext := kubeContext
	kubeContext = "fixture-context"
	t.Cleanup(func() { kubeContext = origContext })
	stamp := time.Date(2026, 10, 6, 1, 2, 3, 0, time.UTC)
	var calls [][]string
	query := func(ctx context.Context, args ...string) ([]byte, error) {
		calls = append(calls, args)
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("query has no deadline")
		}
		if strings.Contains(strings.Join(args, " "), "applications.argoproj.io") {
			return fixture, nil
		}
		return []byte(`{"items":[{"summary":{"pass":5,"fail":1,"error":0,"warn":2,"skip":3}}]}`), nil
	}
	command := newStatusCommand(query, func() time.Time { return stamp })
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{file, "--format=json", "--require-healthy"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	var report platformstatus.Report
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != "mantl.io/status/v1alpha1" || report.Platform != "test-platform" || report.Context != kubeContext || report.ObservedAt != stamp || report.Applications.State != platformstatus.Healthy || len(report.Applications.Applications) != 2 || report.Policies.Counts.Fail != 1 {
		t.Fatalf("report=%+v", report)
	}
	if len(calls) != 2 || calls[0][0] != "--context" || calls[0][1] != kubeContext {
		t.Fatal(calls)
	}
	// Policy failures do not turn application convergence into certification or
	// make this explicit application gate a hidden compliance gate.
	if report.Policies.State != "Observed" {
		t.Fatal(report.Policies)
	}
}

func TestStatusQueryFailureAndHealthGate(t *testing.T) {
	for _, gate := range []bool{false, true} {
		t.Run(stringBool(gate), func(t *testing.T) {
			command := newStatusCommand(func(context.Context, ...string) ([]byte, error) {
				return []byte("secret-plugin-output"), errors.New("secret-error")
			}, time.Now)
			var output bytes.Buffer
			command.SetOut(&output)
			command.SetErr(io.Discard)
			args := []string{statusSpec(t), "--format=json"}
			if gate {
				args = append(args, "--require-healthy")
			}
			command.SetArgs(args)
			err := command.Execute()
			if (err != nil) != gate {
				t.Fatalf("gate=%v err=%v", gate, err)
			}
			var report platformstatus.Report
			if err := json.Unmarshal(output.Bytes(), &report); err != nil {
				t.Fatal(err)
			}
			if report.Applications.State != platformstatus.Unknown || report.Policies.Counts != nil || strings.Contains(output.String(), "secret") {
				t.Fatal(output.String())
			}
		})
	}
}

func stringBool(value bool) string {
	if value {
		return "gate"
	}
	return "observation"
}

func TestStatusValidationAndCancellation(t *testing.T) {
	file := statusSpec(t)
	for _, args := range [][]string{{file, "--format=yaml"}, {file, "--timeout=0s"}, {file, "--timeout=2m"}, {"missing-spec"}, {file, "extra"}} {
		calls := 0
		command := newStatusCommand(func(context.Context, ...string) ([]byte, error) { calls++; return nil, nil }, time.Now)
		command.SetArgs(args)
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)
		if err := command.Execute(); err == nil || calls != 0 {
			t.Fatalf("args=%v calls=%d err=%v", args, calls, err)
		}
	}
	calls := 0
	command := newStatusCommand(func(context.Context, ...string) ([]byte, error) { calls++; return nil, nil }, time.Now)
	command.SetArgs([]string{file})
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := command.ExecuteContext(ctx); !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestStatusDeadlinePropagates(t *testing.T) {
	command := newStatusCommand(func(ctx context.Context, _ ...string) ([]byte, error) { <-ctx.Done(); return nil, ctx.Err() }, time.Now)
	command.SetArgs([]string{statusSpec(t), "--timeout=1ms", "--format=json", "--require-healthy"})
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	if err := command.Execute(); err == nil {
		t.Fatal("deadline must not satisfy health gate")
	}
}

type failingStatusWriter struct{}

func (failingStatusWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestStatusOutputFailureAndBounds(t *testing.T) {
	for _, format := range []string{"text", "json"} {
		if err := writeStatus(failingStatusWriter{}, format, platformstatus.Report{}); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatal(err)
		}
	}
	var output boundedStatusOutput
	if _, err := output.Write(make([]byte, maxStatusOutput)); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Write([]byte("extra")); err == nil || output.buffer.Len() != maxStatusOutput {
		t.Fatal("unbounded query output")
	}
	var copied boundedStatusOutput
	if _, err := io.Copy(&copied, io.LimitReader(strings.NewReader(strings.Repeat("x", maxStatusOutput+1)), maxStatusOutput+1)); err == nil || copied.buffer.Len() > maxStatusOutput {
		t.Fatal("io.Copy bypassed status response bound")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := queryStatus(ctx, "get", "applications"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
