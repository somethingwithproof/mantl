# ADR 009: GitHub Actions validation ownership

Status: Accepted

## Context

Mantl's workflows duplicate Go and Python testing, provision optional cloud
runners for work that GitHub-hosted runners can perform, and mix infrastructure
validation with tests requiring clusters that CI does not create.

## Decision

GitHub-hosted runners execute the default CI pipeline. One core workflow owns
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
component-specific failures. New CI runs create no DigitalOcean droplets. The
manual orphan-runner cleanup remains available for legacy tagged CI droplets.
Fork pull requests still require a reviewed same-repository branch for Sonar
authentication. Optional example applications retain separate validation scope.
