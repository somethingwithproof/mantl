// Package v1alpha1 contains API Schema definitions for the compliance v1alpha1 API group
// +kubebuilder:object:generate=true
// +groupName=compliance.mantl.io
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// =============================================================================
// ComplianceFramework - Defines a compliance standard
// =============================================================================

// ComplianceFrameworkSpec defines a compliance framework like SOC2 or HIPAA
type ComplianceFrameworkSpec struct {
	// DisplayName is the human-readable name
	DisplayName string `json:"displayName"`

	// Description explains the framework
	Description string `json:"description,omitempty"`

	// Version of the framework (e.g., "2017" for SOC2)
	Version string `json:"version"`

	// Authority is the governing body (e.g., "AICPA", "HHS")
	Authority string `json:"authority"`

	// Categories organize controls into logical groups
	Categories []ControlCategory `json:"categories"`

	// Controls lists all controls in this framework
	Controls []Control `json:"controls"`
}

// ControlCategory groups related controls
type ControlCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Control represents a single compliance requirement
type Control struct {
	// ID is the unique identifier (e.g., "CC6.1")
	ID string `json:"id"`

	// CategoryID links to the parent category
	CategoryID string `json:"categoryId"`

	// Name is a short title
	Name string `json:"name"`

	// Description is the full control text
	Description string `json:"description"`

	// Severity indicates importance: critical, high, medium, low
	Severity string `json:"severity"`

	// Automated indicates if this can be automatically verified
	Automated bool `json:"automated"`

	// Mappings link to technical implementations
	Mappings []ControlMapping `json:"mappings,omitempty"`
}

// ControlMapping links a control to Kubernetes resources
type ControlMapping struct {
	// Type of mapping: policy, evidence, attestation
	Type string `json:"type"`

	// PolicyRef references a Kyverno policy template
	PolicyRef *PolicyReference `json:"policyRef,omitempty"`

	// EvidenceCollector defines how to collect proof
	EvidenceCollector *EvidenceCollectorSpec `json:"evidenceCollector,omitempty"`

	// AttestationRequired indicates human sign-off needed
	AttestationRequired bool `json:"attestationRequired,omitempty"`
}

// PolicyReference points to a Kyverno policy
type PolicyReference struct {
	// Template is the policy template name
	Template string `json:"template"`

	// Parameters to customize the policy
	Parameters map[string]string `json:"parameters,omitempty"`

	// Mode: enforce or audit
	Mode string `json:"mode"`
}

// EvidenceCollectorSpec defines evidence collection
type EvidenceCollectorSpec struct {
	// Type: config-snapshot, log-query, metric-query, scan-result
	Type string `json:"type"`

	// Schedule in cron format
	Schedule string `json:"schedule,omitempty"`

	// Query for log/metric collection
	Query string `json:"query,omitempty"`

	// Resources to snapshot
	Resources []string `json:"resources,omitempty"`

	// RetentionDays for evidence storage
	RetentionDays int `json:"retentionDays,omitempty"`
}

// ComplianceFrameworkStatus defines the observed state
type ComplianceFrameworkStatus struct {
	// ControlCount is the total number of controls
	ControlCount int `json:"controlCount"`

	// AutomatedCount is controls that can be auto-verified
	AutomatedCount int `json:"automatedCount"`

	// LastUpdated timestamp
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Controls",type=integer,JSONPath=`.status.controlCount`
// +kubebuilder:printcolumn:name="Automated",type=integer,JSONPath=`.status.automatedCount`

// ComplianceFramework is the Schema for compliance frameworks
type ComplianceFramework struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComplianceFrameworkSpec   `json:"spec,omitempty"`
	Status ComplianceFrameworkStatus `json:"status,omitempty"`
}

// =============================================================================
// ComplianceProfile - Selected controls for an organization
// =============================================================================

