# Architecture Rationale and Alternatives (ARA)
Purpose: capture key architectural decisions, rationale, and considered alternatives for Mantl 2026.

Conventions
- ID format: ARA-XXXX
- Status: Proposed | Accepted | Superseded | Deprecated
- Date: 2025-12-27

Index
- ARA-0001: Kubernetes-first platform
- ARA-0002: GitOps with ArgoCD (platform) + Spinnaker (app delivery)
- ARA-0003: Policy-as-Code with Kyverno
- ARA-0004: Secrets via External Secrets Operator backed by Vault/cloud stores
- ARA-0005: Observability with OpenTelemetry + Grafana Stack
- ARA-0006: CI/CD choices (GitHub Actions, Jenkins X) and triggers
- ARA-0007: Multi-cloud and failover patterns
- ARA-0008: Supply-chain security (SLSA, SBOM, signing)
