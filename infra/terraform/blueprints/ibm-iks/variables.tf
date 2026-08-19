# IBM Cloud Provider Configuration
variable "ibmcloud_api_key" {
  description = "IBM Cloud API key"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "IBM Cloud region (e.g., us-south, us-east, eu-de, eu-gb, jp-tok, au-syd)"
  type        = string
  default     = "us-south"
}

variable "resource_group" {
  description = "IBM Cloud resource group name"
  type        = string
  default     = "default"
}

# Core Configuration
variable "cluster_name" {
  description = "Name of the IKS cluster"
  type        = string
  default     = "mantl-iks"
}

variable "environment" {
  description = "Environment name (e.g., production, staging, development)"
  type        = string
  default     = "production"
}

variable "kubernetes_version" {
  description = "Kubernetes version for IKS cluster"
  type        = string
  default     = "1.31" # Latest stable
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = list(string)
  default     = ["mantl", "kubernetes"]
}

# Network Configuration
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.240.0.0/16"
}

variable "zones_count" {
  description = "Number of zones to deploy to (1-3)"
  type        = number
  default     = 3
  validation {
    condition     = var.zones_count >= 1 && var.zones_count <= 3
    error_message = "zones_count must be between 1 and 3"
  }
}

variable "enable_public_gateway" {
  description = "Create public gateways for subnets (for internet access)"
  type        = bool
  default     = true
}

# Security Configuration
variable "disable_public_endpoint" {
  description = "Disable public service endpoint (private only)"
  type        = bool
  default     = true # Private by default
}

variable "allowed_ssh_cidrs" {
  description = "CIDR blocks allowed to SSH to nodes"
  type        = list(string)
  default     = [] # No SSH by default
}

variable "enable_flow_logs" {
  description = "Enable VPC Flow Logs for security monitoring"
  type        = bool
  default     = true
}

variable "flow_logs_retention_days" {
  description = "Number of days to retain flow logs"
  type        = number
  default     = 90
}

variable "enable_key_protect" {
  description = "Enable IBM Key Protect for secrets encryption"
  type        = bool
  default     = true
}

variable "key_protect_private_endpoint" {
  description = "Use private endpoint for Key Protect"
  type        = bool
  default     = true
}

variable "key_protect_force_delete" {
  description = "Force delete Key Protect keys on destroy"
  type        = bool
  default     = false
}

# System Worker Pool Configuration
variable "system_worker_flavor" {
  description = "Worker node flavor for system pool (e.g., bx2.4x16 = 4 vCPU, 16 GB RAM)"
  type        = string
  default     = "bx2.4x16"
}

variable "system_workers_per_zone" {
  description = "Number of system workers per zone"
  type        = number
  default     = 1
}

variable "taint_system_pool" {
  description = "Apply CriticalAddonsOnly taint to system pool"
  type        = bool
  default     = true
}

# Workload Worker Pool Configuration
variable "workload_worker_flavor" {
  description = "Worker node flavor for workload pool"
  type        = string
  default     = "bx2.4x16" # 4 vCPU, 16 GB RAM
}

variable "workload_workers_per_zone" {
  description = "Number of workload workers per zone"
  type        = number
  default     = 1
}

# Spot Worker Pool Configuration (optional cost savings)
variable "enable_spot_pool" {
  description = "Enable spot instances worker pool for cost savings"
  type        = bool
  default     = false
}

variable "spot_worker_flavor" {
  description = "Worker node flavor for spot pool"
  type        = string
  default     = "bx2.4x16"
}

variable "spot_workers_per_zone" {
  description = "Number of spot workers per zone"
  type        = number
  default     = 1
}

# Cluster Configuration
variable "entitlement" {
  description = "Entitlement for OCP on IBM Cloud (cloud_pak for OCP)"
  type        = string
  default     = null
}

# Backup Storage
variable "enable_backup_bucket" {
  description = "Create Cloud Object Storage bucket for backups (S3-compatible)"
  type        = bool
  default     = true
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30
}

# Logging and Monitoring
variable "enable_log_analysis" {
  description = "Create IBM Log Analysis instance for cluster logging"
  type        = bool
  default     = true
}

variable "log_analysis_plan" {
  description = "IBM Log Analysis plan (7-day, 14-day, 30-day, lite)"
  type        = string
  default     = "7-day"
}

variable "enable_monitoring" {
  description = "Create IBM Cloud Monitoring instance (Sysdig)"
  type        = bool
  default     = true
}

variable "monitoring_plan" {
  description = "IBM Cloud Monitoring plan (graduated-tier, lite)"
  type        = string
  default     = "graduated-tier"
}
