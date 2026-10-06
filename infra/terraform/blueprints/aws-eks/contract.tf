# SPDX-License-Identifier: Apache-2.0
# Stable, non-secret interface consumed by Mantl tooling.
output "platform_contract" {
  description = "Provider-neutral platform identity; deployment verification is separate"
  value = {
    provider           = "aws"
    cluster_name       = module.eks.cluster_name
    endpoint           = module.eks.cluster_endpoint
    kubernetes_version = module.eks.cluster_version
    workload_identity  = module.eks.cluster_oidc_issuer_url
    evidence_backend   = "external-s3-object-lock"
  }
}
