package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComplianceProfileSpec defines the desired state of ComplianceProfile
type ComplianceProfileSpec struct {
	Framework string `json:"framework"`
	Version   string `json:"version"`
	// BundleDigest pins a verified control-content manifest, independent of framework version.
	// +kubebuilder:validation:Pattern="^sha256:[a-f0-9]{64}$"
	BundleDigest string `json:"bundleDigest,omitempty"`
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:Maximum=31536000
	MaxEvidenceAgeSeconds int64 `json:"maxEvidenceAgeSeconds,omitempty"`
	// Namespaces restrict evidence collection. Defaults to the profile namespace.
	Namespaces      []string `json:"namespaces,omitempty"`
	IncludeControls []string `json:"includeControls,omitempty"`
}

// ComplianceProfileStatus defines the observed state of ComplianceProfile
type ComplianceProfileStatus struct {
	State          string `json:"state,omitempty"`
	ActivePolicies int32  `json:"activePolicies,omitempty"`

	// UnresolvedTemplates lists framework templates with no matching policy.
	// +optional
	UnresolvedTemplates []string `json:"unresolvedTemplates,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ComplianceProfile is the Schema for the complianceprofiles API
type ComplianceProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComplianceProfileSpec   `json:"spec,omitempty"`
	Status ComplianceProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ComplianceProfileList contains a list of ComplianceProfile
type ComplianceProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComplianceProfile `json:"items"`
}

// FindingSpec defines the details of a compliance finding
type FindingSpec struct {
	ID        string `json:"id"`
	ControlID string `json:"controlId"`
	Framework string `json:"framework"`
	Severity  string `json:"severity"`
	Resource  string `json:"resource"`
	Message   string `json:"message"`
	Status    string `json:"status"` // pass, fail, warn
}

type FindingTransition struct {
	ID        string      `json:"id"`
	State     string      `json:"state"`
	At        metav1.Time `json:"at"`
	Control   string      `json:"control"`
	Framework string      `json:"framework"`
	Resource  string      `json:"resource"`
}
type HistoryReference struct {
	URI     string `json:"uri"`
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
}
type FindingStatus struct {
	LastQueuedState string            `json:"lastQueuedState,omitempty"`
	HistoryHead     *HistoryReference `json:"historyHead,omitempty"`
	// +kubebuilder:validation:MaxItems=64
	Pending    []FindingTransition `json:"pending,omitempty"`
	HistoryGap bool                `json:"historyGap,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Finding is the Schema for the findings API
type Finding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FindingSpec   `json:"spec,omitempty"`
	Status FindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FindingList contains a list of Finding
type FindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Finding `json:"items"`
}

// ComplianceAuditSpec defines the desired state of a Compliance Audit
type ComplianceAuditSpec struct {
	Profile string `json:"profile"`
	// +kubebuilder:validation:Enum=manual;daily;weekly;framework
	Frequency string `json:"frequency,omitempty"` // manual, daily, weekly, framework
}

// ComplianceAuditStatus defines the observed state of a Compliance Audit
type ComplianceAuditStatus struct {
	ActiveRun      string                 `json:"activeRun,omitempty"`
	Phase          string                 `json:"phase,omitempty"` // Pending, Running, Completed, PartiallyCompleted, Failed, NoEvidence
	StartTime      *metav1.Time           `json:"startTime,omitempty"`
	EndTime        *metav1.Time           `json:"endTime,omitempty"`
	FindingCount   int32                  `json:"findingCount,omitempty"`
	FailedCount    int32                  `json:"failedCount,omitempty"`
	EvidenceURI    string                 `json:"evidenceUri,omitempty"`
	ManifestHash   string                 `json:"manifestHash,omitempty"`
	CollectorTimes map[string]metav1.Time `json:"collectorTimes,omitempty"`
	CoverageGaps   []string               `json:"coverageGaps,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ComplianceAudit is the Schema for the complianceaudits API
type ComplianceAudit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComplianceAuditSpec   `json:"spec,omitempty"`
	Status ComplianceAuditStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ComplianceAuditList contains a list of ComplianceAudit
type ComplianceAuditList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComplianceAudit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComplianceProfile{}, &ComplianceProfileList{})
	SchemeBuilder.Register(&Finding{}, &FindingList{})
	SchemeBuilder.Register(&ComplianceAudit{}, &ComplianceAuditList{})
}
