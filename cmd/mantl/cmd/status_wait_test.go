// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/thomasvincent/mantl/pkg/platformstatus"
)

func TestStatusWaitConvergesAndPrintsOneReport(t *testing.T) {
	file := statusSpec(t)
	healthy := statusFixture(t, file)
	observations := [][]byte{
		[]byte(`{"items":[]}`),
		bytes.ReplaceAll(healthy, []byte(`"Synced"`), []byte(`"OutOfSync"`)),
		healthy,
	}
	attempts, policyQueries, pauses := 0, 0, 0
	query := func(ctx context.Context, args ...string) ([]byte, error) {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 2*time.Second {
			t.Fatal("query did not inherit its bounded deadline")
		}
		if strings.Contains(strings.Join(args, " "), "applications.argoproj.io") {
			observation := observations[attempts]
			attempts++
			return observation, nil
		}
		policyQueries++
		return []byte(`{"items":[{"summary":{"pass":1,"fail":2,"error":0,"warn":0,"skip":0}}]}`), nil
	}
	pause := func(ctx context.Context, interval time.Duration) error {
		if interval != 7*time.Second || ctx.Err() != nil {
			t.Fatal("configured polling interval or context lost")
		}
		pauses++
		return nil
	}
	stamp := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	command := newStatusWaitCommand(query, func() time.Time { return stamp.Add(time.Duration(attempts) * time.Second) }, pause)
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{file, "--format=json", "--interval=7s", "--query-timeout=2s"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	var report platformstatus.Report
	decoder := json.NewDecoder(&output)
	if err := decoder.Decode(&report); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatal("polling wrote more than one final JSON report")
	}
	if attempts != 3 || pauses != 2 || policyQueries != 3 || report.Applications.State != platformstatus.Healthy || report.Policies.Counts.Fail != 2 || report.ObservedAt != stamp.Add(3*time.Second) {
		t.Fatalf("unexpected convergence: %+v, attempts=%d pauses=%d", report, attempts, pauses)
	}
}

