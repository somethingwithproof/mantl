<!-- SPDX-License-Identifier: Apache-2.0 -->
# Mantl documentation

<img src="_static/mantl-mark.svg" alt="Mantl" width="240" />

Use this index for the current Kubernetes compiler and beta compliance runtime.
The [project README](../README.md#capabilities-and-maturity) has the single capability
and maturity table. Published CLI examples target **v0.4.0**; main-only features are
identified in their guides. Design records and framework/chart assets are not
proof of deployed capabilities.

## Evaluate and get started

- [Release quickstart](quickstart-platform.md): verify a download and generate configuration offline.
- [Platform specification and planning](platform-planning.md): fields, examples, generated outputs, and main-only previews, saved plans, comparisons, and reviewed apply.
- [Release assets and verification](releases.md): archives, DEB/RPM, signatures, SBOMs, provenance, and install manifests.
- [Bootstrap prerequisites](quickstart-platform.md#deployment-is-a-separate-step): infrastructure, credentials, Git publication, cluster access, and costs.
- [Platform status](platform-status.md): scoped Application identities, JSON output, and the health gate.

## Understand responsibilities and compliance

- [Architecture diagrams](architecture-diagram.md): configuration ownership and evidence lifecycle.
- [Runtime architecture](architecture-runtime.md): durable collection, identities, exceptions, fleet boundaries, and OSCAL limitations.
- [Compliance operations](compliance-operations.md): control coverage, findings, freshness, evidence storage/retention, verification/export, and installation.
- [Compliance content](../compliance/README.md): catalogs, profiles, policies, and content pins.
- [Provider scope](adr/005-cloud-parity-scope.md) and [acceptance limits](architecture-runtime.md#provider-acceptance-and-oscal): AWS/GCP/Azure static contracts; other providers experimental; AWS S3 evidence backend.
- [Policy ownership](adr/006-policy-application-ownership.md) and [evidence storage contract](adr/003-evidence-storage-contract.md).

## Operate and contribute

- [Operations, upgrades, backup, recovery](runbooks.md): operator procedures and required environment-specific validation.
- [Troubleshooting](troubleshooting.md): generation, GitOps, findings, storage, and diagnostic boundaries.
- [Contributing](../CONTRIBUTING.md) and [testing](testing.md): pinned toolchain, scoped checks, locks, devcontainers, and disposable integration fixtures.
- [GitHub Actions inventory](../.github/workflows/README.md), [required checks](ci-required-checks.md), and [SonarCloud](sonarcloud.md).
- [License declarations and maintenance](licensing.md): SPDX headers, generated files and upstream exceptions.
- [Security reporting](../SECURITY.md), [architecture decisions](adr), and [roadmap/design specs](superpowers/specs).

## Additional assets and documentation gaps

[Backstage](backstage.md), [secrets and issuers](secrets-and-issuers.md),
[provider overlays](overlays.md), and [ExternalDNS identity](external-dns-identity.md)
describe optional configuration assets that need operator integration.
The directory name `clusters/production` is a configuration layout, not a maturity rating.

The older [ARA designs](ara/README.md), [stack decision](CNCF-STACK-DECISION.md),
[platform-app catalog](platform-apps.rst), and [multi-cloud notes](multicloud.rst)
are design/reference material. They do not establish current support, automatic
failover, or Crossplane integration. [Legacy documentation](legacy/README.md)
covers the original Mesos/Marathon/Ansible system and is not a current setup path.

Still needed: recorded cloud deployment/identity/upgrade/recovery acceptance,
validated environment-specific runbooks, complete framework coverage assessments,
and native GCS/Azure immutable evidence backends. The current guides explain
implemented boundaries; these gaps are not filled by architecture plans.
