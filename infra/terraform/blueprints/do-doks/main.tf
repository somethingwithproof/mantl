# SPDX-License-Identifier: Apache-2.0
provider "digitalocean" {}

provider "helm" {
  kubernetes = {
    host  = digitalocean_kubernetes_cluster.main.endpoint
    token = digitalocean_kubernetes_cluster.main.kube_config[0].token
    cluster_ca_certificate = base64decode(
      digitalocean_kubernetes_cluster.main.kube_config[0].cluster_ca_certificate
    )
  }
}

provider "kubernetes" {
  host  = digitalocean_kubernetes_cluster.main.endpoint
  token = digitalocean_kubernetes_cluster.main.kube_config[0].token
  cluster_ca_certificate = base64decode(
    digitalocean_kubernetes_cluster.main.kube_config[0].cluster_ca_certificate
  )
}

locals {
  name = var.cluster_name
  project_environment = contains(
    ["development", "staging", "production"],
    var.environment,
  ) ? var.environment : "development"

  common_tags = {
    managed_by  = "terraform"
    platform    = "mantl"
    environment = var.environment
    blueprint   = "digitalocean"
  }

  # DigitalOcean resources take a list of tag strings. Derive the standard tags
  # from common_tags (single source of truth) and append caller-supplied tags.
  tags = concat([for k, v in local.common_tags : "${k}:${v}"], var.tags)
}

# VPC for Kubernetes cluster
resource "digitalocean_vpc" "main" {
  name     = "${local.name}-vpc"
  region   = var.region
  ip_range = var.vpc_cidr
}

# Firewall for Kubernetes nodes
resource "digitalocean_firewall" "kubernetes_nodes" {
  name = "${local.name}-nodes-fw"
  tags = local.tags

  # Inbound rules
  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = var.allowed_ssh_cidrs
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "1-65535"
    source_addresses = [var.vpc_cidr]
  }

  inbound_rule {
    protocol         = "udp"
    port_range       = "1-65535"
    source_addresses = [var.vpc_cidr]
  }

  inbound_rule {
    protocol         = "icmp"
    source_addresses = [var.vpc_cidr]
  }

  # Outbound rules
  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

# Container Registry for private images
resource "digitalocean_container_registry" "main" {
  name                   = replace(local.name, "-", "")
  subscription_tier_slug = var.registry_tier
  region                 = var.registry_region
}

# System node pool (for platform components)
resource "digitalocean_kubernetes_cluster" "main" {
  name     = local.name
  region   = var.region
  version  = var.kubernetes_version
  vpc_uuid = digitalocean_vpc.main.id

  tags = local.tags

  # Auto-upgrade configuration
  auto_upgrade  = var.enable_auto_upgrade
  surge_upgrade = true

  # Maintenance window
  maintenance_policy {
    start_time = var.maintenance_start_time
    day        = var.maintenance_day
  }

  # System node pool (non-autoscaling, for platform stability)
  node_pool {
    name       = "system"
    size       = var.system_node_size
    node_count = var.system_node_count
    auto_scale = false

    tags = concat(local.tags, ["node-pool:system"])

    labels = {
      "node.kubernetes.io/role" = "system"
      "workload-type"           = "platform"
    }

    taint {
      key    = "CriticalAddonsOnly"
      value  = "true"
      effect = "NoSchedule"
    }
  }

  # Destroy all resources when cluster is destroyed
  destroy_all_associated_resources = var.destroy_all_resources
}

# Workload node pool (auto-scaling for application workloads)
resource "digitalocean_kubernetes_node_pool" "workload" {
  cluster_id = digitalocean_kubernetes_cluster.main.id

  name       = "workload"
  size       = var.workload_node_size
  auto_scale = true
  min_nodes  = var.workload_min_nodes
  max_nodes  = var.workload_max_nodes

  tags = concat(local.tags, ["node-pool:workload"])

  labels = {
    "node.kubernetes.io/role" = "workload"
    "workload-type"           = "application"
  }
}

