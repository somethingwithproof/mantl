# Mantl 2026 Documentation Index

- Quickstart: quickstart-platform.md
- Platform Apps (ArgoCD): platform-apps.rst
- Secrets & Issuers: secrets-and-issuers.md
- Overlays Guide: overlays.md
  - Provider overlays: clusters/production/overlays/{aws,gke,aks}
  - Workload Identity overlays: clusters/production/overlays/{gke-wi,aks-wi}
- Required Checks (CI): ci-required-checks.md
- Kyverno Policies: policies/kyverno/ (see docs/security.md)
- Security Baseline: security.md
- Architecture Decisions (ARA): ara/README.md
- Spinnaker Registries: spinnaker-registries.md
- ExternalDNS Identity: external-dns-identity.md
- Backstage Catalog: ../catalog-info.yaml
- Backstage Guide: backstage.md

Note: The repository contains older Sphinx/reStructuredText documentation for pre‑Kubernetes components (e.g., Mesos/Marathon). Treat those as legacy/historical references; the Kubernetes‑first path above is authoritative.
