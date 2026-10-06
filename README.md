# Mantl

Mantl is a Kubernetes platform compiler and compliance operator built around
OpenTofu/Terraform, ArgoCD, and Kyverno. It generates platform configuration,
reports policy coverage and findings, and captures scheduled, hashed evidence
outside Kubernetes etcd. The current compliance runtime is beta.

| Capability | Current scope |
| --- | --- |
| Go CLI | Offline spec previews and artifact hashes, bootstrap, scoped JSON status, static cloud validation |
| GitOps | ArgoCD owns policy deployment and lifecycle |
| SOC2 | Packaged controls/policies, profile coverage, findings and config snapshots; unsupported collectors are reported as gaps |
| Audit execution (beta) | Durable AuditRuns and isolated, scoped collector Jobs |
| Evaluations (beta) | Coverage/freshness states, approved exceptions and linked finding history |
| Fleet API (beta) | Enrolled mTLS metadata ingestion, tenant PostgreSQL RLS and replay |
| Evidence | Versioned S3 Object Lock objects, minimum retention, manifests and verified export |
| Cloud blueprints | AWS/GCP/Azure static contracts; deployment and recovery acceptance remain environment-specific |
| Other clouds | Experimental; no parity guarantee |
| Tenant isolation | Namespace RBAC, default-deny network policy, quotas and limits |
| Runtime signals | Falco JSON-to-Finding ingestion with reviewed mappings |
| Developer experience | Backstage pull-request scaffold; portal deployment is not included |
| HIPAA/PCI/CIS | Framework assets exist; each needs separate end-to-end acceptance |

Policy checks and evidence collection support an audit workflow; they do not
constitute SOC2, HIPAA, or PCI certification.

## Installation

The current published release is [v0.4.0](https://github.com/somethingwithproof/mantl/releases/tag/v0.4.0).
It provides Linux/macOS amd64/arm64 CLI tar.gz archives, Linux deb/rpm packages,
and operator/platform assets with signatures, checksums, SBOMs and provenance.
Download the matching CLI and authenticate the signed checksum manifest before
checking downloaded files. Use the matching platform bundle through `--source-dir`
for bootstrap, and the digest-pinned installation manifest for the operator.
See [release installation and verification](docs/releases.md). Local snapshot assets do not constitute a published release.

## Plan and inspect a platform

The next CLI release adds offline previews and a JSON artifact inventory. These
commands currently require a build from main after this feature is merged;
v0.4.0 supports the original `mantl plan SPEC` generation command.

```sh
mantl plan examples/mantl-spec.yaml --dry-run --format json
mantl plan examples/mantl-spec.yaml --output-dir build/production
mantl status examples/mantl-spec.yaml --context TARGET --format json --require-healthy
```

Previewing writes no files and makes no cluster or Terraform calls. Generation
lists exact artifact hashes. Commit tenant manifests to the configured GitOps
repository before bootstrap. Planning does not establish cloud readiness or
compliance. Status checks the generated Application topology; policy counts are
reported separately from application health.

## Development

```sh
mise install
mise exec -- make go-test
mise exec -- go install ./cmd/mantl
mise exec -- go run ./cmd/mantl --help
mise exec -- go run ./cmd/mantl validate-clouds
```

Use `make go-test` to exclude vendored chart tests. Regenerate API artifacts with
`mise exec -- make manifests generate` after changing CRDs. `platform/` contains
vendored components; maintain Mantl behavior under `pkg`, `controllers`, and `deploy`.

Python dependency inputs are `requirements.in` and `requirements-test.in`; each
Python example has its own `requirements.in`. After updating a direct pin, run
`mise exec -- python -m ci.lock_python_dependencies` to regenerate the complete
hash-locked requirements. CI checks the locks for drift. Install with
`python -m pip install --only-binary :all: --require-hashes -r requirements.txt -r requirements-test.txt`.

The devcontainer builds the pinned mise toolchain and runs as `mantl`. It mounts
the checkout into `/workspace`, creates a project `.venv`, and installs local
commit hooks. Cloud credentials remain host-mounted; setup does not create a
cluster. Optional kind setup requires a pinned kind installation through mise
and an explicit `CREATE_KIND_CLUSTER=true`.

Kubectl invocations use fixed installation directories, without searching `PATH`.
For a mise-managed binary, set `MANTL_KUBECTL_PATH="$(mise which kubectl)"` before
using cluster-facing CLI commands. The override must be an absolute path to a
regular executable without group or world write permission.

The supported entrypoint is the Go CLI. `bin/mantl` and wizard/Make installers are
legacy paths and do not define the current operator contract.

- [Installation, evidence export, and recovery](docs/compliance-operations.md)
- [Offline platform planning](docs/platform-planning.md)
- [Features and architecture plan](docs/superpowers/plans/2026-10-04-features-and-architecture.md)
- [Architecture decisions](docs/adr/)
- [Codex repository instructions](AGENTS.md)
- [Contribution guidance](CONTRIBUTING.md)
- [Security reporting](SECURITY.md)

[Runtime architecture and migration](docs/architecture-runtime.md) describes the
new beta adapters and their operational prerequisites. Do not infer production readiness from
a chart being present or a Terraform configuration passing static validation.

The beta CLI also provides [scoped platform status](docs/platform-status.md) with
JSON output and an explicit application health gate for CI.
