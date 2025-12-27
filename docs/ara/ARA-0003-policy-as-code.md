# ARA-0003 Policy-as-Code
Status: Accepted
Date: 2025-12-27

Context
Security baselines must be enforced consistently across clouds and clusters.

Decision
Adopt Kyverno for admission-time policy (non-root, no :latest, registry allowlist, resource limits, image signatures later).

Consequences
- Fewer drift/exception channels than manual review
- Policies versioned and tested in CI

Alternatives
- OPA Gatekeeper: also viable, Kyverno chosen for native-K8s CRD ergonomics
