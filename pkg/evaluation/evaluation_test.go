// SPDX-License-Identifier: Apache-2.0

package evaluation

import (
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestFailClosedAssessment(t *testing.T) {
	now := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name        string
		obs         []Observation
		unsupported bool
		want        string
	}{
		{"missing", nil, false, "unknown"},
		{"fresh pass", []Observation{{"policy/team", "pass", now, ""}}, false, "pass"},
		{"stale pass", []Observation{{"policy/team", "pass", now.Add(-48 * time.Hour), ""}}, false, "unknown"},
		{"collector gap", []Observation{{"policy/team", "pass", now, ""}}, true, "unknown"},
		{"failure", []Observation{{"policy/team", "fail", now, ""}}, false, "fail"},
		{"error", []Observation{{"policy/team", "error", now, ""}}, false, "error"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := Assess(Input{Expected: []string{"policy/team"}, Observations: test.obs, Unsupported: test.unsupported, Now: now})
			if got.Result != test.want {
				t.Fatalf("got %s want %s", got.Result, test.want)
			}
		})
	}
}
func TestExceptionExpiryAndApprovalRevision(t *testing.T) {
	now := time.Now().UTC()
	approved := metav1.NewTime(now.Add(-time.Hour))
	e := api.ComplianceException{ObjectMeta: metav1.ObjectMeta{Name: "waiver", Generation: 2}, Spec: api.ComplianceExceptionSpec{Profile: "profile", Control: "CC6.1", Owner: "owner", Justification: "reviewed", ExpiresAt: metav1.NewTime(now.Add(time.Hour))}, Status: api.ComplianceExceptionStatus{Approved: true, ApprovedBy: "approver", ApprovedAt: &approved, ApprovedGeneration: 2}}
	in := Input{Expected: []string{"policy"}, Observations: []Observation{{"policy", "fail", now, ""}}, Now: now, Profile: "profile", Control: "CC6.1", Exceptions: []api.ComplianceException{e}}
	got := Assess(in)
	if got.Result != "fail" || len(got.ExceptionNames) != 1 {
		t.Fatal("waiver hid violation or was not recorded")
	}
	e.Generation++
	if ActiveException(e, "profile", "CC6.1", "", now) {
		t.Fatal("changed scope retained approval")
	}
	e.Generation = 2
	e.Spec.ExpiresAt = metav1.NewTime(now)
	if ActiveException(e, "profile", "CC6.1", "", now) {
		t.Fatal("expired waiver accepted")
	}
}

func TestEveryResourceAndRuleMustBeEvaluated(t *testing.T) {
	now := time.Now().UTC()
	in := Input{Expected: []string{"policy/ns/rule1", "policy/ns/rule2"}, Now: now,
		Observations: []Observation{{"policy/ns/rule1", "fail", now, "Pod/ns/a"}, {"policy/ns/rule1", "pass", now, "Pod/ns/a"}, {"policy/ns/rule1", "pass", now, "Pod/ns/b"}, {"policy/ns/rule2", "pass", now, "Pod/ns/a"}}}
	if got := Assess(in); got.Result != "fail" || got.FindingCount != 1 {
		t.Fatal("a pass masked a failure at the same timestamp", got)
	}
	in.Observations = in.Observations[1:]
	if got := Assess(in); got.Result != "pass" {
		t.Fatal(got)
	}
	in.Observations = in.Observations[:2]
	if got := Assess(in); got.Result != "unknown" {
		t.Fatal("missing rule was treated as passing", got)
	}
}
