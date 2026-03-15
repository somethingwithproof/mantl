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

func init() {
	SchemeBuilder.Register(&ComplianceProfile{}, &ComplianceProfileList{})
}
