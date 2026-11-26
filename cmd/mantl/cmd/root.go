// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/i18n"
	"github.com/thomasvincent/mantl/pkg/version"
)

var rootCmd = &cobra.Command{
	Version: version.Version + " (" + version.Commit + ")",
	Use:     "mantl",
	Short:   "Mantl is a spec-driven platform compiler and runtime",
	Long: `Mantl takes a declarative platform specification and compiles it into
infrastructure, GitOps topology, and compliance policies.`,
}

func Execute() {
	applyLocalization(extractLangArg(os.Args[1:]))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := rootCmd.ExecuteContext(ctx)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var kubeContext string
var sourceDir string
var langFlag string

func init() {
	rootCmd.PersistentFlags().StringVar(&kubeContext, "context", "", "Explicit Kubernetes context")
	rootCmd.PersistentFlags().StringVar(&sourceDir, "source-dir", ".", "Mantl source directory for infrastructure blueprints")
	rootCmd.PersistentFlags().StringVar(&langFlag, "lang", "", "Display language (BCP-47 or POSIX locale, e.g. de, ja, pt-BR)")
}

func extractLangArg(args []string) string {
	var lang string
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			break
		}
		if args[i] == "--lang" && i+1 < len(args) {
			lang = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--lang=") {
			lang = strings.TrimPrefix(args[i], "--lang=")
		}
	}
	return lang
}

func applyLocalization(override string) *i18n.Catalog {
	cat := i18n.Init(override)
	rootCmd.Short = cat.T("root.short")
	rootCmd.Long = cat.T("root.long")

	if f := rootCmd.PersistentFlags().Lookup("context"); f != nil {
		f.Usage = cat.T("flag.context")
	}
	if f := rootCmd.PersistentFlags().Lookup("source-dir"); f != nil {
		f.Usage = cat.T("flag.source_dir")
	}
	if f := rootCmd.PersistentFlags().Lookup("lang"); f != nil {
		f.Usage = cat.T("flag.lang")
	}

	shortKeys := map[string]string{
		"plan":                    "cmd.plan.short",
		"apply":                   "cmd.apply.short",
		"status":                  "cmd.status.short",
		"status wait":             "cmd.status_wait.short",
		"verify":                  "cmd.verify.short",
		"diff":                    "cmd.diff.short",
		"compliance":              "cmd.compliance.short",
		"compliance status":       "cmd.compliance.status.short",
		"compliance export":       "cmd.compliance.export.short",
		"compliance ingest-falco": "cmd.compliance.ingest_falco.short",
		"compliance remediation":  "cmd.compliance.remediation.short",
		"compliance history":      "cmd.compliance.history.short",
		"fleet":                   "cmd.fleet.short",
		"fleet status":            "cmd.fleet.status.short",
		"fleet replay":            "cmd.fleet.replay.short",
		"exception":               "cmd.exception.short",
		"bundle":                  "cmd.bundle.short",
		"collector-rbac":          "cmd.collector_rbac.short",
		"acceptance":              "cmd.acceptance.short",
		"oscal":                   "cmd.oscal.short",
		"oscal import":            "cmd.oscal.import.short",
		"validate-clouds":         "cmd.validate_clouds.short",
	}
	localizeSubcommands(rootCmd, "", shortKeys, cat)
	return cat
}

func localizeSubcommands(parent *cobra.Command, prefix string, keys map[string]string, cat *i18n.Catalog) {
	for _, sub := range parent.Commands() {
		path := sub.Name()
		if prefix != "" {
			path = prefix + " " + path
		}
		if key, ok := keys[path]; ok {
			sub.Short = cat.T(key)
		}
		localizeSubcommands(sub, path, keys, cat)
	}
}

func kubectlArgs(args ...string) []string {
	if kubeContext != "" {
		return append([]string{"--context", kubeContext}, args...)
	}
	return args
}
