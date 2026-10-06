package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thomasvincent/mantl/pkg/compiler"
)

func convergenceFixture(name, revision string) []byte {
	source := map[string]string{"repoURL": "https://example.com/platform.git", "targetRevision": revision, "path": "deploy/gitops/core"}
	destination := map[string]string{"server": "https://kubernetes.default.svc", "namespace": "argocd"}
	identity := map[string]any{"source": source, "destination": destination}
	app := map[string]any{
		"metadata": map[string]string{"name": name, "namespace": "argocd"},
		"spec":     identity,
		"status": map[string]any{
			"sync":   map[string]any{"status": "Synced", "comparedTo": identity},
			"health": map[string]string{"status": "Healthy"},
		},
	}
	data, _ := json.Marshal(map[string]any{"items": []any{app}})
	return data
}

func TestConvergenceRequiresExpectedIdentity(t *testing.T) {
	expected := []compiler.GitOpsApplication{{Name: "root-platform", Repository: "https://example.com/platform.git", Revision: "v1", Path: "deploy/gitops/core"}}
	for _, data := range [][]byte{convergenceFixture("unrelated", "v1"), convergenceFixture("root-platform", "old"), []byte(`{"items":[]}`), []byte("invalid")} {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		query := func(context.Context) ([]byte, error) {
			calls++
			cancel()
			return data, nil
		}
		if err := waitForApplications(ctx, expected, query, time.Millisecond); !errors.Is(err, context.Canceled) {
			t.Fatalf("incorrect convergence accepted: %v", err)
		}
		if calls != 1 {
			t.Fatalf("query ignored cancellation: %d", calls)
		}
	}
	query := func(context.Context) ([]byte, error) { return convergenceFixture("root-platform", "v1"), nil }
	if err := waitForApplications(context.Background(), expected, query, time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestConvergenceRetriesQueriesAndHonorsDeadline(t *testing.T) {
	expected := []compiler.GitOpsApplication{{Name: "root-platform", Repository: "https://example.com/platform.git", Revision: "v1", Path: "deploy/gitops/core"}}
	calls := 0
	query := func(context.Context) ([]byte, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("fixture unavailable")
		}
		return convergenceFixture("root-platform", "v1"), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := waitForApplications(ctx, expected, query, time.Millisecond); err != nil || calls != 2 {
		t.Fatalf("read failure did not retry: %d, %v", calls, err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	query = func(ctx context.Context) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if err := waitForApplications(ctx, expected, query, time.Millisecond); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline lost: %v", err)
	}
	if err := waitForApplications(context.Background(), nil, query, time.Millisecond); err == nil {
		t.Fatal("missing desired topology accepted")
	}
}

func TestApplicationOutputCannotBypassLimit(t *testing.T) {
	var output applicationOutput
	if _, err := output.Write(bytes.Repeat([]byte("x"), 8<<20)); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Write([]byte("x")); err == nil {
		t.Fatal("oversized response accepted")
	}
}

func TestApplicationQueryIsBoundedReadOnlyAndContextSpecific(t *testing.T) {
	logfile := fakeKubectl(t)
	binary := filepath.Join(filepath.Dir(logfile), "kubectl")
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$KUBECTL_FIXTURE_LOG"
printf '%s' "$APPLICATION_FIXTURE"
`
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	data := convergenceFixture("root-platform", "v1")
	t.Setenv("APPLICATION_FIXTURE", string(data))
	dag := &ExecutionDAG{KubeContext: "fixture"}
	observed, err := dag.queryApplications(context.Background())
	if err != nil || !bytes.Equal(observed, data) {
		t.Fatalf("query differs: %v", err)
	}
	calls, err := os.ReadFile(logfile)
	if err != nil || string(calls) != "--context fixture get applications.argoproj.io -n argocd -o json\n" {
		t.Fatalf("query performed unexpected work: %s, %v", calls, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := dag.queryApplications(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("query cancellation lost: %v", err)
	}
}
