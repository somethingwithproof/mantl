# SPDX-License-Identifier: Apache-2.0
# Core Configuration
variable "cluster_name" {
  description = "Name of the LKE cluster"
  type        = string
  default     = "mantl-lke"
}

variable "region" {
  description = "Linode region (e.g., us-east, us-west, eu-west, ap-south)"
  type        = string
  default     = "us-east"
}

variable "environment" {
  description = "Environment name (e.g., production, staging, development)"
  type        = string
  default     = "production"
}

variable "kubernetes_version" {
  description = "Kubernetes version for the cluster"
  type        = string
  default     = "1.31" # Latest stable
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = list(string)
  default     = ["mantl", "kubernetes"]
}

# High Availability
variable "enable_ha_control_plane" {
  description = "Enable high availability control plane"
  type        = bool
  default     = true
}

variable "cluster_cidr" {
  description = "CIDR block for Kubernetes cluster internal networking"
  type        = string
  default     = "10.2.0.0/16"
}

# Security Configuration
variable "allowed_ssh_cidrs" {
  description = "CIDR blocks allowed to SSH to nodes"
  type        = list(string)
  default     = ["10.0.0.0/8"] # Restrict to internal only
}

variable "allowed_nodeport_cidrs" {
  description = "CIDR blocks allowed to access NodePort services"
  type        = list(string)
  default     = ["0.0.0.0/0"] # Restrict as needed
}

# System Node Pool Configuration
variable "system_node_type" {
  description = "Linode type for system nodes (g6-standard-4 = 4 vCPU, 8 GB RAM)"
  type        = string
  default     = "g6-standard-4"
}

variable "system_node_count" {
  description = "Number of system nodes (non-autoscaling)"
  type        = number
  default     = 3
}

# Workload Node Pools Configuration
variable "workload_pools" {
  description = "List of workload node pools with autoscaling"
  type = list(object({
    name       = string
    node_type  = string
    node_count = number
    min_nodes  = number
    max_nodes  = number
  }))
  default = [
    {
      name       = "workload"
      node_type  = "g6-standard-4" # 4 vCPU, 8 GB RAM
      node_count = 2
      min_nodes  = 2
      max_nodes  = 10
    }
  ]
}

# Load Balancer Configuration
variable "enable_ingress_lb" {
  description = "Create a NodeBalancer for ingress traffic"
  type        = bool
  default     = true
}

variable "lb_throttle_connections" {
  description = "Connection throttle for NodeBalancer (connections per second)"
  type        = number
  default     = 0 # 0 = no throttling
}

# Backup Storage
variable "enable_backup_bucket" {
  description = "Create Object Storage bucket for backups (S3-compatible)"
  type        = bool
  default     = true
}

variable "object_storage_region" {
  description = "Region for Object Storage bucket (us-east-1, eu-central-1, ap-south-1)"
  type        = string
  default     = "us-east-1"
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30
}

# Persistent Storage
variable "enable_platform_volume" {
  description = "Create a Block Storage volume for platform data"
  type        = bool
  default     = false
}

variable "platform_volume_size" {
  description = "Size of platform volume in GB"
  type        = number
  default     = 100
}

################################################################################
# Platform Security Variables (Enterprise Feature Alternatives)
################################################################################

# Cilium + Hubble (Flow Logs Alternative)
variable "enable_cilium" {
  description = "Enable Cilium CNI with Hubble for network observability (replaces missing VPC Flow Logs)"
  type        = bool
  default     = true
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
  description = "Export Cilium flow logs to Object Storage bucket"
  type        = bool
  default     = true
}

# Sealed Secrets (KMS Alternative)
variable "enable_sealed_secrets" {
  description = "Enable Sealed Secrets for encrypted secrets in Git (replaces missing KMS)"
  type        = bool
  default     = true
}

# External Secrets Operator
variable "enable_external_secrets" {
  description = "Enable External Secrets Operator for external secrets management"
  type        = bool
  default     = false
}

# Falco (Runtime Security)
variable "enable_falco" {
  description = "Enable Falco for runtime security monitoring (enhanced security)"
  type        = bool
  default     = true
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

# Tailscale VPN (Private Endpoint Alternative)
variable "enable_tailscale" {
  description = "Enable Tailscale VPN for secure cluster access (no private endpoint available)"
  type        = bool
  default     = false
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

# Security Dashboard
variable "enable_security_dashboard" {
  description = "Enable Grafana dashboard for security monitoring"
  type        = bool
  default     = true
}
