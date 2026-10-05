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

var kubeContext string
var sourceDir string

func init() {
	rootCmd.PersistentFlags().StringVar(&kubeContext, "context", "", "Explicit Kubernetes context")
	rootCmd.PersistentFlags().StringVar(&sourceDir, "source-dir", ".", "Mantl source directory for infrastructure blueprints")
}
func kubectlArgs(args ...string) []string {
	if kubeContext != "" {
		return append([]string{"--context", kubeContext}, args...)
	}
	return args
}
