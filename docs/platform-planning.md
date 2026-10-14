<!-- SPDX-License-Identifier: Apache-2.0 -->
# Platform specification and offline planning

`mantl plan` validates a MantlCluster specification and compiles Terraform variables,
ArgoCD Applications, and tenant manifests. It neither provisions resources nor
resolves the referenced Git content.

## Release generation

Save [the README example](../README.md#a-small-example) as `platform.yaml`, or obtain
`examples/mantl-spec.yaml` through the tagged clone in [the quickstart](quickstart-platform.md#generate-configuration-offline).
The release command is:

```sh
mantl plan platform.yaml
```

It writes `.mantl/build/terraform.tfvars.json`, an Application for
`deploy/gitops/core`, Applications for enabled features, and optional tenant files.
Variables are **inputs to existing blueprints**, not complete infrastructure code.
The AWS variables are only cluster name, region, Kubernetes version, and environment;
GCP uses `project` from `accountId`, and Azure uses `resource_group_name` and `location`.
Profile size selects the environment label; it does not size cloud node groups.
Networking fields do not automatically configure domains, VPCs, or public access.
Review provider blueprint variables separately before provisioning.

## Specification reference

The CLI reads a YAML file into [the platform API types](../apis/platform/v1alpha1/types.go)
and applies [strict parsing and validation](../pkg/compiler/spec_parser.go). It is
not a reconciled platform provisioning controller; do not assume that applying a
`MantlCluster` object provisions a cluster. Unknown fields are rejected.

| Field | Current behavior |
| --- | --- |
| `metadata.name` | Required DNS label; used as cluster name. |
| `provider.kind`, `kubernetes.distribution` | Valid pairs: local/kind, aws/eks, gcp/gke, azure/aks, do/doks, linode/lke, oci/oke, ibm/iks, openstack/kubernetes. Validation is not provider acceptance. |
| `provider.region`, `kubernetes.version` | Required strings. Local generation also requires them; no live provider version lookup occurs. |
| `provider.accountId` | Required for GCP project or Azure resource group. AWS account selection comes from credentials, not this field. |
| `profile.size` | Required: small, medium, full. Both tenant labels and Terraform use dev, staging, production. |
| `environment` | Optional override: dev, staging, production. |
| `networking.domain` | Required input; domain/exposure/vpcId are not wired into the generated Terraform variables. |
| `features` | Boolean selections generate feature Applications at the paths below; omitted values are false. |
| `gitops.repository`, `revision` | Optional; defaults to this Mantl repository and main. Explicit HTTPS repository URLs must omit credentials. Pin a reviewed revision. |
| `gitops.operatorPath` | Overrides the compliance Application source path; use an environment-specific overlay. |
| `gitops.tenantPath` | Required when tenants exist; relative path in the Git repository where generated files must be committed. |
| `controlNamespace` | Optional DNS label for the compliance operator namespace. Tenants cannot use it. kube-system, kube-public, kube-node-lease, default, argocd, and mantl-compliance-system are always refused. |
| `tenants` | Unique DNS-label names/namespaces; namespace defaults to name. The namespaces above, and `controlNamespace` when set, are refused. Admin strings become Kubernetes User subjects bound to namespace admin. |
| `profile.compliance` | Present in the API but not used to select/install a compliance profile by the compiler. Configure ComplianceProfile resources separately. |

Feature paths are `observability` → `platform/observability/prometheus/overlays/dev`,
`security` → `platform/security/base`, `secrets` → `platform/secrets/base`,
`compliance` → `deploy/operator/base` (or `operatorPath`), and
`progressiveDelivery` → `platform/progressive-delivery/base`.
These Applications reference repository assets; feature names do not prove every
component described in an old design is integrated. The root always references
`deploy/gitops/core`, including its Kyverno, cert-manager, and policy children.

Tenant output includes a Kustomization listing only desired tenants. Each tenant
gets a Namespace, optional admin RoleBinding, default-deny ingress/egress policy,
ResourceQuota, and LimitRange. Limits are fixed defaults, not profile-sized.
Add reviewed traffic rules and confirm CNI enforcement before running workloads.
Keeping tenantPath after removing the last tenant permits reconciliation of an
empty desired set; pruning can delete namespaces and workloads.

## Preview and inventory

These flags are available in **v0.5.0** and absent from v0.4.0. Install the
[verified release](quickstart-platform.md#install-a-verified-release) and obtain
the tagged example checkout as described in the quickstart. From that checkout,
preview without creating files, contacting a cluster, or running Terraform:

```sh
mantl plan examples/mantl-spec.yaml --dry-run
mantl plan examples/mantl-spec.yaml --dry-run --format json
```

Generate the reviewed artifacts in a chosen directory:

```sh
mantl plan examples/mantl-spec.yaml --output-dir build/production --format json
```

The default directory remains `.mantl/build`. Text output lists artifact paths,
SHA256 hashes, byte counts and planning notices. JSON uses the versioned
`mantl.io/plan/v1alpha2` schema and includes platform/provider/distribution,
application repository/revision/path identities, artifact metadata and notices.
Artifact contents are omitted. Preview and generation report identical metadata
for identical specs and compiler versions. JSON identifies the desired artifacts;
use the command exit status to determine whether writing succeeded.

The v0.5.0 CLI also reports limited-effect fields in `notices`: profile
size labels do not size nodes, networking fields do not configure DNS/VPC/access,
and a nonempty compliance profile does not install a `ComplianceProfile`. An AWS
account ID does not override the runtime credential account. Experimental provider
pairs carry an additional maturity notice. These notices do not change generated
artifacts; saved-plan comparisons still bind the full spec, including these fields.
The published v0.4.0 CLI does not expose this structured notice output.

The entire spec and artifact set are compiled before output files are written.
Invalid specs return an error without generating files. Disk errors can still
leave partial output; generation is not a transactional directory replacement.
Existing unrelated files are preserved. Known generated Application files are
refreshed to the selected feature set. Tenant files omitted from the generated
Kustomization index are inactive and may remain on disk.

Commit the generated tenant directory at `gitops.tenantPath` in the configured
repository before bootstrap. Review namespace admin bindings and isolation
policies as part of that Git change. The plan does not publish or apply them.

This is a compiler preview, not a Terraform resource-change plan, repository
validation, compliance evaluation or deployment acceptance check. Hashes identify
the generated bytes; they do not sign artifacts or pin remote Git contents. A
moving revision such as `main` can resolve to different source contents later.
Use `mantl status` for observed application health and the compliance commands
for separately scoped compliance results. Runtime and cloud acceptance remain beta.

## Save and compare reviewed plans

Generate artifacts and save their inventory in a separate review location:

```sh
mantl plan examples/mantl-spec.yaml --output-dir build/production --save-plan reviewed-plan.json
mantl plan examples/mantl-spec.yaml --dry-run --compare reviewed-plan.json
mantl plan examples/mantl-spec.yaml --dry-run --compare reviewed-plan.json --format json --fail-on-change
```

The comparison lists added, changed and removed artifact paths. JSON includes
before/after hashes and a `comparison` object. `inputChanged` identifies a changed
platform name or spec, including fields that currently produce no different
artifacts. An unchanged comparison exits successfully; `--fail-on-change` returns
a failure after printing the report when the input or artifacts differ. This flag
requires `--compare` and `--dry-run`, so a CI change gate leaves files unchanged.

Saved plans use `mantl.io/plan/v1alpha2`, with a `specSHA256` over the platform name
and complete spec. Regenerate earlier v1alpha1 inventories with this CLI. Save the
inventory outside the generated artifact directory and separately from the input
spec. `--save-plan` writes atomically after artifact generation succeeds and cannot
be combined with `--dry-run`. Its parent directory must already exist. Keep the
inventory in a trusted review location: checksums establish integrity, not reviewer
identity or authenticity.

## Review saved artifacts offline

The following subcommands are available in current source and are not included
in the published v0.5.0 CLI. They use the same bounded inventory and artifact
validation as reviewed apply, without needing the original specification,
kubectl, Terraform, cloud credentials or a running cluster.

Verify that the listed files still match a saved inventory before handing a
build to another operator or an apply pipeline:

```sh
mantl plan verify reviewed-plan.json --output-dir build/production
mantl plan verify reviewed-plan.json --output-dir build/production --format json
```

Missing files, changed sizes or hashes, symlinked artifact files, escaping paths
and unsupported inventory schemas return an error without a success report.
Unlisted files are ignored and remain unverified. Verification does not modify
files or recompile the spec. A successful JSON report uses
`mantl.io/plan-verification/v1alpha1`, identifies its scope as `artifact-integrity`,
and includes `verified: true` plus the inventory metadata. Artifact contents are
omitted. Integrity verification does not authenticate approval, prove that the
current spec matches, or establish infrastructure health or deployment acceptance.

Compare a baseline and candidate inventory even when their specs and generated
directories are unavailable:

```sh
mantl plan diff reviewed-plan.json candidate-plan.json
mantl plan diff reviewed-plan.json candidate-plan.json --format json --fail-on-change
```

Generate the candidate with `plan --save-plan` into a separate output directory
first. The diff reports full-spec hash changes and added, changed or removed
artifact identities; it does not read or verify the artifact files. JSON uses
`mantl.io/plan-comparison/v1alpha1` and includes before/after platform/spec
identities and a `comparison` object. `--fail-on-change` prints the complete report
before returning a failure for any input or artifact difference, so CI can retain
the report. An unchanged comparison succeeds. Both subcommands leave inventories
and generated files untouched. Check the exit status as well as the report.

## Apply reviewed inputs

Apply the reviewed inputs using an explicitly selected target:

```sh
mantl --context TARGET --source-dir PLATFORM_SOURCE apply examples/mantl-spec.yaml \
  --plan reviewed-plan.json --output-dir build/production --timeout 30m
```

Apply requires `--plan` and recompiles the current spec. The full inventory must match the
saved plan. Apply without that file stops before preflight. Every listed regular artifact must match its size and SHA256. Missing,
tampered, symlinked or escaping files stop apply before preflight or provisioning.
Verified bytes are copied into a private temporary directory for execution, so
later changes to reviewed files cannot change execution inputs. Extra YAML files
in the output directory are excluded from execution. The private directory is
removed after success or failure.

Before generating files or changing infrastructure, preflight checks kubectl,
kind/container-runtime availability for local clusters, or the source blueprint,
Terraform/OpenTofu executable and locally configured cloud kube-context. Cloud
bootstrap requires that named context to be prepared already. Verify its actual
cluster identity independently. Local bootstrap derives `kind-PLATFORM_NAME`
and rejects a conflicting explicit context without changing global CLI state.
The total deadline is configurable from a positive duration up to two hours;
convergence also retains its ten-minute stage limit. Each failed or cancelled
stage stops subsequent stages and preserves its underlying error. Ctrl+C and
SIGTERM cancel the CLI context, stop managed child commands and run snapshot cleanup. Partial changes
remain available for diagnosis and recovery; rollback is an operator decision.

Apply without `--plan` generates fresh inputs after preflight and uses the same
private execution snapshot. Prior review is established only by the saved-plan
workflow. Bootstrap applies the expected Applications only and verifies every
expected source identity and explicit Synced/Healthy status, sharing the contract
with `mantl status`. Removed live Applications need a reviewed GitOps cleanup;
comparison removals alone do not delete them. Tenant removals continue through
the committed tenant Kustomization and ArgoCD pruning.

The saved inventory binds generated inputs, not Terraform resource changes,
remote Git contents, source blueprints, dependency binaries or provider state.
Review those separately for deployment. Cloud deployment, recovery and production
acceptance remain pending; the new CLI capabilities remain beta.
