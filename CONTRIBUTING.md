# Contributing

Thanks for helping modernize Mantl. This guide covers the day‑to‑day contributor workflow for the 2026 Kubernetes‑first toolkit.

## TL;DR
- Base branch: main
- Install hooks: `pipx install pre-commit && pre-commit install` (or `pip install pre-commit`)
- Format/lint/test locally, then open a PR. All CI checks must pass.

## Prerequisites
- GitHub CLI (gh) for PRs and release utilities
- Docker/Podman
- kind, kubectl, helm
- Terraform/OpenTofu
- Python 3.11+ (Black, Ruff, pytest)
- Tools used in CI (recommended locally): conftest, kyverno-cli, kubeconform, cosign, syft, trivy

Install methods vary by OS; see your package manager or the projects’ docs.

## Setup
- Clone and create a feature branch: `git checkout -b feat/short-description`
- Install dev tools: `pip install -r requirements-dev.txt` if present; otherwise `pip install -r requirements.txt`
- Enable hooks:
  ```
  pre-commit install
  pre-commit run -a
  ```

## Dev loop
- Format/lint: `make fmt && make lint` (Black, Ruff, yamllint, terraform fmt, etc.)
- Unit/policy tests: `make test` or `pytest -q` if available; `conftest test` on policies/manifests
- Validate manifests: `kubeconform -strict -summary -ignore-missing-schemas -k8s-version 1.30 -schema-location default` against rendered YAML
- Kyverno: `kyverno apply policies/kyverno -r <rendered_yaml_dir> --audit-warn` and `kyverno test` if tests exist
- kind smoke: `make kind-smoke` or follow docs/quickstart-platform.md

## CI parity and required checks
CI runs policy tests, kyverno validation, kubeconform, kind smoke/e2e, sbom, and security scans. Required checks are listed in docs/ci-required-checks.md. Use `scripts/set-required-checks.sh main` to configure branch protection with gh.

## Security policy promotion (Audit → Enforce)
Kyverno policies live in policies/kyverno and are applied by the kyverno-policies ArgoCD app.
- Begin in Audit, review PolicyReports, then flip to Enforce.
- Signature verification policies ship in Enforce with namespace exceptions for core system namespaces; narrow exceptions over time.
- Key‑based: inject your Cosign public key: `scripts/set-cosign-key.sh path/to/cosign.pub`.
- Keyless: set your OIDC subject: `scripts/set-keyless-subject.sh "https://github.com/ORG/REPO/.github/workflows/ci.yaml@refs/heads/main"`.
See docs/security.md.

## Overlays and bootstrap
- Baseline app set: `kubectl apply -k clusters/production`
- Provider overlays: see docs/overlays.md for AWS/GKE/AKS and WI overlays
- Domain customization for Gateway sample: see docs/overlays.md (gateway-sample-custom)

## Release
- Tag vX.Y.Z to trigger release workflow (signs OCI bundles, publishes SBOM)
- Images and bundles are signed with Cosign; verify with `cosign verify --certificate-oidc-issuer https://token.actions.githubusercontent.com <ref>`

## Pull requests
- Keep PRs focused; include motivation and testing notes
- Ensure pre-commit and all CI checks pass
- Use Conventional Commits for messages (feat:, fix:, chore:, docs:, refactor:, perf:, test:, build:, ci:). Messages are linted via pre-commit (Commitizen).
- Link issues where applicable

## Releases and versioning
- We follow Semantic Versioning (MAJOR.MINOR.PATCH) and Conventional Commits.
- Release automation uses release-please. It opens a release PR from commits; merging that PR tags vX.Y.Z and updates CHANGELOG.md.
- Our release workflow reacts to the published release to render/sign bundles and push OCI artifacts; no manual release creation is needed.

## Code of Conduct
Please respect our [Code of Conduct](code-of-conduct.md).
