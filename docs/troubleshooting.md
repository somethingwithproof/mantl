# Troubleshooting Mantl

Start with [the release quickstart](quickstart-platform.md),
[platform planning](platform-planning.md), and [compliance operations](compliance-operations.md).
Record `mantl --version`, the specification/revision, and the exact error.
Do not paste credentials, raw evidence, or sensitive manifests into public issues.

## Offline generation

**Unknown flag:** v0.4.0 `plan` has no `--dry-run`, `--format`, or `--output-dir`.
Use `mantl plan platform.yaml`, or build main for the extra flags. Saved-plan
comparison and `apply --plan` also require main. Main inventories use v1alpha2;
regenerate older v1alpha1 inventories. The legacy
`bin/mantl` shell installer is not the supported Go CLI.

**Invalid specification:** check required fields, provider/distribution pairing,
size, tenant uniqueness, HTTPS repository URL, and relative Git paths.
GCP/Azure require `provider.accountId`; tenants require `gitops.tenantPath`.
Unknown fields are rejected. Generation does not validate whether a cloud accepts
the requested Kubernetes version or whether a Git path exists.

**Unexpected files:** v0.4.0 writes `.mantl/build` in your current directory.
Known generated Application files are refreshed, while unrelated/stale tenant
files can remain. The tenant Kustomization lists only desired files. Use separate
working directories for separate specs; inspect the index before committing output.
Main output inventories identify desired bytes, not every file already on disk.

**Terraform cannot provision from the build directory:** the generated JSON is
only variables. Use the matching verified source bundle's blueprint, review its
other inputs and state, and distinguish `mantl plan` from Terraform's resource plan.

## Bootstrap and GitOps

Cloud `apply` needs OpenTofu/Terraform, kubectl, an explicit authenticated context,
provider identity, and a matching `--source-dir`. It does not install outputs into
kubeconfig. Private endpoints need working network access. Local apply needs
Docker and kind and targets `kind-<metadata.name>`.

**Tenant Application cannot sync:** commit the generated tenant directory at the
configured repository/revision/path, not just on local disk. Check ArgoCD repository
credentials and path existence. The README's tenant path is illustrative and is
not included in the upstream release. Namespace pruning can delete workloads.

**Application looks healthy but status fails:** `status --require-healthy` checks
all directly generated expected Application names and source/destination identities.
Missing, duplicated, incomplete, or mismatched observations fail the gate.
Unrelated Applications do not satisfy the expected set. Review
[status semantics](platform-status.md); nested workloads require their own checks.

**Timeout/query unknown:** explicit `--timeout` is bounded to one minute (default
five seconds). Verify context, read permissions, API reachability, and report CRDs.
Without the application gate, unknown observations can still exit successfully;
inspect JSON state rather than exit status alone.

## Compliance and evidence

**Operator does not start:** Kyverno report CRDs are required by default. Install
reporting before relying on findings. The compatibility flag
`--require-policy-reports=false` disables that watch and emits a disabled-watch
metric; it is not a fix for missing compliance visibility.

**Coverage unknown/partial or stale:** check configured mappings, supported report
version (`wgpolicyk8s.io/v1alpha2`), namespace scope, timestamps, collector support,
RBAC, network egress, and upload failures. An installed policy or a framework file
alone cannot establish coverage. Unsupported log/metrics/query collectors remain gaps.

**Audit stuck Running:** unavailable storage can leave a run retryable. Restore
versioned, Object-Locked S3 access and scoped workload identity; observe the same
persisted run rather than treating pod restart as completion. Failed collectors
retain gaps and do not advance successful collection timestamps.

**Upload/export/history verification fails:** verify bucket versioning, Object Lock
COMPLIANCE retention (at least 2557 days), exact object versions/hashes, prefix
permissions, and optional KMS permissions. Native GCS/Azure storage adapters do not
exist. Do not bypass hash/retention checks or substitute a latest unversioned object.

**Exception does not make a result pass:** expected behavior. Approval requires a
separate approver role and current spec generation; it annotates the failure.
Resource-scoped exceptions do not currently annotate whole-control evaluations.

**Fleet cannot start/connect:** the base deliberately needs image, enrollment,
TLS, database, storage, identity, and network overlays. Check verified PostgreSQL
TLS and non-bypass RLS roles; TCP readiness does not prove recovered fleet state.
Follow [fleet responsibilities](architecture-runtime.md#optional-fleet-metadata-service).

## Reporting a problem

Open an [issue](https://github.com/somethingwithproof/mantl/issues) with version,
redacted inputs, reproducible commands, error/status dimensions, and expected behavior.
Use [private security reporting](../SECURITY.md) for vulnerabilities.
Old [component runbooks](runbooks/troubleshooting-guide.md) are historical examples;
check installed versions and ownership before applying their procedures.
