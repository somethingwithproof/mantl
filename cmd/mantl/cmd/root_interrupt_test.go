// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/thomasvincent/mantl/pkg/bootstrap"
)

// The subprocess runs the real CLI entry point with an injected executor. No
// kubectl, container runtime, cluster or cloud credentials are needed.
func TestCLIInterruptFixture(t *testing.T) {
	if os.Getenv("MANTL_TEST_INTERRUPT_PROCESS") != "1" {
		return
	}
	kubeContext = "fixture"
	executor := &recordingApply{}
	rootCmd = newApplyCommand(func(dag *bootstrap.ExecutionDAG) applyExecutor {
		executor.dag = dag
		return executor
	})
	executor.inspect = func(ctx context.Context, stage string) error {
		if stage == "provision" {
			if err := os.WriteFile(os.Getenv("MANTL_TEST_INTERRUPT_MARKER"), []byte(executor.dag.BuildDir), 0600); err != nil {
				return err
			}
			<-ctx.Done()
			return ctx.Err()
		}
		return nil
	}
	rootCmd.SetArgs([]string{os.Getenv("MANTL_TEST_INTERRUPT_SPEC"), "--output-dir", os.Getenv("MANTL_TEST_INTERRUPT_OUTPUT")})
	Execute()
	t.Fatal("interrupted CLI returned successfully")
}

func TestCLIInterruptCleansPrivateSnapshot(t *testing.T) {
	for _, interrupt := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run(interrupt.String(), func(t *testing.T) { verifyCLIInterrupt(t, interrupt) })
	}
}

func verifyCLIInterrupt(t *testing.T, interrupt os.Signal) {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	marker := filepath.Join(t.TempDir(), "started")
	command := exec.CommandContext(ctx, binary, "-test.run=^TestCLIInterruptFixture$")
	command.Env = append(os.Environ(), "MANTL_TEST_INTERRUPT_PROCESS=1", "MANTL_TEST_INTERRUPT_MARKER="+marker, "MANTL_TEST_INTERRUPT_SPEC="+planSpec(t), "MANTL_TEST_INTERRUPT_OUTPUT="+filepath.Join(t.TempDir(), "generated"))
	var stderr bytes.Buffer
	command.Stderr = &stderr
	command.WaitDelay = time.Second
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- command.Wait() }()
	snapshot := waitForInterruptFixture(t, ctx, marker, finished)
	if err := command.Process.Signal(interrupt); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-finished:
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(stderr.String(), "context canceled") {
			t.Fatalf("CLI did not report cancellation: %v, %s", err, stderr.String())
		}
	case <-ctx.Done():
		<-finished
		t.Fatal("interrupted CLI did not exit before the deadline")
	}
	if _, err := os.Stat(snapshot); !os.IsNotExist(err) {
		t.Fatal("interrupted CLI leaked its private snapshot")
	}
}

func waitForInterruptFixture(t *testing.T, ctx context.Context, marker string, finished <-chan error) string {
	t.Helper()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			<-finished
			t.Fatal("CLI fixture did not start before the deadline")
		case err := <-finished:
			t.Fatalf("CLI fixture exited before interruption: %v", err)
		case <-ticker.C:
			if data, err := os.ReadFile(marker); err == nil {
				if _, err := os.Stat(string(data)); err != nil {
					t.Fatalf("fixture has no private snapshot: %v", err)
				}
				return string(data)
			}
		}
	}
}
