# ARA-0004 Secrets and key management
Status: Accepted
Date: 2025-12-27

Context
We require cross-cloud consistency with minimal developer exposure to raw secrets.

Decision
Use External Secrets Operator to sync secrets into clusters. Backends: Vault (preferred for cross-cloud) or cloud secret stores. Use workload identity/OIDC for short-lived credentials.

Consequences
- Consistent secret consumption pattern across clouds
- Vault adds operational overhead but centralizes policy and audit

Alternatives
- KMS + sealed-secrets only: simpler, less flexible across providers
