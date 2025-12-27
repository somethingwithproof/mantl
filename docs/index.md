# Mantl Platform Documentation

**Last Updated**: 2025-12-27

## Getting Started

- **[Quickstart Guide](quickstart-platform.md)** - Deploy the full platform in minutes
- **[Platform Apps](platform-apps.rst)** - ArgoCD applications and sync waves
- **[CNCF Stack Decision](CNCF-STACK-DECISION.md)** - Component selection rationale

## Core Platform

- **[Secrets & Issuers](secrets-and-issuers.md)** - External Secrets and cert-manager
- **[Overlays Guide](overlays.md)** - Multi-cloud and environment configurations
  - Provider overlays: `clusters/production/overlays/{aws,gke,aks,digitalocean,linode}`
  - Workload Identity overlays: `clusters/production/overlays/{gke-wi,aks-wi}`
- **[Security Baseline](security.md)** - Kyverno policies and security controls

## Observability

- **[Observability Stack](ara/ARA-0005-observability-stack.md)** - Prometheus, Tempo, OTel, Loki
  - Metrics: Prometheus + Grafana
  - Traces: Tempo + OpenTelemetry Collector
  - Logs: Loki

## CI/CD & Delivery

- **[Required Checks](ci-required-checks.md)** - GitHub Actions CI configuration
- **[Spinnaker Registries](spinnaker-registries.md)** - Multi-cloud deployment pipelines
- **[ExternalDNS Identity](external-dns-identity.md)** - Cloud provider DNS automation

## Architecture

- **[Architecture Decision Records](ara/README.md)** - Platform design decisions
  - [ARA-0001: Kubernetes First](ara/ARA-0001-kubernetes-first.md)
  - [ARA-0002: GitOps and Delivery](ara/ARA-0002-gitops-and-delivery.md)
  - [ARA-0003: Policy as Code](ara/ARA-0003-policy-as-code.md)
  - [ARA-0004: Secrets Strategy](ara/ARA-0004-secrets-strategy.md)
  - [ARA-0005: Observability Stack](ara/ARA-0005-observability-stack.md)
  - [ARA-0006: CI/CD Strategy](ara/ARA-0006-ci-cd-strategy.md)
  - [ARA-0007: Multi-Cloud Failover](ara/ARA-0007-multi-cloud-failover.md)
  - [ARA-0008: Supply Chain Security](ara/ARA-0008-supply-chain-security.md)

## Developer Portal

- **[Backstage Catalog](../catalog-info.yaml)** - Service catalog metadata
- **[Backstage Guide](backstage.md)** - Developer portal setup

## Reference

- **Multi-Cloud**: [multicloud.rst](multicloud.rst) - Cross-cloud deployment patterns
- **OIDC/TLS/DNS**: [oidc-tls-dns.rst](oidc-tls-dns.rst) - Authentication and certificates
- **Kyverno Policies**: `policies/kyverno/` - Policy enforcement rules

---

**Note**: This repository contains legacy documentation for pre-Kubernetes components (Mesos/Marathon). Those are historical references only—the Kubernetes-first architecture above is authoritative.
