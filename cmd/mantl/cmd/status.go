package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/bootstrap"
	"github.com/thomasvincent/mantl/pkg/compiler"
)

var statusCmd = &cobra.Command{
	Use:   "status [spec-file]",
	Short: "Show the live status of the Mantl platform",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		specFile := args[0]
		
		// 1. Parse spec to know what we are looking at
		cluster, err := compiler.ParseSpec(specFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Fetching status for Mantl platform: %s\n", cluster.Name)
		fmt.Println("-------------------------------------------")

		dag := &bootstrap.ExecutionDAG{
			BuildDir: ".mantl/build",
		}

		// 2. Check Platform Convergence
		if err := dag.VerifyConvergence(); err != nil {
			fmt.Printf("Platform Status: DEGRADED (%v)\n", err)
		} else {
			fmt.Println("Platform Status: HEALTHY")
		}

		// 3. Show feature status
		fmt.Println("\nEnabled Features:")
		fmt.Printf("  Observability: %v\n", cluster.Spec.Features.Observability)
		fmt.Printf("  Security:      %v\n", cluster.Spec.Features.Security)
		fmt.Printf("  Compliance:    %v\n", cluster.Spec.Features.Compliance)

		// 4. TODO: Fetch latest compliance audit result from cluster
		fmt.Println("\nLatest Compliance Audit: No audits found.")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
