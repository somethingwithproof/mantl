# ARA-0002 GitOps and delivery
Status: Accepted
Date: 2025-12-27

Context
We need consistent, auditable platform management and flexible app deployment flows.

Decision
- Use ArgoCD for GitOps of cluster and platform components (infra add-ons).
- Use Spinnaker for application release orchestration (approvals, rollouts, later canaries).

Consequences
- Clear separation of concerns: platform vs. app delivery
- Works with multiple CI systems (GHA, Jenkins X) via triggers/webhooks

Alternatives
- ArgoCD only for apps: weaker approvals/canary UX for some teams
- Spinnaker only: less native GitOps ergonomics for cluster/platform state
