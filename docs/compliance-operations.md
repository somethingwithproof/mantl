# Compliance installation and operations

The runtime is beta. Terraform validation is static evidence; deployment,
identity federation, restore, and cloud audit delivery require environment-specific
acceptance. SOC2 policy coverage is not SOC2 certification.

## Install

Use `mise install`, then `mise exec -- go install ./cmd/mantl`. The shell
`bin/mantl` is a legacy installer; the supported Go entrypoint provides `plan`,
`apply`, `status`, `validate-clouds`, and `compliance` commands.

Build `Dockerfile.compliance-operator` and push the image to your registry. Record
its digest. Create an operator overlay that includes `deploy/operator/base`, sets
that digest, and generates the `evidence-storage` ConfigMap with `bucket` and
`kms-key` literals. Neither value is an API credential. Configure workload identity
for the operator ServiceAccount; do not put access keys in manifests.

The bucket must have versioning and Object Lock enabled. Uploads explicitly use
COMPLIANCE retention for at least 2557 days and encryption. Optional KMS requires
scoped GenerateDataKey/Decrypt permissions. The operator requires bucket
GetBucketVersioning/GetObjectLockConfiguration and object PutObject/GetObjectVersion
on its `audits/` prefix, with PutObjectRetention permission for explicit retention.
It requires no object deletion or retention-bypass permission.

Install Kyverno and `deploy/gitops/policies` through ArgoCD before starting the
operator. The base includes generated compliance CRDs and runtime RBAC.
Apply the SOC2 standard profile; edit its namespaces for your environment.
The operator defaults to namespaced reads. To collect ClusterRole/ClusterRoleBinding
evidence, include `deploy/operator/cluster-evidence` and explicitly add
`--allow-cluster-evidence` to the operator arguments.

Set `spec.gitops.repository`, `revision`, `operatorPath`, and `tenantPath` in a
MantlCluster spec. Commit generated `.mantl/build/tenants` at tenantPath before
bootstrap; ArgoCD does not read local build files. The built-in topology installs
Kyverno/cert-manager and the reviewed SOC2 policy bundle. Review namespace
exclusions and enforcement modes before adoption. Existing source-directory
blueprints are needed for cloud `apply`; release packaging is separate work.

Cloud bootstrap requires `--context` and accepts `--source-dir`. The target
context must already authenticate to the provisioned cluster; Terraform outputs
are not silently installed into the user's kubeconfig. Local bootstrap targets
`kind-<cluster-name>`. ArgoCD is pinned to v3.5.3; repeated namespace installation
is declarative and the full apply has a 30-minute deadline.

## Observe and export

```
mantl --context TARGET compliance status --namespace mantl-compliance-system
mantl --context TARGET compliance remediation
mantl --context TARGET compliance export soc2-scheduled --output evidence.tar.gz
mantl validate-clouds --source-dir .
```

Status returns separate profile coverage, findings, and audits as JSON. Existing
`mantl status SPEC` reports live ArgoCD health and aggregate PolicyReport results;
that score does not claim framework certification. Audit CoverageGaps identifies
unsupported collectors, denied reads, and upload failures. Framework cron schedules
are UTC, coalesce missed intervals, and do not overlap for one audit. Unsupported
log/metrics/query collectors remain visible rather than being executed as scripts.

Export retrieves exact object versions, validates the manifest and every SHA256,
and atomically publishes an archive containing `manifest.json` and hashed objects.
Use runtime AWS credentials through workload identity or the standard credential
chain; export never prints credentials or raw evidence.

Install `deploy/operator/monitoring` only where Prometheus Operator CRDs exist.
Configure its ServiceMonitor selector in your Prometheus deployment. Alerts cover
missing/stale evidence, coverage gaps, failed audits, and metadata inspection errors.
The evidence-age threshold is a default to tune for your collector schedules.

## Recovery checks

- An unavailable object store leaves the audit Running and retryable, with no new
  success URI. Restore bucket access and observe the same StartTime run complete.
- Failed snapshots stay visible as failed/partial audits; collector clocks do not
  advance for a failed collector. Repair scoped permissions and requeue the audit.
- An operator restart preserves persisted run identity and collector timestamps.
- Report deletion marks prior findings unknown. A current passing report resolves
  a finding; losing its source is not evidence of compliance.
- Startup requires Kyverno report CRDs by default. Kubernetes retries the operator
  until they exist, so bootstrap cannot silently omit finding watches. The explicit
  `--require-policy-reports=false` compatibility mode exposes a disabled-watch
  metric; install the CRDs and restart before relying on findings.
- Before an upgrade, back up profile/audit/finding metadata and record immutable
  evidence versions. Restore metadata into an isolated environment, verify archive
  hashes, and confirm scoped identity access before resuming scheduled audits.

These are operational procedures, not a claim that cloud recovery has been run.

## Runtime events and developer workflows

`mantl compliance ingest-falco EVENT.json` validates one bounded JSON event and
emits a Finding manifest with a reviewed SOC2 rule mapping. Review/import it through
your trusted event delivery path. The command does not create an unauthenticated
receiver or apply manifests. Raw Falco command lines/output are not stored.

The Backstage `templates/backstage/web-service/template.yaml` renders a real
skeleton and opens a GitHub pull request. Enable the fetch:template and
publish:github:pull-request actions with a scoped GitHub integration. Register the
resulting catalog-info.yaml after review; this does not deploy a Backstage server.
Generated services require a ghcr.io image digest, probes, restricted containers,
and resource limits. Tenant namespaces have quotas, LimitRanges, and default-deny
network policies; add explicit DNS and service traffic rules through GitOps before
running an application. These protections require a NetworkPolicy-enforcing CNI.

Replanning lists only the current spec's tenants in the generated Kustomization;
stale and unrelated YAML files are left on disk but excluded. Keep `tenantPath`
configured when removing the last tenant so its ArgoCD application can reconcile
the empty desired state. Review tenant removal before committing: automated
pruning can delete the tenant namespace and its workloads, not just access grants.

## Implementation review boundaries

The local implementation review checked that GitOps remains the policy owner,
Secret/ConfigMap collection is denied, workload payloads omit credential-bearing
pod templates and annotations, cluster evidence requires an opt-in, retention is
verified after upload, exports verify hashes, and report loss is unknown rather
than passing. Signing and release publication were not changed.

Cloud identity and encryption edits still require the repository's pre-merge
security review. Static validation cannot establish actual role propagation,
private endpoint reachability, admission behavior, or recovery success.

CI retains Checkov SARIF reports as `terraform-security-sarif` artifacts for 14
days. Set the repository variable `ENABLE_CODE_SCANNING_SARIF=true` only after
GitHub code scanning is available to upload those reports into the Security tab.
Scan execution and report retention are required even when that integration is off.