# Optional: Spot instances node pool for cost savings
resource "digitalocean_kubernetes_node_pool" "spot" {
  count      = var.enable_spot_pool ? 1 : 0
  cluster_id = digitalocean_kubernetes_cluster.main.id

  name       = "spot"
  size       = var.spot_node_size
  auto_scale = true
  min_nodes  = var.spot_min_nodes
  max_nodes  = var.spot_max_nodes

  tags = concat(local.tags, ["node-pool:spot"])

  labels = {
    "node.kubernetes.io/role" = "workload"
    "workload-type"           = "application"
    "spot-instance"           = "true"
    "kubernetes.io/lifecycle" = "spot"
  }

  taint {
    key    = "spot-instance"
    value  = "true"
    effect = "NoSchedule"
  }
}

# Load Balancer for ingress
resource "digitalocean_loadbalancer" "ingress" {
  count = var.enable_ingress_lb ? 1 : 0

  name     = "${local.name}-ingress-lb"
  region   = var.region
  vpc_uuid = digitalocean_vpc.main.id

  forwarding_rule {
    entry_protocol  = "https"
    entry_port      = 443
    target_protocol = "https"
    target_port     = 443
    tls_passthrough = true
  }

  forwarding_rule {
    entry_protocol  = "http"
    entry_port      = 80
    target_protocol = "http"
    target_port     = 80
  }

  healthcheck {
    protocol                 = "http"
    port                     = 80
    path                     = "/healthz"
    check_interval_seconds   = 10
    response_timeout_seconds = 5
    unhealthy_threshold      = 3
    healthy_threshold        = 3
  }

  droplet_tag = "k8s:${digitalocean_kubernetes_cluster.main.id}"
}

# Monitoring - VPC Flow Logs (requires DO monitoring integration)
# Note: DigitalOcean doesn't have native VPC Flow Logs like AWS/GCP/Azure
# Instead, we enable monitoring at the cluster level and use DO Monitoring
resource "digitalocean_monitor_alert" "cluster_cpu_high" {
  count = var.enable_monitoring ? 1 : 0

  alerts {
    email = var.alert_emails
  }
  window      = "5m"
  type        = "v1/insights/droplet/cpu"
  compare     = "GreaterThan"
  value       = 90
  enabled     = true
  tags        = local.tags
  description = "${local.name} - High CPU usage detected"
}

resource "digitalocean_monitor_alert" "cluster_memory_high" {
  count = var.enable_monitoring ? 1 : 0

  alerts {
    email = var.alert_emails
  }
  window      = "5m"
  type        = "v1/insights/droplet/memory_utilization_percent"
  compare     = "GreaterThan"
  value       = 90
  enabled     = true
  tags        = local.tags
  description = "${local.name} - High memory usage detected"
}

resource "digitalocean_monitor_alert" "cluster_disk_high" {
  count = var.enable_monitoring ? 1 : 0

  alerts {
    email = var.alert_emails
  }
  window      = "5m"
  type        = "v1/insights/droplet/disk_utilization_percent"
  compare     = "GreaterThan"
  value       = 85
  enabled     = true
  tags        = local.tags
  description = "${local.name} - High disk usage detected"
}

# Project to organize resources
resource "digitalocean_project" "main" {
  name        = local.name
  description = "Mantl Kubernetes Platform - ${var.environment}"
  purpose     = "Kubernetes cluster and supporting infrastructure"
  environment = local.project_environment

  resources = concat(
    [digitalocean_kubernetes_cluster.main.urn],
    var.enable_ingress_lb ? [digitalocean_loadbalancer.ingress[0].urn] : []
  )
}

# Database cluster for platform services (optional)
resource "digitalocean_database_cluster" "platform" {
  count = var.enable_platform_db ? 1 : 0

  name       = "${local.name}-db"
  engine     = "pg"
  version    = var.postgres_version
  size       = var.postgres_size
  region     = var.region
  node_count = var.postgres_node_count

  private_network_uuid = digitalocean_vpc.main.id

  maintenance_window {
    day  = var.db_maintenance_day
    hour = var.db_maintenance_hour
  }

  tags = concat(local.tags, ["service:database"])
}

# Database firewall - only allow Kubernetes cluster
resource "digitalocean_database_firewall" "platform" {
  count      = var.enable_platform_db ? 1 : 0
  cluster_id = digitalocean_database_cluster.platform[0].id

  rule {
    type  = "k8s"
    value = digitalocean_kubernetes_cluster.main.id
  }
}

