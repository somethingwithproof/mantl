package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Snapshot represents a piece of point-in-time evidence.
type Snapshot struct {
	Resource    string    `json:"resource"`
	CapturedAt  time.Time `json:"capturedAt"`
	ContentHash string    `json:"contentHash"`
	Data        string    `json:"data"` // JSON representation of the resource
}

// CaptureResource uses kubectl to get a JSON representation of a resource.
func CaptureResource(kind, name, namespace string) (*Snapshot, error) {
	args := []string{"get", kind, name, "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to capture resource: %w (output: %s)", err, string(output))
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(output)
	hashStr := hex.EncodeToString(hash[:])

	return &Snapshot{
		Resource:    fmt.Sprintf("%s/%s", kind, name),
		CapturedAt:  time.Now(),
		ContentHash: hashStr,
		Data:        string(output),
	}, nil
}

// StoreEvidence handles the upload to object storage (stubbed for now).
func StoreEvidence(s *Snapshot, bucket string) (string, error) {
	fmt.Printf("Evidence stored: %s (Hash: %s)\n", s.Resource, s.ContentHash)
	return fmt.Sprintf("s3://%s/evidence/%s/%s.json", bucket, s.Resource, s.CapturedAt.Format(time.RFC3339)), nil
}
