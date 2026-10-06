<!-- SPDX-License-Identifier: Apache-2.0 -->
# Provider & WI Overlays

These are example overlays, not a verified deployment recipe. Review installed
components and identities before applying them: the commands modify a cluster and
the Terraform step creates cloud resources/costs. Provider acceptance and native
immutable-storage parity are not established; start with [bootstrap prerequisites](quickstart-platform.md#deployment-is-a-separate-step).

## Production directory examples
- AWS: `kubectl apply -k clusters/production/overlays/aws`
- GKE: `kubectl apply -k clusters/production/overlays/gke`
- AKS: `kubectl apply -k clusters/production/overlays/aks`
- GKE (WI): `kubectl apply -k clusters/production/overlays/gke-wi`
- AKS (WI):
  1. Run terraform example to create UAMI and FIC: `terraform -chdir=infra/terraform/examples/aks-external-dns-wi apply`
  2. Set the client id in `clusters/production/overlays/aks-wi/kustomization.yaml` (uamiClientId literal)
  3. `kubectl apply -k clusters/production/overlays/aks-wi`

## Gateway sample custom domain
- Use `clusters/production/apps/gateway-sample-custom.yaml`. Set domain in `clusters/production/gateway-sample-custom/kustomization.yaml` via the `gateway-domain` ConfigMap generator.

## Dev / Staging (examples)
- Dev AWS: `kubectl apply -k clusters/dev/overlays/aws`
- Dev GKE: `kubectl apply -k clusters/dev/overlays/gke`
- Dev AKS: `kubectl apply -k clusters/dev/overlays/aks`
- Staging AWS: `kubectl apply -k clusters/staging/overlays/aws`
- Staging GKE: `kubectl apply -k clusters/staging/overlays/gke`
- Staging AKS: `kubectl apply -k clusters/staging/overlays/aks`

Notes
- Enable only one ExternalDNS app per cluster.
- For GKE WI, prefer `external-dns-gcp-wi.yaml` (no key mount).
- For AKS WI, use the aks-wi overlay to inject the client-id via kustomize replacements.
