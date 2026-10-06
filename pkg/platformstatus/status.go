// Package platformstatus interprets bounded, read-only platform observations.
// Application convergence and policy result counts are separate from certification.
package platformstatus

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/thomasvincent/mantl/pkg/compiler"
)

const (
	Healthy  = "Healthy"
	Degraded = "Degraded"
	Unknown  = "Unknown"
)

// Report is the versioned CLI observation contract. Context is the requested
// kube-context; an empty value means kubectl's current context was used.
type Report struct {
	SchemaVersion string    `json:"schemaVersion"`
	ObservedAt    time.Time `json:"observedAt"`
	Platform      string    `json:"platform"`
	Context       string    `json:"context"`
	Applications  Health    `json:"applications"`
	Policies      Policies  `json:"policies"`
}

type Health struct {
	Scope        string        `json:"scope"`
	State        string        `json:"state"`
	Reason       string        `json:"reason,omitempty"`
	Applications []Application `json:"applications"`
}

type Application struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Sync   string `json:"sync,omitempty"`
	Health string `json:"health,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type source struct {
	Repository string `json:"repoURL"`
	Revision   string `json:"targetRevision"`
	Path       string `json:"path"`
}

type destination struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

type applicationMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type applicationIdentity struct {
	Source      source      `json:"source"`
	Sources     []source    `json:"sources"`
	Destination destination `json:"destination"`
}

type syncObservation struct {
	Status     string              `json:"status"`
	ComparedTo applicationIdentity `json:"comparedTo"`
}

type healthObservation struct {
	Status string `json:"status"`
}
type applicationCondition struct {
	Type string `json:"type"`
}
type operationObservation struct {
	Phase string `json:"phase"`
}

type applicationStatus struct {
	Sync           syncObservation        `json:"sync"`
	Health         healthObservation      `json:"health"`
	Conditions     []applicationCondition `json:"conditions"`
	OperationState operationObservation   `json:"operationState"`
}

type observation struct {
	Metadata applicationMetadata `json:"metadata"`
	Spec     applicationIdentity `json:"spec"`
	Status   applicationStatus   `json:"status"`
}

// InterpretApplications requires every generated application, its configured and
// compared source identity, and explicit Synced/Healthy observations. Unrelated
// applications cannot replace missing expected ones. No raw specs are returned.
func InterpretApplications(expected []compiler.GitOpsApplication, raw []byte) Health {
	result := Health{Scope: "generated-applications", State: Unknown, Applications: []Application{}}
	var list struct {
		Items *[]observation `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil || list.Items == nil {
		result.Reason = "invalid application list"
		return result
	}
	if len(expected) == 0 {
		result.Reason = "no desired applications"
		return result
	}
	indexed := make(map[string][]observation)
	for _, app := range *list.Items {
		if app.Metadata.Namespace == "argocd" {
			indexed[app.Metadata.Name] = append(indexed[app.Metadata.Name], app)
		}
	}
	result.State = Healthy
	for _, desired := range expected {
		app := evaluateApplication(desired, indexed[desired.Name])
		result.Applications = append(result.Applications, app)
		if app.State == Unknown {
			result.State = Unknown
		} else if app.State == Degraded && result.State == Healthy {
			result.State = Degraded
		}
	}
	return result
}

