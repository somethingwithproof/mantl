output "cluster_id" {
  description = "ID of the IKS cluster"
  value       = ibm_container_vpc_cluster.iks_cluster.id
}

output "cluster_name" {
  description = "Name of the IKS cluster"
  value       = ibm_container_vpc_cluster.iks_cluster.name
}

output "cluster_endpoint" {
  description = "Public service endpoint for the cluster"
  value       = ibm_container_vpc_cluster.iks_cluster.public_service_endpoint_url
}

output "cluster_ingress_hostname" {
  description = "Ingress hostname for the cluster"
  value       = ibm_container_vpc_cluster.iks_cluster.ingress_hostname
}

output "cluster_ingress_secret" {
  description = "Ingress secret name"
  value       = ibm_container_vpc_cluster.iks_cluster.ingress_secret
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = ibm_is_vpc.iks_vpc.id
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
