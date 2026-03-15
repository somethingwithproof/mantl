package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/compiler"
)

var planCmd = &cobra.Command{
	Use:   "plan [spec-file]",
	Short: "Compile the Mantl spec into Terraform and GitOps manifests",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		specFile := args[0]
		fmt.Printf("Planning Mantl platform from spec: %s\n", specFile)

		// 1. Parse spec
		cluster, err := compiler.ParseSpec(specFile)
		if err != nil {
			fmt.Printf("Error parsing spec: %v\n", err)
			os.Exit(1)
		}

		// 2. Render Terraform variables
		buildDir := ".mantl/build"
		err = compiler.RenderTerraform(cluster, buildDir)
		if err != nil {
			fmt.Printf("Error rendering terraform: %v\n", err)
			os.Exit(1)
		}

		// 3. Render GitOps manifests
		err = compiler.RenderGitOps(cluster, buildDir)
		if err != nil {
			fmt.Printf("Error rendering gitops: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully planned platform into: %s\n", buildDir)
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
