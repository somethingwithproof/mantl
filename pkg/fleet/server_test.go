package fleet

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fixtureStore struct{ objects map[string][]byte }

func (s *fixtureStore) Put(_ context.Context, key string, b []byte, at time.Time) (evidence.ObjectRef, error) {
	hash := evidence.Hash(b)
	uri := "s3://bucket/" + key + "/" + hash + ".json"
	s.objects[uri] = b
	retained := at.AddDate(0, 0, evidence.MinimumRetentionDays)
	return evidence.ObjectRef{URI: uri, Hash: hash, Version: "1", RetainUntil: &retained}, nil
}
func (s *fixtureStore) Get(_ context.Context, ref evidence.ObjectRef) ([]byte, error) {
	b, ok := s.objects[ref.URI]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return b, nil
}

type fixtureIndex struct{ entries []Entry }

func (i *fixtureIndex) Put(_ context.Context, scope Scope, events []Event, source evidence.ObjectRef) error {
	for _, e := range events {
		i.entries = append(i.entries, Entry{scope, e, source})
	}
	return nil
}
func (i *fixtureIndex) List(_ context.Context, scope Scope, _ int, _ string, _ bool) ([]Entry, error) {
	var entries []Entry
	for _, entry := range i.entries {
		if entry.Scope.Tenant == scope.Tenant && (scope.Cluster == "" || entry.Scope.Cluster == scope.Cluster) {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}
func TestCertificateScopeAndJournalVerification(t *testing.T) {
	now := time.Now().UTC()
	cert := &x509.Certificate{Raw: []byte("enrolled certificate")}
	sum := sha256.Sum256(cert.Raw)
	fingerprint := hex.EncodeToString(sum[:])
	scope := Scope{Tenant: "tenant-a", Cluster: "cluster-a", Role: "collector"}
	store := &fixtureStore{objects: map[string][]byte{}}
	index := &fixtureIndex{}
	server := &Server{Enrollments: map[string]Scope{fingerprint: scope}, Store: store, Index: index, Now: func() time.Time { return now }}
	event := Event{ID: evidence.Hash([]byte("event")), Kind: "Finding", ResourceUID: "uid", Namespace: "team", Name: "finding", ObservedAt: now, Result: "fail"}
	payload, err := Canonical(scope.Cluster, []Event{event})
	if err != nil {
		t.Fatal(err)
	}
	ref, err := store.Put(context.Background(), Prefix(scope.Cluster), payload, now)
	if err != nil {
		t.Fatal(err)
	}
	invoke := func(ref evidence.ObjectRef, certificate *x509.Certificate) *httptest.ResponseRecorder {
		body, err := json.Marshal(ref)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/v1/journals", bytes.NewReader(body))
		if certificate != nil {
			req.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{certificate}, VerifiedChains: [][]*x509.Certificate{{certificate}}}
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		return w
	}
	if got := invoke(ref, nil).Code; got != http.StatusUnauthorized {
		t.Fatalf("unauthenticated request: %d", got)
	}
	if got := invoke(ref, &x509.Certificate{Raw: []byte("not enrolled")}).Code; got != http.StatusForbidden {
		t.Fatalf("unenrolled request: %d", got)
	}
	if got := invoke(ref, cert).Code; got != http.StatusAccepted {
		t.Fatalf("valid journal: %d", got)
	}
	if len(index.entries) != 1 || index.entries[0].Scope.Tenant != "tenant-a" {
		t.Fatal("authenticated scope was not used")
	}
	foreign, err := Canonical("cluster-b", []Event{event})
	if err != nil {
		t.Fatal(err)
	}
	other, err := store.Put(context.Background(), Prefix("cluster-b"), foreign, now)
	if err != nil {
		t.Fatal(err)
	}
	if got := invoke(other, cert).Code; got != http.StatusBadRequest {
		t.Fatalf("cross-cluster journal accepted: %d", got)
	}
	store.objects[ref.URI] = []byte(`{"changed":true}`)
	if got := invoke(ref, cert).Code; got != http.StatusBadRequest {
		t.Fatal("tampered journal accepted")
	}
	server.Enrollments[fingerprint] = Scope{Tenant: "tenant-a", Role: "viewer"}
	if got := invoke(ref, cert).Code; got != http.StatusForbidden {
		t.Fatal("viewer could ingest")
	}
}