# Spaces bucket for backups (S3-compatible)
resource "digitalocean_spaces_bucket" "backups" {
  count  = var.enable_backup_bucket ? 1 : 0
  name   = "${replace(local.name, "-", "")}-backups"
  region = var.spaces_region
  acl    = "private"

  versioning {
    enabled = true
  }

  lifecycle_rule {
    id      = "delete-old-backups"
    enabled = true

    expiration {
      days = var.backup_retention_days
    }

    noncurrent_version_expiration {
      days = 7
    }
  }
}

# Outputs
output "cluster_id" {
  description = "The ID of the Kubernetes cluster"
  value       = digitalocean_kubernetes_cluster.main.id
}

output "cluster_name" {
  description = "The name of the Kubernetes cluster"
  value       = digitalocean_kubernetes_cluster.main.name
}

output "endpoint" {
  description = "The endpoint of the Kubernetes API server"
  value       = digitalocean_kubernetes_cluster.main.endpoint
}

output "cluster_ca_certificate" {
  description = "The base64 encoded cluster CA certificate"
  value       = digitalocean_kubernetes_cluster.main.kube_config[0].cluster_ca_certificate
  sensitive   = true
}

output "vpc_id" {
  description = "The ID of the VPC"
  value       = digitalocean_vpc.main.id
}

output "vpc_cidr" {
  description = "The CIDR block of the VPC"
  value       = digitalocean_vpc.main.ip_range
}

output "registry_endpoint" {
  description = "The endpoint of the container registry"
  value       = digitalocean_container_registry.main.endpoint
}

output "loadbalancer_ip" {
  description = "The IP address of the ingress load balancer"
  value       = var.enable_ingress_lb ? digitalocean_loadbalancer.ingress[0].ip : null
}

output "database_host" {
  description = "The host of the platform database"
  value       = var.enable_platform_db ? digitalocean_database_cluster.platform[0].host : null
  sensitive   = true
}

output "database_port" {
  description = "The port of the platform database"
  value       = var.enable_platform_db ? digitalocean_database_cluster.platform[0].port : null
}

output "database_uri" {
  description = "The connection URI of the platform database"
  value       = var.enable_platform_db ? digitalocean_database_cluster.platform[0].uri : null
  sensitive   = true
}

output "backup_bucket" {
  description = "The name of the backup bucket"
  value       = var.enable_backup_bucket ? digitalocean_spaces_bucket.backups[0].name : null
}

output "backup_bucket_region" {
  description = "The region of the backup bucket"
  value       = var.enable_backup_bucket ? digitalocean_spaces_bucket.backups[0].region : null
}

################################################################################
# Platform Security Module (Enterprise Features)
################################################################################

module "platform_security" {
  count  = var.enable_platform_security ? 1 : 0
  source = "../../modules/platform-security"

  # Cilium + Hubble (Flow Logs Alternative)
  enable_cilium             = var.enable_cilium
  hubble_ui_ingress_enabled = var.hubble_ui_ingress_enabled
  hubble_ui_hosts           = var.hubble_ui_hosts
  enable_flow_logs_export   = var.enable_flow_logs_export
  flow_logs_s3_bucket       = var.enable_backup_bucket ? digitalocean_spaces_bucket.backups[0].name : ""

  # Sealed Secrets (KMS Alternative)
  enable_sealed_secrets = var.enable_sealed_secrets

  # External Secrets Operator
  enable_external_secrets = var.enable_external_secrets

  # Falco (Runtime Security)
  enable_falco      = var.enable_falco
  falco_webhook_url = var.falco_webhook_url
  falco_ui_enabled  = var.falco_ui_enabled

  # Tailscale VPN (Private Endpoint Alternative)
  enable_tailscale        = var.enable_tailscale
  tailscale_client_id     = var.tailscale_client_id
  tailscale_client_secret = var.tailscale_client_secret

  # Cert-Manager
  enable_cert_manager = var.enable_cert_manager

  # Security Dashboard
  enable_security_dashboard = var.enable_security_dashboard

  depends_on = [
    digitalocean_kubernetes_cluster.main,
    digitalocean_kubernetes_node_pool.workload,
  ]
}

output "platform_security_enabled_features" {
  description = "Summary of enabled platform security features"
  value       = var.enable_platform_security ? module.platform_security[0].security_features_summary : {}
}
