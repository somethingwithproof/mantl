################################################################################
# General
################################################################################

variable "resource_group_name" {
  description = "Name of the Azure Resource Group"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "East US"
}

variable "cluster_name" {
  description = "Name of the AKS cluster"
  type        = string
  default     = "mantl"
}

variable "environment" {
  description = "Environment name (e.g., production, staging, development)"
  type        = string
  default     = "production"
}

variable "kubernetes_version" {
  description = "Kubernetes version for AKS cluster"
  type        = string
  default     = "1.31" # Latest stable version
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}

################################################################################
# Network
################################################################################

variable "vnet_cidr" {
  description = "CIDR block for VNet"
  type        = string
  default     = "10.0.0.0/16"
}

variable "aks_subnet_cidr" {
  description = "CIDR block for AKS subnet"
  type        = string
  default     = "10.0.0.0/20"
}

variable "service_cidr" {
  description = "CIDR block for Kubernetes services"
  type        = string
  default     = "10.1.0.0/16"
}

variable "dns_service_ip" {
  description = "IP address for Kubernetes DNS service (must be within service_cidr)"
  type        = string
  default     = "10.1.0.10"
}

################################################################################
# Security
################################################################################

variable "enable_private_endpoint" {
  description = "Enable private endpoint (NOT recommended for development)"
  type        = bool
  default     = true # Default to true for production security
}

variable "authorized_ip_ranges" {
  description = "List of authorized IP ranges that can access the Kubernetes API server (when not using private endpoint)"
  type        = list(string)
  default     = []
}

variable "admin_group_object_ids" {
  description = "Azure AD group object IDs for cluster admin access"
  type        = list(string)
  default     = []
}

################################################################################
# Monitoring & Logging
################################################################################

variable "log_retention_days" {
  description = "Log retention in days for Log Analytics workspace"
  type        = number
  default     = 30
}

variable "flow_logs_retention_days" {
  description = "Flow logs retention in days"
  type        = number
  default     = 7
}

variable "key_vault_key_expiration_date" {
  description = "RFC 3339 expiration date for the AKS Key Vault encryption key"
  type        = string
  default     = null
}

################################################################################
# Node Pools - System
################################################################################

variable "system_node_vm_size" {
  description = "VM size for system node pool"
  type        = string
  default     = "Standard_D4s_v5"
}

variable "system_node_count" {
  description = "Initial number of system nodes"
  type        = number
  default     = 2
}

variable "system_node_min_count" {
  description = "Minimum number of system nodes (auto-scaling)"
  type        = number
  default     = 2
}

variable "system_node_max_count" {
  description = "Maximum number of system nodes (auto-scaling)"
  type        = number
  default     = 4
}

################################################################################
# Node Pools - Workload
################################################################################

variable "workload_node_vm_size" {
  description = "VM size for workload node pool"
  type        = string
  default     = "Standard_D4s_v5"
}

variable "workload_node_min_count" {
  description = "Minimum number of workload nodes (auto-scaling)"
  type        = number
  default     = 2
}

variable "workload_node_max_count" {
  description = "Maximum number of workload nodes (auto-scaling)"
  type        = number
  default     = 10
}

variable "workload_node_spot" {
  description = "Use spot instances for workload nodes (cost savings)"
  type        = bool
  default     = false
}
