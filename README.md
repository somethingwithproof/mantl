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
| Evidence | Versioned S3 Object Lock objects, minimum retention, manifests and verified export |
| Cloud blueprints | AWS/GCP/Azure static contracts; deployment and recovery acceptance remain environment-specific |
| Other clouds | Experimental; no parity guarantee |
| Tenant isolation | Namespace RBAC, default-deny network policy, quotas and limits |
| Runtime signals | Falco JSON-to-Finding ingestion with reviewed mappings |
| Developer experience | Backstage pull-request scaffold; portal deployment is not included |
| HIPAA/PCI/CIS | Framework assets exist; each needs separate end-to-end acceptance |

Policy checks and evidence collection support an audit workflow; they do not
constitute SOC2, HIPAA, or PCI certification.

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

The supported entrypoint is the Go CLI. `bin/mantl` and wizard/Make installers are
legacy paths and do not define the current operator contract.

- [Installation, evidence export, and recovery](docs/compliance-operations.md)
- [Features and architecture plan](docs/superpowers/plans/2026-10-04-features-and-architecture.md)
- [Architecture decisions](docs/adr/)
- [Contribution guidance](CONTRIBUTING.md)
- [Security reporting](SECURITY.md)

Release packaging is tracked separately. Do not infer production readiness from
a chart being present or a Terraform configuration passing static validation.
