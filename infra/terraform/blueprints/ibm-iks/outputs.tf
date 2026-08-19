# Cluster identity, VPC, ingress, and the core cluster outputs are defined in
# main.tf. This file holds only the remaining, non-duplicate outputs.

output "cluster_endpoint" {
  description = "Public service endpoint for the cluster"
  value       = ibm_container_vpc_cluster.iks_cluster.public_service_endpoint_url
}

output "subnet_ids" {
  description = "IDs of created subnets"
  value       = ibm_is_subnet.iks_subnet[*].id
}

output "kube_version" {
  description = "Kubernetes version"
  value       = ibm_container_vpc_cluster.iks_cluster.kube_version
}

output "resource_group_id" {
  description = "Resource group ID"
  value       = data.ibm_resource_group.resource_group.id
}

output "kubeconfig_command" {
  description = "Command to configure kubectl"
  value       = "ibmcloud ks cluster config --cluster ${ibm_container_vpc_cluster.iks_cluster.id} --admin"
}
