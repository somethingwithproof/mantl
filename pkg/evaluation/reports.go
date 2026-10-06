// SPDX-License-Identifier: Apache-2.0

package evaluation

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"time"
)

// PolicyReportObservations parses policy report metadata without assuming that
// malformed or absent results are passing evaluations.
func PolicyReportObservations(report unstructured.Unstructured) ([]Observation, error) {
	var observations []Observation
	results, found, err := unstructured.NestedSlice(report.Object, "results")
	if err != nil || !found {
		return nil, fmt.Errorf("invalid evaluation report")
	}
	for _, item := range results {
		result, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid evaluation result")
		}
		policy, _, _ := unstructured.NestedString(result, "policy")
		rule, _, _ := unstructured.NestedString(result, "rule")
		state, _, _ := unstructured.NestedString(result, "result")
		seconds, _, _ := unstructured.NestedInt64(result, "timestamp", "seconds")
		resources, _, _ := unstructured.NestedSlice(result, "resources")
		for _, item := range resources {
			resource, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid evaluation scope")
			}
			ns, _, _ := unstructured.NestedString(resource, "namespace")
			name, _, _ := unstructured.NestedString(resource, "name")
			kind, _, _ := unstructured.NestedString(resource, "kind")
			at := time.Time{}
			if seconds > 0 {
				at = time.Unix(seconds, 0).UTC()
			}
			observations = append(observations, Observation{Key: policy + "/" + ns + "/" + rule, Result: state, At: at, Resource: kind + "/" + ns + "/" + name})
		}
	}
	return observations, nil
}
