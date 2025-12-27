# ARA-0006 CI/CD strategy and triggers
Status: Accepted
Date: 2025-12-27

Context
Teams use GitHub Actions and Jenkins; we want portability.

Decision
- Default CI: GitHub Actions with OIDC for cloud auth; Jenkins X optional.
- Delivery: Spinnaker pipelines with Jenkins/GHA triggers or Gate webhooks; image triggers supported.

Consequences
- CI-agnostic delivery model
- Clear promotion/approval patterns; canaries added later

Alternatives
- Single CI vendor lock-in; or Argo Rollouts only (OK, but fewer enterprise approvals features)
