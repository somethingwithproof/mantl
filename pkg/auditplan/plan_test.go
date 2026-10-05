package auditplan

import (
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/framework"
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
