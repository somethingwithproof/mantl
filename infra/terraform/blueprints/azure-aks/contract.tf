# Stable, non-secret interface consumed by Mantl tooling.
output "platform_contract" {
  description = "Provider-neutral platform identity; deployment verification is separate"
  value = {
    provider           = "azure"
    cluster_name       = azurerm_kubernetes_cluster.main.name
    endpoint           = azurerm_kubernetes_cluster.main.kube_config[0].host
    kubernetes_version = var.kubernetes_version
    workload_identity  = azurerm_kubernetes_cluster.main.oidc_issuer_url
    evidence_backend   = "external-s3-object-lock"
  }
}
