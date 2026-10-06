// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/bootstrap"
	"github.com/thomasvincent/mantl/pkg/compiler"
)

type applyExecutor interface {
	Preflight(context.Context, string, string) error
	ProvisionInfra(context.Context, string, string) error
	InstallArgoCD(context.Context) error
	BootstrapGitOps(context.Context) error
	VerifyConvergence(context.Context) error
}

type applyFactory func(*bootstrap.ExecutionDAG) applyExecutor

type applyOptions struct {
	outputDir string
	planFile  string
	timeout   time.Duration
}

var applyCmd = newApplyCommand(func(dag *bootstrap.ExecutionDAG) applyExecutor { return dag })

func newApplyCommand(factory applyFactory) *cobra.Command {
	options := applyOptions{}
	command := &cobra.Command{
		Use: "apply [spec-file]", Short: "Preflight and bootstrap generated or reviewed platform artifacts",
		Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(command *cobra.Command, args []string) error {
			return executeApply(command, args[0], options, factory)
		},
	}
	command.Flags().StringVar(&options.outputDir, "output-dir", ".mantl/build", "Generated or reviewed artifact directory")
	command.Flags().StringVar(&options.planFile, "plan", "", "Saved inventory to verify before any bootstrap changes")
	command.Flags().DurationVar(&options.timeout, "timeout", 30*time.Minute, "Total apply deadline (maximum 2h)")
	return command
}

func executeApply(command *cobra.Command, specFile string, options applyOptions, factory applyFactory) error {
	if options.outputDir == "" || options.timeout <= 0 || options.timeout > 2*time.Hour {
		return fmt.Errorf("apply requires an output directory and a positive timeout of at most 2h")
	}
	ctx, cancel := context.WithTimeout(command.Context(), options.timeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	cluster, err := compiler.ParseSpec(specFile)
	if err != nil {
		return fmt.Errorf("read apply spec: %w", err)
	}
	plan, err := compiler.Compile(cluster)
	if err != nil {
		return fmt.Errorf("compile apply plan: %w", err)
	}
	plan, err = reviewedApplyPlan(plan, options)
	if err != nil {
		return err
	}
	target, err := applyTarget(plan, kubeContext)
	if err != nil {
		return err
	}

	dag := &bootstrap.ExecutionDAG{SpecFile: specFile, KubeContext: target, SourceDir: sourceDir, Distribution: plan.Distribution, Applications: compiler.GitOpsApplications(cluster)}
	executor := factory(dag)
	if err := executor.Preflight(ctx, plan.Provider, plan.Platform); err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return fmt.Errorf("apply preflight: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Keep the familiar generated output, but execute only private snapshot bytes.
	if err := generatePlanArtifacts(cluster, plan, planOptions{outputDir: options.outputDir, dryRun: options.planFile != ""}); err != nil {
		return fmt.Errorf("generate apply artifacts: %w", err)
	}

	directory, err := stageApplyPlan(plan)
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(directory) }()
	dag.BuildDir = directory
	return runApplySteps(ctx, executor, plan, command.OutOrStdout())
}

func reviewedApplyPlan(current compiler.Plan, options applyOptions) (compiler.Plan, error) {
	if options.planFile == "" {
		return current, nil
	}
	reviewed, err := compiler.ReadPlan(options.planFile)
	if err != nil {
		return compiler.Plan{}, err
	}
	if err := compiler.MatchPlan(reviewed, current); err != nil {
		return compiler.Plan{}, err
	}
	return compiler.ReadPlanArtifacts(reviewed, options.outputDir)
}

func stageApplyPlan(plan compiler.Plan) (string, error) {
	directory, err := os.MkdirTemp("", "mantl-apply-*")
	if err != nil {
		return "", fmt.Errorf("create private apply directory: %w", err)
	}
	for _, artifact := range plan.Artifacts {
		filename := filepath.Join(directory, filepath.FromSlash(artifact.Path))
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			_ = os.RemoveAll(directory)
			return "", fmt.Errorf("create private artifact directory: %w", err)
		}
		if err := os.WriteFile(filename, artifact.Content, 0600); err != nil {
			_ = os.RemoveAll(directory)
			return "", fmt.Errorf("stage reviewed artifact: %w", err)
		}
	}
	return directory, nil
}

func runApplySteps(ctx context.Context, executor applyExecutor, plan compiler.Plan, writer io.Writer) error {
	type applyStep struct {
		name string
		run  func(context.Context) error
	}
	steps := []applyStep{
		{"Provision infrastructure", func(ctx context.Context) error { return executor.ProvisionInfra(ctx, plan.Provider, plan.Platform) }},
		{"Install ArgoCD", executor.InstallArgoCD},
		{"Bootstrap desired Applications", executor.BootstrapGitOps},
		{"Verify desired Application convergence", executor.VerifyConvergence},
	}
	for _, step := range steps {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer, step.name); err != nil {
			return fmt.Errorf("write apply progress: %w", err)
		}
		if err := step.run(ctx); err != nil {
			if ctx.Err() != nil {
				err = ctx.Err()
			}
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(writer, "Platform bootstrap completed; desired Applications are synced and healthy.")
	return err
}

func init() { rootCmd.AddCommand(applyCmd) }

func applyTarget(plan compiler.Plan, requested string) (string, error) {
	if plan.Provider == "local" {
		target := "kind-" + plan.Platform
		if requested != "" && requested != target {
			return "", fmt.Errorf("local apply context must be %s", target)
		}
		return target, nil
	}
	if requested == "" {
		return "", fmt.Errorf("--context is required for cloud bootstrap")
	}
	return requested, nil
}
