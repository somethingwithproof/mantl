// SPDX-License-Identifier: Apache-2.0

//go:build integration

package controlapi

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/controllers/compliance"
	"github.com/thomasvincent/mantl/pkg/collection"
	"github.com/thomasvincent/mantl/pkg/evidence"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// This fixture executes real API reads, reconcilers, S3Store verification, and
// export. Reports, Job completion, and the S3 API are simulated; no Kyverno
// admission, kubelet, workload identity, or AWS acceptance is claimed.
func TestLocalComplianceLifecycle(t *testing.T) {
	c, cfg, ctx := testControlAPIConfig(t, reportFixtureCRDs())
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"control", "team"} {
		must(c.Create(ctx, &core.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}))
	}
	dir := t.TempDir()
	must(os.MkdirAll(filepath.Join(dir, "fixture", "policies"), 0700))
	must(os.WriteFile(filepath.Join(dir, "fixture", "framework.yaml"), []byte(`spec:
  version: "1"
  controls:
    - id: C1
      mappings:
        - policyRef:
            template: isolation
        - evidenceCollector:
            type: config-snapshot
            resources: [networkpolicies]
`), 0600))
	must(os.WriteFile(filepath.Join(dir, "fixture", "policies", "isolation.yaml"), []byte("metadata:\n  name: isolation\nspec:\n  rules:\n    - name: default-deny\n"), 0600))
	now := time.Now().UTC().Truncate(time.Second)
	profile := &api.ComplianceProfile{ObjectMeta: metav1.ObjectMeta{Name: "profile", Namespace: "control"}, Spec: api.ComplianceProfileSpec{Framework: "fixture", Version: "1", Namespaces: []string{"team"}, MaxEvidenceAgeSeconds: 60}}
	must(c.Create(ctx, profile))
	network := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "networking.k8s.io/v1", "kind": "NetworkPolicy",
		"metadata": map[string]interface{}{"name": "isolation", "namespace": "team"},
		"spec":     map[string]interface{}{"podSelector": map[string]interface{}{}, "policyTypes": []interface{}{"Ingress", "Egress"}, "ingress": []interface{}{map[string]interface{}{}}, "egress": []interface{}{map[string]interface{}{}}},
	}}
	must(c.Create(ctx, network))
	report := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "wgpolicyk8s.io/v1alpha2", "kind": "PolicyReport",
		"metadata": map[string]interface{}{"name": "isolation", "namespace": "team"},
		"results": []interface{}{map[string]interface{}{
			"policy": "isolation", "rule": "default-deny", "result": "pass",
			"timestamp": map[string]interface{}{"seconds": now.Unix()},
			"resources": []interface{}{map[string]interface{}{"kind": "NetworkPolicy", "namespace": "team", "name": "isolation", "uid": string(network.GetUID())}},
		}},
	}}
	must(c.Create(ctx, report))
	backend := &fixtureS3{objects: map[string]fixtureObject{}}
	store := &evidence.S3Store{Client: backend, Bucket: "fixture"}
	evaluations := &compliance.EvaluationReconciler{Client: c, FrameworkDir: dir, Now: func() time.Time { return now }}
	evaluate := func() api.ControlEvaluationStatus {
		t.Helper()
		_, err := evaluations.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(profile)})
		must(err)
		var list api.ControlEvaluationList
		must(c.List(ctx, &list, client.InNamespace("control")))
		if len(list.Items) != 1 {
			t.Fatalf("expected one evaluation: %d", len(list.Items))
		}
		return list.Items[0].Status
	}
	if status := evaluate(); status.Result != "unknown" || status.Freshness != "stale" || status.EvidenceURI != "" {
		t.Fatalf("missing collection established a pass: %+v", status)
	}
	setResult := func(value string) {
		t.Helper()
		report.Object["results"].([]interface{})[0].(map[string]interface{})["result"] = value
		must(c.Update(ctx, report))
	}
	setResult("fail")
	findings := &compliance.FindingReconciler{Client: c, Scheme: c.Scheme(), FrameworkDir: dir, Store: store, ClusterID: "fixture-cluster"}
	reconcileFinding := func() api.Finding {
		t.Helper()
		_, err := findings.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(report)})
		must(err)
		var list api.FindingList
		must(c.List(ctx, &list, client.InNamespace("team")))
		if len(list.Items) != 1 {
			t.Fatalf("expected one finding: %d", len(list.Items))
		}
		return list.Items[0]
	}
	if finding := reconcileFinding(); finding.Spec.Status != "fail" || finding.Spec.ControlID != "C1" || finding.Status.HistoryHead == nil {
		t.Fatalf("violation lacks mapped retained history: %+v", finding)
	}
	dc, err := dynamic.NewForConfig(cfg)
	must(err)
	reader := &evidence.KubernetesReader{Client: dc}
	image := "ghcr.io/somethingwithproof/mantl-operator@sha256:" + strings.Repeat("0", 64)
	scheduler := &compliance.AuditScheduler{Client: c, FrameworkDir: dir, ControlNamespace: "control", ClusterID: "fixture-cluster", CollectorImage: image, CollectorServiceAccount: "collector", Bucket: "fixture", Now: func() time.Time { return now }}
	runner := &compliance.AuditRunReconciler{Client: c, Store: store, ControlNamespace: "control", ClusterID: "fixture-cluster", Image: image, ServiceAccount: "collector", Bucket: "fixture", Now: func() time.Time { return now }}
	collect := func(name string) api.AuditRun {
		t.Helper()
		audit := &api.ComplianceAudit{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "control"}, Spec: api.ComplianceAuditSpec{Profile: "profile", Frequency: "manual"}}
		must(c.Create(ctx, audit))
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(audit)}
		_, err := scheduler.Reconcile(ctx, req)
		must(err)
		must(c.Get(ctx, req.NamespacedName, audit))
		var run api.AuditRun
		must(c.Get(ctx, client.ObjectKey{Name: audit.Status.ActiveRun, Namespace: "control"}, &run))
		runReq := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&run)}
		_, err = runner.Reconcile(ctx, runReq)
		must(err)
		var jobs batch.JobList
		must(c.List(ctx, &jobs, client.InNamespace("team"), client.MatchingLabels{collection.RunLabel: string(run.UID)}))
		if len(jobs.Items) != 1 || len(run.Spec.Tasks) != 1 {
			t.Fatal("incomplete collection scope")
		}
		receipt, err := collection.Execute(ctx, reader, store, run.Spec.Tasks[0], string(run.UID), run.Spec.ClusterID, now)
		must(err)
		payload, err := json.Marshal(receipt)
		must(err)
		job := &jobs.Items[0]
		pod := &core.Pod{ObjectMeta: metav1.ObjectMeta{Name: name + "-collector", Namespace: "team", Labels: job.Labels, OwnerReferences: []metav1.OwnerReference{{APIVersion: "batch/v1", Kind: "Job", Name: job.Name, UID: job.UID}}}, Spec: job.Spec.Template.Spec}
		must(c.Create(ctx, pod))
		pod.Status.Phase = core.PodSucceeded
		pod.Status.ContainerStatuses = []core.ContainerStatus{{Name: "collector", State: core.ContainerState{Terminated: &core.ContainerStateTerminated{ExitCode: 0, Message: string(payload)}}}}
		must(c.Status().Update(ctx, pod))
		job.Status.Succeeded = 1
		must(c.Status().Update(ctx, job))
		_, err = runner.Reconcile(ctx, runReq)
		must(err)
		must(c.Get(ctx, runReq.NamespacedName, &run))
		if run.Status.Phase != "Completed" || run.Status.ManifestHash == "" || len(run.Status.CoverageGaps) != 0 {
			t.Fatalf("unverified run completed: %+v", run.Status)
		}
		_, err = scheduler.Reconcile(ctx, req)
		must(err)
		must(c.Get(ctx, req.NamespacedName, audit))
		if audit.Status.EvidenceURI != run.Status.EvidenceURI {
			t.Fatal("schedule did not retain completed run reference")
		}
		return run
	}
	collect("violation")
	if status := evaluate(); status.Result != "fail" || status.Freshness != "current" || status.EvidenceURI == "" {
		t.Fatalf("violation hidden by collected evidence: %+v", status)
	}
	exception := &api.ComplianceException{ObjectMeta: metav1.ObjectMeta{Name: "reviewed", Namespace: "control"}, Spec: api.ComplianceExceptionSpec{Profile: "profile", Control: "C1", Owner: "fixture-owner", Justification: "temporary fixture exception", ExpiresAt: metav1.NewTime(now.Add(time.Hour))}}
	must(c.Create(ctx, exception))
	approvedAt := metav1.NewTime(now)
	exception.Status = api.ComplianceExceptionStatus{Approved: true, ApprovedBy: "fixture-approver", ApprovedAt: &approvedAt, ApprovedGeneration: exception.Generation}
	must(c.Status().Update(ctx, exception))
	if status := evaluate(); status.Result != "fail" || len(status.ExceptionNames) != 1 {
		t.Fatalf("approval erased failure: %+v", status)
	}
	network.Object["spec"] = map[string]interface{}{"podSelector": map[string]interface{}{}, "policyTypes": []interface{}{"Ingress", "Egress"}}
	must(c.Update(ctx, network))
	now = now.Add(time.Second)
	setResult("pass")
	if finding := reconcileFinding(); finding.Spec.Status != "resolved" {
		t.Fatal("current repair did not resolve finding")
	}
	completed := collect("repaired")
	if status := evaluate(); status.Result != "pass" || status.Freshness != "current" {
		t.Fatalf("fresh repaired scope unavailable: %+v", status)
	}
	uri, err := url.Parse(completed.Status.EvidenceURI)
	must(err)
	version := uri.Query().Get("versionId")
	uri.RawQuery = ""
	ref := evidence.ObjectRef{URI: uri.String(), Hash: completed.Status.ManifestHash, Version: version}
	var archive bytes.Buffer
	must(evidence.Export(ctx, store, ref, &archive))
	verifyFixtureArchive(t, archive.Bytes())
	now = now.Add(2 * time.Minute)
	if status := evaluate(); status.Result == "pass" || status.Freshness != "stale" {
		t.Fatalf("stale scope established a current pass: %+v", status)
	}
	must(c.Delete(ctx, report))
	finding := reconcileFinding()
	if finding.Spec.Status != "unknown" {
		t.Fatal("deleted report resolved finding")
	}
	history, err := evidence.ReadFindingHistory(ctx, store, finding)
	must(err)
	if len(history) != 3 {
		t.Fatalf("incomplete finding chain: %d", len(history))
	}
	if status := evaluate(); status.Result == "pass" {
		t.Fatal("missing report established a pass")
	}
	backend.corrupt = true
	if err := evidence.Export(ctx, store, ref, io.Discard); err == nil {
		t.Fatal("corrupted retained bytes exported")
	}
	t.Log("local fixture: violation, retained history, collection, exception, repair, verified export, stale/missing evidence and corruption checks passed")
}

