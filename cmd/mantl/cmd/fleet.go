// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/fleet"
	"io"
	"os"
)

func init() {
	var endpoint, cert, key, ca string
	command := &cobra.Command{Use: "fleet", Short: "Query or replay metadata through an enrolled mTLS identity"}
	command.PersistentFlags().StringVar(&endpoint, "endpoint", "", "HTTPS fleet API origin")
	command.PersistentFlags().StringVar(&cert, "certificate", "", "Runtime client certificate path")
	command.PersistentFlags().StringVar(&key, "key", "", "Runtime private key path")
	command.PersistentFlags().StringVar(&ca, "ca", "", "Trusted server CA path")
	command.AddCommand(&cobra.Command{Use: "status", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := fleet.NewClient(endpoint, cert, key, ca)
		if err != nil {
			return err
		}
		entries, err := client.State(cmd.Context())
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(entries)
	}})
	command.AddCommand(&cobra.Command{Use: "replay RECEIPTS.json", Short: "Rebuild indexed metadata from immutable, verified journal references", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := fleet.NewClient(endpoint, cert, key, ca)
		if err != nil {
			return err
		}
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if info.Size() > 8<<20 {
			return fmt.Errorf("replay file exceeds 8 MiB")
		}
		var refs []evidence.ObjectRef
		if err = json.NewDecoder(io.LimitReader(file, 8<<20)).Decode(&refs); err != nil {
			return err
		}
		if len(refs) > 10000 {
			return fmt.Errorf("too many journal references")
		}
		for _, ref := range refs {
			if err = client.Publish(cmd.Context(), ref); err != nil {
				return err
			}
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]int{"replayed": len(refs)})
	}})
	rootCmd.AddCommand(command)
}
