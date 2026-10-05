# Mantl GitHub Actions

Core CI owns first-party tests and coverage. Optional component checks are selected
inside the workflow, so required checks exist even when a pull request changes
only documentation. ADR 009 records the ownership and test-scope rules.

| Workflow | Responsibility | Trigger |
| --- | --- | --- |
| ci.yml | Go/Python unit tests and coverage, component selection, final CI result | PR, main push, manual |
| sonarcloud.yml | Project/visibility verification and server quality gate; consumes CI coverage | Reusable, manual |
| pre-commit.yml | Changed-file hooks and actionlint, required by core CI | Reusable, manual |
| compliance-runtime.yml | Disposable Kubernetes/PostgreSQL integration, operator image and package snapshots | Reusable for runtime changes, manual |
| terraform-validate.yml | Eight blueprints, the security module and experimental GCE mock plans, supported-provider TFLint, security report | Reusable for infrastructure changes, manual |
| helm-validate.yml | Locked chart dependencies, first-party charts and stable GitOps manifests | Reusable for manifest changes, manual |
| developer-image.yml | Cached development-image build and tool/user smoke checks | Reusable for Dockerfile/dependency changes, manual |
| example-frontend.yml | Locked npm install/audit, tests, build and container smoke checks | Reusable for frontend changes, weekly, manual |
| nightly.yml | Disposable integration/package validation and recorded benchmarks | Daily, manual |
| mutation.yml | Pinned mutation tool and mutation baseline | Weekly, manual |
| dependency-audit.yml | Pinned Python dependency audit with retained report | Weekly, manual |
| release-please.yml | SemVer release PRs/tags and direct release publication call | Main push, manual on main |
| release.yml | Exact-tag archives/deb/rpm, multiarchitecture image, content, SBOMs and signatures | Reusable, published release, manual |
| slsa-provenance.yml | Generated artifact provenance for the release | Reusable |
| control-bundle.yml | Independently versioned, signed immutable control content | Manual on main |
| dependabot-auto-merge.yml | Native protected auto-merge for Dependabot PRs | Dependabot metadata events, hourly, manual |
| auto-request-reviewers.yml | Request reviews from non-author assignees without checking out PR code | PR metadata events, manual |
| runner-reaper.yml | Legacy tagged orphan-runner cleanup on GitHub-hosted infrastructure | Manual |
| digitalocean-runner-smoke.yml | Dedicated single-job runner prerequisites, mise and Docker validation | Relevant same-repository PR, manual |
| wp-ci.yml | Legacy WordPress PHP compatibility/syntax and custom-code checks | Relevant application changes, manual |

Release Please dispatches preflight CI on its generated PR head branch. CI verifies
the PR is open and the dispatched commit matches its current head before selecting
checks or attaching Sonar analysis. GitHub does not count workflow_dispatch job
checks toward required PR checks. For a PR created or updated with GITHUB_TOKEN,
a maintainer reviews the generated changes and approves its pending PR workflows
using GitHub’s “Approve workflows to run” control. Those pull_request checks must
pass before merging; the preflight does not replace those gates. See
[GitHub’s token-triggered workflow rules](https://docs.github.com/en/actions/concepts/security/github_token).

## Runtime and cache policy

Eligible Linux jobs can use the dedicated ephemeral DigitalOcean pool after the
provisioning smoke test succeeds and `DO_RUNNERS_ENABLED=true` is configured.
Fork/metadata events and the legacy reaper retain GitHub-hosted routing; the
external provenance generator owns its runner selection. See
[the rollout and rollback contract](../../docs/digitalocean-runners.md).

Core runtimes and validation tools are pinned in mise.toml; the shared setup action
installs only the tools a job uses. The example frontend has its own mise
configuration. WordPress checks use mise-pinned PHP 8.4.26 and Composer 2.10.3,
validate and audit the lockfile, and check custom code when present. PHP source-build
headers are installed only in that job; mise caches the compiled runtime.
External actions are pinned to full commit IDs, with native Dependabot maintaining
them. The SLSA generator deliberately uses its supported
version tag so its verifier can identify the trusted reusable workflow.

Caches hold Go modules/build output, pip/npm/Composer downloads, pre-commit
environments, Terraform providers, pinned TFLint rulesets, chart dependencies and
Docker layers. They do not hold credentials or Terraform state. Coverage,
package snapshots, audit reports and benchmarks have bounded retention.

## Required checks and permissions

Main should require DCO, CI, and the actual reusable SonarCloud quality-gate check
reported by GitHub Actions. CI requires Go, Python, repository checks and Sonar success, plus success
for every selected component; an unselected component may be skipped. Do not
require a path-filtered workflow whose check might never be created.

Sonar authentication lives in the SONAR_TOKEN repository secret. The project is
somethingwithproof_mantl under somethingwithproof. Automatic analysis must be
disabled for this CI pipeline. See ../../docs/sonarcloud.md for setup and fork PR
limitations.

Default permissions are contents: read. Write permissions are restricted to
release publishing, provenance, optional SARIF publishing and PR metadata
automation. PR metadata workflows never check out or execute PR code. Dependabot
uses GitHub's native auto-merge and the required branch protections; it does not
bypass checks or require every optional third-party integration to be available.

## Validation boundaries

Normal and scheduled workflows use local contracts, rendering, unit tests and
explicit disposable Kubernetes/PostgreSQL fixtures. They do not deploy cloud
infrastructure or claim acceptance for live clusters. Existing Checkov findings
remain informational and are retained as reports; Trivy's configured changed-file
hook continues to enforce the repository's existing severity policy.

The former Dagger/no-module pipeline, duplicate container/Go pipelines, custom
Dependabot updater, separate benchmark schedule, nonexistent Cloudflare path,
and duplicate DigitalOcean validation were removed or consolidated. The Terraform
helper validates real infra/terraform paths and retains failures across modules.

Release publication retains exact-tag validation, immutable registry/content
versions, least-privilege signing, checksums, SBOMs, and generated provenance.
See ../../docs/releases.md for the supported artifacts and release process.

The core Python job also checks the ML example’s pinned dependencies and health
handler when its source, runtime, shared setup, or CI configuration changes. Its
coverage is appended to the same report used by SonarCloud.
