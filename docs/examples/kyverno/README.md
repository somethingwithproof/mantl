# Kyverno policy examples (templates)

These are template policies that require user-specific values (cosign public
keys for key-based image verification) before use. They are not deployed.
Mantl signs its own images keyless (see `policies/kyverno/verify-image-keyless.yaml`).
Copy a template into `policies/kyverno/`, replace the placeholder keys, and let
GitOps apply it (ADR 006).
