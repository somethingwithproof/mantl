<!-- SPDX-License-Identifier: Apache-2.0 -->
# ARA-0001 Kubernetes-first platform
Status: Accepted
Date: 2025-12-27

Context
Mantl historically supported Mesos/Marathon and Swarm. Those ecosystems are deprecated and fragmented. The 2015-era architecture required manual cluster provisioning and Ansible-based configuration management.

Decision
Adopt **managed Kubernetes** as the sole orchestrator. All platform and app primitives target Kubernetes APIs.

- **Cluster Provisioning**: Use managed Kubernetes services (AWS EKS, GCP GKE, Azure AKS, DigitalOcean DOKS, Linode LKE)
- **Bootstrapping**: Terraform blueprints in `terraform/blueprints/` for cluster creation
- **Platform Management**: GitOps with ArgoCD replaces Ansible configuration management
- **Infrastructure as Code**: Crossplane for Day 2 operations (databases, storage, networking)

Consequences
- Simplifies platform surface area and talent hiring
- Enables reuse of CNCF ecosystem (Helm, Kustomize, Gateway API, Crossplane)
- Eliminates operational overhead of managing control planes
- Cloud-native workload identity (IRSA, Workload Identity, Managed Identity)
- Legacy Mesos/Marathon/Chronos/Swarm components removed from repository (commit 57b028bb)
- Legacy Ansible configuration removed in favor of declarative Kubernetes manifests

Alternatives
- **Self-managed Kubernetes (kubeadm, k3s)**: Higher operational burden; reserved for bare-metal/edge use cases
- **Keep multi-orchestrator support**: Higher maintenance, limited value in 2026, no CNCF ecosystem benefits
