package bootstrap

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func preflightFixture(t *testing.T) (*ExecutionDAG, string) {
	t.Helper()
	logfile := fakeKubectl(t)
	binary := filepath.Join(filepath.Dir(logfile), "kubectl")
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$KUBECTL_FIXTURE_LOG"
if [ "$PREFLIGHT_FAIL" = "true" ]; then exit 1; fi
printf '%s\n' "$PREFLIGHT_CONTEXT"
`
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	tools := t.TempDir()
	for _, name := range []string{"terraform", "kind"} {
		if err := os.WriteFile(filepath.Join(tools, name), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", tools)
	t.Setenv("PREFLIGHT_CONTEXT", "fixture")
	t.Setenv("PREFLIGHT_FAIL", "false")
	source := t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "infra/terraform/blueprints/aws-eks"), 0700); err != nil {
		t.Fatal(err)
	}
	return &ExecutionDAG{SourceDir: source, Distribution: "eks", KubeContext: "fixture"}, logfile
}

func TestPreflightUsesOnlyLocalContextConfiguration(t *testing.T) {
	dag, logfile := preflightFixture(t)
	if err := dag.Preflight(context.Background(), "aws", "fixture"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(logfile)
	if err != nil || strings.TrimSpace(string(data)) != "--context fixture config get-contexts fixture -o name" {
		t.Fatalf("preflight performed unexpected work: %s, %v", data, err)
	}
	if err := dag.Preflight(context.Background(), "local", "fixture"); err != nil {
		t.Fatal(err)
	}
}

func TestPreflightRejectsMissingPrerequisites(t *testing.T) {
	for _, missing := range []string{"context", "distribution", "blueprint", "blueprint-directory", "terraform", "kind", "kubectl", "local-name", "context-query", "context-configuration"} {
		t.Run(missing, func(t *testing.T) {
			dag, _ := preflightFixture(t)
			provider, name := "aws", "fixture"
			invalidatePreflightFixture(t, dag, missing, &provider, &name)
			if err := dag.Preflight(context.Background(), provider, name); err == nil {
				t.Fatalf("missing %s passed preflight", missing)
			}
		})
	}
}

func invalidatePreflightFixture(t *testing.T, dag *ExecutionDAG, missing string, provider, name *string) {
	t.Helper()
	switch missing {
	case "context":
		dag.KubeContext = ""
	case "distribution":
		dag.Distribution = ""
	case "blueprint":
		dag.SourceDir = t.TempDir()
	case "blueprint-directory":
		file := filepath.Join(dag.SourceDir, "infra/terraform/blueprints/aws-eks")
		if err := os.Remove(file); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("not a directory"), 0600); err != nil {
			t.Fatal(err)
		}
	case "terraform":
		t.Setenv("PATH", t.TempDir())
	case "kind":
		*provider = "local"
		t.Setenv("PATH", t.TempDir())
	case "kubectl":
		t.Setenv("MANTL_KUBECTL_PATH", filepath.Join(t.TempDir(), "missing"))
	case "local-name":
		*provider, *name = "local", "Invalid"
	case "context-query":
		t.Setenv("PREFLIGHT_FAIL", "true")
	case "context-configuration":
		t.Setenv("PREFLIGHT_CONTEXT", "other")
	default:
		t.Fatalf("unknown fixture %s", missing)
	}
}

func TestPreflightPreservesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (&ExecutionDAG{}).Preflight(ctx, "local", "fixture"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}