func evaluateApplication(desired compiler.GitOpsApplication, observations []observation) Application {
	result := Application{Name: desired.Name, State: Unknown}
	if len(observations) == 0 {
		result.Reason = "application missing"
		return result
	}
	if len(observations) != 1 {
		result.Reason = "duplicate application observations"
		return result
	}
	app := observations[0]
	expectedSource := source{Repository: desired.Repository, Revision: desired.Revision, Path: desired.Path}
	expectedDestination := destination{Server: "https://kubernetes.default.svc", Namespace: "argocd"}
	if app.Spec.Source != expectedSource || len(app.Spec.Sources) != 0 || app.Spec.Destination != expectedDestination {
		result.Reason = "configured application identity differs from spec"
		return result
	}
	if app.Status.Sync.ComparedTo.Source != expectedSource || len(app.Status.Sync.ComparedTo.Sources) != 0 || app.Status.Sync.ComparedTo.Destination != expectedDestination {
		result.Reason = "application has not been compared against the desired identity"
		return result
	}
	result.Sync, result.Health = app.Status.Sync.Status, app.Status.Health.Status
	if result.Sync == "" || result.Health == "" || result.Sync == Unknown || result.Health == Unknown {
		result.Reason = "application status incomplete or unknown"
		return result
	}
	result.State = Degraded
	if result.Sync != "Synced" || result.Health != Healthy {
		result.Reason = "application not synced and healthy"
		return result
	}
	for _, condition := range app.Status.Conditions {
		if strings.HasSuffix(condition.Type, "Error") {
			result.Reason = "application reports an error condition"
			return result
		}
	}
	phase := app.Status.OperationState.Phase
	if phase != "" && phase != "Succeeded" {
		result.Reason = "application operation not successful"
		return result
	}
	result.State = Healthy
	return result
}

// Counts are policy engine result totals, not framework compliance coverage.
type Counts struct {
	Pass  int64 `json:"pass"`
	Fail  int64 `json:"fail"`
	Error int64 `json:"error"`
	Warn  int64 `json:"warn"`
	Skip  int64 `json:"skip"`
}

type Policies struct {
	Scope   string  `json:"scope"`
	State   string  `json:"state"`
	Reason  string  `json:"reason,omitempty"`
	Reports int     `json:"reports"`
	Counts  *Counts `json:"counts"`
}

// InterpretPolicies rejects absent, fractional, negative or overflowing counts.
// Zero results and missing reports remain unknown, never an invented pass rate.
func InterpretPolicies(raw []byte) Policies {
	result := Policies{Scope: "all-readable-policy-reports", State: Unknown}
	var list struct {
		Items *[]struct {
			Summary map[string]json.RawMessage `json:"summary"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil || list.Items == nil {
		result.Reason = "invalid policy report list"
		return result
	}
	result.Reports = len(*list.Items)
	if result.Reports == 0 {
		result.Reason = "no policy reports"
		return result
	}
	totals := map[string]int64{"pass": 0, "fail": 0, "error": 0, "warn": 0, "skip": 0}
	var total int64
	for _, report := range *list.Items {
		for _, key := range []string{"pass", "fail", "error", "warn", "skip"} {
			var count int64
			value, ok := report.Summary[key]
			if !ok || string(value) == "null" || json.Unmarshal(value, &count) != nil || count < 0 || count > math.MaxInt64-total {
				result.Reason = "invalid or overflowing policy summary"
				return result
			}
			total += count
			totals[key] += count
		}
	}
	result.Counts = &Counts{Pass: totals["pass"], Fail: totals["fail"], Error: totals["error"], Warn: totals["warn"], Skip: totals["skip"]}
	if total == 0 {
		result.Reason = "no policy results"
		return result
	}
	result.State = "Observed"
	return result
}

// Text provides a human-readable view of the same structured observation.
func (r Report) Text() string {
	var out strings.Builder
	fmt.Fprintf(&out, "Mantl platform: %s\nApplication status: %s (generated applications)\n", r.Platform, r.Applications.State)
	if r.Applications.Reason != "" {
		fmt.Fprintf(&out, "  %s\n", r.Applications.Reason)
	}
	for _, app := range r.Applications.Applications {
		fmt.Fprintf(&out, "  %s: %s", app.Name, app.State)
		if app.Reason != "" {
			fmt.Fprintf(&out, " (%s)", app.Reason)
		}
		out.WriteByte('\n')
	}
	fmt.Fprintf(&out, "Policy reports: %s (all readable reports; not certification)\n", r.Policies.State)
	if r.Policies.Counts != nil {
		counts := r.Policies.Counts
		fmt.Fprintf(&out, "  pass=%d fail=%d error=%d warn=%d skip=%d\n", counts.Pass, counts.Fail, counts.Error, counts.Warn, counts.Skip)
	}
	if r.Policies.Reason != "" {
		fmt.Fprintf(&out, "  %s\n", r.Policies.Reason)
	}
	return out.String()
}
