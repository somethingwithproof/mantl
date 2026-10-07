<!-- SPDX-License-Identifier: Apache-2.0 -->
# ADR 009: GitHub Actions validation ownership

Status: Accepted

## Context

Mantl's workflows duplicate Go and Python testing, provision optional cloud
runners for work that GitHub-hosted runners can perform, and mix infrastructure
validation with tests requiring clusters that CI does not create.

## Decision

GitHub-hosted runners execute the default CI pipeline. A repository variable can
opt eligible trusted Linux events into a dedicated, ephemeral DigitalOcean pool
after provisioning, runtime compatibility and deletion have been verified.
Fork and metadata events retain GitHub-hosted routing. A separate controller
owns capacity and cleanup; workflows receive no cloud provisioning credentials.
See [the runner rollout contract](../digitalocean-runners.md).
One core workflow owns
first-party Go testing and coverage, Python unit testing and coverage, and the
SonarCloud quality gate. Sonar consumes those reports rather than rerunning tests.
The final CI check fails when an unconditional prerequisite fails or a selected
component check fails; intentionally unselected component jobs may be skipped.

Reusable component workflows own infrastructure, chart, disposable integration,
image and package validation. Relevant paths select these checks inside the core
workflow so required checks are always created. Runtime versions come from mise;
dependencies and build layers are cached without persisting credentials.

Mutation testing and benchmarks run on schedules or manual dispatches. Cluster
acceptance requires an explicitly provisioned environment and recorded receipts;
scheduled CI does not claim cloud acceptance from unprovisioned cluster tests.

Native Dependabot owns dependency updates. Release Please owns SemVer tags and
invokes the existing exact-tag artifact release workflow. Signing, immutable
content publication, provenance and scoped write permissions remain required.

## Consequences

Maintainers get stable CI and quality-gate checks, fewer duplicate builds, and
component-specific failures. Opted-in jobs allocate one-job DigitalOcean VMs
through the external controller, capped independently of workflow concurrency. The
manual orphan-runner cleanup remains available for legacy tagged CI droplets.
Fork pull requests still require a reviewed same-repository branch for Sonar
authentication. Optional example applications retain separate validation scope.


## PR selection and scheduling refinement

Keep path selection and the final aggregate check on GitHub-hosted runners so
control jobs do not wait for or consume ephemeral DigitalOcean capacity. Runtime
and application work retains the existing trusted-event routing and capacity cap.
Do not change the controller or cancel other runs to prioritize a PR.

Go/Python core coverage and Sonar remain mandatory. PRs test affected Python
examples; shared fixtures, dependencies, runtime pins, and core workflow changes
select every example. Main pushes and standalone validation run all examples so
main's coverage remains complete. PR Sonar input contains coverage only for the
examples tested by that PR; a documentation-only PR has no example artifact.
Sonar and the final CI check must explicitly accept only an intentionally
unselected/skipped example matrix, and reject selected skipped, failed, or cancelled
jobs. Markdown/RST reference edits do not select runtime/package checks, except
framework and template content whose bytes are part of a delivered bundle.

## Tool-aware selection and release preflight deduplication

The core Go, Python, repository, Sonar and aggregate checks remain mandatory.
Changes limited to known tool pins select additional components that consume
those tools. Compare parsed `mise.toml` at the verified base and checked-out
commit, including additions and removals. Unknown tools, unavailable/invalid
configuration, or changes outside `[tools]` retain the complete component
selection. Shared action and selector changes also retain full validation.
Python runtime changes select every Python example; other known tool changes
do not select that matrix. Main and standalone runs keep the full example baseline.

Release Please's preflight remains advisory: manual dispatch does not satisfy
required PR checks. Before dispatch, look for an existing `pull_request` CI run
for the same repository PR, branch and head. Reuse that run regardless of its
state; approve or rerun the eligible PR workflow when needed instead of starting
a parallel manual run. A short bounded lookup allows PR-run creation to settle.
Never dispatch after the PR closes or its head changes. Retain manual preflight
as a fallback only when no matching PR run exists.

Release Please and Dependabot auto-merge perform repository metadata operations
on GitHub-hosted runners. Artifact publication retains its existing runner
routing, exact-tag checks, permissions, signing and provenance requirements.
