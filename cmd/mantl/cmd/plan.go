package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/compiler"
)

var planCmd = newPlanCommand()

func newPlanCommand() *cobra.Command {
	var outputDir, format string
	var dryRun bool
	command := &cobra.Command{
		Use: "plan [spec-file]", Short: "Compile or preview Terraform and GitOps artifacts offline",
		Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(command *cobra.Command, args []string) error {
			if format != "text" && format != "json" {
				return fmt.Errorf("plan format must be text or json")
			}
			if outputDir == "" {
				return fmt.Errorf("plan output directory must not be empty")
			}
			if err := command.Context().Err(); err != nil {
				return err
			}
			cluster, err := compiler.ParseSpec(args[0])
			if err != nil {
				return fmt.Errorf("read plan spec: %w", err)
			}
			plan, err := compiler.Compile(cluster)
			if err != nil {
				return fmt.Errorf("compile plan: %w", err)
			}
			if err := command.Context().Err(); err != nil {
				return err
			}
			if !dryRun {
				if err := compiler.RenderTerraform(cluster, outputDir); err != nil {
					return err
				}
				if err := compiler.RenderGitOps(cluster, outputDir); err != nil {
					return err
				}
			}
			return writePlan(command.OutOrStdout(), format, plan, outputDir, dryRun)
		},
	}
	command.Flags().StringVar(&outputDir, "output-dir", ".mantl/build", "Directory for generated files")
	command.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Preview exact artifact hashes without writing files")
	return command
}

func writePlan(writer io.Writer, format string, plan compiler.Plan, outputDir string, dryRun bool) error {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(plan)
	}
	action := "Generated"
	if dryRun {
		action = "Would generate"
	}
	if _, err := fmt.Fprintf(writer, "%s %d artifacts for %s in %s\n", action, len(plan.Artifacts), plan.Platform, outputDir); err != nil {
		return err
	}
	for _, artifact := range plan.Artifacts {
		if _, err := fmt.Fprintf(writer, "  %s  %s (%d bytes)\n", artifact.SHA256, artifact.Path, artifact.Size); err != nil {
			return err
		}
	}
	for _, notice := range plan.Notices {
		if _, err := fmt.Fprintln(writer, notice); err != nil {
			return err
		}
	}
	return nil
}

func init() { rootCmd.AddCommand(planCmd) }
