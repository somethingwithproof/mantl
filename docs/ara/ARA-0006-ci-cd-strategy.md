<!-- SPDX-License-Identifier: Apache-2.0 -->
# ARA-0006 CI/CD strategy and triggers
Status: Accepted
Date: 2025-12-27

Context
Teams use GitHub Actions and Jenkins for CI. We need a deployment strategy that's CI-agnostic and supports GitOps principles.

Decision
**CI (Build & Test)**:
- **Default**: GitHub Actions with OIDC for passwordless cloud authentication
- **Alternative**: Jenkins with cloud workload identity
- **Output**: Push container images to registry (Harbor, ECR, GCR, ACR, GHCR)

**CD (Deployment)**:
- **GitOps Engine**: ArgoCD for automated deployment from Git
- **Image Updates**: ArgoCD Image Updater or external tools update Git manifests
- **Progressive Delivery**: Argo Rollouts for canary and blue-green deployments (optional)
- **Approvals**: ArgoCD RBAC + Git-based approvals (PR reviews)
- **Multi-Environment**: Promote by updating Git refs per environment (`dev` → `staging` → `production`)

**Workflow**:
1. CI builds image and pushes to registry
2. CI creates PR to update image tag in Git manifests
3. PR reviewed and merged (approval gate)
4. ArgoCD detects Git change and deploys to cluster
5. Optional: Argo Rollouts performs canary analysis

Consequences
- CI-agnostic delivery model (any CI can update Git)
- Git as source of truth for all deployments
- Clear promotion/approval patterns via Git workflow
- Automated rollback by reverting Git commits
- Native Kubernetes resources (no proprietary pipeline DSL)
- Removed Spinnaker (commit 57b028bb) - ArgoCD provides equivalent functionality with lower overhead

Alternatives
- **Tekton**: Kubernetes-native CI/CD; more complex than GitHub Actions + ArgoCD
- **Flux CD**: Similar to ArgoCD; ArgoCD chosen for superior UI and RBAC
- **Spinnaker**: Enterprise-grade but high operational overhead; deprecated in favor of ArgoCD + Argo Rollouts
