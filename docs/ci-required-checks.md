<!-- SPDX-License-Identifier: Apache-2.0 -->
# Required GitHub checks

Main currently requires these exact check contexts:

- `DCO`: contributor sign-off.
- `CI`: the final result of the Mantl CI workflow.
- `SonarCloud / SonarCloud quality gate`: the reusable SonarCloud server gate.

Branch protection requires an up-to-date branch and applies to administrators.
These contexts were verified against repository settings on 2026-10-06. Settings
can change; compare actual PR checks and branch protection before changing them.

[Core CI](../.github/workflows/ci.yml) requires Go, Python, repository checks, and
Sonar success, plus selected component checks. Component selection keeps optional
jobs from becoming required contexts that are absent on documentation-only PRs.
[The workflow inventory](../.github/workflows/README.md) describes selection and
permissions; [SonarCloud setup](sonarcloud.md) explains project and token scope.

Do not bypass required checks to merge. Release Please PRs created/updated by
GITHUB_TOKEN can require approval of their PR workflows; dispatched preflight
checks alone do not replace required pull-request checks. Follow the existing
workflow guidance rather than the historical list of removed kind/E2E job names.
