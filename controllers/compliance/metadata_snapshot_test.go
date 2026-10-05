package compliance

import (
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestAuditMetricsDistinguishFailuresAndMissingEvidence(t *testing.T) {
	if gaps, failed, age := auditMetrics(nil); gaps != 0 || failed != 0 || age != -1 {
		t.Fatalf("missing evidence inferred fresh: %d %d %f", gaps, failed, age)
	}
	older := metav1.NewTime(time.Now().Add(-2 * time.Hour))
	newer := metav1.NewTime(time.Now().Add(-time.Hour))
	audits := []api.ComplianceAudit{{Status: api.ComplianceAuditStatus{Phase: "Failed", CoverageGaps: []string{"a"}}}, {Status: api.ComplianceAuditStatus{Phase: "PartiallyCompleted", CoverageGaps: []string{"b", "c"}, EvidenceURI: "s3://bucket/old", EndTime: &older}}, {Status: api.ComplianceAuditStatus{Phase: "Completed", EvidenceURI: "s3://bucket/new", EndTime: &newer}}}
	gaps, failed, age := auditMetrics(audits)
	if gaps != 3 || failed != 2 || age < 7200 || age > 7210 {
		t.Fatalf("incorrect aggregate: %d %d %f", gaps, failed, age)
	}
}

func TestFleetSnapshotRequiresObservationIdentityAndKeepsVersionsDistinct(t *testing.T) {
	now := metav1.NewTime(time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC))
	meta := metav1.ObjectMeta{Name: "object", Namespace: "team", UID: "uid", ResourceVersion: "1"}
	run := api.AuditRun{ObjectMeta: meta, Spec: api.AuditRunSpec{Profile: "profile"}, Status: api.AuditRunStatus{EndTime: &now, Phase: "Completed", EvidenceURI: "s3://bucket/manifest", ManifestHash: "digest"}}
	evaluation := api.ControlEvaluation{ObjectMeta: meta, Spec: api.ControlEvaluationSpec{Profile: "profile", Control: "C"}, Status: api.ControlEvaluationStatus{ObservedAt: &now, Result: "fail", Coverage: "complete", Freshness: "current", EvidenceURI: "s3://bucket/manifest"}}
	finding := api.Finding{ObjectMeta: meta, Spec: api.FindingSpec{ControlID: "C", Status: "fail"}}
	finding.Annotations = map[string]string{observedAtAnnotation: now.Format(time.RFC3339Nano)}
	events := snapshotFleetEvents([]api.AuditRun{{}, run}, []api.ControlEvaluation{{}, evaluation}, []api.Finding{{}, finding})
	if len(events) != 3 {
		t.Fatalf("missing observation included or valid lost: %+v", events)
	}
	if events[0].ManifestHash != "digest" || events[0].Profile != "profile" || events[1].Control != "C" || events[1].Coverage != "complete" || events[2].Result != "fail" {
		t.Fatalf("metadata lost: %+v", events)
	}
	for i, e := range events {
		if e.ResourceUID != "uid" || e.Namespace != "team" || !e.ObservedAt.Equal(now.Time) {
			t.Fatalf("identity lost: %+v", e)
		}
		for j := 0; j < i; j++ {
			if e.ID == events[j].ID {
				t.Fatal("event kinds share identity")
			}
		}
	}
	run.ResourceVersion = "2"
	next := snapshotFleetEvents([]api.AuditRun{run}, nil, nil)
	if next[0].ID == events[0].ID {
		t.Fatal("resource version update cannot be published")
	}
}
