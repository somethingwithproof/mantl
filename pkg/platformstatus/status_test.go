package platformstatus

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/thomasvincent/mantl/pkg/compiler"
)

var desired = compiler.GitOpsApplication{Name: "root-platform", Repository: "https://example.com/mantl.git", Revision: "v1.0.0", Path: "deploy/gitops/core"}

func healthyObservation() observation {
	var app observation
	app.Metadata.Name = desired.Name
	app.Metadata.Namespace = "argocd"
	app.Spec.Source = source{desired.Repository, desired.Revision, desired.Path}
	app.Spec.Destination = destination{"https://kubernetes.default.svc", "argocd"}
	app.Status.Sync.ComparedTo.Source = app.Spec.Source
	app.Status.Sync.ComparedTo.Destination = app.Spec.Destination
	app.Status.Sync.Status = "Synced"
	app.Status.Health.Status = Healthy
	return app
}

func applicationList(t *testing.T, apps ...observation) []byte {
	t.Helper()
	if apps == nil {
		apps = []observation{}
	}
	data, err := json.Marshal(map[string]any{"items": apps})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestApplicationScopeAndIdentity(t *testing.T) {
	cases := []struct {
		name, state string
		mutate      func(*observation)
	}{
		{"healthy", Healthy, func(*observation) {}},
		{"unrelated application", Unknown, func(a *observation) { a.Metadata.Name = "other" }},
		{"different namespace", Unknown, func(a *observation) { a.Metadata.Namespace = "tenant" }},
		{"old configured revision", Unknown, func(a *observation) { a.Spec.Source.Revision = "main" }},
		{"other repository", Unknown, func(a *observation) { a.Spec.Source.Repository = "https://other.example/repo" }},
		{"other path", Unknown, func(a *observation) { a.Spec.Source.Path = "other" }},
		{"external destination", Unknown, func(a *observation) { a.Spec.Destination.Server = "https://other.cluster" }},
		{"multi source", Unknown, func(a *observation) { a.Spec.Sources = []source{a.Spec.Source} }},
		{"old compared revision", Unknown, func(a *observation) { a.Status.Sync.ComparedTo.Source.Revision = "old" }},
		{"missing compared identity", Unknown, func(a *observation) { a.Status.Sync.ComparedTo.Source = source{} }},
		{"old compared destination", Unknown, func(a *observation) { a.Status.Sync.ComparedTo.Destination.Namespace = "other" }},
		{"multiple compared sources", Unknown, func(a *observation) { a.Status.Sync.ComparedTo.Sources = []source{a.Spec.Source} }},
		{"missing health", Unknown, func(a *observation) { a.Status.Health.Status = "" }},
		{"unknown sync", Unknown, func(a *observation) { a.Status.Sync.Status = Unknown }},
		{"out of sync", Degraded, func(a *observation) { a.Status.Sync.Status = "OutOfSync" }},
		{"unhealthy", Degraded, func(a *observation) { a.Status.Health.Status = "Progressing" }},
		{"error condition", Degraded, func(a *observation) {
			a.Status.Conditions = append(a.Status.Conditions, applicationCondition{Type: "ComparisonError"})
		}},
		{"successful operation", Healthy, func(a *observation) { a.Status.OperationState.Phase = "Succeeded" }},
		{"unknown health", Unknown, func(a *observation) { a.Status.Health.Status = Unknown }},
		{"running operation", Degraded, func(a *observation) { a.Status.OperationState.Phase = "Running" }},
		{"failed operation", Degraded, func(a *observation) { a.Status.OperationState.Phase = "Failed" }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app := healthyObservation()
			test.mutate(&app)
			result := InterpretApplications([]compiler.GitOpsApplication{desired}, applicationList(t, app))
			if result.State != test.state || len(result.Applications) != 1 {
				t.Fatalf("result=%+v, want %s", result, test.state)
			}
		})
	}
	app := healthyObservation()
	if result := InterpretApplications([]compiler.GitOpsApplication{desired}, applicationList(t, app, app)); result.State != Unknown {
		t.Fatal("duplicate accepted", result)
	}
	other := app
	other.Metadata.Name = "unrelated"
	other.Status.Health.Status = Degraded
	if result := InterpretApplications([]compiler.GitOpsApplication{desired}, applicationList(t, app, other)); result.State != Healthy {
		t.Fatal("unrelated app affected desired topology", result)
	}
	for _, raw := range []string{"", `{}`, `{"items":null}`, `{"items":[]}`, `{"items":[null]}`, `{"items":[]} trailing`} {
		if result := InterpretApplications([]compiler.GitOpsApplication{desired}, []byte(raw)); result.State != Unknown {
			t.Fatalf("accepted invalid/missing observations: %s", raw)
		}
	}
	if result := InterpretApplications(nil, applicationList(t, app)); result.State != Unknown {
		t.Fatal("empty desired topology accepted")
	}
}

func TestUnknownCannotBeOverriddenByDegraded(t *testing.T) {
	missing := desired
	missing.Name = "missing"
	app := healthyObservation()
	app.Status.Sync.Status = "OutOfSync"
	for _, expected := range [][]compiler.GitOpsApplication{{missing, desired}, {desired, missing}} {
		if result := InterpretApplications(expected, applicationList(t, app)); result.State != Unknown {
			t.Fatal(result)
		}
	}
}

func TestPolicySummaryCountsAndBoundaries(t *testing.T) {
	valid := `{"items":[{"summary":{"pass":8,"fail":2,"error":1,"warn":3,"skip":4}},{"summary":{"pass":2,"fail":0,"error":0,"warn":0,"skip":0}}]}`
	result := InterpretPolicies([]byte(valid))
	if result.State != "Observed" || result.Reports != 2 || result.Counts == nil || *result.Counts != (Counts{10, 2, 1, 3, 4}) {
		t.Fatal(result)
	}
	for _, raw := range []string{
		"", `{}`, `{"items":null}`, `{"items":[]}`, `{"items":[{}]}`,
		`{"items":[{"summary":{"pass":0,"fail":0,"error":0,"warn":0,"skip":0}}]}`,
		strings.Replace(valid, `"pass":8`, `"pass":-1`, 1),
		strings.Replace(valid, `"pass":8`, `"pass":0.5`, 1),
		strings.Replace(valid, `"pass":8`, `"pass":null`, 1),
		strings.Replace(valid, `"pass":8`, `"pass":"8"`, 1),
		strings.Replace(valid, `"pass":8`, `"pass":9223372036854775807`, 1),
		strings.Replace(valid, `"pass":8`, `"pass":9223372036854775808`, 1),
	} {
		result := InterpretPolicies([]byte(raw))
		if result.State != Unknown {
			t.Fatalf("accepted invalid/empty summary %s: %+v", raw, result)
		}
	}
	boundary := []byte(`{"items":[{"summary":{"pass":9223372036854775807,"fail":0,"error":0,"warn":0,"skip":0}}]}`)
	if result := InterpretPolicies(boundary); result.Counts == nil || result.Counts.Pass != math.MaxInt64 {
		t.Fatal(result)
	}
}

func TestTextDoesNotClaimCompliance(t *testing.T) {
	text := (Report{Platform: "demo", Applications: Health{State: Unknown, Reason: "missing"}, Policies: Policies{State: "Observed", Counts: &Counts{Pass: 1, Fail: 1, Error: 2, Warn: 3, Skip: 4}, Reason: "fixture"}}).Text()
	if !strings.Contains(text, "pass=1 fail=1 error=2 warn=3 skip=4") || strings.Contains(text, "% compliant") {
		t.Fatal(text)
	}
}
