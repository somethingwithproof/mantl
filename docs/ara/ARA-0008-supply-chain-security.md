<!-- SPDX-License-Identifier: Apache-2.0 -->
# ARA-0008 Supply-chain security
Status: Accepted
Date: 2025-12-27

Context
Enterprises require provenance, SBOMs, and deploy-time verification to protect against supply chain attacks.

Decision
**Image Security**:
- **Sign images** with Cosign using keyless (OIDC) or key-based signatures
- **Generate SBOMs** with Syft for vulnerability tracking
- **Vulnerability scanning** with Trivy or Grype in CI pipelines
- **SLSA provenance** attestations generated in GitHub Actions

**Admission Control**:
- **Kyverno policy** (`verify-image-signatures.yaml`) enforces signature verification at deploy time
- Only signed images from approved registries are allowed to run
- Policy mode: Audit (default) or Enforce (production)

**Registry Strategy**:
- **Cloud-native registries** (ECR, GCR, ACR, GHCR) for cloud deployments
- **Harbor** (CNCF Graduated) for on-prem/air-gapped environments with built-in:
  - Vulnerability scanning (Trivy integration)
  - Content trust (Notary/Cosign)
  - RBAC and project quotas
  - Replication across registries

**CI/CD Integration**:
1. CI builds container image
2. Syft generates SBOM
3. Trivy scans for vulnerabilities (fail on HIGH/CRITICAL)
4. Cosign signs image with OIDC identity
5. Push to registry (Harbor, ECR, GCR, etc.)
6. ArgoCD deploys; Kyverno verifies signature

Consequences
- Stronger integrity guarantees; protection against tampered images
- Audit trail of who built and signed each image
- Vulnerability tracking across deployed images
- Slight build time increase (~30-60 seconds per image)
- Harbor provides enterprise-grade registry for on-prem deployments

Alternatives
- **Scan-only approach**: Detects vulnerabilities but no tamper protection
- **TUF/Notary**: More complex than Cosign; less GitHub Actions integration
- **Cloud registry only**: Doesn't support air-gapped or on-prem deployments
