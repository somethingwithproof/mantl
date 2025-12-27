variable "ibmcloud_api_key" {
  description = "IBM Cloud API key"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "IBM Cloud region (e.g., us-south, us-east, eu-de, jp-tok)"
  type        = string
  default     = "us-south"
}

variable "resource_group" {
  description = "IBM Cloud resource group name"
  type        = string
  default     = "default"
}

variable "cluster_name" {
  description = "Name of the IKS cluster"
  type        = string
  default     = "mantl-iks"
}

variable "kubernetes_version" {
  description = "Kubernetes version for IKS cluster"
  type        = string
  default     = "1.31"
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

variable "worker_flavor" {
  description = "Worker node flavor (e.g., bx2.4x16, bx2.8x32)"
  type        = string
  default     = "bx2.4x16" # 4 vCPU, 16 GB RAM
}

variable "workers_per_zone" {
  description = "Number of worker nodes per zone"
  type        = number
  default     = 1
}

variable "disable_public_endpoint" {
  description = "Disable public service endpoint (private only)"
  type        = bool
  default     = false
}

variable "entitlement" {
  description = "Entitlement for OCP on IBM Cloud (cloud_pak for OCP)"
  type        = string
  default     = null
}

variable "create_additional_pool" {
  description = "Create an additional worker pool"
  type        = bool
  default     = false
}

variable "additional_pool_flavor" {
  description = "Flavor for additional worker pool"
  type        = string
  default     = "bx2.4x16"
}

variable "additional_pool_workers_per_zone" {
  description = "Workers per zone in additional pool"
  type        = number
  default     = 1
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = list(string)
  default     = ["mantl", "kubernetes"]
}
