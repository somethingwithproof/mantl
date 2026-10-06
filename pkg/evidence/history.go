package evidence

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
)

// FindingHistoryBucket validates completeness before callers configure storage.
func FindingHistoryBucket(finding api.Finding) (string, error) {
	if finding.UID == "" || finding.Status.HistoryGap || len(finding.Status.Pending) > 0 {
		return "", fmt.Errorf("finding history is incomplete; inspect its durable pending queue")
	}
	if finding.Status.HistoryHead == nil {
		return "", fmt.Errorf("finding has no recorded lifecycle")
	}
	uri, err := url.Parse(finding.Status.HistoryHead.URI)
	if err != nil {
		return "", fmt.Errorf("parse finding history reference: %w", err)
	}
	if uri.Scheme != "s3" || uri.Host == "" {
		return "", fmt.Errorf("invalid finding history reference")
	}
	return uri.Host, nil
}

// ReadFindingHistory verifies a bounded immutable lifecycle chain, newest first.
func ReadFindingHistory(ctx context.Context, store Store, finding api.Finding) ([]json.RawMessage, error) {
	if _, err := FindingHistoryBucket(finding); err != nil {
		return nil, err
	}
	head := finding.Status.HistoryHead
	seen := map[string]bool{}
	var records []json.RawMessage
	total := 0
	for head != nil {
		identity := head.URI + "/" + head.Version
		if len(records) >= 10000 || seen[identity] {
			return nil, fmt.Errorf("finding history is cyclic or exceeds the export limit")
		}
		seen[identity] = true
		if err := findingHistoryScope(head, finding); err != nil {
			return nil, err
		}
		payload, err := store.Get(ctx, ObjectRef{URI: head.URI, Version: head.Version, Hash: head.SHA256})
		if err != nil {
			return nil, fmt.Errorf("read finding history: %w", err)
		}
		total += len(payload)
		if total > 16<<20 {
			return nil, fmt.Errorf("finding history exceeds 16 MiB")
		}
		if Hash(payload) != head.SHA256 {
			return nil, fmt.Errorf("finding history hash mismatch")
		}
		previous, err := previousFindingHistory(payload, finding)
		if err != nil {
			return nil, err
		}
		records = append(records, payload)
		head = previous
	}
	return records, nil
}

func findingHistoryScope(head *api.HistoryReference, finding api.Finding) error {
	uri, err := url.Parse(head.URI)
	if err != nil {
		return fmt.Errorf("parse finding history scope: %w", err)
	}
	if uri.Scheme != "s3" || uri.Host == "" || uri.User != nil || uri.RawQuery != "" || uri.Fragment != "" {
		return fmt.Errorf("finding history scope mismatch")
	}
	parts := strings.Split(strings.TrimPrefix(uri.Path, "/"), "/")
	if (len(parts) != 6 && len(parts) != 7) || parts[0] != "audits" || parts[1] != "findings" || parts[2] == "" || parts[3] != finding.Namespace || parts[4] != string(finding.UID) || parts[5] == "" {
		return fmt.Errorf("finding history scope mismatch")
	}
	// S3Store adds a content-addressed filename to the supplied event prefix.
	// Retain older event-key references, but permit no arbitrary extra suffix.
	if len(parts) == 7 && (!hashPattern.MatchString(head.SHA256) || parts[6] != head.SHA256+".json") {
		return fmt.Errorf("finding history scope mismatch")
	}
	return nil
}

func previousFindingHistory(payload []byte, finding api.Finding) (*api.HistoryReference, error) {
	var record struct {
		Schema   int                   `json:"schemaVersion"`
		Finding  string                `json:"finding"`
		UID      string                `json:"uid"`
		Event    api.FindingTransition `json:"event"`
		Previous *api.HistoryReference `json:"previous"`
	}
	if err := json.Unmarshal(payload, &record); err != nil {
		return nil, fmt.Errorf("decode finding history: %w", err)
	}
	if record.Schema != 1 || record.UID != string(finding.UID) || record.Finding != finding.Namespace+"/"+finding.Name {
		return nil, fmt.Errorf("finding history identity mismatch")
	}
	return record.Previous, nil
}
