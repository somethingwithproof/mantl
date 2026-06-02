---
name: mantl-compliance
description: Load when working on the compliance operator. Covers CRD types under apis/compliance, the controllers that reconcile them, framework YAML under compliance/frameworks, Kyverno policies under policies/kyverno, and the evidence capture path in pkg/evidence, plus the framework-policy-evidence-finding-audit wiring and the ADR 003 evidence contract.
---

# mantl compliance skill

## The data flow

A compliance framework is YAML data. The operator turns it into running enforcement and durable evidence:

```
ComplianceFramework (YAML)         controls reference policy templates + evidence collectors
        ↓ profile selects + sets mode (enforce/audit) + namespace excludes
ComplianceProfile (CRD)            reconciled by profile_controller.go
        ↓ policy templates applied as Kyverno policies
Kyverno PolicyReport               violations surface as report entries
        ↓ finding_controller.go maps reports → Findings
Finding (CRD)                      one per policy/rule/resource, status pass|fail|warn
        ↓ evidence collectors snapshot cluster state
pkg/evidence snapshot → S3         Finding holds the s3:// URI, not the payload (ADR 003)
        ↓ audit_controller.go rolls findings into a point-in-time audit
ComplianceAudit (CRD)
```

## CRD types (`apis/compliance/v1alpha1`)

Three kinds, each with the standard `Spec` / `Status` split and a `List` type:

- `ComplianceProfile` / `ComplianceProfileSpec` / `ComplianceProfileStatus`
- `Finding` / `FindingSpec` / `FindingStatus` (`FindingSpec.Status` is the string `pass`, `fail`, or `warn`)
- `ComplianceAudit` / `ComplianceAuditSpec` / `ComplianceAuditStatus`

Each kind carries kubebuilder markers:

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ComplianceProfile struct { ... }

// +kubebuilder:object:root=true
type ComplianceProfileList struct { ... }
```

Rules when editing types:

- Root objects need `+kubebuilder:object:root=true`. Objects with a status subresource need `+kubebuilder:subresource:status`.
- `zz_generated.deepcopy.go` is generated. After changing any type, regenerate it (controller-gen) rather than hand-editing. A type without matching DeepCopy methods will not satisfy `runtime.Object` and the controller will not compile.
- `groupversion_info.go` registers the group/version (`compliance.mantl.io/v1alpha1`) and the scheme. New kinds must be added to the scheme builder there.

## Controllers (`controllers/compliance`)

- `profile_controller.go` reconciles a `ComplianceProfile`: resolves its `frameworkRef`, applies the referenced policy templates as Kyverno policies in the requested `mode`, honoring `excludeNamespaces`.
- `finding_controller.go` reconciles Kyverno `PolicyReport` resources into `Finding` CRs. It builds a stable finding name from policy/rule/namespace/resource and normalizes severity. This is the read side of enforcement.
- `audit_controller.go` reconciles a `ComplianceAudit`, aggregating findings into a point-in-time report.

Reconcilers are idempotent and must converge from any starting state. Do not assume a single pass; assume re-entry.

## Evidence contract (ADR 003)

`pkg/evidence/snapshot.go` captures a resource snapshot and uploads it:

- `CaptureResource` / `CaptureResourceWithContext` produce a `Snapshot` for a kind/name/namespace.
- `UploadToS3` uses `DefaultServerSideEncryption` (`AES256`). `UploadToS3WithEncryption` accepts `aws:kms` for KMS-backed buckets (profiles can set `evidenceStore.encryption.type: aws-kms`).
- The upload returns an `s3://bucket/key` URI. That URI is what lands in the `Finding`, per the contract.

Hard rule from ADR 003: never put raw evidence (full JSON dumps of NetworkPolicies, RBAC, etc.) into etcd or into a CRD. CRDs are metadata and a pointer. The bucket carries versioning plus a WORM/Object-Lock retention policy; the operator authenticates via IRSA. etcd has a ~1MB value limit and is not document storage.

## The framework→policy gap (known and load-bearing)

Framework YAML controls reference policy templates by name, for example:

```yaml
- type: policy
  policyRef:
    template: require-network-policy
    mode: enforce
```

Across `compliance/frameworks/*/framework.yaml` roughly thirty distinct templates are referenced (`require-rbac-least-privilege`, `disallow-privileged-containers`, `require-seccomp-profile`, and so on). Each referenced template must resolve to an actual Kyverno policy. The shipped set is incomplete:

- `policies/kyverno/policies/` ships only `require-labels`, `restrict-image-registries`, `verify-image-signatures`.
- Some templates live framework-locally under `compliance/frameworks/<fw>/policies/` (SOC2 has `require-network-policy`, `require-pod-security-context`, `require-resource-limits`, `require-image-signatures`, and a few others).
- Many referenced templates have no implementation in either location.

When adding or wiring a control: confirm the named template exists as a real Kyverno policy. If it does not, the control is a dangling reference that produces no enforcement and no findings. Either implement the template or do not reference it. Name the template identically in the framework YAML and the policy `metadata.name`.

## When you change a CRD type

1. Edit the struct and its kubebuilder markers in `types.go`.
2. Regenerate `zz_generated.deepcopy.go` with controller-gen; do not hand-edit it.
3. If you added a kind, register it in `groupversion_info.go`.
4. Update any controller that lists or watches the type.
5. Run `make go-test`.
