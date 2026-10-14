// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/compiler"
)

type planVerificationReport struct {
	SchemaVersion string        `json:"schemaVersion"`
	Scope         string        `json:"scope"`
	Verified      bool          `json:"verified"`
	Inventory     compiler.Plan `json:"inventory"`
}

type savedPlanIdentity struct {
	Platform   string `json:"platform"`
	SpecSHA256 string `json:"specSHA256"`
}

type savedPlanDiffReport struct {
	SchemaVersion string                  `json:"schemaVersion"`
	Before        savedPlanIdentity       `json:"before"`
	After         savedPlanIdentity       `json:"after"`
	Comparison    compiler.PlanComparison `json:"comparison"`
}

func newPlanVerifyCommand() *cobra.Command {
	var directory, format string
	command := &cobra.Command{
		Use: "verify PLAN", Short: "Verify saved artifact hashes offline without applying them",
		Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(command *cobra.Command, args []string) error {
			if err := validateReviewFormat(format); err != nil {
				return err
			}
			if directory == "" {
				return errors.New("artifact directory must not be empty")
			}
			if err := command.Context().Err(); err != nil {
				return err
			}
			plan, err := compiler.ReadPlan(args[0])
			if err != nil {
				return fmt.Errorf("read verification inventory: %w", err)
			}
			verified, err := compiler.ReadPlanArtifacts(plan, directory)
			if err != nil {
				return fmt.Errorf("verify saved artifacts: %w", err)
			}
			if err := command.Context().Err(); err != nil {
				return err
			}
			report := planVerificationReport{
				SchemaVersion: "mantl.io/plan-verification/v1alpha1",
				Scope:         "artifact-integrity", Verified: true, Inventory: verified,
			}
			if err := writePlanVerification(command.OutOrStdout(), format, report); err != nil {
				return fmt.Errorf("write verification report: %w", err)
			}
			return nil
		},
	}
	command.Flags().StringVar(&directory, "output-dir", ".mantl/build", "Directory containing the saved plan's artifacts")
	command.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	return command
}

func newPlanDiffCommand() *cobra.Command {
	var format string
	var failOnChange bool
	command := &cobra.Command{
		Use: "diff BEFORE AFTER", Short: "Compare two saved inventories offline without recompiling",
		Args: cobra.ExactArgs(2), SilenceUsage: true, SilenceErrors: true,
		RunE: func(command *cobra.Command, args []string) error {
			if err := validateReviewFormat(format); err != nil {
				return err
			}
			if err := command.Context().Err(); err != nil {
				return err
			}
			before, err := compiler.ReadPlan(args[0])
			if err != nil {
				return fmt.Errorf("read baseline inventory: %w", err)
			}
			after, err := compiler.ReadPlan(args[1])
			if err != nil {
				return fmt.Errorf("read candidate inventory: %w", err)
			}
			comparison, err := compiler.ComparePlans(before, after)
			if err != nil {
				return fmt.Errorf("compare saved inventories: %w", err)
			}
			if err := command.Context().Err(); err != nil {
				return err
			}
			report := savedPlanDiffReport{
				SchemaVersion: "mantl.io/plan-comparison/v1alpha1",
				Before:        savedPlanIdentity{Platform: before.Platform, SpecSHA256: before.SpecSHA256},
				After:         savedPlanIdentity{Platform: after.Platform, SpecSHA256: after.SpecSHA256},
				Comparison:    comparison,
			}
			if err := writeSavedPlanDiff(command.OutOrStdout(), format, report); err != nil {
				return fmt.Errorf("write saved-plan comparison: %w", err)
			}
			if failOnChange && comparison.Changed {
				return errors.New("saved platform plans differ")
			}
			return nil
		},
	}
	command.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	command.Flags().BoolVar(&failOnChange, "fail-on-change", false, "Fail after printing input or artifact identity changes")
	return command
}

func validateReviewFormat(format string) error {
	if format != "text" && format != "json" {
		return errors.New("plan format must be text or json")
	}
	return nil
}

func writePlanVerification(writer io.Writer, format string, report planVerificationReport) error {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	if _, err := fmt.Fprintf(writer, "Verified %d artifacts for %s (spec SHA256 %s)\n", len(report.Inventory.Artifacts), report.Inventory.Platform, report.Inventory.SpecSHA256); err != nil {
		return err
	}
	_, err := fmt.Fprintln(writer, "Artifact integrity only; does not authenticate approval, recompile the spec or verify live infrastructure.")
	return err
}

func writeSavedPlanDiff(writer io.Writer, format string, report savedPlanDiffReport) error {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	if _, err := fmt.Fprintf(writer, "Saved plans: %s (%s) -> %s (%s)\n", report.Before.Platform, report.Before.SpecSHA256, report.After.Platform, report.After.SpecSHA256); err != nil {
		return err
	}
	return writePlanComparison(writer, &report.Comparison)
}
