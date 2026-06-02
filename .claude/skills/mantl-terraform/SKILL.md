---
name: mantl-terraform
description: Load when writing or modifying Terraform under infra/terraform, whether cluster blueprints, modules, or examples. Covers blueprint layout conventions, the AWS-GA / GCP-Azure-beta cloud-parity scope (ADR 005), the common hardening baseline, and provider pinning.
---

# mantl terraform skill

## Layout

```
infra/terraform/
  blueprints/   one directory per cloud target; the user-facing entry point
    aws-eks/    GA
    gcp-gke/    beta (hardening to AWS baseline per ADR 005)
    azure-aks/  beta (same)
    do-doks/, linode-lke/, oci-oke/, ibm-iks/   experimental
  modules/      reusable building blocks (modules/core, ...)
  examples/     worked usage (e.g. gke-external-dns-wi, aks-external-dns-wi)
```

A blueprint directory follows a fixed file convention. Match it exactly:

```
main.tf                  resources and module calls
variables.tf             inputs
outputs.tf               outputs
versions.tf              required_version + required_providers
terraform.tfvars.example sample inputs
README.md                usage
```

## Cloud-parity scope (ADR 005)

Three clouds are in scope for parity; the rest are explicitly experimental.

| Cloud | Blueprint | Status | Obligation |
|---|---|---|---|
| AWS | `aws-eks` | GA | reference baseline; other clouds match it |
| GCP | `gcp-gke` | beta | hardened to the AWS baseline |
| Azure | `azure-aks` | beta | hardened to the AWS baseline |
| Linode, Oracle, IBM, DigitalOcean | respective dirs | experimental | no parity or CI guarantee |

Do not present an experimental blueprint as production-ready, and do not silently upgrade its maturity label. Maturity labels in docs and directory layout are a deliberate honesty signal (ADR 005). Touching an experimental blueprint is fine; promoting one to beta/GA is an ADR-worthy change.

## Hardening baseline

Every GA/beta blueprint must deliver the common baseline. When editing aws-eks, or bringing gcp-gke or azure-aks toward parity, check all of these are present:

- Private cluster (no public control-plane endpoint, or restricted/authorized networks only).
- Control-plane audit logging enabled and shipped off-cluster.
- Encryption at rest for cluster secrets/etcd (KMS-backed: AWS KMS, Cloud KMS, or Azure Key Vault).
- Workload identity (IRSA on EKS, Workload Identity on GKE, Azure AD Workload Identity on AKS). No long-lived node-attached cloud credentials.
- Least-privilege IAM: scope roles to what the workload needs; no wildcard policies, no cluster-admin handed to controllers.

The compliance evidence path depends on this: the operator authenticates to the evidence bucket via workload identity (ADR 003), so the workload-identity wiring is not optional on a GA cluster.

## Provider pinning

`versions.tf` pins both the Terraform core version and each provider. Current aws-eks pins:

```hcl
terraform {
  required_version = ">= 1.14.0"
  required_providers {
    aws        = { version = ">= 5.40.0" }
    kubernetes = { version = ">= 2.27.0" }
    helm       = { version = ">= 2.12.0" }
    tls        = { version = ">= 4.0.0" }
  }
}
```

Module versions are pinned with `~>` constraints in `main.tf` (the EKS module at `~> 20.8`, AWS submodules at `~> 5.37`). When adding a provider or module, pin it; do not leave a constraint open-ended. Keep `kubernetes` and `helm` provider versions consistent across blueprints so the same chart install behaves identically on every cloud.

## When adding or changing a blueprint

1. Keep the six-file convention; do not invent new top-level file names.
2. Pin every new provider and module in `versions.tf` / `main.tf`.
3. Verify the hardening baseline items are all wired for GA/beta blueprints.
4. Update `README.md` and `terraform.tfvars.example`.
5. `terraform fmt` and `terraform validate` locally. Do not run `apply`/`destroy` from here.
