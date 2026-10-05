package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
)

func init() {
	complianceCmd.AddCommand(&cobra.Command{Use: "history FINDING", Short: "Verify and print the immutable finding lifecycle, newest first", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
		defer cancel()
		data, err := queryCompliance(ctx, "finding", args[0], complianceNamespace)
		if err != nil {
			return err
		}
		var finding api.Finding
		if err = json.Unmarshal(data, &finding); err != nil {
			return err
		}
		if finding.UID == "" || finding.Status.HistoryGap || len(finding.Status.Pending) > 0 {
			return fmt.Errorf("finding history is incomplete; inspect its durable pending queue")
		}
		if finding.Status.HistoryHead == nil {
			return fmt.Errorf("finding has no recorded lifecycle")
		}
		head := finding.Status.HistoryHead
		u, err := url.Parse(head.URI)
		if err != nil || u.Scheme != "s3" || u.Host == "" {
			return fmt.Errorf("invalid finding history reference")
		}
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return err
		}
		store := &evidence.S3Store{Client: s3.NewFromConfig(cfg), Bucket: u.Host}
		seen := map[string]bool{}
		var records []json.RawMessage
		total := 0
		for head != nil {
			if len(records) >= 10000 || seen[head.URI+"/"+head.Version] {
				return fmt.Errorf("finding history is cyclic or exceeds the export limit")
			}
			seen[head.URI+"/"+head.Version] = true
			uri, parseErr := url.Parse(head.URI)
			if parseErr != nil || !strings.HasPrefix(uri.Path, "/audits/findings/") || !strings.Contains(uri.Path, "/"+string(finding.UID)+"/") {
				return fmt.Errorf("finding history scope mismatch")
			}
			payload, readErr := store.Get(ctx, evidence.ObjectRef{URI: head.URI, Version: head.Version, Hash: head.SHA256})
			if readErr != nil {
				return readErr
			}
			total += len(payload)
			if total > 16<<20 {
				return fmt.Errorf("finding history exceeds 16 MiB")
			}
			var record struct {
				Schema   int                   `json:"schemaVersion"`
				Finding  string                `json:"finding"`
				UID      string                `json:"uid"`
				Event    api.FindingTransition `json:"event"`
				Previous *api.HistoryReference `json:"previous"`
			}
			if err = json.Unmarshal(payload, &record); err != nil {
				return err
			}
			if record.Schema != 1 || record.UID != string(finding.UID) || record.Finding != finding.Namespace+"/"+finding.Name {
				return fmt.Errorf("finding history identity mismatch")
			}
			records = append(records, payload)
			head = record.Previous
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(records)
	}})
}
