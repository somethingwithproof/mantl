# Security Baseline

Kyverno policies:
- require-labels.yaml
- restrict-registries.yaml
- verify-image-signatures.yaml (starts in Audit; replace cosign public key and switch to Enforce)

Recommended flow:
1. Apply policies in Audit and review policy reports
2. Update registries and keys as needed
3. Change validationFailureAction to Enforce (this repo now ships signature policies in Enforce with exceptions for core namespaces)

Cosign signing:
- Key-based (public key): use verify-image-signatures*.yaml and set publicKeys
- Keyless (OIDC): use verify-image-keyless.yaml and set subject to your issuer identity (e.g., https://github.com/YOUR_ORG/YOUR_REPO/.github/workflows/<wf>@refs/heads/main), issuer https://token.actions.githubusercontent.com, rekor https://rekor.sigstore.dev. You can update the subject via `scripts/set-keyless-subject.sh <subject>`.
- Our CI signs the dev image; extend to application images in your org pipelines
- Distribute public keys securely and reference them in verify-image-signatures.yaml

Injecting your Cosign public key:

Pre-commit and CI:
- A pre-commit hook (kyverno-no-placeholders) and CI job (policy-placeholder-check) block merges if placeholder keys/subjects are present in policies/kyverno.
1. Obtain `cosign.pub` for your signing key (or your OIDC keyless root if using key-based fallbacks)
2. Run `scripts/set-cosign-key.sh path/to/cosign.pub`
3. Commit and let ArgoCD sync `kyverno-policies`
