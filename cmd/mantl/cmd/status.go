package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/compiler"
	"github.com/thomasvincent/mantl/pkg/platformstatus"
	"github.com/thomasvincent/mantl/pkg/toolcommand"
)

const statusCheckTimeout = 5 * time.Second
const maxStatusOutput = 8 << 20

// statusQuery is injected for command tests; production uses read-only kubectl.
type statusQuery func(context.Context, ...string) ([]byte, error)

var statusCmd = newStatusCommand(queryStatus, time.Now)

func newStatusCommand(query statusQuery, now func() time.Time) *cobra.Command {
	var format string
	var requireHealthy bool
	var timeout time.Duration
	command := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		Use:           "status [spec-file]",
		Short:         "Observe generated applications and policy report counts",
		Args:          cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if format != "text" && format != "json" {
				return fmt.Errorf("status format must be text or json")
			}
			if timeout <= 0 || timeout > time.Minute {
				return fmt.Errorf("status timeout must be positive and at most 1m")
			}
			cluster, err := compiler.ParseSpec(args[0])
			if err != nil {
				return fmt.Errorf("read status spec: %w", err)
			}
			if err := command.Context().Err(); err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(command.Context(), timeout)
			defer cancel()
			apps, appErr := query(ctx, kubectlArgs("get", "applications.argoproj.io", "-n", "argocd", "-o", "json")...)
			policies, policyErr := query(ctx, kubectlArgs("get", "policyreports,clusterpolicyreports", "-A", "-o", "json")...)
			report := platformstatus.Report{
				SchemaVersion: "mantl.io/status/v1alpha1", ObservedAt: now().UTC(), Platform: cluster.Name, Context: kubeContext,
				Applications: platformstatus.InterpretApplications(compiler.GitOpsApplications(cluster), apps),
				Policies:     platformstatus.InterpretPolicies(policies),
			}
			// External stderr may contain credential-plugin output. Publish only a stable
			// query failure reason, never raw command errors or application source specs.
			if appErr != nil {
				report.Applications.State = platformstatus.Unknown
				report.Applications.Reason = "application query failed"
			}
			if policyErr != nil {
				report.Policies = platformstatus.Policies{Scope: "all-readable-policy-reports", State: platformstatus.Unknown, Reason: "policy query failed"}
			}
			if err := writeStatus(command.OutOrStdout(), format, report); err != nil {
				return fmt.Errorf("write status: %w", err)
			}
			if err := command.Context().Err(); err != nil {
				return err
			}
			if requireHealthy && report.Applications.State != platformstatus.Healthy {
				return errors.New("generated applications are not healthy")
			}
			return nil
		},
	}
	command.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	command.Flags().BoolVar(&requireHealthy, "require-healthy", false, "Exit unsuccessfully unless all generated applications are synced and healthy")
	command.Flags().DurationVar(&timeout, "timeout", statusCheckTimeout, "Total cluster query deadline (maximum 1m)")
	return command
}

func writeStatus(writer io.Writer, format string, report platformstatus.Report) error {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	_, err := io.WriteString(writer, report.Text())
	return err
}

type boundedStatusOutput struct{ buffer bytes.Buffer }

func (output *boundedStatusOutput) Write(data []byte) (int, error) {
	if len(data) > maxStatusOutput-output.buffer.Len() {
		return 0, errors.New("status response exceeds size limit")
	}
	return output.buffer.Write(data)
}

func queryStatus(ctx context.Context, args ...string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	command := toolcommand.Kubectl(ctx, args...)
	// Bound inherited pipes from authentication plugins as well as the direct
	// process. Stderr is discarded so plugin credentials cannot enter reports.
	command.WaitDelay = time.Second
	var output boundedStatusOutput
	command.Stdout = &output
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("query platform status: %w", err)
	}
	return output.buffer.Bytes(), nil
}

func init() { rootCmd.AddCommand(statusCmd) }
