// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/controlbundle"
	"os"
)

func init() {
	root := &cobra.Command{Use: "bundle", Short: "Build and verify independently versioned control content"}
	var version, output, expected string
	build := &cobra.Command{Use: "build", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		m, buildErr := controlbundle.Build(sourceDir, version, f)
		closeErr := f.Close()
		if buildErr != nil {
			_ = os.Remove(output)
			return buildErr
		}
		if closeErr != nil {
			_ = os.Remove(output)
			return closeErr
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{"version": m.Version, "digest": m.Digest()})
	}}
	build.Flags().StringVar(&version, "version", "", "Independent content SemVer")
	build.Flags().StringVar(&output, "output", "controls.tar.gz", "New archive path")
	verify := &cobra.Command{Use: "verify DIRECTORY", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		m, err := controlbundle.Verify(args[0], expected)
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{"version": m.Version, "digest": m.Digest()})
	}}
	verify.Flags().StringVar(&expected, "digest", "", "Required manifest digest")
	unpack := &cobra.Command{Use: "unpack ARCHIVE DIRECTORY", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		m, err := controlbundle.Unpack(f, args[1], expected)
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{"version": m.Version, "digest": m.Digest()})
	}}
	unpack.Flags().StringVar(&expected, "digest", "", "Required manifest digest")
	root.AddCommand(build, verify, unpack)
	rootCmd.AddCommand(root)
}
