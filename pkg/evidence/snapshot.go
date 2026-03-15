package evidence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// allowedEvidenceKinds lists resource kinds permitted for evidence collection.
// Secret and ConfigMap are excluded to prevent credential exfiltration.
var allowedEvidenceKinds = map[string]bool{
	"CronJob":             true,
	"DaemonSet":           true,
	"Deployment":          true,
	"Ingress":             true,
	"Job":                 true,
	"LimitRange":          true,
	"NetworkPolicy":       true,
	"Pod":                 true,
	"PodDisruptionBudget": true,
	"ReplicaSet":          true,
	"ResourceQuota":       true,
	"Role":                true,
	"RoleBinding":         true,
	"Service":             true,
	"ServiceAccount":      true,
	"StatefulSet":         true,
}

const (
	maxKubeNameLen   = 253 // RFC 1123 DNS subdomain max length
	maxNamespaceLen  = 63  // Kubernetes namespace max length
	defaultCaptureTimeout = 30 * time.Second
)

var validKubeKindRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9.-]*$`)
var validKubeNameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)
var validNamespaceRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// Snapshot represents a piece of point-in-time evidence.
type Snapshot struct {
	Resource    string    `json:"resource"`
	CapturedAt  time.Time `json:"capturedAt"`
	ContentHash string    `json:"contentHash"`
	Data        string    `json:"data"` // JSON representation of the resource
}

// CaptureResource uses kubectl to get a JSON representation of a resource.
// kind must be in the allowedEvidenceKinds set, and namespace is required
// to prevent unscoped cluster-wide collection.
// Uses a default 30-second timeout. For caller-controlled deadlines, use
// CaptureResourceWithContext.
func CaptureResource(kind, name, namespace string) (*Snapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCaptureTimeout)
	defer cancel()
	return CaptureResourceWithContext(ctx, kind, name, namespace)
}

// CaptureResourceWithContext is like CaptureResource but accepts a
// caller-provided context for deadline and cancellation control.
func CaptureResourceWithContext(ctx context.Context, kind, name, namespace string) (*Snapshot, error) {
	if !allowedEvidenceKinds[kind] {
		return nil, fmt.Errorf("resource kind %q is not permitted for evidence collection", kind)
	}
	if !validKubeKindRe.MatchString(kind) {
		return nil, fmt.Errorf("invalid resource kind %q", kind)
	}
	if name != "" {
		if len(name) > maxKubeNameLen {
			return nil, fmt.Errorf("resource name exceeds maximum length of %d", maxKubeNameLen)
		}
		if !validKubeNameRe.MatchString(name) {
			return nil, fmt.Errorf("invalid resource name %q", name)
		}
	}
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required for evidence collection")
	}
	if len(namespace) > maxNamespaceLen {
		return nil, fmt.Errorf("namespace exceeds maximum length of %d", maxNamespaceLen)
	}
	if !validNamespaceRe.MatchString(namespace) {
		return nil, fmt.Errorf("invalid namespace %q", namespace)
	}

	// Place our flags before "--" so user-controlled kind/name are treated
	// strictly as positional arguments by kubectl.
	args := []string{"get", "-o", "json", "-n", namespace, "--", kind}
	if name != "" {
		args = append(args, name)
	}

	slog.Info("capturing resource evidence", "kind", kind, "name", name, "namespace", namespace)

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		stderrStr := stderr.String()
		slog.Error("kubectl capture failed", "kind", kind, "name", name, "namespace", namespace, "error", err, "stderr", stderrStr)
		if stderrStr == "" {
			stderrStr = "(no stderr output)"
		}
		return nil, fmt.Errorf("failed to capture resource: %w (stderr: %s)", err, stderrStr)
	}

	hash := sha256.Sum256(output)
	hashStr := hex.EncodeToString(hash[:])

	resourceID := kind
	if name != "" {
		resourceID = fmt.Sprintf("%s/%s", kind, name)
	}

	return &Snapshot{
		Resource:    resourceID,
		CapturedAt:  time.Now().UTC(),
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