// ComplianceProfileSpec defines which controls to enforce
type ComplianceProfileSpec struct {
	// FrameworkRef references the compliance framework
	FrameworkRef string `json:"frameworkRef"`

	// Mode: enforce or audit
	Mode string `json:"mode"`

	// IncludeControls lists control IDs to include (empty = all)
	IncludeControls []string `json:"includeControls,omitempty"`

	// ExcludeControls lists control IDs to exclude
	ExcludeControls []string `json:"excludeControls,omitempty"`

	// Namespaces to apply controls (empty = all)
	Namespaces []string `json:"namespaces,omitempty"`

	// ExcludeNamespaces to skip
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`

	// EvidenceStore configuration
	EvidenceStore EvidenceStoreSpec `json:"evidenceStore"`

	// NotificationChannels for alerts
	NotificationChannels []NotificationChannel `json:"notificationChannels,omitempty"`
}

// EvidenceStoreSpec defines where to store evidence
type EvidenceStoreSpec struct {
	// Type: s3, gcs, azure-blob
	Type string `json:"type"`

	// Bucket name
	Bucket string `json:"bucket"`

	// Prefix for object keys
	Prefix string `json:"prefix,omitempty"`

	// SecretRef for credentials
	SecretRef string `json:"secretRef,omitempty"`

	// Encryption configuration
	Encryption *EncryptionSpec `json:"encryption,omitempty"`
}

// EncryptionSpec for evidence encryption
type EncryptionSpec struct {
	// Type: aws-kms, gcp-kms, azure-keyvault
	Type string `json:"type"`

	// KeyID for encryption
	KeyID string `json:"keyId"`
}

// NotificationChannel for compliance alerts
type NotificationChannel struct {
	// Type: slack, teams, email, pagerduty
	Type string `json:"type"`

	// SecretRef containing webhook URL or credentials
	SecretRef string `json:"secretRef"`

	// Severity threshold: critical, high, medium, low
	MinSeverity string `json:"minSeverity"`
}

// ComplianceProfileStatus defines the observed state
type ComplianceProfileStatus struct {
	// State: Active, Degraded, Error
	State string `json:"state"`

	// ActiveControls count
	ActiveControls int `json:"activeControls"`

	// ComplianceScore percentage (0-100)
	ComplianceScore int `json:"complianceScore"`

	// Findings count by severity
	CriticalFindings int `json:"criticalFindings"`
	HighFindings     int `json:"highFindings"`
	MediumFindings   int `json:"mediumFindings"`
	LowFindings      int `json:"lowFindings"`

	// LastAudit timestamp
	LastAudit metav1.Time `json:"lastAudit,omitempty"`

	// Conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Framework",type=string,JSONPath=`.spec.frameworkRef`
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.mode`
// +kubebuilder:printcolumn:name="Score",type=integer,JSONPath=`.status.complianceScore`
// +kubebuilder:printcolumn:name="Critical",type=integer,JSONPath=`.status.criticalFindings`

// ComplianceProfile is the Schema for compliance profiles
type ComplianceProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComplianceProfileSpec   `json:"spec,omitempty"`
	Status ComplianceProfileStatus `json:"status,omitempty"`
}

// =============================================================================
// ComplianceAudit - Point-in-time assessment
// =============================================================================

// ComplianceAuditSpec defines an audit run
type ComplianceAuditSpec struct {
	// ProfileRef references the compliance profile
	ProfileRef string `json:"profileRef"`

	// Type: scheduled, manual, continuous
	Type string `json:"type"`

	// Schedule in cron format (for scheduled type)
	Schedule string `json:"schedule,omitempty"`

	// IncludeEvidence in the report
	IncludeEvidence bool `json:"includeEvidence"`

	// ReportFormats to generate: pdf, html, json
	ReportFormats []string `json:"reportFormats,omitempty"`
}

// ComplianceAuditStatus defines the observed state
type ComplianceAuditStatus struct {
	// Phase: Pending, Running, Completed, Failed
	Phase string `json:"phase"`

	// StartTime of the audit
	StartTime metav1.Time `json:"startTime,omitempty"`

	// CompletionTime of the audit
	CompletionTime metav1.Time `json:"completionTime,omitempty"`

	// ControlsAssessed count
	ControlsAssessed int `json:"controlsAssessed"`

	// ControlsPassed count
	ControlsPassed int `json:"controlsPassed"`

	// ControlsFailed count
	ControlsFailed int `json:"controlsFailed"`

	// ComplianceScore percentage
	ComplianceScore int `json:"complianceScore"`

	// ReportURLs for generated reports
	ReportURLs map[string]string `json:"reportUrls,omitempty"`

	// Findings discovered in this audit
	FindingRefs []string `json:"findingRefs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Profile",type=string,JSONPath=`.spec.profileRef`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Score",type=integer,JSONPath=`.status.complianceScore`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ComplianceAudit is the Schema for compliance audits
type ComplianceAudit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComplianceAuditSpec   `json:"spec,omitempty"`
	Status ComplianceAuditStatus `json:"status,omitempty"`
}

// =============================================================================
// Finding - Individual compliance violation
// =============================================================================

// FindingSpec defines a compliance finding
type FindingSpec struct {
	// ControlID that was violated
	ControlID string `json:"controlId"`

	// ProfileRef that detected this
	ProfileRef string `json:"profileRef"`

	// Severity: critical, high, medium, low
	Severity string `json:"severity"`

	// Title is a short description
	Title string `json:"title"`

	// Description of the violation
	Description string `json:"description"`

	// Resource that violated the control
	Resource ResourceReference `json:"resource"`

	// Evidence collected
	Evidence []EvidenceItem `json:"evidence,omitempty"`

	// Remediation guidance
	Remediation RemediationSpec `json:"remediation"`
}

// ResourceReference identifies a Kubernetes resource
type ResourceReference struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name"`
}

