# Cluster identity and VCN outputs are defined in main.tf. This file holds
# only the remaining, non-duplicate outputs.

output "cluster_endpoint" {
  description = "Kubernetes API endpoint"
  value       = oci_containerengine_cluster.oke_cluster.endpoints[0].public_endpoint
}

output "system_node_pool_id" {
  description = "OCID of the system node pool"
  value       = oci_containerengine_node_pool.system_pool.id
}

output "workload_node_pool_id" {
  description = "OCID of the workload node pool"
  value       = oci_containerengine_node_pool.workload_pool.id
}

output "region" {
  description = "OCI region"
  value       = var.region
}

output "kubeconfig_command" {
  description = "Command to generate kubeconfig"
  value       = "oci ce cluster create-kubeconfig --cluster-id ${oci_containerengine_cluster.oke_cluster.id} --file $HOME/.kube/config --region ${var.region} --token-version 2.0.0"
}
