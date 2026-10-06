<!-- SPDX-License-Identifier: Apache-2.0 -->
# Compliance installation and operations

The runtime is beta. Terraform validation is static evidence; deployment,
identity federation, restore, and cloud audit delivery require environment-specific
acceptance. SOC2 policy coverage is not SOC2 certification.

## Reading compliance results

A **control** is a catalog requirement. A **profile** selects controls and namespace
scope; reviewed mappings bind those controls to policies/rules and collectors.
The operator evaluates that configured scope, not every requirement of an external
framework or every workload in your organization.

- **Coverage** records available mappings, policy/rule observations, and collector
  support. Missing mappings/reports, denied reads, unsupported collectors, and
  upload failures are gaps. Configured coverage is not proof of complete inventory coverage.
- **Findings** record observed violations and their lifecycle. Losing a source
  report makes its prior finding unknown; it does not resolve the violation.
- **Freshness** records whether observations/evidence meet their age requirements.
  Missing timestamps and stale evidence cannot establish a current passing result.
- **Exceptions** need owner, justification, scope, expiry, and separate approval
  of the current spec generation. They annotate a failure without converting it
  to pass; whole-control exceptions currently drive evaluation annotation.
- **Evidence** is a retained point-in-time snapshot with exact version/hash
  references. It supports review of what was captured, not certification or proof
  that unobserved controls are satisfied.

`ControlEvaluation` keeps result, coverage, freshness, evidence, and exceptions
separate. Inspect all dimensions rather than treating a single score as compliance.
`mantl status` observes Application health and report counts; its health gate is
not a compliance gate. See [the lifecycle diagram](architecture-diagram.md#evidence-lifecycle).

## Install

Install a verified CLI release archive or Linux deb/rpm package; see
[release verification](releases.md). For source development use `mise install`,
then `mise exec -- go install ./cmd/mantl`. The shell
`bin/mantl` is a legacy installer; the supported Go entrypoint provides `plan`,
`apply`, `status`, `validate-clouds`, and `compliance` commands.

For v0.5.0, use the verified `operator-install.yaml` and image digest recorded in
`release-identity.json`, together with your environment configuration. For source
customization, build `Dockerfile.compliance-operator` and record your own image digest.
Create a reviewed overlay including `deploy/operator/base`, setting that digest,
and generating the `evidence-storage` ConfigMap with `bucket` and `kms-key` literals.
Neither value is an API credential. Configure workload identity
for the operator ServiceAccount; do not put access keys in manifests.

The implemented backend uses the AWS S3 SDK and S3 Object Lock for **all cluster
providers**. No native GCS or Azure immutable-storage adapter is implemented.
You provision the bucket and runtime identity; bootstrap does not create them.
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
exclusions and enforcement modes before adoption. The matching release platform bundle supplies the source-directory blueprints
needed for cloud `apply`. Tenant environment labels follow profile size by default
(small/dev, medium/staging, full/production); `spec.environment` overrides this.

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

Status returns profiles, findings, audit schedules, historical runs, control
evaluations and exceptions as separate JSON dimensions. Existing
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

## Evidence retention and recovery

Retention is write-once, read-many (WORM): every uploaded version receives
COMPLIANCE retention for at least 2557 days, independently of Kubernetes object
lifetime. The current operator does not expose a MantlCluster retention setting.
Account for storage costs and legal requirements before enabling collection;
deleting a schedule or uninstalling Mantl does not remove retained objects.
There is no automatic evidence-retirement workflow. Back up version/reference
inventories and metadata so exports/history remain discoverable after recovery.

Hashes detect changed bytes, not whether a control assessment was complete or
correct. Export validates the stored manifest and referenced versions; preserve
source image/content identities and the approved scope alongside the archive.
Keep archives access-controlled because resource/RBAC snapshots can be sensitive.

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
than passing. The published release path and verification requirements are documented in
[releases.md](releases.md); a snapshot build does not establish hosted signing or publication.

Cloud identity and encryption edits still require the repository's pre-merge
security review. Static validation cannot establish actual role propagation,
private endpoint reachability, admission behavior, or recovery success.

CI retains Checkov SARIF reports as `terraform-security-sarif` artifacts for 14
days. Set the repository variable `ENABLE_CODE_SCANNING_SARIF=true` only after
GitHub code scanning is available to upload those reports into the Security tab.
Scan execution and report retention are required even when that integration is off.

See [the new runtime architecture](architecture-runtime.md) for collector identities,
content pins, exception approvals, immutable history, optional fleet setup and
provider acceptance limitations.
