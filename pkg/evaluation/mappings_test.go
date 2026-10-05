package evaluation

import (
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/auditplan"
	"github.com/thomasvincent/mantl/pkg/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestMappingExpectationsRejectUnresolvedAndMalformedPolicies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.yaml")
	writeTestPolicy(t, path, "metadata:\n  name: actual\nspec:\n  rules:\n  - name: rule\n")
	policy := framework.Mapping{PolicyRef: &framework.PolicyRef{Template: "template"}, EvidenceCollector: &framework.EvidenceCollector{Type: "config-snapshot"}}
	control := framework.Control{ID: "C", Mappings: []framework.Mapping{policy}}
	index := map[string]string{"actual": path}
	aliases := map[string]string{"template": "actual"}
	expected, unsupported, required, err := MappingExpectations(control, []string{"team", "other"}, index, aliases)
	if err != nil || unsupported || !required || !reflect.DeepEqual(expected, []string{"actual/team/rule", "actual/other/rule"}) {
		t.Fatalf("valid policy scope: %v %v %v %v", expected, unsupported, required, err)
	}
	for _, body := range []string{"[broken", "metadata: {}", "metadata:\n  name: actual\nspec:\n  rules: []"} {
		writeTestPolicy(t, path, body)
		expected, unsupported, required, err = MappingExpectations(control, []string{"team"}, index, aliases)
		if err != nil || !unsupported || required || len(expected) != 0 {
			t.Fatalf("malformed policy inferred coverage: %v %v %v %v", expected, unsupported, required, err)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err = MappingExpectations(control, []string{"team"}, index, aliases); err == nil {
		t.Fatal("read failure became passing mapping")
	}
	for _, mappings := range [][]framework.Mapping{nil, {{}}, {policy}} {
		_, unsupported, required, err = MappingExpectations(framework.Control{Mappings: mappings}, nil, nil, nil)
		if err != nil || !unsupported || required {
			t.Fatalf("unresolved mapping state: %v %v %v", unsupported, required, err)
		}
	}
	_, unsupported, required, err = MappingExpectations(framework.Control{Mappings: []framework.Mapping{{EvidenceCollector: &framework.EvidenceCollector{Type: "config-snapshot"}}}}, nil, nil, nil)
	if err != nil || unsupported || !required {
		t.Fatalf("evidence-only mapping: %v %v %v", unsupported, required, err)
	}
}

func TestCollectionEvidenceUsesMatchingRunAndOldestCapture(t *testing.T) {
	now := metav1.NewTime(time.Now().UTC())
	older := metav1.NewTime(now.Add(-time.Hour))
	tasks := []api.CollectorTask{{ID: "a", Control: "C"}, {ID: "b", Control: "C"}, {ID: "other", Control: "other"}}
	valid := api.AuditRun{Spec: api.AuditRunSpec{Profile: "profile", Framework: "soc2", BundleDigest: "digest"}, Status: api.AuditRunStatus{EndTime: &now, EvidenceURI: "s3://bucket/manifest", Results: map[string]api.TaskResult{"a": {Phase: "Completed", CapturedAt: &now}, "b": {Phase: "Completed", CapturedAt: &older}}}}
	input := Input{Control: "C"}
	AttachCollectionEvidence(&input, auditplan.Plan{Tasks: tasks, Gaps: []string{"other: unsupported"}}, []api.AuditRun{valid}, "profile", "soc2", "digest")
	if input.Unsupported || input.EvidenceAt == nil || !input.EvidenceAt.Equal(older.Time) || input.EvidenceURI != valid.Status.EvidenceURI {
		t.Fatalf("wrong capture selection: %+v", input)
	}
	for _, mutate := range []func(*api.AuditRun){
		func(r *api.AuditRun) { r.Spec.Profile = "foreign" }, func(r *api.AuditRun) { r.Spec.Framework = "foreign" }, func(r *api.AuditRun) { r.Spec.BundleDigest = "foreign" }, func(r *api.AuditRun) { r.Status.EndTime = nil }, func(r *api.AuditRun) { r.Status.Results = nil }, func(r *api.AuditRun) { r.Status.EvidenceURI = "" }, func(r *api.AuditRun) { r.Status.Results["a"] = api.TaskResult{Phase: "Failed"} }, func(r *api.AuditRun) { r.Status.Results["a"] = api.TaskResult{Phase: "Completed"} },
	} {
		run := valid.DeepCopy()
		mutate(run)
		input = Input{Control: "C"}
		AttachCollectionEvidence(&input, auditplan.Plan{Tasks: tasks}, []api.AuditRun{*run}, "profile", "soc2", "digest")
		if !input.Unsupported {
			t.Fatalf("untrusted evidence passed: %+v", input)
		}
	}
	failed := valid.DeepCopy()
	failed.Status.Results["a"] = api.TaskResult{Phase: "Failed"}
	input = Input{Control: "C"}
	AttachCollectionEvidence(&input, auditplan.Plan{Tasks: tasks}, []api.AuditRun{*failed, valid}, "profile", "soc2", "digest")
	if !input.Unsupported {
		t.Fatal("newer failed receipt fell back to older success")
	}
	input = Input{Control: "C"}
	AttachCollectionEvidence(&input, auditplan.Plan{Gaps: []string{"C: unsupported"}}, nil, "profile", "soc2", "digest")
	if !input.Unsupported {
		t.Fatal("coverage gap lost")
	}
}

func TestPolicyReportParsingKeepsMissingTimeAndRejectsMalformedScopes(t *testing.T) {
	result := map[string]interface{}{"policy": "p", "rule": "r", "result": "fail", "timestamp": map[string]interface{}{"seconds": int64(123)}, "resources": []interface{}{map[string]interface{}{"namespace": "team", "name": "object", "kind": "Pod"}}}
	report := unstructured.Unstructured{Object: map[string]interface{}{"results": []interface{}{result}}}
	got, err := PolicyReportObservations(report)
	if err != nil || len(got) != 1 || got[0].Key != "p/team/r" || got[0].Resource != "Pod/team/object" || got[0].Result != "fail" || !got[0].At.Equal(time.Unix(123, 0)) {
		t.Fatalf("invalid observation: %+v %v", got, err)
	}
	delete(result, "timestamp")
	got, err = PolicyReportObservations(report)
	if err != nil || !got[0].At.IsZero() {
		t.Fatal("missing timestamp inferred current evidence")
	}
	for _, object := range []map[string]interface{}{{}, {"results": "bad"}, {"results": []interface{}{"bad"}}, {"results": []interface{}{map[string]interface{}{"resources": []interface{}{"bad"}}}}} {
		if _, err := PolicyReportObservations(unstructured.Unstructured{Object: object}); err == nil {
			t.Fatalf("malformed report passed: %+v", object)
		}
	}
}

func writeTestPolicy(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceURITracksOldestCaptureAcrossDifferentRuns(t *testing.T) {
	newer := metav1.NewTime(time.Now().UTC())
	older := metav1.NewTime(newer.Add(-time.Hour))
	run := func(id, uri string, at metav1.Time) api.AuditRun {
		return api.AuditRun{Spec: api.AuditRunSpec{Profile: "p", Framework: "f", BundleDigest: "d"}, Status: api.AuditRunStatus{EndTime: &newer, EvidenceURI: uri, Results: map[string]api.TaskResult{id: {Phase: "Completed", CapturedAt: &at}}}}
	}
	runs := []api.AuditRun{run("new", "s3://bucket/new", newer), run("old", "s3://bucket/old", older)}
	for _, tasks := range [][]api.CollectorTask{{{ID: "old", Control: "C"}, {ID: "new", Control: "C"}}, {{ID: "new", Control: "C"}, {ID: "old", Control: "C"}}} {
		input := Input{Control: "C"}
		AttachCollectionEvidence(&input, auditplan.Plan{Tasks: tasks}, runs, "p", "f", "d")
		if input.Unsupported || input.EvidenceAt == nil || !input.EvidenceAt.Equal(older.Time) || input.EvidenceURI != "s3://bucket/old" {
			t.Fatalf("freshness and manifest disagree: %+v", input)
		}
	}
}
