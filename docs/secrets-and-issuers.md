# Secrets and Issuers

This repo expects External Secrets Operator (ESO) to sync secrets from Vault.

## Vault paths and keys (KV v2)

Example secret paths for External Secrets Operator:

- cert-manager/route53
  - aws_access_key_id
  - aws_secret_access_key
- cert-manager/cloudflare
  - api_token
- cert-manager/gcp
  - service_account_json

## Selecting an ACME solver

### DigitalOcean
- Vault path: `cert-manager/digitalocean` with key `api_token`
- Enable app `clusters/production/apps/cert-manager-issuer-dns-digitalocean.yaml` and (optionally) `external-dns-digitalocean.yaml`

### Linode
- Cert-manager DNS-01 webhook for Linode is not bundled; prefer HTTP-01 on Linode or bring your own webhook.
- ExternalDNS Linode is supported via `clusters/production/apps/external-dns-linode.yaml` with Vault path `external-dns/linode` key `api_token`.

### Azure DNS
- Vault path: `cert-manager/azure` with keys: tenant_id, client_id, client_secret, subscription_id, resource_group, hosted_zone
- Enable app `clusters/production/apps/cert-manager-issuer-dns-azure.yaml` (or provider overlay `clusters/production/overlays/aks`).

Choose one DNS-01 issuer Application in `clusters/production/apps/` and add it to the production kustomization.
Alternatively, use HTTP-01 (ClusterIssuer in platform/base/cert-manager).