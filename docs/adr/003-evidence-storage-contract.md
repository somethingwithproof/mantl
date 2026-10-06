# ADR 003: Compliance Evidence Storage Contract

## Status
Accepted

## Context
Mantl's differentiator is its continuous compliance subsystem. For this subsystem to satisfy enterprise security and audit requirements (such as SOC2, HIPAA, PCI-DSS), it must collect and store point-in-time evidence of the cluster's state. We need to decide where and how this evidence is stored.

## Decision
1. **Control Plane Separation:** We will NOT store raw compliance evidence (e.g., full JSON dumps of network policies or RBAC roles) in Kubernetes `etcd` or directly within the `Finding` Custom Resource Definitions (CRDs).
2. **Object Storage Backend:** All raw evidence payloads will be uploaded to an S3-compatible Object Storage bucket (e.g., AWS S3, Google Cloud Storage).
3. **CRD Linkage:** The Kubernetes `Finding` CRD will act as metadata only, storing a URI/pointer (e.g., `s3://mantl-evidence-prod-us-east/2026/03/14/policy-snapshot-xyz.json`) to the actual evidence blob.
4. **Immutability and Retention:** The designated Object Storage bucket MUST have versioning enabled and a WORM (Write Once, Read Many) policy enforced. Evidence should be retained for a minimum of 7 years, depending on the regulatory framework (configurable via `MantlCluster` spec).

## Rationale
*   **etcd Limits:** Kubernetes `etcd` is designed for state management, not document storage. Storing large JSON snapshots would quickly exhaust etcd's 1MB value limit and bloat the cluster database.
*   **Audit Trail:** Auditors require evidence to be immutable. Cloud-native object storage provides built-in Object Lock and retention policies, removing the burden of implementing WORM at the application layer.
*   **Performance:** The Compliance Operator can reconcile `Finding` metadata rapidly without carrying the payload overhead.

## Consequences
*   **Dependency:** The Mantl bootstrap process must ensure the Object Storage bucket is provisioned (e.g., via Crossplane) before the Compliance Operator can become active.
*   **IAM Complexity:** The Compliance Operator will require specific IAM roles (e.g., IRSA in AWS) to securely authenticate and write to the bucket.


## Implementation status clarification (2026-10-06)

The decision above records the original storage intention. The implemented backend
is AWS S3 through the AWS SDK with versioned objects and Object Lock COMPLIANCE
retention, including for GCP/Azure Kubernetes clusters. Native GCS/Azure WORM
adapters are not implemented; GCS should not be read as a working interchangeable
backend. `pkg/evidence/store.go` enforces a minimum 2557 days. The current operator
has no MantlCluster retention setting. Operators provision the bucket and identity;
the current bootstrap does not provision it through Crossplane.

See [compliance operations](../compliance-operations.md) for the implemented
contract, permissions, retention costs, and export verification. This clarification
does not add a storage adapter or establish provider/cloud acceptance.
