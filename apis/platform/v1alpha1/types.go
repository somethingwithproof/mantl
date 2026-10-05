// Package v1alpha1 contains API Schema definitions for the platform v1alpha1 API group
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MantlClusterSpec defines the desired state of a MantlCluster
type MantlClusterSpec struct {
	// Environment labels generated namespaces for reviewed admission policies.
	// +kubebuilder:validation:Enum=dev;staging;production
	// +optional
	Environment string `json:"environment,omitempty"`
	// Provider configuration
	Provider ProviderSpec `json:"provider"`

	// Kubernetes configuration
	Kubernetes KubernetesSpec `json:"kubernetes"`

	// Profile configuration (sizing and compliance)
	Profile ProfileSpec `json:"profile"`

	// Networking configuration
	Networking NetworkingSpec `json:"networking"`

	// Features to enable on the platform
	Features FeatureSpec `json:"features"`

	GitOps GitOpsSpec `json:"gitops,omitempty"`

	// Tenants to onboard to the cluster
	Tenants []TenantSpec `json:"tenants,omitempty"`
}

// GitOpsSpec binds rendered applications to a versioned repository.
type GitOpsSpec struct {
	Repository string `json:"repository,omitempty"`
	Revision   string `json:"revision,omitempty"`
	// TenantPath is the committed path of generated tenant files in Repository.
	TenantPath string `json:"tenantPath,omitempty"`
	// OperatorPath selects an overlay containing evidence settings and image digest.
	OperatorPath string `json:"operatorPath,omitempty"`
}

// TenantSpec defines a platform tenant (team/app)
type TenantSpec struct {
	// Name of the tenant
	Name string `json:"name"`

	// Namespace for the tenant (defaults to tenant name if empty)
	Namespace string `json:"namespace,omitempty"`

	// Admins for the tenant (mapped to RBAC)
	Admins []string `json:"admins,omitempty"`
}

// ProviderSpec defines the cloud provider configuration
type ProviderSpec struct {
	// Kind: aws, gcp, azure, do, linode, local
	Kind string `json:"kind"`

	// Region to deploy to
	Region string `json:"region"`

	// AccountID or project name (optional)
	AccountID string `json:"accountId,omitempty"`
}

// KubernetesSpec defines the Kubernetes distribution and version
type KubernetesSpec struct {
	// Distribution: eks, gke, aks, doks, lke, kind
	Distribution string `json:"distribution"`

	// Version of Kubernetes
	Version string `json:"version"`
}

// ProfileSpec defines sizing and compliance profiles
type ProfileSpec struct {
	// Size: small, medium, full
	Size string `json:"size"`

	// Compliance profile to apply: soc2, hipaa, pci, cis
	Compliance string `json:"compliance,omitempty"`
}

// NetworkingSpec defines cluster networking
type NetworkingSpec struct {
	// Domain for platform services
	Domain string `json:"domain"`

	// Exposure: public, private, hybrid
	Exposure string `json:"exposure"`

	// VPC/VNet ID (optional)
	VpcID string `json:"vpcId,omitempty"`
}

// FeatureSpec defines platform features to enable
type FeatureSpec struct {
	// Observability: Prometheus, Grafana, Tempo, Loki
	Observability bool `json:"observability"`

	// ProgressiveDelivery: Argo Rollouts, Flagger
	ProgressiveDelivery bool `json:"progressiveDelivery"`

	// Security: Falco, Kyverno, Trivy
	Security bool `json:"security"`

	// Secrets: External Secrets, Vault
	Secrets bool `json:"secrets"`

	// Compliance: Compliance Operator, Evidence Collection
	Compliance bool `json:"compliance"`
}

// MantlClusterStatus defines the observed state of MantlCluster
type MantlClusterStatus struct {
	// Phase: Pending, Provisioning, Running, Degraded, Failed
	Phase string `json:"phase"`

	// Conditions represent the current state of the cluster
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Endpoint of the Kubernetes API
	Endpoint string `json:"endpoint,omitempty"`

	// Version of the platform software
	Version string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MantlCluster is the Schema for the mantlclusters API
type MantlCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MantlClusterSpec   `json:"spec,omitempty"`
	Status MantlClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MantlClusterList contains a list of MantlCluster
type MantlClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MantlCluster `json:"items"`
}