// EvidenceItem is a piece of evidence
type EvidenceItem struct {
	// Type: config, log, metric, screenshot
	Type string `json:"type"`

	// Description of what this proves
	Description string `json:"description"`

	// URL to the evidence in object storage
	URL string `json:"url"`

	// Hash for integrity verification
	Hash string `json:"hash"`

	// CollectedAt timestamp
	CollectedAt metav1.Time `json:"collectedAt"`
}

// RemediationSpec defines how to fix a finding
type RemediationSpec struct {
	// Description of manual steps
	Description string `json:"description"`

	// AutoRemediable indicates if auto-fix is possible
	AutoRemediable bool `json:"autoRemediable"`

	// AutoRemediationRef references the fix
	AutoRemediationRef string `json:"autoRemediationRef,omitempty"`

	// EstimatedEffort: low, medium, high
	EstimatedEffort string `json:"estimatedEffort"`
}

// FindingStatus defines the observed state
type FindingStatus struct {
	// State: Open, InProgress, Resolved, Accepted, FalsePositive
	State string `json:"state"`

	// AssignedTo user/team
	AssignedTo string `json:"assignedTo,omitempty"`

	// DueDate for remediation
	DueDate metav1.Time `json:"dueDate,omitempty"`

	// ResolvedAt timestamp
	ResolvedAt metav1.Time `json:"resolvedAt,omitempty"`

	// Resolution notes
	Resolution string `json:"resolution,omitempty"`

	// History of state changes
	History []FindingHistoryEntry `json:"history,omitempty"`
}

// FindingHistoryEntry records a state change
type FindingHistoryEntry struct {
	Timestamp metav1.Time `json:"timestamp"`
	FromState string      `json:"fromState"`
	ToState   string      `json:"toState"`
	Actor     string      `json:"actor"`
	Comment   string      `json:"comment,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Control",type=string,JSONPath=`.spec.controlId`
// +kubebuilder:printcolumn:name="Severity",type=string,JSONPath=`.spec.severity`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Finding is the Schema for compliance findings
type Finding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FindingSpec   `json:"spec,omitempty"`
	Status FindingStatus `json:"status,omitempty"`
}

// =============================================================================
// Evidence - Stored compliance proof
// =============================================================================

// EvidenceSpec defines a piece of compliance evidence
type EvidenceSpec struct {
	// ControlID this evidence supports
	ControlID string `json:"controlId"`

	// ProfileRef that collected this
	ProfileRef string `json:"profileRef"`

	// Type: config-snapshot, log-extract, metric-sample, scan-result, attestation
	Type string `json:"type"`

	// Description of what this proves
	Description string `json:"description"`

	// StorageURL in object storage
	StorageURL string `json:"storageUrl"`

	// ContentHash for integrity
	ContentHash string `json:"contentHash"`

	// HashAlgorithm used (sha256, sha512)
	HashAlgorithm string `json:"hashAlgorithm"`

	// CollectedAt timestamp
	CollectedAt metav1.Time `json:"collectedAt"`

	// ExpiresAt for retention
	ExpiresAt metav1.Time `json:"expiresAt,omitempty"`

	// Attestation for human sign-offs
	Attestation *AttestationSpec `json:"attestation,omitempty"`
}

// AttestationSpec for human attestations
type AttestationSpec struct {
	// Attestor identity (email)
	Attestor string `json:"attestor"`

	// AttestedAt timestamp
	AttestedAt metav1.Time `json:"attestedAt"`

	// Statement of attestation
	Statement string `json:"statement"`

	// Signature (base64 encoded)
	Signature string `json:"signature,omitempty"`
}

// EvidenceStatus defines the observed state
type EvidenceStatus struct {
	// State: Pending, Stored, Verified, Expired, Corrupted
	State string `json:"state"`

	// VerifiedAt timestamp
	VerifiedAt metav1.Time `json:"verifiedAt,omitempty"`

	// Size in bytes
	Size int64 `json:"size,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Control",type=string,JSONPath=`.spec.controlId`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Collected",type=date,JSONPath=`.spec.collectedAt`

// Evidence is the Schema for compliance evidence
type Evidence struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EvidenceSpec   `json:"spec,omitempty"`
	Status EvidenceStatus `json:"status,omitempty"`
}
