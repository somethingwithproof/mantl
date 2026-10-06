# SPDX-License-Identifier: Apache-2.0
output "cluster_name" {
  description = "Name of the cluster"
  value       = var.cluster_name
}

output "kubernetes_version" {
  description = "Version of Kubernetes"
  value       = var.kubernetes_version
}

output "environment" {
  description = "Deployment environment"
  value       = var.environment
}

output "region" {
  description = "Configured cloud region"
  value       = var.region
}

output "tags" {
  description = "Configured resource tags"
  value       = var.tags
}
