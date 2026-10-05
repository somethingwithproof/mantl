package evidence

import (
	"context"
	"encoding/json"
	"errors"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
	"time"
)

type historyStore struct {
	data  map[string][]byte
	err   error
	reads []ObjectRef
}

func (s *historyStore) Put(context.Context, string, []byte, time.Time) (ObjectRef, error) {
	return ObjectRef{}, errors.New("read only")
}
func (s *historyStore) Get(_ context.Context, ref ObjectRef) ([]byte, error) {
	s.reads = append(s.reads, ref)
	if s.err != nil {
		return nil, s.err
	}
	data, ok := s.data[ref.URI]
	if !ok {
		return nil, errors.New("missing version")
	}
	return data, nil
}
func historyFixture(t *testing.T) (api.Finding, *historyStore) {
	t.Helper()
	finding := api.Finding{ObjectMeta: metav1.ObjectMeta{Namespace: "tenant", Name: "finding", UID: "uid"}}
	store := &historyStore{data: map[string][]byte{}}
	var previous *api.HistoryReference
	for _, id := range []string{"old", "new"} {
		data, err := json.Marshal(map[string]any{"schemaVersion": 1, "finding": "tenant/finding", "uid": "uid", "event": map[string]string{"id": id}, "previous": previous})
		if err != nil {
			t.Fatal(err)
		}
		uri := "s3://bucket/audits/findings/cluster/tenant/uid/" + id
		store.data[uri] = data
		previous = &api.HistoryReference{URI: uri, Version: "immutable-" + id, SHA256: Hash(data)}
	}
	finding.Status.HistoryHead = previous
	return finding, store
}
func TestFindingHistoryExactVersionsAndNewestFirst(t *testing.T) {
	finding, store := historyFixture(t)
	records, err := ReadFindingHistory(context.Background(), store, finding)
	if err != nil || len(records) != 2 {
		t.Fatalf("records=%v err=%v", records, err)
	}
	if !strings.Contains(string(records[0]), `"id":"new"`) || !strings.Contains(string(records[1]), `"id":"old"`) {
		t.Fatal("lifecycle order lost")
	}
	if store.reads[0].Version != "immutable-new" || store.reads[1].Version != "immutable-old" || store.reads[0].Hash != finding.Status.HistoryHead.SHA256 {
		t.Fatal("immutable identity lost")
	}
}
func TestFindingHistoryRejectsIncompleteMetadata(t *testing.T) {
	for _, mutate := range []func(*api.Finding){
		func(f *api.Finding) { f.UID = "" }, func(f *api.Finding) { f.Status.HistoryGap = true }, func(f *api.Finding) { f.Status.Pending = []api.FindingTransition{{ID: "pending"}} }, func(f *api.Finding) { f.Status.HistoryHead = nil }, func(f *api.Finding) { f.Status.HistoryHead.URI = "https://example.com/history" },
	} {
		finding, store := historyFixture(t)
		mutate(&finding)
		if _, err := ReadFindingHistory(context.Background(), store, finding); err == nil {
			t.Fatal("incomplete history accepted")
		}
		if len(store.reads) != 0 {
			t.Fatal("invalid metadata caused storage access")
		}
	}
}
func TestFindingHistoryRejectsForeignOrTamperedRecords(t *testing.T) {
	for _, data := range []string{`not json`, `{"schemaVersion":2,"uid":"uid","finding":"tenant/finding"}`, `{"schemaVersion":1,"uid":"foreign","finding":"tenant/finding"}`, `{"schemaVersion":1,"uid":"uid","finding":"other/finding"}`} {
		finding, store := historyFixture(t)
		head := finding.Status.HistoryHead
		store.data[head.URI] = []byte(data)
		head.SHA256 = Hash([]byte(data))
		if _, err := ReadFindingHistory(context.Background(), store, finding); err == nil {
			t.Fatal("foreign/invalid record accepted")
		}
	}
	finding, store := historyFixture(t)
	store.data[finding.Status.HistoryHead.URI] = []byte("tampered")
	if _, err := ReadFindingHistory(context.Background(), store, finding); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatal(err)
	}
	finding, store = historyFixture(t)
	finding.Status.HistoryHead.URI = "s3://bucket/audits/findings/cluster/tenant/other/new"
	if _, err := ReadFindingHistory(context.Background(), store, finding); err == nil || !strings.Contains(err.Error(), "scope mismatch") {
		t.Fatal(err)
	}
}
func TestFindingHistoryPreservesReadErrorsAndBoundsPayloads(t *testing.T) {
	finding, store := historyFixture(t)
	cause := errors.New("storage unavailable")
	store.err = cause
	if _, err := ReadFindingHistory(context.Background(), store, finding); !errors.Is(err, cause) {
		t.Fatal(err)
	}
	store.err = nil
	store.data[finding.Status.HistoryHead.URI] = make([]byte, 17<<20)
	if _, err := ReadFindingHistory(context.Background(), store, finding); err == nil || !strings.Contains(err.Error(), "16 MiB") {
		t.Fatal(err)
	}
}
func TestFindingHistoryDetectsCyclicReferences(t *testing.T) {
	finding, store := historyFixture(t)
	head := finding.Status.HistoryHead
	data, err := json.Marshal(map[string]any{"schemaVersion": 1, "finding": "tenant/finding", "uid": "uid", "previous": head})
	if err != nil {
		t.Fatal(err)
	}
	store.data[head.URI] = data
	head.SHA256 = Hash(data)
	// The embedded previous reference has the same URI and immutable version.
	if _, err := ReadFindingHistory(context.Background(), store, finding); err == nil || !strings.Contains(err.Error(), "cyclic") {
		t.Fatal(err)
	}
}

func TestFindingHistoryRejectsNoncanonicalPathsBeforeRead(t *testing.T) {
	for _, uri := range []string{
		"s3://bucket/audits/findings/cluster/other/uid/new",
		"s3://bucket/audits/findings/cluster/tenant/other/uid/new",
		"s3://bucket/audits/findings//tenant/uid/new",
		"s3://bucket/audits/findings/cluster/tenant/uid/",
		"s3://bucket/audits/findings/cluster/tenant/uid/new/extra",
		"s3://bucket/audits/findings/cluster/tenant/uid/new?versionId=other",
		"s3://bucket/audits/findings/cluster/tenant/uid/new#fragment",
	} {
		finding, store := historyFixture(t)
		finding.Status.HistoryHead.URI = uri
		if _, err := ReadFindingHistory(context.Background(), store, finding); err == nil || !strings.Contains(err.Error(), "scope mismatch") {
			t.Fatalf("noncanonical path %s: %v", uri, err)
		}
		if len(store.reads) != 0 {
			t.Fatalf("invalid path read storage: %s", uri)
		}
	}
}
