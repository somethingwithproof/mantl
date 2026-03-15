################################################################################
# General
################################################################################

variable "cluster_name" {
  description = "Name of the GKE cluster"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "kubernetes_version" {
  description = "Kubernetes version for GKE cluster"
  type        = string
  default     = "1.31"
}

variable "environment" {
  description = "Deployment environment (dev, staging, prod)"
  type        = string
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "tags" {
  description = "Additional labels for all resources"
  type        = map(string)
  default     = {
    project    = "mantl"
    managed_by = "terraform"
  }
}
