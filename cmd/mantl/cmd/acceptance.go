package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/acceptance"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"os"
	"time"
)

func init() {
	var bucket string
	command := &cobra.Command{Use: "acceptance", Short: "Verify immutable deployment and recovery acceptance records"}
	verify := &cobra.Command{Use: "verify RECORD.json", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if info.Size() > 1<<20 {
			return fmt.Errorf("acceptance record too large")
		}
		var record acceptance.Record
		decoder := json.NewDecoder(file)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&record); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
		defer cancel()
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return err
		}
		report, verifyErr := acceptance.Verify(ctx, &evidence.S3Store{Client: s3.NewFromConfig(cfg), Bucket: bucket}, record, time.Now().UTC())
		if err = json.NewEncoder(cmd.OutOrStdout()).Encode(report); err != nil {
			return err
		}
		return verifyErr
	}}
	verify.Flags().StringVar(&bucket, "bucket", "", "Authorized acceptance evidence bucket")
	command.AddCommand(verify)
	rootCmd.AddCommand(command)
}
