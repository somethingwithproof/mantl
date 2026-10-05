package auditplan

import (
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestInlineScopesPreserveDuplicatesAndCoverageGaps(t *testing.T) {
	collector := &framework.EvidenceCollector{Type: "config-snapshot", Resources: []string{"roles", "clusterroles", "unsupported"}}
	fw := &framework.Framework{FrameworkSpec: framework.FrameworkSpec{Controls: []framework.Control{
		{ID: "selected", Mappings: []framework.Mapping{{}, {EvidenceCollector: collector}}},
		{ID: "excluded", Mappings: []framework.Mapping{{EvidenceCollector: collector}}},
	}}}
	profile := &api.ComplianceProfile{Spec: api.ComplianceProfileSpec{IncludeControls: []string{"selected"}, Namespaces: []string{"team", "team"}}}
	plan := BuildInline(fw, profile, &api.ComplianceAudit{}, time.Now())
	if len(plan.Tasks) != 3 || len(plan.Gaps) != 1 || !plan.Incomplete["selected/1"] || !plan.Due["selected/1"] {
		t.Fatalf("invalid plan: %+v", plan)
	}
	if plan.Tasks[0].Namespace != "team" || plan.Tasks[1].Namespace != "team" || plan.Tasks[2].Namespace != "" {
		t.Fatalf("scope changed: %+v", plan.Tasks)
	}
	profile.Spec.Namespaces = nil
	profile.Namespace = "fallback"
	plan = BuildInline(fw, profile, &api.ComplianceAudit{}, time.Now())
	if len(plan.Tasks) != 2 || plan.Tasks[0].Namespace != "fallback" {
		t.Fatalf("fallback scope lost: %+v", plan)
	}
}

func TestInlineUnsupportedCollectorsNeverProduceEvidence(t *testing.T) {
	for _, collector := range []framework.EvidenceCollector{
		{Type: "log-query", Resources: []string{"roles"}},
		{Type: "config-snapshot", Query: "untrusted", Resources: []string{"roles"}},
		{Type: "config-snapshot"},
	} {
		fw := &framework.Framework{FrameworkSpec: framework.FrameworkSpec{Controls: []framework.Control{{ID: "C", Mappings: []framework.Mapping{{EvidenceCollector: &collector}}}}}}
		plan := BuildInline(fw, &api.ComplianceProfile{}, &api.ComplianceAudit{}, time.Now())
		if len(plan.Tasks) != 0 || len(plan.Gaps) != 1 || !plan.Due["C/0"] {
			t.Fatalf("unsupported collector inferred evidence: %+v", plan)
		}
	}
}

func TestInlineSchedulingDefersAndCoalescesMissedIntervals(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	collector := &framework.EvidenceCollector{Type: "config-snapshot", Schedule: "0 * * * *", Resources: []string{"roles"}}
	fw := &framework.Framework{FrameworkSpec: framework.FrameworkSpec{Controls: []framework.Control{{ID: "C", Mappings: []framework.Mapping{{EvidenceCollector: collector}}}}}}
	audit := &api.ComplianceAudit{Spec: api.ComplianceAuditSpec{Frequency: "framework"}, Status: api.ComplianceAuditStatus{CollectorTimes: map[string]metav1.Time{"C/0": metav1.NewTime(now)}}}
	plan := BuildInline(fw, &api.ComplianceProfile{}, audit, now)
	if len(plan.Tasks) != 0 || plan.Next != time.Hour {
		t.Fatalf("early collection: %+v", plan)
	}
	audit.Status.CollectorTimes["C/0"] = metav1.NewTime(now.Add(-3 * time.Hour))
	plan = BuildInline(fw, &api.ComplianceProfile{}, audit, now)
	if len(plan.Tasks) != 1 || plan.Next != time.Hour {
		t.Fatalf("missed interval not coalesced: %+v", plan)
	}
	audit.Status.CollectorTimes = nil
	plan = BuildInline(fw, &api.ComplianceProfile{}, audit, now)
	if len(plan.Tasks) != 1 {
		t.Fatalf("first observation deferred: %+v", plan)
	}
	collector.Schedule = "invalid"
	plan = BuildInline(fw, &api.ComplianceProfile{}, audit, now)
	if len(plan.Tasks) != 0 || len(plan.Gaps) != 1 {
		t.Fatalf("invalid cron passed: %+v", plan)
	}
	plan.waitInline(10 * time.Minute)
	plan.waitInline(time.Hour)
	if plan.Next != 10*time.Minute {
		t.Fatalf("lost earliest retry: %+v", plan)
	}
}
