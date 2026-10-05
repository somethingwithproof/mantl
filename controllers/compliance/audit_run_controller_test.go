package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/collection"
	"github.com/thomasvincent/mantl/pkg/evidence"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

type runStore struct {
	objects map[string][]byte
	fail    bool
}

func (s *runStore) Put(_ context.Context, key string, b []byte, at time.Time) (evidence.ObjectRef, error) {
	if s.fail {
		return evidence.ObjectRef{}, fmt.Errorf("store unavailable")
	}
	hash := evidence.Hash(b)
	uri := "s3://bucket/" + key + "/" + hash + ".json"
	s.objects[uri] = b
	retained := at.AddDate(7, 0, 0)
	return evidence.ObjectRef{URI: uri, Hash: hash, Version: "v1", RetainUntil: &retained}, nil
}
func (s *runStore) Get(_ context.Context, ref evidence.ObjectRef) ([]byte, error) {
	b, ok := s.objects[ref.URI]
	if !ok || evidence.Hash(b) != ref.Hash {
		return nil, fmt.Errorf("invalid object")
	}
	return b, nil
}
func runFixture(t *testing.T, count int) (*AuditRunReconciler, *api.AuditRun, ctrl.Request, *runStore) {
	t.Helper()
	scheme := runtime.NewScheme()
	requireNoError(t, api.AddToScheme(scheme))
	requireNoError(t, core.AddToScheme(scheme))
	requireNoError(t, batch.AddToScheme(scheme))
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	run := &api.AuditRun{ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "system", UID: "run-uid"}, Spec: api.AuditRunSpec{ClusterID: "cluster-uid", Audit: "audit", Profile: "profile", ScheduledAt: metav1.NewTime(now), BundleDigest: "sha256:bundle"}}
	for i := 0; i < count; i++ {
		id := evidence.Hash([]byte(fmt.Sprint(i)))[:32]
		run.Spec.Tasks = append(run.Spec.Tasks, api.CollectorTask{ID: id, Collector: "C1/0", Control: "C1", Resource: "networkpolicies", Namespace: "team"})
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&api.AuditRun{}, &batch.Job{}).WithObjects(run).Build()
	store := &runStore{objects: map[string][]byte{}}
	return &AuditRunReconciler{Client: c, Store: store, ControlNamespace: "system", ClusterID: "cluster-uid", Image: "registry/operator@sha256:digest", Bucket: "bucket", ServiceAccount: "collector", MaxConcurrent: 2, Now: func() time.Time { return now.Add(time.Minute) }}, run, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(run)}, store
}
func TestRunConcurrencyAndRestart(t *testing.T) {
	r, _, req, _ := runFixture(t, 5)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, err := r.Reconcile(ctx, req)
		requireNoError(t, err)
		var jobs batch.JobList
		requireNoError(t, r.List(ctx, &jobs))
		if len(jobs.Items) != 2 {
			t.Fatalf("retry exceeded concurrency: %d", len(jobs.Items))
		}
	}
}
func TestRunVerifiesReceiptBeforeCompletion(t *testing.T) {
	r, run, req, store := runFixture(t, 1)
	ctx := context.Background()
	_, err := r.Reconcile(ctx, req)
	requireNoError(t, err)
	var jobs batch.JobList
	requireNoError(t, r.List(ctx, &jobs))
	job := &jobs.Items[0]
	job.UID = "job-uid"
	requireNoError(t, r.Update(ctx, job))
	job.Status.Succeeded = 1
	requireNoError(t, r.Status().Update(ctx, job))
	payload := []byte(`{"resource":"networkpolicies","namespace":"team","items":[]}`)
	ref, err := store.Put(ctx, collection.Prefix(run.Spec.ClusterID, run.Spec.Tasks[0].Namespace, string(run.UID), run.Spec.Tasks[0].ID), payload, r.Now())
	requireNoError(t, err)
	receipt, err := json.Marshal(collection.Receipt{Ref: ref, CapturedAt: r.Now()})
	requireNoError(t, err)
	pod := &core.Pod{ObjectMeta: metav1.ObjectMeta{Name: "collector", Namespace: "team", Labels: job.Labels, OwnerReferences: []metav1.OwnerReference{{Kind: "Job", UID: job.UID}}}, Status: core.PodStatus{ContainerStatuses: []core.ContainerStatus{{Name: "collector", State: core.ContainerState{Terminated: &core.ContainerStateTerminated{ExitCode: 0, Message: string(receipt)}}}}}}
	requireNoError(t, r.Create(ctx, pod))
	store.fail = true
	_, err = r.Reconcile(ctx, req)
	if err == nil {
		t.Fatal("manifest upload failure accepted")
	}
	var persisted api.AuditRun
	requireNoError(t, r.Get(ctx, req.NamespacedName, &persisted))
	if persisted.Status.EndTime != nil || persisted.Status.EvidenceURI != "" {
		t.Fatal("failed upload published success")
	}
	store.fail = false
	_, err = r.Reconcile(ctx, req)
	requireNoError(t, err)
	requireNoError(t, r.Get(ctx, req.NamespacedName, &persisted))
	if persisted.Status.Phase != "Completed" || persisted.Status.ManifestHash == "" {
		t.Fatalf("run not completed: %+v", persisted.Status)
	}
}
func TestRunRejectsChangedScope(t *testing.T) {
	r, _, req, _ := runFixture(t, 1)
	ctx := context.Background()
	_, err := r.Reconcile(ctx, req)
	requireNoError(t, err)
	var jobs batch.JobList
	requireNoError(t, r.List(ctx, &jobs))
	jobs.Items[0].Spec.Template.Spec.Containers[0].Args = []string{"--resource", "secrets"}
	requireNoError(t, r.Update(ctx, &jobs.Items[0]))
	if _, err = r.Reconcile(ctx, req); err == nil {
		t.Fatal("modified collector accepted")
	}
}

func TestRunCannotSelectAnotherNamespaceOrCollectorIdentity(t *testing.T) {
	for _, attack := range []string{"namespace", "cluster", "image"} {
		t.Run(attack, func(t *testing.T) {
			r, run, req, _ := runFixture(t, 1)
			switch attack {
			case "namespace":
				r.ControlNamespace = "another-system"
			case "cluster":
				r.ClusterID = "another-cluster"
			case "image":
				run.Spec.CollectorImage = "untrusted/image"
				requireNoError(t, r.Update(context.Background(), run))
			}
			if _, err := r.Reconcile(context.Background(), req); err == nil {
				t.Fatal("untrusted run accepted")
			}
			var jobs batch.JobList
			requireNoError(t, r.List(context.Background(), &jobs))
			if len(jobs.Items) != 0 {
				t.Fatal("untrusted run launched a job")
			}
		})
	}
}
