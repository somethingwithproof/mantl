# Stable, non-secret interface consumed by Mantl tooling.
output "platform_contract" {
  description = "Provider-neutral platform identity; deployment verification is separate"
  value = {
    provider           = "gcp"
    cluster_name       = module.gke.name
    endpoint           = module.gke.endpoint
    kubernetes_version = var.kubernetes_version
    workload_identity  = "${var.project}.svc.id.goog"
    evidence_backend   = "external-s3-object-lock"
  }
}
