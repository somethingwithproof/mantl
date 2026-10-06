<!-- SPDX-License-Identifier: Apache-2.0 -->
# Linode cert-manager DNS-01 Webhook (optional)

This repo does not bundle a Linode DNS-01 solver by default. If you choose to use Linode for ACME:
- Install the Linode cert-manager webhook from its official Helm repository
- Pin a specific chart version and verify values
- Configure cert-manager ClusterIssuer to use the webhook solver

A stub ArgoCD Application is provided at `clusters/production/apps/cert-manager-webhook-linode.yaml` with placeholder fields (repoURL, chart, version). Update these to official values before enabling.
