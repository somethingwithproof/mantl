# Core Configuration
variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
  default     = "mantl-doks"
}

variable "region" {
  description = "DigitalOcean region (e.g., nyc1, nyc3, sfo3, sgp1, lon1, fra1)"
  type        = string
  default     = "nyc3"
}

variable "environment" {
  description = "Environment name (e.g., production, staging, development)"
  type        = string
  default     = "production"
}

variable "kubernetes_version" {
  description = "Kubernetes version for the cluster"
  type        = string
  default     = "1.31.3-do.0" # Latest stable
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = list(string)
  default     = ["mantl", "kubernetes"]
}

# VPC Configuration
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.10.0.0/16"
}

# Security Configuration
variable "allowed_ssh_cidrs" {
  description = "CIDR blocks allowed to SSH to nodes"
  type        = list(string)
  default     = ["10.0.0.0/8"] # Restrict to internal only
}

# System Node Pool Configuration
variable "system_node_size" {
  description = "Droplet size for system nodes (platform components)"
  type        = string
  default     = "s-4vcpu-8gb" # 4 vCPU, 8 GB RAM
}

variable "system_node_count" {
  description = "Number of system nodes (non-autoscaling)"
  type        = number
  default     = 3
}

# Workload Node Pool Configuration
variable "workload_node_size" {
  description = "Droplet size for workload nodes (application workloads)"
  type        = string
  default     = "s-4vcpu-8gb" # 4 vCPU, 8 GB RAM
}

variable "workload_min_nodes" {
  description = "Minimum number of workload nodes (autoscaling)"
  type        = number
  default     = 2
}

variable "workload_max_nodes" {
  description = "Maximum number of workload nodes (autoscaling)"
  type        = number
  default     = 10
}

# Spot Node Pool Configuration (optional cost savings)
variable "enable_spot_pool" {
  description = "Enable spot instances node pool for cost savings"
  type        = bool
  default     = false
}

variable "spot_node_size" {
  description = "Droplet size for spot nodes"
  type        = string
  default     = "s-4vcpu-8gb"
}

variable "spot_min_nodes" {
  description = "Minimum number of spot nodes"
  type        = number
  default     = 0
}

variable "spot_max_nodes" {
  description = "Maximum number of spot nodes"
  type        = number
  default     = 5
}

# Cluster Management
variable "enable_auto_upgrade" {
  description = "Enable automatic Kubernetes version upgrades"
  type        = bool
  default     = false # Manual control recommended for production
}

variable "maintenance_day" {
  description = "Day of week for maintenance (monday, tuesday, etc.)"
  type        = string
  default     = "sunday"
}

variable "maintenance_start_time" {
  description = "Start time for maintenance window (HH:MM in UTC)"
  type        = string
  default     = "02:00"
}

variable "destroy_all_resources" {
  description = "Destroy all associated resources when cluster is destroyed"
  type        = bool
  default     = false
}

# Container Registry
variable "registry_tier" {
  description = "Container registry tier (starter, basic, professional)"
  type        = string
  default     = "basic"
}

variable "registry_region" {
  description = "Region for container registry"
  type        = string
  default     = "nyc3"
}

# Load Balancer
variable "enable_ingress_lb" {
  description = "Create a load balancer for ingress traffic"
  type        = bool
  default     = true
}

# Monitoring
variable "enable_monitoring" {
  description = "Enable DigitalOcean monitoring and alerts"
  type        = bool
  default     = true
}

variable "alert_emails" {
  description = "Email addresses for monitoring alerts"
  type        = list(string)
  default     = []
}

# Platform Database (optional)
variable "enable_platform_db" {
  description = "Create a managed PostgreSQL database for platform services"
  type        = bool
  default     = false
}

variable "postgres_version" {
  description = "PostgreSQL version"
  type        = string
  default     = "16"
}

variable "postgres_size" {
  description = "Database cluster size slug"
  type        = string
  default     = "db-s-2vcpu-4gb" # 2 vCPU, 4 GB RAM
}

variable "postgres_node_count" {
  description = "Number of database nodes (1 for single, 2 for HA)"
  type        = number
  default     = 2 # HA by default
}

variable "db_maintenance_day" {
  description = "Day of week for database maintenance"
  type        = string
  default     = "sunday"
}

variable "db_maintenance_hour" {
  description = "Hour of day for database maintenance (00-23 UTC)"
  type        = string
  default     = "03:00"
}

# Backup Storage
variable "enable_backup_bucket" {
  description = "Create a Spaces bucket for backups (S3-compatible)"
  type        = bool
  default     = true
}

variable "spaces_region" {
  description = "Region for Spaces bucket (nyc3, sfo3, sgp1, fra1, ams3)"
  type        = string
  default     = "nyc3"
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30
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
  description = "Export Cilium flow logs to Spaces bucket"
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
  description = "Enable Tailscale VPN for secure cluster access (replaces missing private endpoint)"
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
