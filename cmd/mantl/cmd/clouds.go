package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/cloudcontract"
)

var cloudsCmd = &cobra.Command{Use: "validate-clouds", Short: "Validate static AWS/GCP/Azure contracts without provisioning", RunE: func(cmd *cobra.Command, args []string) error {
	results, err := cloudcontract.Validate(sourceDir)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(cmd.OutOrStdout()).Encode(results); err != nil {
		return err
	}
	if failures := cloudcontract.Failures(results); len(failures) > 0 {
		return fmt.Errorf("cloud contract checks failed: %v", failures)
	}
	return nil
}}

func init() { rootCmd.AddCommand(cloudsCmd) }
