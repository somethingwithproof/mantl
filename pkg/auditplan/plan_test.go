// SPDX-License-Identifier: Apache-2.0

package auditplan

import (
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestSchedulesScopesAndUnsupportedCollectors(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	fw := &framework.Framework{FrameworkSpec: framework.FrameworkSpec{Controls: []framework.Control{{ID: "C1", Mappings: []framework.Mapping{{EvidenceCollector: &framework.EvidenceCollector{Type: "config-snapshot", Schedule: "0 * * * *", Resources: []string{"roles", "secrets"}}}}}, {ID: "C2", Mappings: []framework.Mapping{{EvidenceCollector: &framework.EvidenceCollector{Type: "log-query", Schedule: "0 * * * *"}}}}}}}
	profile := &api.ComplianceProfile{Spec: api.ComplianceProfileSpec{Namespaces: []string{"team"}}}
	audit := &api.ComplianceAudit{Spec: api.ComplianceAuditSpec{Frequency: "framework"}}
	first := Build(fw, profile, audit, now)
	again := Build(fw, profile, audit, now)
	if len(first.Tasks) != 1 || len(first.Gaps) != 2 || first.Tasks[0].Namespace != "team" || first.Tasks[0].ID != again.Tasks[0].ID {
		t.Fatalf("invalid plan: %+v", first)
	}
	profile.Spec.IncludeControls = []string{"C2"}
	selected := Build(fw, profile, audit, now)
	if len(selected.Tasks) != 0 || len(selected.Gaps) != 1 {
		t.Fatal("selection ignored")
	}
}

func TestPlannerDefersScheduledCollectorsAndRejectsInvalidSchedules(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	collector := &framework.EvidenceCollector{Type: "config-snapshot", Schedule: "0 * * * *", Resources: []string{"roles"}}
	fw := &framework.Framework{FrameworkSpec: framework.FrameworkSpec{Controls: []framework.Control{{ID: "C1", Mappings: []framework.Mapping{{EvidenceCollector: collector}, {}}}}}}
	profile := &api.ComplianceProfile{}
	profile.Namespace = "team"
	audit := &api.ComplianceAudit{Spec: api.ComplianceAuditSpec{Frequency: "framework"}, Status: api.ComplianceAuditStatus{CollectorTimes: map[string]metav1.Time{"C1/0": metav1.NewTime(now)}}}
	deferred := Build(fw, profile, audit, now)
	if len(deferred.Tasks) != 0 || deferred.Next != time.Hour {
		t.Fatalf("collector ran before scheduled time: %+v", deferred)
	}
	collector.Schedule = "invalid"
	invalid := Build(fw, profile, audit, now)
	if len(invalid.Tasks) != 0 || len(invalid.Gaps) != 1 {
		t.Fatalf("invalid schedule was not reported as a gap: %+v", invalid)
	}
}

func TestPlannerDeduplicatesNamespacesAndKeepsClusterScope(t *testing.T) {
	collector := &framework.EvidenceCollector{Type: "config-snapshot", Resources: []string{"roles", "clusterroles"}}
	fw := &framework.Framework{FrameworkSpec: framework.FrameworkSpec{Controls: []framework.Control{{ID: "C1", Mappings: []framework.Mapping{{EvidenceCollector: collector}}}}}}
	profile := &api.ComplianceProfile{Spec: api.ComplianceProfileSpec{Namespaces: []string{"team", "team"}}}
	plan := Build(fw, profile, &api.ComplianceAudit{}, time.Now().UTC())
	if len(plan.Tasks) != 2 || len(plan.Gaps) != 0 {
		t.Fatalf("invalid deduplicated scopes: %+v", plan)
	}
	if plan.Tasks[0].Namespace != "team" || plan.Tasks[1].Namespace != "" {
		t.Fatalf("cluster scope was replaced with tenant scope: %+v", plan.Tasks)
	}
}
