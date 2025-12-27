Platform Apps via ArgoCD
========================

Overview
--------
The production cluster is bootstrapped by ArgoCD Applications defined in ``clusters/production/apps``. Each app installs or configures a core platform component.

Apps
----
- cert-manager: Installs cert-manager via Helm.
- cert-manager-issuer: Applies HTTP-01 ClusterIssuer; DNS-01 alternatives are available under:
  - cert-manager-issuer-dns-cloudflare
  - cert-manager-issuer-dns-route53
  - cert-manager-issuer-dns-gcp
- external-secrets: Installs External Secrets Operator.
- kyverno: Installs Kyverno; kyverno-policies applies repo policies.
- gateway-api-crds: Installs Gateway API CRDs.
- cilium: Installs Cilium with Hubble and Gateway API support enabled.
- gateway-sample: Deploys a sample Gateway and HTTPRoute using the Cilium GatewayClass.
- envoy-gateway (optional): Installs Envoy Gateway controller.
- envoy-gateway-sample (optional): Deploys a sample Gateway/HTTPRoute using Envoy GatewayClass.
- example-hello: Deploys a simple nginx service named `app` in the default namespace for routing tests.
- spinnaker-operator: Installs the Spinnaker Operator.
- spinnaker-secrets: Applies ESO secrets needed by Spinnaker (OIDC, docker registries).
- spinnaker: Applies the base SpinnakerService.
- spinnaker-docker-registries: Patches clouddriver and SpinnakerService to add multiple docker registry artifact accounts.
- external-dns-digitalocean (optional): ExternalDNS using DO token via ESO.
- external-dns-linode (optional): ExternalDNS using Linode token via ESO.
- cert-manager-issuer-dns-digitalocean (optional): DNS-01 via DigitalOcean API token.

Notes
-----
- Secrets are sourced via ESO from Vault (see docs/ara/ARA-0004-secrets-strategy.md).
- Enable one DNS-01 issuer app at a time per cluster.