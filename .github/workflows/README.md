# Mantl GitHub Actions

Core CI owns first-party tests and coverage. Optional component checks are selected
inside the workflow, so required checks exist even when a pull request changes
only documentation. ADR 009 records the ownership and test-scope rules.

| Workflow | Responsibility | Trigger |
| --- | --- | --- |
| ci.yml | Go/Python unit tests and coverage, component selection, final CI result | PR, main push, manual |
| sonarcloud.yml | Project/visibility verification and server quality gate; consumes CI coverage | Reusable, manual |
| pre-commit.yml | Changed-file hooks and actionlint | PR, main push, manual |
| compliance-runtime.yml | Disposable Kubernetes/PostgreSQL integration, operator image and package snapshots | Reusable for runtime changes, manual |
| terraform-validate.yml | Eight blueprints plus the security module, supported-provider TFLint, security report | Reusable for infrastructure changes, manual |
| helm-validate.yml | Locked chart dependencies, first-party charts and stable GitOps manifests | Reusable for manifest changes, manual |
| developer-image.yml | Cached development-image build and tool/user smoke checks | Reusable for Dockerfile/dependency changes, manual |
| example-frontend.yml | Locked npm install/audit, tests, build and container smoke checks | Reusable for frontend changes, weekly, manual |
| nightly.yml | Disposable integration/package validation and recorded benchmarks | Daily, manual |
| mutation.yml | Pinned mutation tool and mutation baseline | Weekly, manual |
| dependency-audit.yml | Pinned Python dependency audit with retained report | Weekly, manual |
| release-please.yml | SemVer release PRs/tags and direct release publication call | Main push, manual |
| release.yml | Exact-tag archives/deb/rpm, multiarchitecture image, content, SBOMs and signatures | Reusable, published release, manual |
| slsa-provenance.yml | Generated artifact provenance for the release | Reusable |
| control-bundle.yml | Independently versioned, signed immutable control content | Manual on main |
| dependabot-auto-merge.yml | Native protected auto-merge for Dependabot PRs | Dependabot metadata events, hourly, manual |
| auto-request-reviewers.yml | Request reviews from non-author assignees without checking out PR code | PR metadata events, manual |
| runner-reaper.yml | Legacy tagged orphan-runner cleanup; new CI creates no cloud runners | Manual |
| wp-ci.yml | Legacy WordPress PHP compatibility/syntax and custom-code checks | Relevant application changes, manual |

Release Please explicitly dispatches CI on its generated PR head branch. This
allows release PRs created with GITHUB_TOKEN to satisfy required checks without
a personal access token. CI verifies the PR is open and the dispatched commit
matches its current head before selecting checks or attaching Sonar analysis.

## Runtime and cache policy

Core runtimes and validation tools are pinned in mise.toml; the shared setup action
installs only the tools a job uses. The example frontend has its own mise
configuration. The legacy PHP compatibility matrix uses its pinned native PHP
setup action. External actions are pinned to full commit IDs, with native
Dependabot maintaining them. The SLSA generator deliberately uses its supported
version tag so its verifier can identify the trusted reusable workflow.

Caches hold Go modules/build output, pip/npm/Composer downloads, pre-commit
environments, Terraform providers, pinned TFLint rulesets, chart dependencies and
Docker layers. They do not hold credentials or Terraform state. Coverage,
package snapshots, audit reports and benchmarks have bounded retention.

## Required checks and permissions

Main should require DCO, CI, and the actual reusable SonarCloud quality-gate check
reported by GitHub Actions. CI requires Go, Python and Sonar success, plus success
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
