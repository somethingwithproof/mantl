provider "openstack" {
  # Credentials are typically provided via OS_* environment variables
  # or clouds.yaml configuration file
}

locals {
  name = var.cluster_name
  common_tags = [
    "managed-by:terraform",
    "platform:mantl",
    "environment:${var.environment}"
  ]
}

# Get available Magnum cluster template (or create one if needed)
data "openstack_containerinfra_clustertemplate_v1" "k8s_template" {
  count = var.use_existing_template ? 1 : 0
  name  = var.existing_template_name
}

# Create cluster template if not using existing
resource "openstack_containerinfra_clustertemplate_v1" "k8s_template" {
  count = var.use_existing_template ? 0 : 1

  name                  = "${local.name}-template"
  coe                   = "kubernetes"
  image                 = var.node_image_name
  flavor                = var.master_flavor
  master_flavor         = var.master_flavor
  docker_storage_driver = "overlay2"
  docker_volume_size    = var.docker_volume_size

  # Network configuration
  network_driver      = var.network_driver
  volume_driver       = "cinder"
  external_network_id = data.openstack_networking_network_v2.external.id
  fixed_network       = openstack_networking_network_v2.cluster_network.id
  fixed_subnet        = openstack_networking_subnet_v2.cluster_subnet.id
  dns_nameserver      = var.dns_nameserver
  http_proxy          = var.http_proxy
  https_proxy         = var.https_proxy
  no_proxy            = var.no_proxy

  # Security features
  tls_disabled        = false
  registry_enabled    = var.enable_registry
  floating_ip_enabled = var.enable_floating_ip
  master_lb_enabled   = var.enable_master_lb

  # Additional features
  labels = merge(
    var.template_labels,
    {
      "kube_tag"               = var.kubernetes_version
      "cloud_provider_enabled" = "true"
      "cinder_csi_enabled"     = "true"
      "octavia_provider"       = "amphora"
      "auto_healing_enabled"   = var.enable_auto_healing ? "true" : "false"
      "auto_scaling_enabled"   = var.enable_auto_scaling ? "true" : "false"
      "min_node_count"         = tostring(var.min_node_count)
      "max_node_count"         = tostring(var.max_node_count)
    }
  )

  # Hide sensitive information
  hidden = false
  public = false
}

# Get external network for floating IPs
data "openstack_networking_network_v2" "external" {
  name     = var.external_network_name
  external = true
}

# Create private network for cluster
resource "openstack_networking_network_v2" "cluster_network" {
  name           = "${local.name}-network"
  admin_state_up = true
  tags           = local.common_tags
}

# Create subnet for cluster nodes
resource "openstack_networking_subnet_v2" "cluster_subnet" {
  name            = "${local.name}-subnet"
  network_id      = openstack_networking_network_v2.cluster_network.id
  cidr            = var.cluster_cidr
  ip_version      = 4
  dns_nameservers = [var.dns_nameserver]
  tags            = local.common_tags

  allocation_pool {
    start = cidrhost(var.cluster_cidr, 10)
    end   = cidrhost(var.cluster_cidr, 250)
  }
}

# Create router for external connectivity
resource "openstack_networking_router_v2" "cluster_router" {
  name                = "${local.name}-router"
  external_network_id = data.openstack_networking_network_v2.external.id
  admin_state_up      = true
  tags                = local.common_tags
}

# Attach subnet to router
resource "openstack_networking_router_interface_v2" "cluster_router_interface" {
  router_id = openstack_networking_router_v2.cluster_router.id
  subnet_id = openstack_networking_subnet_v2.cluster_subnet.id
}

# Security group for cluster nodes
resource "openstack_networking_secgroup_v2" "cluster_nodes" {
  name        = "${local.name}-nodes-sg"
  description = "Security group for Kubernetes cluster nodes"
  tags        = local.common_tags
}

# Allow all traffic within cluster network
resource "openstack_networking_secgroup_rule_v2" "intra_cluster" {
  direction         = "ingress"
  ethertype         = "IPv4"
  security_group_id = openstack_networking_secgroup_v2.cluster_nodes.id
  remote_group_id   = openstack_networking_secgroup_v2.cluster_nodes.id
}

# Allow SSH from specific CIDRs (if provided)
resource "openstack_networking_secgroup_rule_v2" "ssh" {
  count             = length(var.allowed_ssh_cidrs) > 0 ? length(var.allowed_ssh_cidrs) : 0
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = var.allowed_ssh_cidrs[count.index]
  security_group_id = openstack_networking_secgroup_v2.cluster_nodes.id
}

