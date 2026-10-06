<!-- SPDX-License-Identifier: Apache-2.0 -->
# ARA-0003 Policy-as-Code
Status: Accepted
Date: 2025-12-27

Context
Security baselines must be enforced consistently across clouds and clusters. Teams need policy-as-code that's easy to write, test, and version in Git.

Decision
Adopt **Kyverno** (CNCF Graduated) for admission-time policy enforcement.

**Included Policies** (`policies/kyverno/policies/`):
- **require-labels.yaml**: Enforce `app.kubernetes.io/name` and `app.kubernetes.io/managed-by` labels
- **restrict-registries.yaml**: Allow only approved registries (ghcr.io, gcr.io, quay.io, docker.io, harbor)
- **verify-image-signatures.yaml**: Verify Sigstore/Cosign signatures on container images

**Policy Modes**:
- **Audit** (default): Log violations without blocking
- **Enforce**: Block non-compliant resources (enable per-policy)

Consequences
- Fewer drift/exception channels than manual review
- Policies versioned and tested in CI
- Kubernetes-native YAML format (no Rego learning curve)
- Built-in policy reports and Prometheus metrics
- Native support for image verification with Cosign/Sigstore
- Removed OPA/Gatekeeper (commit 71b98ac3) - Kyverno provides equivalent functionality

Alternatives
- **OPA Gatekeeper**: Also CNCF Graduated, powerful but requires Rego language expertise
- **jsPolicy**: Policies in JavaScript; less mature ecosystem
- **Kubewarden**: Policies in WebAssembly; emerging technology
