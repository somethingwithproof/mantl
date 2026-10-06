// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/framework"
	"github.com/thomasvincent/mantl/pkg/oscal"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"time"
)

func init() {
	var directory, output string
	command := &cobra.Command{Use: "oscal", Short: "Exchange control catalogs; enforcement mappings require separate review"}
	export := &cobra.Command{Use: "export FRAMEWORK", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("framework-dir") {
			directory = filepath.Join(sourceDir, "compliance/frameworks")
		}
		fw, err := framework.LoadPinned(directory, args[0], "", "")
		if err != nil {
			return err
		}
		document, err := oscal.Export(fw, time.Now().UTC())
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(document)
	}}
	export.Flags().StringVar(&directory, "framework-dir", "compliance/frameworks", "Verified framework content directory")
	imported := &cobra.Command{Use: "import CATALOG.json", Short: "Create a new unmapped framework draft from an OSCAL catalog", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		info, err := os.Stat(args[0])
		if err != nil {
			return err
		}
		if info.Size() > 4<<20 {
			return fmt.Errorf("OSCAL catalog too large")
		}
		data, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		fw, err := oscal.Import(data)
		if err != nil {
			return err
		}
		payload, err := yaml.Marshal(fw)
		if err != nil {
			return err
		}
		file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(payload)
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		return closeErr
	}}
	imported.Flags().StringVar(&output, "output", "framework-draft.yaml", "New draft path; never overwrites existing content")
	command.AddCommand(export, imported)
	rootCmd.AddCommand(command)
}