func reportFixtureCRDs() []*apiext.CustomResourceDefinition {
	preserve := true
	var out []*apiext.CustomResourceDefinition
	for _, kind := range []string{"PolicyReport", "ClusterPolicyReport"} {
		plural, scope := "policyreports", apiext.NamespaceScoped
		if kind == "ClusterPolicyReport" {
			plural, scope = "clusterpolicyreports", apiext.ClusterScoped
		}
		out = append(out, &apiext.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: plural + ".wgpolicyk8s.io"}, Spec: apiext.CustomResourceDefinitionSpec{Group: "wgpolicyk8s.io", Scope: scope, Names: apiext.CustomResourceDefinitionNames{Plural: plural, Singular: strings.ToLower(kind), Kind: kind, ListKind: kind + "List"}, Versions: []apiext.CustomResourceDefinitionVersion{{Name: "v1alpha2", Served: true, Storage: true, Schema: &apiext.CustomResourceValidation{OpenAPIV3Schema: &apiext.JSONSchemaProps{Type: "object", XPreserveUnknownFields: &preserve}}}}}})
	}
	return out
}

func verifyFixtureArchive(t *testing.T, data []byte) {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = gz.Close() }()
	archive := tar.NewReader(gz)
	entries := map[string][]byte{}
	for {
		h, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(archive)
		if err != nil {
			t.Fatal(err)
		}
		entries[h.Name] = content
	}
	var manifest evidence.Manifest
	if err := json.Unmarshal(entries["manifest.json"], &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Objects) != 1 || len(entries) != 2 {
		t.Fatal("archive omitted captured scope")
	}
	object := manifest.Objects[0]
	captured := entries["objects/"+object.Hash+".json"]
	if evidence.Hash(captured) != object.Hash || object.Version == "" || object.RetainUntil == nil || !bytes.Contains(captured, []byte("networkpolicies")) {
		t.Fatal("archive did not preserve verified object identity")
	}
	var snapshot struct {
		Resource  string `json:"resource"`
		Namespace string `json:"namespace"`
		Items     []struct {
			Spec struct {
				Ingress []any `json:"ingress"`
				Egress  []any `json:"egress"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := json.Unmarshal(captured, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Namespace != "team" || snapshot.Resource != "networkpolicies" || len(snapshot.Items) != 1 || len(snapshot.Items[0].Spec.Ingress) != 0 || len(snapshot.Items[0].Spec.Egress) != 0 {
		t.Fatal("archive does not contain the repaired namespace isolation snapshot")
	}

}

type fixtureObject struct {
	data      []byte
	retention time.Time
	mode      s3types.ObjectLockMode
}
type fixtureS3 struct {
	objects map[string]fixtureObject
	corrupt bool
}

func (s *fixtureS3) GetBucketVersioning(context.Context, *s3.GetBucketVersioningInput, ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	return &s3.GetBucketVersioningOutput{Status: s3types.BucketVersioningStatusEnabled}, nil
}
func (s *fixtureS3) GetObjectLockConfiguration(context.Context, *s3.GetObjectLockConfigurationInput, ...func(*s3.Options)) (*s3.GetObjectLockConfigurationOutput, error) {
	return &s3.GetObjectLockConfigurationOutput{ObjectLockConfiguration: &s3types.ObjectLockConfiguration{ObjectLockEnabled: s3types.ObjectLockEnabledEnabled}}, nil
}
func (s *fixtureS3) PutObject(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if in.ObjectLockMode != s3types.ObjectLockModeCompliance || in.ObjectLockRetainUntilDate == nil || in.ServerSideEncryption != s3types.ServerSideEncryptionAes256 {
		return nil, fmt.Errorf("fixture requires retained encrypted writes")
	}
	data, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	version := fmt.Sprintf("fixture-%d", len(s.objects)+1)
	s.objects[aws.ToString(in.Key)+"/"+version] = fixtureObject{data: data, retention: *in.ObjectLockRetainUntilDate, mode: in.ObjectLockMode}
	return &s3.PutObjectOutput{VersionId: aws.String(version)}, nil
}
func (s *fixtureS3) GetObject(_ context.Context, in *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	object, ok := s.objects[aws.ToString(in.Key)+"/"+aws.ToString(in.VersionId)]
	if !ok {
		return nil, fmt.Errorf("fixture object version is missing")
	}
	data := object.data
	if s.corrupt {
		data = []byte("corrupted fixture bytes")
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(data)), ObjectLockMode: object.mode, ObjectLockRetainUntilDate: &object.retention}, nil
}
