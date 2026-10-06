# ADR 011: Reviewed platform plans and bootstrap inputs

Status: Accepted for implementation; cloud deployment acceptance remains pending.

## Context

Offline plans identify generated artifact bytes, but apply recompiles without a
reviewed baseline, applies every YAML file in a build directory, and checks
convergence without confirming that every desired Application exists.

## Decision

Keep compiler contracts in `pkg/compiler`, bootstrap execution in `pkg/bootstrap`,
and orchestration in the CLI. Introduce `mantl.io/plan/v1alpha2` inventories with a
SHA256 of the platform name and complete spec. Metadata and artifact comparisons
are deterministic. Earlier v1alpha1 inventories must be regenerated.

`plan --save-plan` writes an inventory after generation succeeds. A dry run never
writes a saved plan. `--compare` reports added, changed and removed artifact
identities; a read-only change gate may return a failure after printing the diff.

`apply --plan` requires the current compiler/spec to reproduce the entire saved
inventory. All listed regular files must match their sizes and hashes. Root-bound
file reads reject escaping paths and symlinked artifacts. Verified bytes are
copied into a private temporary build directory before execution; later changes
to reviewed files cannot alter the bytes passed to bootstrap. Extra build files
are not executed. Apply without a saved plan remains available and uses the same
private execution snapshot, but does not claim prior review.

Preflight validates tools, source blueprint availability and an explicit cloud
kube-context before generating files or provisioning. The local provider derives
its own kind context without changing global CLI state. A configurable total
context deadline covers preflight and execution. Steps stop on error/cancellation;
no automatic destructive rollback is attempted.

Bootstrap receives the compiler's desired Application topology, applies only its
named manifests, and uses the same source-identity and health evaluation as
`mantl status`. Missing or mismatched Applications cannot establish convergence.

## Limits

Local inventories are integrity records, not authenticated approvals or signatures.
Keep them in a trusted review location. They bind generated inputs, not Terraform
resource plans, provider state, source blueprints, remote Git contents or dependency
binaries. A moving Git revision remains mutable. A named kube-context does not
prove a cloud cluster's identity; operators must independently verify the target.
Terraform/OpenTofu may change infrastructure during apply; planning comparisons
are artifact differences, not infrastructure drift detection. Local tests exercise
injected executors and fake tools; cloud deployment and recovery acceptance remain
separate. ArgoCD keeps policy ownership under ADR 006.
