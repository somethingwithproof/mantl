<!-- SPDX-License-Identifier: Apache-2.0 -->
# Required GitHub checks

Main currently requires these exact check contexts:

- `DCO`: contributor sign-off.
- `CI`: the final result of the Mantl CI workflow.
- `SonarCloud / SonarCloud quality gate`: the reusable SonarCloud server gate.

Branch protection requires an up-to-date branch and applies to administrators.
Conversation resolution is required. No approving-review count is configured,
but an outstanding changes-requested review can still block merging.
These contexts were verified against repository settings on 2026-10-06. Settings
can change; compare actual PR checks and branch protection before changing them.

[Core CI](../.github/workflows/ci.yml) requires Go, Python, repository checks, and
Sonar success, plus selected component checks. Component selection keeps optional
jobs from becoming required contexts that are absent on documentation-only PRs.
Once selected, a component must succeed; its failure cannot be ignored. Snyk and
CodeRabbit status contexts are not in the required-check list, although actionable
review requests must be addressed independently of a bot's status context.
[The workflow inventory](../.github/workflows/README.md) describes selection and
permissions; [SonarCloud setup](sonarcloud.md) explains project and token scope.

Do not bypass required checks to merge. Release Please PRs created/updated by
GITHUB_TOKEN can require approval of their PR workflows; dispatched preflight
checks alone do not replace required pull-request checks. Follow the existing
workflow guidance rather than the historical list of removed kind/E2E job names.

Core CI starts on PR opening, reopening and commit pushes to PRs targeting main,
on main pushes, and on explicit manual dispatch. Title/body/label changes do not
restart it. Updating a branch against main changes its head and requires fresh
checks; passing checks from an older head cannot be reused as the new result.
PR concurrency cancels superseded PR runs, but branch-based manual dispatch and
PR-ref runs use different groups. Release Please therefore checks for a matching
PR run before dispatching fallback advisory preflight. See
[GitHub's eligible events and commit requirements](https://docs.github.com/en/pull-requests/how-tos/merge-and-close-pull-requests/troubleshooting-required-status-checks).
