<!-- SPDX-License-Identifier: Apache-2.0 -->
# Repository cleanup audit

Audit baseline: main `5b6214303b1ca6eae63771c4d657ff50ea5895a7`.
The tracked tree, runtime loaders, release packaging, Dockerfiles, workflows,
tests, GitOps manifests and documentation references were inspected. Lack of a
text reference alone is not proof that a public example or tool configuration is
unused.

## Removed

| Path | Reason |
| --- | --- |
| `.omc/` | 39 local assistant session records and 39 license sidecars. No application, build, CI, release or documentation consumer. Session state is not project source. |
| `err` and `err.license` | Captured Kustomize stderr, accidentally tracked rather than source/configuration. |
| Root `frameworks/` | Two superseded flat SOC2/PCI-DSS YAML specimens. The runtime reads `<framework>/framework.yaml` under `compliance/frameworks/`; images, bundles and tests use that canonical tree. |
| `CLOUD_PROVIDER_COMPARISON.md`, `UPDATED_CLOUD_COMPARISON.md`, `FEATURE_PARITY_IMPLEMENTATION.md`, `PLATFORM_SECURITY_QUICK_START.md` | Obsolete guides asserted full cloud parity and production readiness without acceptance evidence. Current provider/runtime boundaries are documented elsewhere. Their only external records were secret-scan baseline entries, now removed solely for the deleted files. |

The removed files remain recoverable in Git history. No current compliance catalog,
policy mapping, runtime code, deployment manifest, license for retained content,
or active secret-scan exception was removed.

## Retained intentionally

| Area | Why it remains |
| --- | --- |
| `apis/`, `cmd/`, `controllers/`, `pkg/`, `config/` | First-party APIs, CLI/operator/services, logic and generated API assets. |
| `compliance/`, `policies/`, `deploy/`, `charts/`, `clusters/`, `templates/` | Canonical content, install bundles, manifests and examples. Vendored content is not automatically supported functionality, but may still be an input. |
| `infra/` | Provider blueprints/modules, static contracts and supported/experimental examples. |
| `infrastructure/` | Prototype platform assets; Cilium/SPIRE content is explicitly referenced by tests and security ADRs. It cannot be deleted as a duplicate of `infra/`. |
| `integration/` | Integration-tagged control API fixtures, directly invoked by runtime CI. |
| `platform/` | Vendored charts and platform manifests. Repository instructions preserve upstream assets; topology and dependency review are needed before retiring components. |
| `apps/`, `examples/`, `bin/`, `scripts/`, `ci/`, `hack/` | Applications/examples, developer/bootstrap tools, release utilities and CI/generated-code inputs. |
| `Jenkinsfile` | Documented compatibility entrypoint in ADR 008. GitHub Actions remains the required CI/release owner. |
| `renovate.json` | Helm dependency maintenance, complementing Dependabot's supported ecosystems. |
| Agent/editor/devcontainer configuration | Intentional developer tooling, distinct from local assistant session state. |
| ADRs, plans, `docs/legacy/`, license/REUSE metadata | Recorded decisions, clearly identified historical material and provenance. Plans do not prove implemented features. |

Current cloud maturity is governed by [ADR 005](adr/005-cloud-parity-scope.md),
[architecture boundaries](architecture-runtime.md), and the [getting-started
guide](quickstart-platform.md). Use those instead of the removed parity claims.

## Keep transient files local

`.gitignore` and `.dockerignore` exclude local `.omc` state, captured `err` output,
test/linter caches and bootstrap logs/backups. The bootstrap tools still produce
their logs/backups locally; these may contain environment details and do not belong
in source control or image build contexts.

This audit removes confirmed residue; it does not declare every retained prototype
deployable or every vendored component integrated. Retiring more platform assets
requires checking relative Kustomize/Helm inputs, release contracts, existing users,
and architecture ownership rather than deleting everything with no direct import.
