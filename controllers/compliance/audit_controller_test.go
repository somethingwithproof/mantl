package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

type memoryStore struct {
	objects map[string][]byte
	fail    bool
}

func (s *memoryStore) Put(_ context.Context, key string, data []byte, _ time.Time) (evidence.ObjectRef, error) {
	if s.fail {
		return evidence.ObjectRef{}, fmt.Errorf("store unavailable")
	}
	hash := evidence.Hash(data)
	s.objects[key] = append([]byte(nil), data...)
	return evidence.ObjectRef{URI: "s3://evidence/" + key, Hash: hash, Version: "v1"}, nil
}
func (s *memoryStore) Get(_ context.Context, ref evidence.ObjectRef) ([]byte, error) {
	return nil, fmt.Errorf("unused")
}

type fakeReader struct {
	fail  bool
	calls int
}

func (f *fakeReader) Capture(_ context.Context, resource, namespace string, at time.Time) (*evidence.Snapshot, error) {
	f.calls++
	if f.fail {
		return nil, fmt.Errorf("read denied")
	}
	return &evidence.Snapshot{Resource: namespace + "/" + resource, CapturedAt: at, Data: `{"items":[]}`}, nil
}
func auditFixture(t *testing.T) (*ComplianceAuditReconciler, client.Client, ctrl.Request, *fakeReader, *memoryStore) {
	t.Helper()
	dir := t.TempDir()
	requireNoError(t, os.Mkdir(filepath.Join(dir, "fixture"), 0700))
	requireNoError(t, os.WriteFile(filepath.Join(dir, "fixture/framework.yaml"), []byte(`spec:
  version: "1"
  controls:
    - id: C1
      mappings:
        - type: evidence
          evidenceCollector:
            type: config-snapshot
            schedule: "0 * * * *"
            resources: [networkpolicies]
`), 0600))
	scheme := runtime.NewScheme()
	requireNoError(t, api.AddToScheme(scheme))
	profile := &api.ComplianceProfile{ObjectMeta: metav1.ObjectMeta{Name: "profile", Namespace: "tenant"}, Spec: api.ComplianceProfileSpec{Framework: "fixture", Version: "1"}}
	audit := &api.ComplianceAudit{ObjectMeta: metav1.ObjectMeta{Name: "audit", Namespace: "tenant", UID: "uid"}, Spec: api.ComplianceAuditSpec{Profile: "profile", Frequency: "framework"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(audit).WithObjects(profile, audit).Build()
	reader := &fakeReader{}
	store := &memoryStore{objects: map[string][]byte{}}
	now := time.Date(2026, 1, 1, 12, 15, 0, 0, time.UTC)
	reconciler := &ComplianceAuditReconciler{Client: c, Scheme: scheme, FrameworkDir: dir, Reader: reader, Store: store, Now: func() time.Time { return now }}
	return reconciler, c, ctrl.Request{NamespacedName: types.NamespacedName{Name: "audit", Namespace: "tenant"}}, reader, store
}
func TestAuditCollectsStoresAndSchedules(t *testing.T) {
	r, c, req, reader, store := auditFixture(t)
	res, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	var audit api.ComplianceAudit
	requireNoError(t, c.Get(context.Background(), req.NamespacedName, &audit))
	if audit.Status.Phase != "Completed" || audit.Status.ManifestHash == "" || audit.Status.EvidenceURI == "" || res.RequeueAfter != 45*time.Minute {
		t.Fatalf("status=%+v result=%+v", audit.Status, res)
	}
	manifestFound := false
	for _, data := range store.objects {
		var m evidence.Manifest
		requireNoError(t, json.Unmarshal(data, &m))
		if m.SchemaVersion == 1 {
			manifestFound = true
			if len(m.Objects) != 1 || m.Objects[0].Control != "C1" || m.Objects[0].Version == "" {
				t.Fatal("manifest lacks linkage")
			}
		}
	}
	if !manifestFound {
		t.Fatal("no manifest stored")
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if reader.calls != 1 {
		t.Fatal("duplicate capture before due")
	}
	r.Now = func() time.Time { return time.Date(2026, 1, 1, 13, 1, 0, 0, time.UTC) }
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if reader.calls != 2 {
		t.Fatal("scheduled capture did not run")
	}
}
func TestAuditCaptureAndManifestFailures(t *testing.T) {
	r, c, req, reader, store := auditFixture(t)
	reader.fail = true
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	var audit api.ComplianceAudit
	requireNoError(t, c.Get(context.Background(), req.NamespacedName, &audit))
	if audit.Status.Phase != "Failed" || len(audit.Status.CoverageGaps) == 0 || len(audit.Status.CollectorTimes) > 0 {
		t.Fatalf("failure hidden: %+v", audit.Status)
	}
	reader.fail = false
	store.fail = true
	if _, err := r.Reconcile(context.Background(), req); err == nil {
		t.Fatal("manifest failure hidden")
	}
	requireNoError(t, c.Get(context.Background(), req.NamespacedName, &audit))
	start := audit.Status.StartTime.Time
	if audit.Status.EvidenceURI != "" {
		t.Fatal("failed run retained success URI")
	}
	store.fail = false
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	requireNoError(t, c.Get(context.Background(), req.NamespacedName, &audit))
	if !audit.Status.StartTime.Time.Equal(start) {
		t.Fatal("restart lost run identity")
	}
}
