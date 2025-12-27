# ARA-0008 Supply-chain security
Status: Accepted
Date: 2025-12-27

Context
Enterprises require provenance, SBOMs, and deploy-time verification.

Decision
- Generate SBOMs (Syft) and sign images/charts (Cosign) with OIDC keys.
- Enforce signature verification at admission (Kyverno/Gatekeeper policy).
- Add SLSA provenance in CI and vulnerability scanning (Trivy/Grype).

Consequences
- Stronger integrity guarantees; slight build time increase

Alternatives
- Scan-only approach (no signing/provenance): insufficient for tamper protection
