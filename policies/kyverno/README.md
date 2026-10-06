<!-- SPDX-License-Identifier: Apache-2.0 -->
# Kyverno Policies

This folder is applied by the ArgoCD app `kyverno-policies`.

Policies included:
- require-labels.yaml
- restrict-registries.yaml
- verify-image-signatures.yaml (Enforce; with system-namespace exceptions)
- verify-image-signatures-ecr-gcr.yaml (Enforce; add your ECR/GCR keys; with exceptions)
- verify-image-keyless.yaml (Enforce; set subject/issuer for keyless; with exceptions)

Promotion flow:
1. Start in Audit (default). Review PolicyReports and fix violations.
2. Replace Cosign public key in `verify-image-signatures.yaml` (see `scripts/set-cosign-key.sh`).
3. Switch `validationFailureAction: Enforce` when ready.
