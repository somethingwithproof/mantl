<!-- SPDX-License-Identifier: Apache-2.0 -->
# ARA-0002 GitOps and delivery
Status: Accepted
Date: 2025-12-27

Context
We need consistent, auditable platform management and flexible app deployment flows. The platform must support both infrastructure GitOps and application progressive delivery.

Decision
Use **ArgoCD** as the unified GitOps engine for both platform and application delivery.

- **Platform Components**: Deployed via app-of-apps pattern (`clusters/production/platform-apps.yaml`)
- **Sync Waves**: Dependency-ordered deployment (Cilium → cert-manager → policies → observability)
- **Application Delivery**: ArgoCD Applications for workloads with automated sync and self-healing
- **Progressive Delivery**: Argo Rollouts for canary and blue-green deployments (optional)
- **Multi-Environment**: Environment-specific overlays in `clusters/{dev,staging,production}/`

Consequences
- Single GitOps tool reduces operational complexity
- Native Kubernetes CRDs for all deployment workflows
- Declarative rollback and drift detection built-in
- ArgoCD UI provides unified view of platform and application state
- Automated sync with prune and self-heal eliminates manual drift management
- Removed Spinnaker (commit 57b028bb) - ArgoCD handles both platform and app delivery

Alternatives
- **Flux CD**: Also viable, ArgoCD chosen for superior UI, RBAC, and multi-cluster management
- **Spinnaker**: Powerful enterprise features but higher operational overhead; deprecated in favor of ArgoCD + Argo Rollouts
- **Jenkins X**: Opinionated CI/CD; teams prefer GitHub Actions with ArgoCD for flexibility
