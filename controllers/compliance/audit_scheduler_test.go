package compliance

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSchedulerReusesRunAndDoesNotAdvanceFailedCollectorClock(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	requireNoError(t, os.Mkdir(filepath.Join(directory, "soc2"), 0700))
	requireNoError(t, os.WriteFile(filepath.Join(directory, "soc2/framework.yaml"), []byte("spec:\n  version: '2017'\n  controls:\n    - id: C1\n      mappings:\n        - evidenceCollector:\n            type: config-snapshot\n            schedule: '0 * * * *'\n            resources: [networkpolicies]\n"), 0600))
	scheme := runtime.NewScheme()
	requireNoError(t, api.AddToScheme(scheme))
	profile := &api.ComplianceProfile{ObjectMeta: metav1.ObjectMeta{Name: "profile", Namespace: "system"}, Spec: api.ComplianceProfileSpec{Framework: "soc2", Namespaces: []string{"team"}}}
	audit := &api.ComplianceAudit{ObjectMeta: metav1.ObjectMeta{Name: "audit", Namespace: "system", UID: "audit-uid"}, Spec: api.ComplianceAuditSpec{Profile: "profile", Frequency: "manual"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&api.ComplianceAudit{}, &api.AuditRun{}).WithObjects(profile, audit).Build()
	now := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	r := &AuditScheduler{Client: c, ControlNamespace: "system", FrameworkDir: directory, ClusterID: "cluster", Now: func() time.Time { return now }}
	request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(audit)}
	_, err := r.Reconcile(ctx, request)
	requireNoError(t, err)
	requireNoError(t, c.Get(ctx, request.NamespacedName, audit))
	firstRun := audit.Status.ActiveRun
	if firstRun == "" || audit.Status.StartTime == nil {
		t.Fatal("run identity was not persisted")
	}
	// A restart with an active run must not allocate another run.
	_, err = r.Reconcile(ctx, request)
	requireNoError(t, err)
	var runs api.AuditRunList
	requireNoError(t, c.List(ctx, &runs))
	if len(runs.Items) != 1 || len(runs.Items[0].OwnerReferences) != 0 {
		t.Fatal("duplicate run or schedule deletion would erase history")
	}
	run := &runs.Items[0]
	end := metav1.NewTime(now.Add(time.Minute))
	run.Status = api.AuditRunStatus{Phase: "Failed", EndTime: &end, CoverageGaps: []string{"C1: collector failed"}, Results: map[string]api.TaskResult{run.Spec.Tasks[0].ID: {Phase: "Failed"}}}
	requireNoError(t, c.Status().Update(ctx, run))
	now = now.Add(2 * time.Minute)
	_, err = r.Reconcile(ctx, request)
	requireNoError(t, err)
	requireNoError(t, c.Get(ctx, request.NamespacedName, audit))
	if audit.Status.Phase != "Failed" || audit.Status.ActiveRun != "" || len(audit.Status.CollectorTimes) != 0 {
		t.Fatal("failure lost or collector clock advanced", audit.Status)
	}
	requireNoError(t, c.List(ctx, &runs))
	if len(runs.Items) != 1 {
		t.Fatal("manual audit restarted after a terminal result")
	}
}
