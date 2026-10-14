// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/compiler"
	"github.com/thomasvincent/mantl/pkg/platformstatus"
)

type statusWaitOptions struct {
	format                          string
	timeout, interval, queryTimeout time.Duration
}

type statusIntervalWait func(context.Context, time.Duration) error

func newStatusWaitCommand(query statusQuery, now func() time.Time, pause statusIntervalWait) *cobra.Command {
	options := statusWaitOptions{}
	command := &cobra.Command{
		Use: "wait SPEC", Short: "Wait for generated applications to become synced and healthy",
		Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(command *cobra.Command, args []string) error {
			if err := validateStatusWaitOptions(options); err != nil {
				return err
			}
			if err := command.Context().Err(); err != nil {
				return err
			}
			cluster, err := compiler.ParseSpec(args[0])
			if err != nil {
				return fmt.Errorf("read status wait spec: %w", err)
			}
			ctx, cancel := context.WithTimeout(command.Context(), options.timeout)
			defer cancel()
			expected := compiler.GitOpsApplications(cluster)
			var report platformstatus.Report
			for {
				if err := ctx.Err(); err != nil {
					if report.SchemaVersion == "" {
						return fmt.Errorf("wait for application health: %w", err)
					}
					return finishStatusWait(ctx, command, options.format, report, err)
				}
				queryContext, stopQuery := context.WithTimeout(ctx, options.queryTimeout)
				report = observePlatformStatus(queryContext, query, now, cluster.Name, expected)
				stopQuery()
				if err := ctx.Err(); err != nil {
					return finishStatusWait(ctx, command, options.format, report, err)
				}
				if report.Applications.State == platformstatus.Healthy {
					return finishStatusWait(ctx, command, options.format, report, nil)
				}
				if err := pause(ctx, options.interval); err != nil {
					return finishStatusWait(ctx, command, options.format, report, err)
				}
			}
		},
	}
	command.Flags().StringVar(&options.format, "format", "text", "Final report format: text or json")
	command.Flags().DurationVar(&options.timeout, "timeout", 10*time.Minute, "Total wait deadline (maximum 2h)")
	command.Flags().DurationVar(&options.interval, "interval", 5*time.Second, "Pause between observations (1s to 1m)")
	command.Flags().DurationVar(&options.queryTimeout, "query-timeout", statusCheckTimeout, "Deadline for each pair of queries (maximum 1m)")
	return command
}

func validateStatusWaitOptions(options statusWaitOptions) error {
	if options.format != "text" && options.format != "json" {
		return fmt.Errorf("status format must be text or json")
	}
	if options.timeout <= 0 || options.timeout > 2*time.Hour {
		return fmt.Errorf("status wait timeout must be positive and at most 2h")
	}
	if options.interval < time.Second || options.interval > time.Minute {
		return fmt.Errorf("status wait interval must be between 1s and 1m")
	}
	if options.queryTimeout <= 0 || options.queryTimeout > time.Minute {
		return fmt.Errorf("status query timeout must be positive and at most 1m")
	}
	return nil
}

func finishStatusWait(ctx context.Context, command *cobra.Command, format string, report platformstatus.Report, waitError error) error {
	if err := writeStatus(command.OutOrStdout(), format, report); err != nil {
		return fmt.Errorf("write final status: %w", err)
	}
	if waitError == nil {
		waitError = ctx.Err()
	}
	if waitError != nil {
		return fmt.Errorf("wait for application health: %w", waitError)
	}
	return nil
}

func waitStatusInterval(ctx context.Context, interval time.Duration) error {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
