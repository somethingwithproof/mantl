<!-- SPDX-License-Identifier: Apache-2.0 -->
# ARA-0004 Secrets and key management
Status: Accepted
Date: 2025-12-27

Context
We require cross-cloud consistency with minimal developer exposure to raw secrets. Secrets must sync from cloud providers into Kubernetes with proper rotation and audit trails.

Decision
Use **External Secrets Operator** (CNCF Graduated) to sync secrets into clusters.

**Primary Backends** (cloud-native):
- **AWS Secrets Manager** with IRSA (IAM Roles for Service Accounts)
- **GCP Secret Manager** with Workload Identity
- **Azure Key Vault** with Managed Identity
- **HashiCorp Vault** (optional for cross-cloud or on-prem)

**Authentication**:
- Cloud workload identity (IRSA, Workload Identity, Managed Identity) for passwordless authentication
- No long-lived credentials stored in clusters
- Automatic credential rotation via cloud IAM

**Configuration** (`platform/secrets/external-secrets/`):
- Provider-specific SecretStore configurations in `providers/`
- ClusterSecretStore for cluster-wide secret sync
- ExternalSecret CRs for application-specific secrets

Consequences
- Consistent secret consumption pattern across clouds (Kubernetes Secret API)
- Secrets stay in cloud-native secret managers (compliance, audit trails)
- Automatic rotation without application restarts (ESO polls for changes)
- GitOps-friendly (ExternalSecret manifests versioned in Git, actual secrets in cloud)
- No Sealed Secrets or manual encryption needed

Alternatives
- **Sealed Secrets**: Git-native but requires custom encryption workflow; less flexible for multi-cloud
- **HashiCorp Vault only**: Adds operational overhead; use when cross-cloud secret sharing is required
- **KMS + manual sync**: No automation, high operational burden
