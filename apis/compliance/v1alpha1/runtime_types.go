package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// CollectorTask is an immutable resource scope; it contains no evidence payloads.
type CollectorTask struct {
	ID        string `json:"id"`
	Collector string `json:"collector"`
	Control   string `json:"control"`
	Resource  string `json:"resource"`
	Namespace string `json:"namespace,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="audit run input is immutable"
type AuditRunSpec struct {
	ClusterID               string `json:"clusterId,omitempty"`
	CollectorImage          string `json:"collectorImage,omitempty"`
	CollectorServiceAccount string `json:"collectorServiceAccount,omitempty"`
	EvidenceBucket          string `json:"evidenceBucket,omitempty"`
	EvidenceKMSKey          string `json:"evidenceKmsKey,omitempty"`

	Audit            string          `json:"audit"`
	AuditUID         string          `json:"auditUid"`
	Profile          string          `json:"profile"`
	Framework        string          `json:"framework"`
	FrameworkVersion string          `json:"frameworkVersion"`
	BundleDigest     string          `json:"bundleDigest"`
	ScheduledAt      metav1.Time     `json:"scheduledAt"`
	Tasks            []CollectorTask `json:"tasks"`
	CoverageGaps     []string        `json:"coverageGaps,omitempty"`
}
type TaskResult struct {
	Phase       string       `json:"phase"`
	URI         string       `json:"uri,omitempty"`
	Version     string       `json:"version,omitempty"`
	SHA256      string       `json:"sha256,omitempty"`
	CapturedAt  *metav1.Time `json:"capturedAt,omitempty"`
	RetainUntil *metav1.Time `json:"retainUntil,omitempty"`
}
type AuditRunStatus struct {
	Phase        string                `json:"phase,omitempty"`
	EndTime      *metav1.Time          `json:"endTime,omitempty"`
	Results      map[string]TaskResult `json:"results,omitempty"`
	EvidenceURI  string                `json:"evidenceUri,omitempty"`
	ManifestHash string                `json:"manifestHash,omitempty"`
	CoverageGaps []string              `json:"coverageGaps,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type AuditRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AuditRunSpec   `json:"spec"`
	Status            AuditRunStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type AuditRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuditRun `json:"items"`
}

type ControlEvaluationSpec struct {
	Profile      string `json:"profile"`
	Control      string `json:"control"`
	Framework    string `json:"framework"`
	BundleDigest string `json:"bundleDigest,omitempty"`
}
type ControlEvaluationStatus struct {
	// +kubebuilder:validation:Enum=pass;fail;unknown;error;not-applicable
	Result         string       `json:"result"`
	Coverage       string       `json:"coverage"`
	Freshness      string       `json:"freshness"`
	ObservedAt     *metav1.Time `json:"observedAt,omitempty"`
	FindingCount   int32        `json:"findingCount,omitempty"`
	ExceptionNames []string     `json:"exceptionNames,omitempty"`
	EvidenceURI    string       `json:"evidenceUri,omitempty"`
	Reason         string       `json:"reason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ControlEvaluation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ControlEvaluationSpec   `json:"spec"`
	Status            ControlEvaluationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ControlEvaluationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlEvaluation `json:"items"`
}
type ComplianceExceptionSpec struct {
	Profile       string      `json:"profile"`
	Control       string      `json:"control"`
	Resource      string      `json:"resource,omitempty"`
	Owner         string      `json:"owner"`
	Justification string      `json:"justification"`
	ExpiresAt     metav1.Time `json:"expiresAt"`
}

// ComplianceExceptionStatus must be written only by a separately authorized approver.
type ComplianceExceptionStatus struct {
	Approved   bool         `json:"approved"`
	ApprovedBy string       `json:"approvedBy,omitempty"`
	ApprovedAt *metav1.Time `json:"approvedAt,omitempty"`
	// ApprovedGeneration binds approval to an exact reviewed spec revision.
	ApprovedGeneration int64 `json:"approvedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ComplianceException struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComplianceExceptionSpec   `json:"spec"`
	Status            ComplianceExceptionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ComplianceExceptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComplianceException `json:"items"`
}
