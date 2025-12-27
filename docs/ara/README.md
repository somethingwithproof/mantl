# Architecture Rationale and Alternatives (ARA)

Purpose: Capture key architectural decisions, rationale, and considered alternatives for Mantl 2026.

## Conventions

- **ID format**: ARA-XXXX
- **Status**: Proposed | Accepted | Superseded | Deprecated
- **Date**: 2025-12-27

## Index

**Infrastructure & Platform**:
- [ARA-0001](ARA-0001-kubernetes-first.md): Managed Kubernetes over Mesos/Marathon
- [ARA-0007](ARA-0007-multi-cloud-failover.md): Multi-cloud strategy with Terraform + Crossplane

**GitOps & Deployment**:
- [ARA-0002](ARA-0002-gitops-and-delivery.md): ArgoCD for unified GitOps (platform + apps)
- [ARA-0006](ARA-0006-ci-cd-strategy.md): GitHub Actions + ArgoCD deployment workflow

**Security & Compliance**:
- [ARA-0003](ARA-0003-policy-as-code.md): Kyverno for policy enforcement
- [ARA-0004](ARA-0004-secrets-strategy.md): External Secrets Operator with cloud-native backends
- [ARA-0008](ARA-0008-supply-chain-security.md): Image signing, SBOMs, and Harbor registry

**Observability**:
- [ARA-0005](ARA-0005-observability-stack.md): Prometheus + Tempo + Loki + OpenTelemetry

## Key Technology Decisions

### CNCF Graduated Projects (9)
Kubernetes, Cilium, ArgoCD, cert-manager, Kyverno, External Secrets, Prometheus, OpenTelemetry, Falco, Harbor

### CNCF Incubating Projects (3)
Crossplane, ExternalDNS

### Removed Components
- **Mesos/Marathon**: Replaced by managed Kubernetes (ARA-0001)
- **Ansible**: Replaced by ArgoCD GitOps (ARA-0001, ARA-0002)
- **Spinnaker**: Replaced by ArgoCD + Argo Rollouts (ARA-0002, ARA-0006)
- **OPA**: Replaced by Kyverno for policy enforcement (ARA-0003)
- **Jaeger**: Replaced by Grafana Tempo for distributed tracing (ARA-0005)
