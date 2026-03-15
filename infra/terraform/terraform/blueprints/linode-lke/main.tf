terraform {
  required_version = ">= 1.10.0"
  required_providers {
    linode     = { source = "linode/linode", version = ">= 3.0" }
    helm       = { source = "hashicorp/helm", version = "~> 2.12" }
    kubernetes = { source = "hashicorp/kubernetes", version = "~> 2.25" }
  }
}

provider "linode" {}

provider "helm" {
  kubernetes {
    host  = linode_lke_cluster.main.api_endpoints[0]
    token = linode_lke_cluster.main.kubeconfig
    cluster_ca_certificate = base64decode(
      linode_lke_cluster.main.kubeconfig
    )
  }
}

provider "kubernetes" {
  host  = linode_lke_cluster.main.api_endpoints[0]
  token = linode_lke_cluster.main.kubeconfig
  cluster_ca_certificate = base64decode(
    linode_lke_cluster.main.kubeconfig
  )
}

locals {
  name = var.cluster_name
  tags = concat(var.tags, ["mantl-platform", "environment:${var.environment}"])
}

# VLAN for private networking (optional, enhanced security)
resource "linode_vlan" "main" {
  count  = var.enable_vlan ? 1 : 0
  label  = "${local.name}-vlan"
  region = var.region

  linodes = []  # Will be populated automatically by LKE
}

# Cloud Firewall for Kubernetes cluster
resource "linode_firewall" "kubernetes" {
  label = "${local.name}-firewall"
  tags  = local.tags

  # Inbound rules
  inbound {
    label    = "ssh-internal"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "22"
    ipv4     = var.allowed_ssh_cidrs
  }

  inbound {
    label    = "https-public"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "443"
    ipv4     = ["0.0.0.0/0"]
    ipv6     = ["::/0"]
  }

  inbound {
    label    = "http-public"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "80"
    ipv4     = ["0.0.0.0/0"]
    ipv6     = ["::/0"]
  }

  inbound {
    label    = "kubelet-api"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "10250"
    ipv4     = [var.cluster_cidr]
  }

  inbound {
    label    = "kube-proxy"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "10256"
    ipv4     = [var.cluster_cidr]
  }

  inbound {
    label    = "nodeport-services"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "30000-32767"
    ipv4     = var.allowed_nodeport_cidrs
  }

  inbound_policy = "DROP"

  # Outbound rules
  outbound {
    label    = "allow-all-outbound"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "1-65535"
    ipv4     = ["0.0.0.0/0"]
    ipv6     = ["::/0"]
  }

  outbound {
    label    = "allow-udp-outbound"
    action   = "ACCEPT"
    protocol = "UDP"
    ports    = "1-65535"
    ipv4     = ["0.0.0.0/0"]
    ipv6     = ["::/0"]
  }

  outbound_policy = "ACCEPT"

  # Will be attached to LKE nodes automatically via tags
  linodes = []
}

# LKE Cluster
resource "linode_lke_cluster" "main" {
  label       = local.name
  k8s_version = var.kubernetes_version
  region      = var.region
  tags        = local.tags

  # High availability control plane
  control_plane {
    high_availability = var.enable_ha_control_plane
  }

  # System node pool (non-autoscaling, for platform components)
  pool {
    type  = var.system_node_type
    count = var.system_node_count

    tags = concat(local.tags, ["node-pool:system"])

    autoscaler {
      min = var.system_node_count
      max = var.system_node_count # Fixed size for stability
    }
  }

  # Workload node pool (autoscaling, for applications)
  dynamic "pool" {
    for_each = var.workload_pools
    content {
      type  = pool.value.node_type
      count = pool.value.node_count

      tags = concat(local.tags, ["node-pool:${pool.value.name}"])

      autoscaler {
        min = pool.value.min_nodes
        max = pool.value.max_nodes
      }
    }
  }
}

# Attach firewall to LKE cluster nodes
resource "null_resource" "attach_firewall" {
  depends_on = [
    linode_lke_cluster.main,
    linode_firewall.kubernetes
  ]

  triggers = {
    cluster_id  = linode_lke_cluster.main.id
    firewall_id = linode_firewall.kubernetes.id
  }

  provisioner "local-exec" {
    command = <<-EOT
      # Get all node IDs from the LKE cluster
      NODE_IDS=$(linode-cli lke pools-list ${linode_lke_cluster.main.id} --json | \
        jq -r '.[].linodes[].id')

      # Attach firewall to each node
      for NODE_ID in $NODE_IDS; do
        linode-cli firewalls device-create ${linode_firewall.kubernetes.id} \
          --id $NODE_ID --type linode || true
      done
    EOT
  }
}

# Object Storage bucket for backups (S3-compatible)
resource "linode_object_storage_bucket" "backups" {
  count  = var.enable_backup_bucket ? 1 : 0
  label  = "${replace(local.name, "-", "")}-backups"
  region = var.object_storage_region
  acl    = "private"

  lifecycle_rule {
    id      = "delete-old-backups"
    enabled = true

    expiration {
      days = var.backup_retention_days
    }
  }

  versioning = true
}

