package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComplianceProfileSpec defines the desired state of ComplianceProfile
type ComplianceProfileSpec struct {
	Framework string `json:"framework"`
	Version   string `json:"version"`
}

// ComplianceProfileStatus defines the observed state of ComplianceProfile
type ComplianceProfileStatus struct {
	State          string `json:"state,omitempty"`
	ActivePolicies int32  `json:"activePolicies,omitempty"`
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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Finding is the Schema for the findings API
type Finding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FindingSpec `json:"spec,omitempty"`
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
	Profile   string `json:"profile"`
	Frequency string `json:"frequency,omitempty"` // manual, daily, weekly
}

// ComplianceAuditStatus defines the observed state of a Compliance Audit
type ComplianceAuditStatus struct {
	Phase        string       `json:"phase,omitempty"` // Pending, Running, Completed, PartiallyCompleted, Failed, NoEvidence
	StartTime    *metav1.Time `json:"startTime,omitempty"`
	EndTime      *metav1.Time `json:"endTime,omitempty"`
	FindingCount int32        `json:"findingCount,omitempty"`
	FailedCount  int32        `json:"failedCount,omitempty"`
	EvidenceURI  string       `json:"evidenceUri,omitempty"`
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
