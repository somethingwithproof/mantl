// SPDX-License-Identifier: Apache-2.0

//go:build integration

package controlapi

import (
	"context"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/controllers/compliance"
	"github.com/thomasvincent/mantl/pkg/evidence"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
	"time"
)

func TestGeneratedCRDsAndImmutableRuns(t *testing.T) {
	c, ctx := testControlAPI(t)
	var err error
	namespace := &core.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fixture"}}
	if err = c.Create(ctx, namespace); err != nil {
		t.Fatal(err)
	}
	run := &api.AuditRun{ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: namespace.Name}, Spec: api.AuditRunSpec{Audit: "audit", AuditUID: "audit-uid", Profile: "profile", Framework: "soc2", FrameworkVersion: "2017", BundleDigest: "sha256:bundle", ScheduledAt: metav1.Now(), Tasks: []api.CollectorTask{}}}
	if err = c.Create(ctx, run); err != nil {
		t.Fatal(err)
	}
	run.Status.Phase = "Running"
	if err = c.Status().Update(ctx, run); err != nil {
		t.Fatal("status update should remain permitted", err)
	}
	run.Spec.Profile = "different-profile"
	if err = c.Update(ctx, run); !apierrors.IsInvalid(err) {
		t.Fatal("immutable run input accepted", err)
	}
	jobRun := &api.AuditRun{ObjectMeta: metav1.ObjectMeta{Name: "job-run", Namespace: namespace.Name}, Spec: api.AuditRunSpec{ClusterID: "cluster", Audit: "audit", AuditUID: "audit-uid", Profile: "profile", Framework: "soc2", FrameworkVersion: "2017", BundleDigest: "sha256:bundle", ScheduledAt: metav1.Now(), Tasks: []api.CollectorTask{{ID: evidence.Hash([]byte("task"))[:32], Collector: "C1/0", Control: "C1", Resource: "networkpolicies", Namespace: namespace.Name}}}}
	if err = c.Create(ctx, jobRun); err != nil {
		t.Fatal(err)
	}
	reconciler := &compliance.AuditRunReconciler{Client: c, Store: unusedStore{}, ControlNamespace: namespace.Name, ClusterID: "cluster", Image: "registry/operator@sha256:fixture", Bucket: "fixture", ServiceAccount: "collector"}
	request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(jobRun)}
	for i := 0; i < 2; i++ {
		if _, err = reconciler.Reconcile(ctx, request); err != nil {
			t.Fatal("Job defaults broke immutable scope validation", err)
		}
	}
	var jobs batch.JobList
	if err = c.List(ctx, &jobs); err != nil || len(jobs.Items) != 1 {
		t.Fatal("API retries created duplicate collector jobs", err)
	}
}

type unusedStore struct{}

func (unusedStore) Put(context.Context, string, []byte, time.Time) (evidence.ObjectRef, error) {
	return evidence.ObjectRef{}, fmt.Errorf("no collector has completed")
}
func (unusedStore) Get(context.Context, evidence.ObjectRef) ([]byte, error) {
	return nil, fmt.Errorf("no collector has completed")
}

func testControlAPI(t *testing.T) (client.Client, context.Context) {
	t.Helper()
	environment := &envtest.Environment{CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")}, ErrorIfCRDPathMissing: true}
	cfg, err := environment.Start()
	if err != nil {
		t.Fatal("start explicitly configured API fixture", err)
	}
	t.Cleanup(func() {
		if err := environment.Stop(); err != nil {
			t.Error(err)
		}
	})
	scheme := runtime.NewScheme()
	if err = core.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err = batch.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err = api.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return c, ctx
}