# Allow Kubernetes API access
resource "openstack_networking_secgroup_rule_v2" "kube_api" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  remote_ip_prefix  = var.api_server_access_cidr
  security_group_id = openstack_networking_secgroup_v2.cluster_nodes.id
}

# Allow HTTPS traffic for ingress
resource "openstack_networking_secgroup_rule_v2" "https" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.cluster_nodes.id
}

# Allow HTTP traffic for ingress
resource "openstack_networking_secgroup_rule_v2" "http" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.cluster_nodes.id
}

# Allow all outbound traffic
resource "openstack_networking_secgroup_rule_v2" "outbound" {
  direction         = "egress"
  ethertype         = "IPv4"
  security_group_id = openstack_networking_secgroup_v2.cluster_nodes.id
}

# Create Magnum Kubernetes cluster
resource "openstack_containerinfra_cluster_v1" "k8s_cluster" {
  name                = local.name
  cluster_template_id = var.use_existing_template ? data.openstack_containerinfra_clustertemplate_v1.k8s_template[0].id : openstack_containerinfra_clustertemplate_v1.k8s_template[0].id
  master_count        = var.master_count
  node_count          = var.initial_node_count
  keypair             = var.ssh_keypair_name
  flavor              = var.worker_flavor

  # Timeouts for cluster creation
  create_timeout = var.cluster_create_timeout

  labels = {
    "environment" = var.environment
    "managed_by"  = "terraform"
  }

  depends_on = [
    openstack_networking_router_interface_v2.cluster_router_interface
  ]
}

# Swift container for backups (S3-compatible via Swift API)
resource "openstack_objectstorage_container_v1" "backups" {
  count        = var.enable_backup_container ? 1 : 0
  name         = "${local.name}-backups"
  content_type = "application/octet-stream"

  metadata = {
    managed-by  = "terraform"
    environment = var.environment
    purpose     = "kubernetes-backups"
  }

  versioning = true
}

# Application credentials for backup access
resource "openstack_identity_application_credential_v3" "backup_credentials" {
  count       = var.enable_backup_container ? 1 : 0
  name        = "${local.name}-backup-access"
  description = "Application credentials for backup access to Swift"

  roles = ["member"]

  # Restrict to Swift only
  unrestricted = false
}

# Outputs
output "cluster_id" {
  description = "The ID of the Kubernetes cluster"
  value       = openstack_containerinfra_cluster_v1.k8s_cluster.id
}

output "cluster_name" {
  description = "The name of the Kubernetes cluster"
  value       = openstack_containerinfra_cluster_v1.k8s_cluster.name
}

output "master_addresses" {
  description = "The IP addresses of master nodes"
  value       = openstack_containerinfra_cluster_v1.k8s_cluster.master_addresses
}

output "node_addresses" {
  description = "The IP addresses of worker nodes"
  value       = openstack_containerinfra_cluster_v1.k8s_cluster.node_addresses
}

output "api_address" {
  description = "The API server address"
  value       = openstack_containerinfra_cluster_v1.k8s_cluster.api_address
}

output "kubeconfig" {
  description = "The kubeconfig for accessing the cluster"
  value       = openstack_containerinfra_cluster_v1.k8s_cluster.kubeconfig
  sensitive   = true
}

output "network_id" {
  description = "The ID of the cluster network"
  value       = openstack_networking_network_v2.cluster_network.id
}

output "subnet_id" {
  description = "The ID of the cluster subnet"
  value       = openstack_networking_subnet_v2.cluster_subnet.id
}

output "security_group_id" {
  description = "The ID of the cluster security group"
  value       = openstack_networking_secgroup_v2.cluster_nodes.id
}

output "backup_container" {
  description = "The name of the backup Swift container"
  value       = var.enable_backup_container ? openstack_objectstorage_container_v1.backups[0].name : null
}

output "backup_credentials_id" {
  description = "The ID of backup application credentials"
  value       = var.enable_backup_container ? openstack_identity_application_credential_v3.backup_credentials[0].id : null
  sensitive   = true
}

output "backup_credentials_secret" {
  description = "The secret for backup application credentials"
  value       = var.enable_backup_container ? openstack_identity_application_credential_v3.backup_credentials[0].secret : null
  sensitive   = true
}
