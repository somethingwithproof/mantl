<!-- SPDX-License-Identifier: Apache-2.0 -->
# Platform specification and offline planning

`mantl plan` validates a MantlCluster specification and compiles Terraform variables,
ArgoCD Applications, and tenant manifests. It neither provisions resources nor
resolves the referenced Git content.

## Published v0.4.0 generation

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
| `profile.size` | Required: small, medium, full; labels map to dev, staging, production (Terraform uses prod). |
| `environment` | Optional override: dev, staging, production. |
| `networking.domain` | Required input; domain/exposure/vpcId are not wired into the generated Terraform variables. |
| `features` | Boolean selections generate feature Applications at the paths below; omitted values are false. |
| `gitops.repository`, `revision` | Optional; defaults to this Mantl repository and main. Explicit HTTPS repository URLs must omit credentials. Pin a reviewed revision. |
| `gitops.operatorPath` | Overrides the compliance Application source path; use an environment-specific overlay. |
| `gitops.tenantPath` | Required when tenants exist; relative path in the Git repository where generated files must be committed. |
| `tenants` | Unique DNS-label names/namespaces; namespace defaults to name. Admin strings become Kubernetes User subjects bound to namespace admin. |
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

## Main-only preview and inventory

These flags are implemented on main and are **absent from published v0.4.0**.
Build using [the contributor instructions](../CONTRIBUTING.md#build-and-test).
From a source checkout, preview without creating files, contacting a cluster,
or running Terraform:

```sh
/tmp/mantl-main plan examples/mantl-spec.yaml --dry-run
/tmp/mantl-main plan examples/mantl-spec.yaml --dry-run --format json
```

Generate the reviewed artifacts in a chosen directory:

```sh
/tmp/mantl-main plan examples/mantl-spec.yaml --output-dir build/production --format json
```

The default directory remains `.mantl/build`. Text output lists artifact paths,
SHA256 hashes, byte counts and planning notices. JSON uses the versioned
`mantl.io/plan/v1alpha2` schema and includes platform/provider/distribution,
application repository/revision/path identities, artifact metadata and notices.
Artifact contents are omitted. Preview and generation report identical metadata
for identical specs and compiler versions. JSON identifies the desired artifacts;
use the command exit status to determine whether writing succeeded.

Current source builds also report limited-effect fields in `notices`: profile
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

## Save and compare reviewed plans (main only)

Generate artifacts and save their inventory in a separate review location:

```sh
/tmp/mantl-main plan examples/mantl-spec.yaml --output-dir build/production --save-plan reviewed-plan.json
/tmp/mantl-main plan examples/mantl-spec.yaml --dry-run --compare reviewed-plan.json
/tmp/mantl-main plan examples/mantl-spec.yaml --dry-run --compare reviewed-plan.json --format json --fail-on-change
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

Apply the reviewed inputs using an explicitly selected target:

```sh
/tmp/mantl-main --context TARGET --source-dir PLATFORM_SOURCE apply examples/mantl-spec.yaml \
  --plan reviewed-plan.json --output-dir build/production --timeout 30m
```

Apply recompiles the current spec and requires its full inventory to match the
saved plan. Every listed regular artifact must match its size and SHA256. Missing,
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
