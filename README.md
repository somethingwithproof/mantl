![image](./docs/_static/mantl-logo.png)

# Mantl — 2026 Platform Engineering Toolkit

Mantl is a Kubernetes‑first, GitOps‑driven, secure‑by‑default platform engineering toolkit for multi‑cloud, distributed services. It keeps the original “batteries‑included” spirit while modernizing for today’s cloud‑native ecosystems.

- Multi‑cloud blueprints (AWS, GCP, Azure, DigitalOcean, Linode)
- GitOps app-of-apps with ArgoCD
- Gateway API (Cilium or Envoy Gateway)
- ExternalDNS + cert‑manager issuers per provider
- Supply‑chain security: SBOM, image signing (Cosign), policy as code (Kyverno/OPA)

Legacy Ansible/Mesos materials are preserved as historical docs. New work centers on Kubernetes and GitOps.

## Quickstart

- Local: kind + ArgoCD + production app set
  ```
  kind create cluster --name mantl
  kubectl apply -k clusters/production
  ```
  See docs/quickstart-platform.md.

- Cloud: use Terraform blueprints under terraform/blueprints/* to provision EKS/GKE/AKS/DOKS/LKE, then apply the ArgoCD app set.

## Documentation

Start here:
- docs/index.md — doc index
- docs/overlays.md — provider and Workload Identity overlays
- docs/security.md — security baseline and signing
- docs/platform-apps.rst — inventory of ArgoCD apps
- docs/spinnaker-registries.md — Spinnaker registry accounts overlay

## Multi‑cloud and failover
See docs/multicloud.rst and terraform/modules/global-dns/route53-failover for patterns using health‑checked DNS and per‑cluster ingress.

## Contributing
See CONTRIBUTING.md for dev setup, CI parity, policy promotion (Audit→Enforce), and release. Run pre‑commit before pushing.

## License
Apache 2.0. See LICENSE.
