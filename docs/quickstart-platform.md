# Mantl 2026 Platform Quickstart

1) Bootstrap ArgoCD
- Apply platform/base/argocd/kustomization.yaml or install ArgoCD via Helm and then apply clusters/production/apps/argocd*.yaml.

2) Apply production app set
```
kubectl apply -k clusters/production
```
This installs cert-manager, ESO, Kyverno (+ policies), Gateway API CRDs, Cilium, Spinnaker Operator, Spinnaker secrets/base, and the sample Gateway/HTTPRoute.

3) DNS/Certificates
- Choose one DNS‑01 issuer app in clusters/production/apps and add it to clusters/production/kustomization.yaml, or use the default HTTP‑01 issuer.

4) Secrets via Vault
- Populate Vault with keys referenced by ExternalSecrets (see platform/base/cert-manager/acme/* and addons/spinnaker/*/externalsecret.yaml).

5) Verify
- ArgoCD UI (namespace argocd) shows apps healthy.
- Gateway: create an A record for app.example.com pointing at the cluster LB and curl through the sample route.
