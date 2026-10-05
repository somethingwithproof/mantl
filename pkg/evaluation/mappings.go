package evaluation

import (
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/auditplan"
	"github.com/thomasvincent/mantl/pkg/framework"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
)

type mappingExpectation struct {
	expected                      []string
	unsupported, evidenceRequired bool
}

// MappingExpectations does not execute policies or infer passing observations.
func MappingExpectations(control framework.Control, namespaces []string, index, aliases map[string]string) ([]string, bool, bool, error) {
	state := mappingExpectation{unsupported: len(control.Mappings) == 0}
	for _, mapping := range control.Mappings {
		next, err := expectationsForMapping(mapping, namespaces, index, aliases)
		if err != nil {
			return nil, false, false, err
		}
		state.expected = append(state.expected, next.expected...)
		state.unsupported = state.unsupported || next.unsupported
		state.evidenceRequired = state.evidenceRequired || next.evidenceRequired
	}
	return state.expected, state.unsupported, state.evidenceRequired, nil
}

func expectationsForMapping(mapping framework.Mapping, namespaces []string, index, aliases map[string]string) (mappingExpectation, error) {
	state := mappingExpectation{unsupported: mapping.PolicyRef == nil && mapping.EvidenceCollector == nil}
	if mapping.PolicyRef != nil {
		name := mapping.PolicyRef.Template
		if alias, ok := aliases[name]; ok {
			name = alias
		}
		file, ok := index[name]
		if !ok {
			return mappingExpectation{unsupported: true}, nil
		}
		expected, unsupported, err := policyExpectations(file, namespaces)
		if err != nil {
			return mappingExpectation{}, err
		}
		if unsupported {
			return mappingExpectation{unsupported: true}, nil
		}
		state.expected = expected
	}
	state.evidenceRequired = mapping.EvidenceCollector != nil
	return state, nil
}

func policyExpectations(file string, namespaces []string) ([]string, bool, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, false, err
	}
	var policy struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			Rules []struct {
				Name string `json:"name"`
			} `json:"rules"`
		} `json:"spec"`
	}
	if yaml.Unmarshal(data, &policy) != nil || policy.Metadata.Name == "" || len(policy.Spec.Rules) == 0 {
		return nil, true, nil
	}
	var expected []string
	for _, namespace := range namespaces {
		for _, rule := range policy.Spec.Rules {
			expected = append(expected, policy.Metadata.Name+"/"+namespace+"/"+rule.Name)
		}
	}
	return expected, false, nil
}

// AttachCollectionEvidence selects the first matching completed run for each
// immutable task, preserving unknown state for failed or unavailable evidence.
func AttachCollectionEvidence(input *Input, plan auditplan.Plan, runs []api.AuditRun, profileName, frameworkName, bundleDigest string) {
	for _, gap := range plan.Gaps {
		if strings.HasPrefix(gap, input.Control+":") {
			input.Unsupported = true
		}
	}
	for _, task := range plan.Tasks {
		if task.Control != input.Control {
			continue
		}
		result, uri, found := selectedTaskEvidence(task, runs, profileName, frameworkName, bundleDigest)
		if !found || result.Phase != "Completed" || result.CapturedAt == nil || uri == "" {
			input.Unsupported = true
			continue
		}
		at := result.CapturedAt.Time
		if input.EvidenceAt == nil || at.Before(*input.EvidenceAt) {
			input.EvidenceAt = &at
		}
		input.EvidenceURI = uri
	}
}

func selectedTaskEvidence(task api.CollectorTask, runs []api.AuditRun, profileName, frameworkName, bundleDigest string) (api.TaskResult, string, bool) {
	for _, run := range runs {
		if run.Spec.Profile != profileName || run.Spec.Framework != frameworkName || run.Spec.BundleDigest != bundleDigest || run.Status.EndTime == nil {
			continue
		}
		result, found := run.Status.Results[task.ID]
		if found {
			return result, run.Status.EvidenceURI, true
		}
	}
	return api.TaskResult{}, "", false
}
