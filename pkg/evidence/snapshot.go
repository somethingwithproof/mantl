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

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"net/url"
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
	maxKubeNameLen        = 253 // RFC 1123 DNS subdomain max length
	maxNamespaceLen       = 63  // Kubernetes namespace max length
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

// validateCaptureArgs checks kind, name, and namespace against the evidence
// collection allowlist and syntax rules. It returns nil when the arguments are
// safe to pass to kubectl. Keeping validation separate from process execution
// lets tests exercise the rules without a live cluster.
func validateCaptureArgs(kind, name, namespace string) error {
	if !validKubeKindRe.MatchString(kind) {
		return fmt.Errorf("invalid resource kind %q", kind)
	}
	if !allowedEvidenceKinds[kind] {
		return fmt.Errorf("resource kind %q is not permitted for evidence collection", kind)
	}
	if name != "" {
		if len(name) > maxKubeNameLen {
			return fmt.Errorf("resource name exceeds maximum length of %d", maxKubeNameLen)
		}
		if !validKubeNameRe.MatchString(name) {
			return fmt.Errorf("invalid resource name %q", name)
		}
	}
	if namespace == "" {
		return fmt.Errorf("namespace is required for evidence collection")
	}
	if len(namespace) > maxNamespaceLen {
		return fmt.Errorf("namespace exceeds maximum length of %d", maxNamespaceLen)
	}
	if !validNamespaceRe.MatchString(namespace) {
		return fmt.Errorf("invalid namespace %q", namespace)
	}
	return nil
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
	if err := validateCaptureArgs(kind, name, namespace); err != nil {
		return nil, err
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

	return &Snapshot{
		Resource:    BuildResourceID(kind, name),
		CapturedAt:  time.Now().UTC(),
		ContentHash: hashStr,
		Data:        string(output),
	}, nil
}

// BuildResourceID returns a resource identifier in the form "kind/name".
// If name is empty, only the kind is returned.
// Slashes in kind or name are replaced with underscores to prevent
// ambiguous identifiers.
func BuildResourceID(kind, name string) string {
	if kind == "" {
		return ""
	}
	kind = strings.ReplaceAll(kind, "/", "__")
	name = strings.ReplaceAll(name, "/", "__")
	if name != "" {
		return fmt.Sprintf("%s/%s", kind, name)
	}
	return kind
}

// DefaultServerSideEncryption is the S3 encryption method used when none is
// specified. Override via UploadToS3WithEncryption for buckets that require
// aws:kms.
const DefaultServerSideEncryption = s3types.ServerSideEncryptionAes256

// UploadToS3 uploads the snapshot data to an S3 bucket using the default
// server-side encryption (AES256). For buckets that require a different
// encryption method (e.g. aws:kms), use UploadToS3WithEncryption.
func UploadToS3(ctx context.Context, s *Snapshot, bucket string) (string, error) {
	return UploadToS3WithEncryption(ctx, s, bucket, DefaultServerSideEncryption)
}

// UploadToS3WithEncryption uploads the snapshot data to an S3 bucket with the
// specified server-side encryption method (e.g. "AES256" or "aws:kms").
// UploadToS3WithEncryption is retained for compatibility. New callers should
// use Store so they retain the object version and manifest linkage.
func UploadToS3WithEncryption(ctx context.Context, s *Snapshot, bucket string, encryption s3types.ServerSideEncryption) (string, error) {
	if s == nil || s.ContentHash != Hash([]byte(s.Data)) {
		return "", fmt.Errorf("invalid snapshot or content hash")
	}
	if encryption != s3types.ServerSideEncryptionAes256 {
		return "", fmt.Errorf("configure a KMS key using S3Store for non-default encryption")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("configure immutable evidence store: %w", err)
	}
	store := &S3Store{Client: s3.NewFromConfig(cfg), Bucket: bucket}
	ref, err := store.Put(ctx, "evidence/"+s.Resource+"/"+s.CapturedAt.UTC().Format(time.RFC3339Nano), []byte(s.Data), time.Now().UTC())
	if err != nil {
		return "", err
	}
	return ref.URI + "?versionId=" + url.QueryEscape(ref.Version), nil
}
