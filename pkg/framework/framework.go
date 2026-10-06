// SPDX-License-Identifier: Apache-2.0

package framework

// Framework mirrors the on-disk compliance framework YAML
// (compliance/frameworks/<name>/framework.yaml). The file is a Kubernetes-style
// object whose controls live under spec, so FrameworkSpec is embedded with the
// "spec" tag: the document nests under spec while Go access stays flat
// (framework.Controls).
type Framework struct {
	FrameworkSpec `json:"spec"`
	ContentDigest string `json:"-"`
}

// FrameworkSpec holds the framework body: controls and their mappings.
type FrameworkSpec struct {
	Name     string    `json:"displayName"`
	Version  string    `json:"version"`
	Controls []Control `json:"controls"`
}

// Control is a single framework control with its policy and evidence mappings.
type Control struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Severity string    `json:"severity"`
	Mappings []Mapping `json:"mappings"`
}

// Mapping is one entry under a control: either a policy reference or an
// evidence collector. Type is "policy" or "evidence".
type Mapping struct {
	Type              string             `json:"type"`
	PolicyRef         *PolicyRef         `json:"policyRef,omitempty"`
	EvidenceCollector *EvidenceCollector `json:"evidenceCollector,omitempty"`
}

// PolicyRef names a Kyverno policy template and its enforcement mode.
type PolicyRef struct {
	Template string `json:"template"`
	Mode     string `json:"mode"`
}

// EvidenceCollector describes periodic evidence capture for a control.
// Collector configuration consumed by audit planning.
type EvidenceCollector struct {
	Type          string   `json:"type"`
	Schedule      string   `json:"schedule"`
	Query         string   `json:"query,omitempty"`
	Resources     []string `json:"resources,omitempty"`
	RetentionDays int      `json:"retentionDays,omitempty"`
}
