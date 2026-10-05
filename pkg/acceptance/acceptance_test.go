package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"testing"
	"time"
)

type fixture struct{ objects map[string][]byte }

func (f fixture) Put(context.Context, string, []byte, time.Time) (evidence.ObjectRef, error) {
	return evidence.ObjectRef{}, fmt.Errorf("read-only")
}
func (f fixture) Get(_ context.Context, ref evidence.ObjectRef) ([]byte, error) {
	b, ok := f.objects[ref.URI]
	if !ok {
		return nil, fmt.Errorf("missing")
	}
	return b, nil
}
func acceptanceFixture(t *testing.T) (Record, fixture) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	hash := evidence.Hash([]byte("digest"))
	record := Record{Schema: 1, Provider: "aws", ClusterID: "cluster", RunID: "run", OperatorDigest: "sha256:" + hash, BundleDigest: "sha256:" + hash, FinishedAt: now, Checks: map[string]evidence.ObjectRef{}}
	store := fixture{objects: map[string][]byte{}}
	artifact := []byte(`{"test":"actual fixture observation"}`)
	source := evidence.ObjectRef{URI: "s3://bucket/artifact", Hash: evidence.Hash(artifact), Version: "1"}
	store.objects[source.URI] = artifact
	for _, check := range Required {
		data, err := json.Marshal(Receipt{OperatorDigest: record.OperatorDigest, BundleDigest: record.BundleDigest, Provider: record.Provider, ClusterID: record.ClusterID, RunID: record.RunID, Check: check, Passed: true, ObservedAt: now, RunnerURI: "https://github.com/somethingwithproof/mantl/actions/runs/1", Artifacts: []evidence.ObjectRef{source}})
		if err != nil {
			t.Fatal(err)
		}
		ref := evidence.ObjectRef{URI: "s3://bucket/" + Prefix(record, check) + "/" + evidence.Hash(data) + ".json", Hash: evidence.Hash(data), Version: "1"}
		store.objects[ref.URI] = data
		record.Checks[check] = ref
	}
	return record, store
}

func TestAcceptanceRequiresAllScopedReceipts(t *testing.T) {
	record, store := acceptanceFixture(t)
	now := record.FinishedAt
	report, err := Verify(context.Background(), store, record, now)
	if err != nil || report.State != "verified-receipts" {
		t.Fatal(report, err)
	}
	delete(record.Checks, "recovery")
	if _, err = Verify(context.Background(), store, record, now); err == nil {
		t.Fatal("missing recovery accepted")
	}
	record.Provider = "gcp"
	if _, err = Verify(context.Background(), store, record, now); err == nil {
		t.Fatal("foreign provider receipt accepted")
	}
}

func rewriteReceipt(t *testing.T, record *Record, store fixture, mutate func(*Receipt)) {
	t.Helper()
	check := Required[0]
	ref := record.Checks[check]
	var receipt Receipt
	if err := json.Unmarshal(store.objects[ref.URI], &receipt); err != nil {
		t.Fatal(err)
	}
	mutate(&receipt)
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	ref.Hash = evidence.Hash(data)
	ref.URI = "s3://bucket/" + Prefix(*record, check) + "/" + ref.Hash + ".json"
	record.Checks[check] = ref
	store.objects[ref.URI] = data
}

func TestAcceptanceRejectsUntrustedOrIncompleteReceipts(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*Receipt)
	}{
		{"failed check", func(r *Receipt) { r.Passed = false }},
		{"different cluster", func(r *Receipt) { r.ClusterID = "foreign" }},
		{"different image", func(r *Receipt) { r.OperatorDigest = "sha256:" + evidence.Hash([]byte("foreign")) }},
		{"future observation", func(r *Receipt) { r.ObservedAt = r.ObservedAt.Add(time.Second) }},
		{"stale observation", func(r *Receipt) { r.ObservedAt = r.ObservedAt.Add(-25 * time.Hour) }},
		{"foreign runner", func(r *Receipt) { r.RunnerURI = "https://github.com/foreign/repo/actions/runs/1" }},
		{"missing artifacts", func(r *Receipt) { r.Artifacts = nil }},
		{"unversioned artifact", func(r *Receipt) { r.Artifacts[0].Version = "" }},
		{"invalid artifact digest", func(r *Receipt) { r.Artifacts[0].Hash = "invalid" }},
		{"unavailable artifact", func(r *Receipt) { r.Artifacts[0].URI = "s3://bucket/missing" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			record, store := acceptanceFixture(t)
			rewriteReceipt(t, &record, store, test.mutate)
			report, err := Verify(context.Background(), store, record, record.FinishedAt)
			if err == nil || report.State == "verified-receipts" {
				t.Fatalf("invalid receipt accepted: %+v, %v", report, err)
			}
		})
	}
}

func TestAcceptanceRejectsInvalidIdentityReferencesAndContent(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*Record, fixture)
	}{
		{"invalid schema", func(r *Record, _ fixture) { r.Schema = 2 }},
		{"stale record", func(r *Record, _ fixture) { r.FinishedAt = r.FinishedAt.Add(-31 * 24 * time.Hour) }},
		{"missing receipt", func(r *Record, s fixture) { delete(s.objects, r.Checks[Required[0]].URI) }},
		{"altered receipt", func(r *Record, s fixture) { s.objects[r.Checks[Required[0]].URI] = []byte("changed") }},
		{"unversioned receipt", func(r *Record, _ fixture) {
			ref := r.Checks[Required[0]]
			ref.Version = ""
			r.Checks[Required[0]] = ref
		}},
		{"altered artifact", func(_ *Record, s fixture) { s.objects["s3://bucket/artifact"] = []byte("changed") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			record, store := acceptanceFixture(t)
			test.mutate(&record, store)
			report, err := Verify(context.Background(), store, record, time.Now().UTC())
			if err == nil || report.State == "verified-receipts" {
				t.Fatalf("invalid acceptance accepted: %+v, %v", report, err)
			}
		})
	}
}
