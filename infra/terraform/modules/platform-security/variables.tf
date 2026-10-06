# SPDX-License-Identifier: Apache-2.0
################################################################################
# Platform Security Module Variables
################################################################################

# Cilium + Hubble (Network Observability)
variable "enable_cilium" {
  description = "Enable Cilium CNI with Hubble for network observability (flow logs alternative)"
  type        = bool
  default     = false
}

variable "hubble_ui_ingress_enabled" {
  description = "Enable ingress for Hubble UI"
  type        = bool
  default     = false
}

variable "hubble_ui_hosts" {
  description = "Hostnames for Hubble UI ingress"
  type        = list(string)
  default     = []
}

variable "enable_flow_logs_export" {
  description = "Enable export of Cilium flow logs to external storage"
  type        = bool
  default     = true
}

variable "flow_logs_s3_bucket" {
  description = "S3-compatible bucket for flow logs export (leave empty to disable)"
  type        = string
  default     = ""
}

# Sealed Secrets (KMS Alternative)
variable "enable_sealed_secrets" {
  description = "Enable Sealed Secrets for encrypted secrets in Git (KMS alternative)"
  type        = bool
  default     = false
}

# External Secrets Operator
variable "enable_external_secrets" {
  description = "Enable External Secrets Operator for external secrets management"
  type        = bool
  default     = false
}

# Falco (Runtime Security)
variable "enable_falco" {
  description = "Enable Falco for runtime security monitoring"
  type        = bool
  default     = false
}

variable "falco_webhook_url" {
  description = "Webhook URL for Falco alerts (e.g., Slack, PagerDuty)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "falco_ui_enabled" {
  description = "Enable Falco UI for alert visualization"
  type        = bool
  default     = false
}

# Tailscale VPN (Private Access Alternative)
variable "enable_tailscale" {
  description = "Enable Tailscale VPN for secure cluster access (private endpoint alternative)"
  type        = bool
  default     = false
}

variable "tailscale_auth_key" {
  description = "Tailscale auth key (deprecated - use OAuth)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "tailscale_client_id" {
  description = "Tailscale OAuth client ID"
  type        = string
  default     = ""
}

variable "tailscale_client_secret" {
  description = "Tailscale OAuth client secret"
  type        = string
  default     = ""
  sensitive   = true
}

# Cert-Manager
variable "enable_cert_manager" {
  description = "Enable cert-manager for automated TLS certificate management"
  type        = bool
  default     = true
}

# Monitoring
variable "enable_security_dashboard" {
  description = "Enable Grafana dashboard for security monitoring"
  type        = bool
  default     = true
}
