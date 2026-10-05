# Mantl

Mantl is a Kubernetes platform compiler and compliance operator built around
OpenTofu/Terraform, ArgoCD, and Kyverno. It generates platform configuration,
reports policy coverage and findings, and captures scheduled, hashed evidence
outside Kubernetes etcd. The current compliance runtime is beta.

| Capability | Current scope |
| --- | --- |
| Go CLI | Spec planning, bootstrap, live status, static cloud validation |
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

Download the CLI tar.gz, deb or rpm from a tagged GitHub release and verify its
checksums and signature. Use the matching platform bundle through `--source-dir`
for bootstrap, and the digest-pinned installation manifest for the operator.
See [release installation and verification](docs/releases.md). The release
workflow is implemented; local snapshot assets do not constitute a published release.

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
`mise exec -- python ci/lock_python_dependencies.py` to regenerate the complete
hash-locked requirements. CI checks the locks for drift. Install with
`python -m pip install --only-binary :all: --require-hashes -r requirements.txt -r requirements-test.txt`.

The devcontainer builds the pinned mise toolchain and runs as `mantl`. It mounts
the checkout into `/workspace`, creates a project `.venv`, and installs local
commit hooks. Cloud credentials remain host-mounted; setup does not create a
cluster. Optional kind setup requires a pinned kind installation through mise
and an explicit `CREATE_KIND_CLUSTER=true`.

The supported entrypoint is the Go CLI. `bin/mantl` and wizard/Make installers are
legacy paths and do not define the current operator contract.

- [Installation, evidence export, and recovery](docs/compliance-operations.md)
- [Features and architecture plan](docs/superpowers/plans/2026-10-04-features-and-architecture.md)
- [Architecture decisions](docs/adr/)
- [Codex repository instructions](AGENTS.md)
- [Contribution guidance](CONTRIBUTING.md)
- [Security reporting](SECURITY.md)

[Runtime architecture and migration](docs/architecture-runtime.md) describes the
new beta adapters and their operational prerequisites. Do not infer production readiness from
a chart being present or a Terraform configuration passing static validation.
