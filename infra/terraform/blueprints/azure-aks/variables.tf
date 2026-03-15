################################################################################
# General
################################################################################

variable "cluster_name" {
  description = "Name of the AKS cluster"
  type        = string
}

variable "region" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

variable "kubernetes_version" {
  description = "Kubernetes version for AKS cluster"
  type        = string
  default     = "1.31"
}

variable "environment" {
  description = "Deployment environment (dev, staging, prod)"
  type        = string
}

variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {
    Project   = "Mantl"
    ManagedBy = "Terraform"
  }
}
