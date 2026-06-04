################################################################################
# General
################################################################################

variable "project" {
  description = "GCP project ID"
  type        = string
}

variable "cluster_name" {
  description = "Name of the GKE cluster"
  type        = string
  default     = "mantl"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "kubernetes_version" {
  description = "Kubernetes version for GKE cluster"
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

variable "network" {
  description = "VPC network name"
  type        = string
  default     = "default"
}

variable "subnet_cidr" {
  description = "CIDR block for GKE subnet"
  type        = string
  default     = "10.0.0.0/20"
}

variable "ip_range_pods" {
  description = "CIDR block for pods"
  type        = string
  default     = "10.4.0.0/14"
}

variable "ip_range_services" {
  description = "CIDR block for services"
  type        = string
  default     = "10.8.0.0/20"
}

################################################################################
# Security
################################################################################

variable "enable_private_endpoint" {
  description = "Enable private endpoint (NOT recommended for development)"
  type        = bool
  default     = true # Default to true for production security
}

variable "master_authorized_networks" {
  description = "List of CIDR blocks that can access the Kubernetes master"
  type = list(object({
    cidr_block   = string
    display_name = string
  }))
  default = []
}

variable "enable_binary_authorization" {
  description = "Enable Binary Authorization for deployment policy enforcement"
  type        = bool
  default     = false
}

################################################################################
# Node Pools - System
################################################################################

variable "system_node_machine_type" {
  description = "Machine type for system node pool"
  type        = string
  default     = "e2-standard-4"
}

variable "system_node_min_count" {
  description = "Minimum number of system nodes"
  type        = number
  default     = 2
}

variable "system_node_max_count" {
  description = "Maximum number of system nodes"
  type        = number
  default     = 4
}

################################################################################
# Node Pools - Workload
################################################################################

variable "workload_node_machine_type" {
  description = "Machine type for workload node pool"
  type        = string
  default     = "e2-standard-4"
}

variable "workload_node_min_count" {
  description = "Minimum number of workload nodes"
  type        = number
  default     = 2
}

variable "workload_node_max_count" {
  description = "Maximum number of workload nodes"
  type        = number
  default     = 10
}

variable "workload_node_spot" {
  description = "Use spot instances for workload nodes (cost savings)"
  type        = bool
  default     = false
}