# Object Storage keys for backup access
resource "linode_object_storage_key" "backups" {
  count  = var.enable_backup_bucket ? 1 : 0
  label  = "${local.name}-backup-key"

  bucket_access {
    bucket_name = linode_object_storage_bucket.backups[0].label
    region      = var.object_storage_region
    permissions = "read_write"
  }
}

# NodeBalancer for ingress (optional)
resource "linode_nodebalancer" "ingress" {
  count  = var.enable_ingress_lb ? 1 : 0
  label  = "${local.name}-ingress-lb"
  region = var.region
  tags   = concat(local.tags, ["service:ingress"])

  client_conn_throttle = var.lb_throttle_connections
}

# NodeBalancer config for HTTPS
resource "linode_nodebalancer_config" "https" {
  count          = var.enable_ingress_lb ? 1 : 0
  nodebalancer_id = linode_nodebalancer.ingress[0].id

  protocol = "tcp"
  port     = 443

  check          = "connection"
  check_interval = 10
  check_timeout  = 5
  check_attempts = 3

  algorithm       = "roundrobin"
  stickiness      = "none"
  cipher_suite    = "recommended"
}

# NodeBalancer config for HTTP
resource "linode_nodebalancer_config" "http" {
  count          = var.enable_ingress_lb ? 1 : 0
  nodebalancer_id = linode_nodebalancer.ingress[0].id

  protocol = "tcp"
  port     = 80

  check          = "connection"
  check_interval = 10
  check_timeout  = 5
  check_attempts = 3

  algorithm  = "roundrobin"
  stickiness = "none"
}

# Block Storage volume for persistent data (optional)
resource "linode_volume" "platform_data" {
  count  = var.enable_platform_volume ? 1 : 0
  label  = "${local.name}-platform-data"
  region = var.region
  size   = var.platform_volume_size
  tags   = concat(local.tags, ["service:platform-data"])
}

# Outputs
output "cluster_id" {
  description = "The ID of the LKE cluster"
  value       = linode_lke_cluster.main.id
}

output "cluster_name" {
  description = "The name of the LKE cluster"
  value       = linode_lke_cluster.main.label
}

output "api_endpoints" {
  description = "The API endpoints of the LKE cluster"
  value       = linode_lke_cluster.main.api_endpoints
}

output "kubeconfig" {
  description = "The kubeconfig for the LKE cluster (base64 encoded)"
  value       = linode_lke_cluster.main.kubeconfig
  sensitive   = true
}

output "status" {
  description = "The status of the LKE cluster"
  value       = linode_lke_cluster.main.status
}

output "firewall_id" {
  description = "The ID of the cluster firewall"
  value       = linode_firewall.kubernetes.id
}

output "backup_bucket" {
  description = "The name of the backup bucket"
  value       = var.enable_backup_bucket ? linode_object_storage_bucket.backups[0].label : null
}

output "backup_bucket_region" {
  description = "The region of the backup bucket"
  value       = var.enable_backup_bucket ? linode_object_storage_bucket.backups[0].region : null
}

output "backup_access_key" {
  description = "The access key for the backup bucket"
  value       = var.enable_backup_bucket ? linode_object_storage_key.backups[0].access_key : null
  sensitive   = true
}

output "backup_secret_key" {
  description = "The secret key for the backup bucket"
  value       = var.enable_backup_bucket ? linode_object_storage_key.backups[0].secret_key : null
  sensitive   = true
}

output "nodebalancer_ip" {
  description = "The IP address of the ingress NodeBalancer"
  value       = var.enable_ingress_lb ? linode_nodebalancer.ingress[0].ipv4 : null
}

output "platform_volume_id" {
  description = "The ID of the platform data volume"
  value       = var.enable_platform_volume ? linode_volume.platform_data[0].id : null
}

################################################################################
# Platform Security Module (Enterprise Features)
################################################################################

module "platform_security" {
  source = "../../modules/platform-security"

  # Cilium + Hubble (Flow Logs Alternative)
  enable_cilium               = var.enable_cilium
  hubble_ui_ingress_enabled   = var.hubble_ui_ingress_enabled
  hubble_ui_hosts             = var.hubble_ui_hosts
  enable_flow_logs_export     = var.enable_flow_logs_export
  flow_logs_s3_bucket         = var.enable_backup_bucket ? linode_object_storage_bucket.backups[0].label : ""

  # Sealed Secrets (KMS Alternative)
  enable_sealed_secrets = var.enable_sealed_secrets

  # External Secrets Operator
  enable_external_secrets = var.enable_external_secrets

  # Falco (Runtime Security)
  enable_falco       = var.enable_falco
  falco_webhook_url  = var.falco_webhook_url
  falco_ui_enabled   = var.falco_ui_enabled

  # Tailscale VPN (Private Endpoint Alternative)
  enable_tailscale        = var.enable_tailscale
  tailscale_client_id     = var.tailscale_client_id
  tailscale_client_secret = var.tailscale_client_secret

  # Cert-Manager
  enable_cert_manager = var.enable_cert_manager

  # Security Dashboard
  enable_security_dashboard = var.enable_security_dashboard

  depends_on = [
    linode_lke_cluster.main,
  ]
}

output "platform_security_enabled_features" {
  description = "Summary of enabled platform security features"
  value       = module.platform_security.security_features_summary
}
