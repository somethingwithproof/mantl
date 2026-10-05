// Package evaluation keeps coverage, freshness, violations and waivers distinct.
package evaluation

import (
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type Observation struct {
	Key, Result string
	At          time.Time
	Resource    string
}
type Input struct {
	Expected                   []string
	Observations               []Observation
	Unsupported                bool
	EvidenceRequired           bool
	EvidenceAt                 *time.Time
	EvidenceURI                string
	Exceptions                 []api.ComplianceException
	Profile, Control, Resource string
	Now                        time.Time
	MaxAge                     time.Duration
}

func ActiveException(exception api.ComplianceException, profile, control, resource string, now time.Time) bool {
	s := exception.Status
	return s.Approved && s.ApprovedBy != "" && s.ApprovedAt != nil && s.ApprovedGeneration == exception.Generation && s.ApprovedAt.Time.Before(now.Add(time.Second)) && exception.Spec.Owner != "" && exception.Spec.Justification != "" && exception.Spec.ExpiresAt.After(now) && exception.Spec.Profile == profile && exception.Spec.Control == control && (exception.Spec.Resource == "" || exception.Spec.Resource == resource)
}
func Assess(in Input) api.ControlEvaluationStatus {
	result := api.ControlEvaluationStatus{Result: "unknown", Coverage: "complete", Freshness: "current", EvidenceURI: in.EvidenceURI}
	if in.MaxAge <= 0 {
		in.MaxAge = 24 * time.Hour
	}
	observed := metav1.NewTime(in.Now)
	result.ObservedAt = &observed
	if in.Unsupported {
		result.Coverage = "partial"
		result.Reason = "unsupported or unresolved mapping"
	}
	state := assessExpected(in.Expected, latestObservations(in.Observations), in)
	missing, failed, errored := state.missing, state.failed, state.errored
	result.FindingCount = state.findings
	if state.stale {
		result.Freshness = "stale"
	}
	if in.EvidenceRequired && (in.EvidenceAt == nil || in.EvidenceURI == "" || in.EvidenceAt.After(in.Now.Add(time.Minute)) || in.Now.Sub(*in.EvidenceAt) > in.MaxAge) {
		missing = true
		result.Freshness = "stale"
	}
	switch {
	case failed:
		result.Result = "fail"
	case errored:
		result.Result = "error"
	case missing || in.Unsupported || len(in.Expected) == 0:
		result.Result = "unknown"
	default:
		result.Result = "pass"
	}
	for _, e := range in.Exceptions {
		if ActiveException(e, in.Profile, in.Control, in.Resource, in.Now) {
			result.ExceptionNames = append(result.ExceptionNames, e.Name)
		}
	}
	// An approved exception annotates a violation; it never changes fail to pass.
	if result.Result == "unknown" && result.Reason == "" {
		result.Reason = "current evaluation is unavailable"
	}
	return result
}

func stateRank(state string) int {
	switch state {
	case "pass":
		return 0
	case "fail", "warn":
		return 3
	case "error":
		return 2
	default:
		return 1
	}
}

type observedState struct {
	missing, failed, errored, stale bool
	findings                        int32
}

func latestObservations(observations []Observation) map[string]Observation {
	latest := map[string]Observation{}
	for _, o := range observations {
		if old, ok := latest[o.Key+"\x00"+o.Resource]; !ok || o.At.After(old.At) || o.At.Equal(old.At) && stateRank(o.Result) > stateRank(old.Result) {
			latest[o.Key+"\x00"+o.Resource] = o
		}
	}
	return latest
}

func (s *observedState) merge(next observedState) {
	s.missing = s.missing || next.missing
	s.failed = s.failed || next.failed
	s.errored = s.errored || next.errored
	s.stale = s.stale || next.stale
	s.findings += next.findings
}

func assessExpected(expected []string, latest map[string]Observation, in Input) observedState {
	var state observedState
	for _, key := range expected {
		state.merge(assessExpectedKey(key, latest, in))
	}
	return state
}

func assessExpectedKey(key string, latest map[string]Observation, in Input) observedState {
	seen := false
	var state observedState
	for _, observation := range latest {
		if observation.Key != key {
			continue
		}
		seen = true
		state.merge(assessObservation(observation, in))
	}
	if !seen {
		state.missing = true
		state.stale = true
	}
	return state
}

func assessObservation(o Observation, in Input) observedState {
	if o.At.IsZero() || o.At.After(in.Now.Add(time.Minute)) || in.Now.Sub(o.At) > in.MaxAge {
		return observedState{missing: true, stale: true}
	}
	switch o.Result {
	case "pass":
		return observedState{}
	case "fail", "warn":
		return observedState{failed: true, findings: 1}
	case "error":
		return observedState{errored: true}
	default:
		return observedState{missing: true}
	}
}
