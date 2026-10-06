package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/thomasvincent/mantl/pkg/compiler"
	"github.com/thomasvincent/mantl/pkg/platformstatus"
)

type applicationQuery func(context.Context) ([]byte, error)

// waitForApplications uses the status command's desired-identity contract.
func waitForApplications(ctx context.Context, expected []compiler.GitOpsApplication, query applicationQuery, interval time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(expected) == 0 {
		return fmt.Errorf("desired application topology is required for convergence")
	}
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for desired application convergence: %w", ctx.Err())
		case <-timer.C:
			if err := ctx.Err(); err != nil {
				return err
			}
			data, err := query(ctx)
			if err == nil && platformstatus.InterpretApplications(expected, data).State == platformstatus.Healthy {
				return ctx.Err()
			}
			timer.Reset(interval)
		}
	}
}

// VerifyConvergence requires every expected Application, its source identity and
// explicit Synced/Healthy observations. Unrelated Applications are ignored.
func (d *ExecutionDAG) VerifyConvergence(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	return waitForApplications(ctx, d.Applications, d.queryApplications, 15*time.Second)
}

type applicationOutput struct{ buffer bytes.Buffer }

func (output *applicationOutput) Write(data []byte) (int, error) {
	if len(data) > (8<<20)-output.buffer.Len() {
		return 0, fmt.Errorf("application response exceeds size limit")
	}
	return output.buffer.Write(data)
}

func (d *ExecutionDAG) queryApplications(ctx context.Context) ([]byte, error) {
	command := d.command(ctx, "kubectl", "get", "applications.argoproj.io", "-n", "argocd", "-o", "json")
	command.WaitDelay = time.Second
	var output applicationOutput
	command.Stdout = &output
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("query desired applications: %w", err)
	}
	return output.buffer.Bytes(), nil
}
