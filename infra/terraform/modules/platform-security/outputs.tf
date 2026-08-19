output "cilium_enabled" {
  description = "Whether Cilium is enabled"
  value       = var.enable_cilium
}

output "hubble_ui_url" {
  description = "Hubble UI URL (if ingress enabled)"
  value = (
    var.enable_cilium && var.hubble_ui_ingress_enabled && length(var.hubble_ui_hosts) > 0
    ? "https://${var.hubble_ui_hosts[0]}"
    : null
  )
}

output "sealed_secrets_enabled" {
  description = "Whether Sealed Secrets is enabled"
  value       = var.enable_sealed_secrets
}

output "external_secrets_enabled" {
  description = "Whether External Secrets Operator is enabled"
  value       = var.enable_external_secrets
}

output "falco_enabled" {
  description = "Whether Falco is enabled"
  value       = var.enable_falco
}

output "tailscale_enabled" {
  description = "Whether Tailscale VPN is enabled"
  value       = var.enable_tailscale
}

output "security_features_summary" {
  description = "Summary of enabled security features"
  value = {
    network_observability = var.enable_cilium ? "Cilium + Hubble (Flow Logs)" : "None"
    secrets_management    = var.enable_sealed_secrets ? "Sealed Secrets" : (var.enable_external_secrets ? "External Secrets" : "None")
    runtime_security      = var.enable_falco ? "Falco" : "None"
    private_access        = var.enable_tailscale ? "Tailscale VPN" : "None"
    tls_management        = var.enable_cert_manager ? "Cert-Manager" : "None"
  }
}
