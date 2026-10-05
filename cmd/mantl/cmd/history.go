package cmd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
)

func init() {
	complianceCmd.AddCommand(&cobra.Command{Use: "history FINDING", Short: "Verify and print the immutable finding lifecycle, newest first", Args: cobra.ExactArgs(1), RunE: findingHistory})
}

func findingHistory(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()
	data, err := queryCompliance(ctx, "finding", args[0], complianceNamespace)
	if err != nil {
		return err
	}
	var finding api.Finding
	if err := json.Unmarshal(data, &finding); err != nil {
		return err
	}
	bucket, err := evidence.FindingHistoryBucket(finding)
	if err != nil {
		return err
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	store := &evidence.S3Store{Client: s3.NewFromConfig(cfg), Bucket: bucket}
	records, err := evidence.ReadFindingHistory(ctx, store, finding)
	if err != nil {
		return err
	}
	return json.NewEncoder(cmd.OutOrStdout()).Encode(records)
}
