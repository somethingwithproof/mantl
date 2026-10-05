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
func TestAcceptanceRequiresAllScopedReceipts(t *testing.T) {
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
