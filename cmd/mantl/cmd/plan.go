// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/compiler"
)

type planOptions struct {
	outputDir, format, saveFile, baseline string
	dryRun, failOnChange                  bool
}

type planReport struct {
	compiler.Plan
	Comparison *compiler.PlanComparison `json:"comparison,omitempty"`
}

var planCmd = newPlanCommand()

func newPlanCommand() *cobra.Command {
	options := planOptions{}
	command := &cobra.Command{
		Use: "plan [spec-file]", Short: "Compile, save or compare platform artifacts offline",
		Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(command *cobra.Command, args []string) error { return executePlan(command, args[0], options) },
	}
	command.Flags().StringVar(&options.outputDir, "output-dir", ".mantl/build", "Directory for generated files")
	command.Flags().StringVar(&options.format, "format", "text", "Output format: text or json")
	command.Flags().BoolVar(&options.dryRun, "dry-run", false, "Preview without writing files")
	command.Flags().StringVar(&options.saveFile, "save-plan", "", "Save an inventory outside the generated artifact directory")
	command.Flags().StringVar(&options.baseline, "compare", "", "Compare generated artifact identities with a saved inventory")
	command.Flags().BoolVar(&options.failOnChange, "fail-on-change", false, "Fail after printing changes (requires --compare and --dry-run)")
	return command
}

func executePlan(command *cobra.Command, specFile string, options planOptions) error {
	if err := validatePlanOptions(options); err != nil {
		return err
	}
	if err := command.Context().Err(); err != nil {
		return err
	}
	cluster, err := compiler.ParseSpec(specFile)
	if err != nil {
		return fmt.Errorf("read plan spec: %w", err)
	}
	plan, err := compiler.Compile(cluster)
	if err != nil {
		return fmt.Errorf("compile plan: %w", err)
	}
	comparison, err := compareSavedPlan(plan, options.baseline)
	if err != nil {
		return err
	}
	if err := validateSavedPlanPath(specFile, options); err != nil {
		return err
	}
	if err := command.Context().Err(); err != nil {
		return err
	}
	if err := generatePlanArtifacts(cluster, plan, options); err != nil {
		return err
	}
	if err := command.Context().Err(); err != nil {
		return err
	}

	if err := writePlanReport(command.OutOrStdout(), options.format, planReport{Plan: plan, Comparison: comparison}, options.outputDir, options.dryRun); err != nil {
		return fmt.Errorf("write plan report: %w", err)
	}
	if options.failOnChange && comparison.Changed {
		return errors.New("platform plan differs from baseline")
	}
	return nil
}

func validatePlanOptions(options planOptions) error {
	if options.format != "text" && options.format != "json" {
		return fmt.Errorf("plan format must be text or json")
	}
	if options.outputDir == "" {
		return fmt.Errorf("plan output directory must not be empty")
	}
	if options.dryRun && options.saveFile != "" {
		return fmt.Errorf("--save-plan cannot be combined with --dry-run")
	}
	if options.failOnChange && (options.baseline == "" || !options.dryRun) {
		return fmt.Errorf("--fail-on-change requires --compare and --dry-run")
	}
	return nil
}

func compareSavedPlan(current compiler.Plan, filename string) (*compiler.PlanComparison, error) {
	if filename == "" {
		return nil, nil
	}
	baseline, err := compiler.ReadPlan(filename)
	if err != nil {
		return nil, fmt.Errorf("read comparison baseline: %w", err)
	}
	comparison, err := compiler.ComparePlans(baseline, current)
	if err != nil {
		return nil, err
	}
	return &comparison, nil
}

func validateSavedPlanPath(specFile string, options planOptions) error {
	if options.saveFile == "" {
		return nil
	}
	saved, err := canonicalPlanPath(options.saveFile)
	if err != nil {
		return err
	}
	spec, err := canonicalPlanPath(specFile)
	if err != nil {
		return err
	}
	output, err := canonicalPlanPath(options.outputDir)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(output, saved)
	if err != nil {
		return err
	}
	if savedInfo, err := os.Stat(options.saveFile); err == nil {
		if specInfo, err := os.Stat(specFile); err == nil && os.SameFile(savedInfo, specInfo) {
			return fmt.Errorf("saved plan must not replace the input spec")
		}
	}

	if saved == spec || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return fmt.Errorf("saved plan must be separate from the spec and outside the generated artifact directory")
	}
	return nil
}

func canonicalPlanPath(filename string) (string, error) {
	absolute, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("resolve plan path: %w", err)
	}
	// Resolve existing parent symlinks without following a destination file symlink;
	// saving replaces that directory entry atomically rather than its target.
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err == nil {
		return filepath.Join(parent, filepath.Base(absolute)), nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("resolve plan parent: %w", err)
	}
	parent, err = canonicalPlanPath(filepath.Dir(absolute))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(absolute)), nil
}

func savePlan(plan compiler.Plan, filename string) error {
	file, err := os.CreateTemp(filepath.Dir(filename), ".mantl-plan-*")
	if err != nil {
		return fmt.Errorf("create saved plan: %w", err)
	}
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(plan); err != nil {
		return fmt.Errorf("encode saved plan: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close saved plan: %w", err)
	}
	if err := os.Rename(file.Name(), filename); err != nil {
		return fmt.Errorf("publish saved plan: %w", err)
	}
	return nil
}

func writePlan(writer io.Writer, format string, plan compiler.Plan, outputDir string, dryRun bool) error {
	return writePlanReport(writer, format, planReport{Plan: plan}, outputDir, dryRun)
}

func writePlanReport(writer io.Writer, format string, report planReport, outputDir string, dryRun bool) error {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	action := "Generated"
	if dryRun {
		action = "Would generate"
	}
	if _, err := fmt.Fprintf(writer, "%s %d artifacts for %s in %s\n", action, len(report.Artifacts), report.Platform, outputDir); err != nil {
		return err
	}
	for _, artifact := range report.Artifacts {
		if _, err := fmt.Fprintf(writer, "  %s  %s (%d bytes)\n", artifact.SHA256, artifact.Path, artifact.Size); err != nil {
			return err
		}
	}
	for _, notice := range report.Notices {
		if _, err := fmt.Fprintln(writer, notice); err != nil {
			return err
		}
	}
	return writePlanComparison(writer, report.Comparison)
}

func init() { rootCmd.AddCommand(planCmd) }

func generatePlanArtifacts(cluster *v1alpha1.MantlCluster, plan compiler.Plan, options planOptions) error {
	if options.dryRun {
		return nil
	}
	if err := compiler.RenderTerraform(cluster, options.outputDir); err != nil {
		return fmt.Errorf("write terraform artifacts to %s: %w", options.outputDir, err)
	}
	if err := compiler.RenderGitOps(cluster, options.outputDir); err != nil {
		return fmt.Errorf("write gitops artifacts to %s: %w", options.outputDir, err)
	}
	if options.saveFile != "" {
		return savePlan(plan, options.saveFile)
	}
	return nil
}

func writePlanComparison(writer io.Writer, comparison *compiler.PlanComparison) error {
	if comparison != nil {
		if _, err := fmt.Fprintf(writer, "Baseline comparison: changed=%t inputChanged=%t\n", comparison.Changed, comparison.InputChanged); err != nil {
			return err
		}
		for _, change := range comparison.Artifacts {
			if _, err := fmt.Fprintf(writer, "  %s %s\n", change.Action, change.Path); err != nil {
				return err
			}
		}
	}
	return nil
}
