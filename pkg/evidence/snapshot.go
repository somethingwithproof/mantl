package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Snapshot represents a piece of point-in-time evidence.
type Snapshot struct {
	Resource    string    `json:"resource"`
	CapturedAt  time.Time `json:"capturedAt"`
	ContentHash string    `json:"contentHash"`
	Data        string    `json:"data"` // JSON representation of the resource
}

// CaptureResource uses kubectl to get a JSON representation of a resource.
// When name is empty, all resources of the given kind are listed. When
// namespace is also empty, the listing spans all namespaces.
func CaptureResource(kind, name, namespace string) (*Snapshot, error) {
	args := []string{"get", kind}
	if name != "" {
		args = append(args, name)
	}
	args = append(args, "-o", "json")
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else if name == "" {
		// Listing without a namespace scope would be limited to the default
		// namespace; span all namespaces so evidence is complete.
		args = append(args, "--all-namespaces")
	}

	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to capture resource: %w (output: %s)", err, string(output))
	}

	hash := sha256.Sum256(output)
	hashStr := hex.EncodeToString(hash[:])

	resourceID := kind
	if name != "" {
		resourceID = fmt.Sprintf("%s/%s", kind, name)
	}

	return &Snapshot{
		Resource:    resourceID,
		CapturedAt:  time.Now(),
		ContentHash: hashStr,
		Data:        string(output),
	}, nil
}

// UploadToS3 uploads the snapshot data to an S3 bucket.
func UploadToS3(ctx context.Context, s *Snapshot, bucket string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	key := fmt.Sprintf("evidence/%s/%s.json", s.Resource, s.CapturedAt.Format(time.RFC3339))

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        strings.NewReader(s.Data),
		ContentType: aws.String("application/json"),
		Metadata: map[string]string{
			"mantl-content-hash": s.ContentHash,
			"mantl-captured-at":  s.CapturedAt.Format(time.RFC3339),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload evidence to S3: %w", err)
	}

	return fmt.Sprintf("s3://%s/%s", bucket, key), nil
}
