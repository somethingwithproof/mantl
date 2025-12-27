# ExternalDNS Identity

## EKS (IRSA)
Use `terraform/modules/iam/external-dns-eks` to create an IAM role trusted by the cluster OIDC provider and attach Route53 permissions. Annotate the external-dns service account with the provided role ARN (`eks.amazonaws.com/role-arn`).

## GKE (Workload Identity)
- Terraform example: terraform/examples/gke-external-dns-wi (grants roles/dns.admin and binds KSA).
- Update clusters/production/apps/external-dns-gcp.yaml to mount credentials only if you are not using Workload Identity.

## AKS (Workload Identity)
- Terraform module: terraform/modules/iam/aks-workload-identity
- Example: terraform/examples/aks-external-dns-wi outputs the managed identity client_id.
- Set that client_id in clusters/production/apps/external-dns-azure.yaml under serviceAccount.annotations.azure.workload.identity/client-id.
- Ensure AKS Workload Identity is enabled on the cluster.
- Create a Google service account (GSA) with DNS permissions.
- Bind the KSA `external-dns` in namespace `external-dns` to the GSA via IAM policy binding `roles/iam.workloadIdentityUser`.
- Annotate the KSA with `iam.gke.io/gcp-service-account: <gsa>@<project>.iam.gserviceaccount.com`.

See clusters/production/apps/external-dns-*.yaml for per-cloud Helm values. For end-to-end overlay commands, see `docs/overlays.md`.