func TestStatusWaitRetriesQueryFailureWithoutLeakingSecrets(t *testing.T) {
	file := statusSpec(t)
	healthy := statusFixture(t, file)
	attempts := 0
	query := func(_ context.Context, args ...string) ([]byte, error) {
		if strings.Contains(strings.Join(args, " "), "applications.argoproj.io") {
			attempts++
			if attempts > 1 {
				return healthy, nil
			}
			// Even valid-looking partial stdout cannot override a failed query.
			return healthy, errors.New("secret-auth-error")
		}
		return []byte("secret-plugin-output"), errors.New("secret-auth-error")
	}
	command := newStatusWaitCommand(query, time.Now, func(context.Context, time.Duration) error { return nil })
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{file, "--format=json"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	var report platformstatus.Report
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 || report.Applications.State != platformstatus.Healthy || report.Policies.State != platformstatus.Unknown || strings.Contains(output.String(), "secret") {
		t.Fatalf("query recovery or sanitization failed: %s", output.String())
	}
}

func TestStatusWaitCannotAcceptIncompleteOrMismatchedApplications(t *testing.T) {
	file := statusSpec(t)
	healthy := statusFixture(t, file)
	for _, apps := range [][]byte{
		[]byte(`{"items":[]}`), []byte(`{"unexpected":"secret"}`),
		bytes.ReplaceAll(healthy, []byte(`"Healthy"`), []byte(`"Degraded"`)),
		bytes.ReplaceAll(healthy, []byte(`"main"`), []byte(`"other-revision"`)),
	} {
		for _, format := range []string{"json", "text"} {
			pauses := 0
			command := newStatusWaitCommand(func(context.Context, ...string) ([]byte, error) { return apps, nil }, time.Now,
				func(context.Context, time.Duration) error { pauses++; return context.DeadlineExceeded })
			var output bytes.Buffer
			command.SetOut(&output)
			command.SetArgs([]string{file, "--format", format})
			if err := command.Execute(); !errors.Is(err, context.DeadlineExceeded) || pauses != 1 || output.Len() == 0 || strings.Contains(output.String(), "secret") {
				t.Fatalf("unhealthy wait was accepted or lost its final report: %v, %s", err, output.String())
			}
			if format == "json" {
				var report platformstatus.Report
				if err := json.Unmarshal(output.Bytes(), &report); err != nil || report.Applications.State == platformstatus.Healthy {
					t.Fatalf("invalid last observation: %s, %v", output.String(), err)
				}
			}
		}
	}
}

func TestStatusWaitDeadlineAndCancellation(t *testing.T) {
	file := statusSpec(t)
	// A query deadline is retryable, but the parent deadline bounds the entire wait.
	command := newStatusWaitCommand(func(ctx context.Context, _ ...string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}, time.Now, waitStatusInterval)
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{file, "--timeout=100ms", "--query-timeout=5ms", "--format=json"})
	if err := command.Execute(); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("overall deadline was lost: %v", err)
	}
	var report platformstatus.Report
	if err := json.Unmarshal(output.Bytes(), &report); err != nil || report.Applications.State != platformstatus.Unknown {
		t.Fatalf("timed-out wait lost the final unknown report: %s, %v", output.String(), err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	queries := 0
	command = newStatusWaitCommand(func(context.Context, ...string) ([]byte, error) {
		queries++
		return []byte(`{"items":[]}`), nil
	}, time.Now, func(ctx context.Context, _ time.Duration) error {
		cancel()
		return waitStatusInterval(ctx, time.Minute)
	})
	output.Reset()
	command.SetOut(&output)
	command.SetArgs([]string{file, "--format=json"})
	if err := command.ExecuteContext(ctx); !errors.Is(err, context.Canceled) || queries != 2 || output.Len() == 0 {
		t.Fatalf("cancelled wait kept polling or lost its report: queries=%d err=%v", queries, err)
	}
}

func TestStatusWaitValidationAndAlreadyCancelledContext(t *testing.T) {
	file := statusSpec(t)
	for _, args := range [][]string{
		{file, "--format=yaml"}, {file, "--timeout=0s"}, {file, "--timeout=3h"},
		{file, "--interval=0s"}, {file, "--interval=1ms"}, {file, "--interval=2m"},
		{file, "--query-timeout=0s"}, {file, "--query-timeout=2m"}, {"missing-spec"}, {file, "extra"},
	} {
		queries := 0
		command := newStatusWaitCommand(func(context.Context, ...string) ([]byte, error) { queries++; return nil, nil }, time.Now, waitStatusInterval)
		command.SetOut(io.Discard)
		command.SetArgs(args)
		if err := command.Execute(); err == nil || queries != 0 {
			t.Fatalf("invalid options reached cluster queries: %v queries=%d err=%v", args, queries, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	command := newStatusCommand(func(context.Context, ...string) ([]byte, error) {
		t.Fatal("cancelled command queried a cluster")
		return nil, nil
	}, time.Now)
	command.SetArgs([]string{"wait", file})
	command.SetOut(io.Discard)
	if err := command.ExecuteContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("subcommand registration or cancellation failed: %v", err)
	}
}

type cancellingStatusWriter struct{ cancel context.CancelFunc }

func (writer cancellingStatusWriter) Write(data []byte) (int, error) {
	writer.cancel()
	return len(data), nil
}

func TestStatusWaitFinalReportErrors(t *testing.T) {
	file := statusSpec(t)
	healthy := statusFixture(t, file)
	for _, format := range []string{"text", "json"} {
		command := newStatusWaitCommand(func(context.Context, ...string) ([]byte, error) { return healthy, nil }, time.Now, waitStatusInterval)
		command.SetOut(failingStatusWriter{})
		command.SetArgs([]string{file, "--format", format})
		if err := command.Execute(); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatalf("final report error lost: %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		command = newStatusWaitCommand(func(context.Context, ...string) ([]byte, error) { return healthy, nil }, time.Now, waitStatusInterval)
		command.SetOut(cancellingStatusWriter{cancel: cancel})
		command.SetArgs([]string{file, "--format", format})
		if err := command.ExecuteContext(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation during final report became success: %v", err)
		}
	}
}
