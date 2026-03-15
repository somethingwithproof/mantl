package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mantl",
	Short: "Mantl is a spec-driven platform compiler and runtime",
	Long: `Mantl takes a declarative platform specification and compiles it into 
infrastructure, GitOps topology, and compliance policies.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
