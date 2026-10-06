// SPDX-License-Identifier: Apache-2.0

package fleet

import (
	"testing"
	"time"

	"github.com/thomasvincent/mantl/pkg/evidence"
)

func TestEventValidationPreservesRetentionAndMetadataBoundaries(t *testing.T) {
	now := time.Now().UTC()
	retained := now.AddDate(0, 0, evidence.MinimumRetentionDays)
	for _, test := range []struct {
		name   string
		mutate func(*Event, *evidence.ObjectRef, map[string]bool)
		valid  bool
	}{
		{"retained metadata", func(_ *Event, _ *evidence.ObjectRef, _ map[string]bool) {}, true},
		{"versioned evidence", func(e *Event, _ *evidence.ObjectRef, _ map[string]bool) {
			e.EvidenceURI = "s3://bucket/artifact?versionId=1"
		}, true},
		{"missing retention", func(_ *Event, r *evidence.ObjectRef, _ map[string]bool) { r.RetainUntil = nil }, false},
		{"short retention", func(_ *Event, r *evidence.ObjectRef, _ map[string]bool) { r.RetainUntil = &now }, false},
		{"duplicate identity", func(e *Event, _ *evidence.ObjectRef, seen map[string]bool) { seen[e.ID] = true }, false},
		{"unsupported kind", func(e *Event, _ *evidence.ObjectRef, _ map[string]bool) { e.Kind = "Secret" }, false},
		{"future observation", func(e *Event, _ *evidence.ObjectRef, _ map[string]bool) { e.ObservedAt = now.Add(6 * time.Minute) }, false},
		{"external evidence", func(e *Event, _ *evidence.ObjectRef, _ map[string]bool) {
			e.EvidenceURI = "https://example.test/artifact"
		}, false},
		{"unexpected query", func(e *Event, _ *evidence.ObjectRef, _ map[string]bool) {
			e.EvidenceURI = "s3://bucket/artifact?token=value"
		}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			event := Event{ID: evidence.Hash([]byte("event")), Kind: "Finding", ResourceUID: "uid", Namespace: "team", Name: "finding", ObservedAt: now}
			ref := evidence.ObjectRef{RetainUntil: &retained}
			seen := map[string]bool{}
			test.mutate(&event, &ref, seen)
			if err := validateEvent(event, seen, ref, now); (err == nil) != test.valid {
				t.Fatalf("metadata validation: %v", err)
			}
		})
	}
}
