# Cluster identity and VCN outputs are defined in main.tf. This file holds
# only the remaining, non-duplicate outputs.

output "cluster_endpoint" {
  description = "Kubernetes API endpoint"
  value       = oci_containerengine_cluster.oke_cluster.endpoints[0].public_endpoint
}

output "node_pool_id" {
  description = "OCID of the node pool"
  value       = oci_containerengine_node_pool.oke_node_pool.id
}

output "region" {
  description = "OCI region"
  value       = var.region
}

output "kubeconfig_command" {
  description = "Command to generate kubeconfig"
  value       = "oci ce cluster create-kubeconfig --cluster-id ${oci_containerengine_cluster.oke_cluster.id} --file $HOME/.kube/config --region ${var.region} --token-version 2.0.0"
}
