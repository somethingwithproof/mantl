package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/bootstrap"
	"github.com/thomasvincent/mantl/pkg/compiler"
)

var applyCmd = &cobra.Command{
	Use:   "apply [spec-file]",
	Short: "Apply the compiled Mantl spec to bootstrap the cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		specFile := args[0]
		fmt.Printf("Applying Mantl platform from spec: %s\n", specFile)

		// 1. Parse and Plan
		cluster, err := compiler.ParseSpec(specFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		buildDir := ".mantl/build"
		fmt.Println("Step 1/5: Compiling spec...")
		if err := compiler.RenderTerraform(cluster, buildDir); err != nil {
			fmt.Printf("Error rendering terraform: %v\n", err)
			os.Exit(1)
		}
		if err := compiler.RenderGitOps(cluster, buildDir); err != nil {
			fmt.Printf("Error rendering gitops: %v\n", err)
			os.Exit(1)
		}

		// 2. Execution DAG
		dag := &bootstrap.ExecutionDAG{
			SpecFile: specFile,
			BuildDir: buildDir,
		}

		fmt.Println("Step 2/5: Provisioning infrastructure...")
		if err := dag.ProvisionInfra(cluster.Spec.Provider.Kind, cluster.Name); err != nil {
			fmt.Printf("Infrastructure error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Step 3/5: Installing GitOps Controller (ArgoCD)...")
		if err := dag.InstallArgoCD(); err != nil {
			fmt.Printf("ArgoCD install error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Step 4/5: Bootstrapping GitOps Applications...")
		if err := dag.BootstrapGitOps(); err != nil {
			fmt.Printf("GitOps bootstrap error: %v\n", err)
			os.Exit(1)
		}

		// 5. Verification
		fmt.Println("Step 5/5: Verifying platform convergence...")
		if err := dag.VerifyConvergence(); err != nil {
			fmt.Printf("Verification failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully applied platform spec!")
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}
