terraform {
  required_version = ">= 1.10.0"
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = ">= 1.65.0"
    }
  }
}

provider "ibm" {
  ibmcloud_api_key = var.ibmcloud_api_key
  region           = var.region
}

locals {
  name = var.cluster_name
  tags = concat(var.tags, [
    "mantl-platform",
    "environment:${var.environment}",
    "managed-by:terraform"
  ])
}

# Get resource group
data "ibm_resource_group" "resource_group" {
  name = var.resource_group
}

# Get available zones in region
data "ibm_is_zones" "zones" {
  region = var.region
}

# Create VPC
resource "ibm_is_vpc" "iks_vpc" {
  name           = "${local.name}-vpc"
  resource_group = data.ibm_resource_group.resource_group.id
  tags           = local.tags

  address_prefix_management = "manual"
}

# Create address prefixes for each zone
resource "ibm_is_vpc_address_prefix" "zone_prefix" {
  count = var.zones_count
  name  = "${local.name}-prefix-${count.index + 1}"
  zone  = data.ibm_is_zones.zones.zones[count.index]
  vpc   = ibm_is_vpc.iks_vpc.id
  cidr  = cidrsubnet(var.vpc_cidr, 4, count.index)
}

# VPC Flow Logs for security monitoring
resource "ibm_is_flow_log" "vpc_flow_logs" {
  count          = var.enable_flow_logs ? 1 : 0
  name           = "${local.name}-flow-logs"
  target         = ibm_is_vpc.iks_vpc.id
  active         = true
  storage_bucket = ibm_cos_bucket.flow_logs[0].bucket_name
  resource_group = data.ibm_resource_group.resource_group.id
  tags           = local.tags
}

# Cloud Object Storage instance for flow logs
resource "ibm_resource_instance" "cos_instance" {
  count             = var.enable_flow_logs ? 1 : 0
  name              = "${local.name}-cos"
  service           = "cloud-object-storage"
  plan              = "standard"
  location          = "global"
  resource_group_id = data.ibm_resource_group.resource_group.id
  tags              = local.tags
}

# COS bucket for flow logs
resource "ibm_cos_bucket" "flow_logs" {
  count                = var.enable_flow_logs ? 1 : 0
  bucket_name          = "${replace(local.name, "-", "")}-flow-logs"
  resource_instance_id = ibm_resource_instance.cos_instance[0].id
  region_location      = var.region
  storage_class        = "smart"

  expire_rule {
    rule_id = "delete-old-logs"
    enable  = true
    days    = var.flow_logs_retention_days
  }
}

# IBM Key Protect instance for secrets encryption
resource "ibm_resource_instance" "key_protect" {
  count             = var.enable_key_protect ? 1 : 0
  name              = "${local.name}-kms"
  service           = "kms"
  plan              = "tiered-pricing"
  location          = var.region
  resource_group_id = data.ibm_resource_group.resource_group.id
  tags              = local.tags
}

# Root key for cluster encryption
resource "ibm_kms_key" "root_key" {
  count        = var.enable_key_protect ? 1 : 0
  instance_id  = ibm_resource_instance.key_protect[0].guid
  key_name     = "${local.name}-root-key"
  standard_key = false # Root key for encryption

  force_delete = var.key_protect_force_delete
}

# Public Gateway for each zone (optional, for internet access)
resource "ibm_is_public_gateway" "pgw" {
  count          = var.enable_public_gateway ? var.zones_count : 0
  name           = "${local.name}-pgw-${count.index + 1}"
  vpc            = ibm_is_vpc.iks_vpc.id
  zone           = data.ibm_is_zones.zones.zones[count.index]
  resource_group = data.ibm_resource_group.resource_group.id
  tags           = local.tags
}

# Security group for cluster nodes
resource "ibm_is_security_group" "cluster_nodes" {
  name           = "${local.name}-nodes-sg"
  vpc            = ibm_is_vpc.iks_vpc.id
  resource_group = data.ibm_resource_group.resource_group.id
  tags           = local.tags
}

# Allow all traffic within VPC
resource "ibm_is_security_group_rule" "inbound_vpc" {
  group     = ibm_is_security_group.cluster_nodes.id
  direction = "inbound"
  remote    = var.vpc_cidr
}

# Allow SSH from specific CIDRs (if provided)
resource "ibm_is_security_group_rule" "inbound_ssh" {
  count     = length(var.allowed_ssh_cidrs) > 0 ? length(var.allowed_ssh_cidrs) : 0
  group     = ibm_is_security_group.cluster_nodes.id
  direction = "inbound"
  remote    = var.allowed_ssh_cidrs[count.index]

  tcp {
    port_min = 22
    port_max = 22
  }
}

# Allow HTTPS traffic
resource "ibm_is_security_group_rule" "inbound_https" {
  group     = ibm_is_security_group.cluster_nodes.id
  direction = "inbound"
  remote    = "0.0.0.0/0"

  tcp {
    port_min = 443
    port_max = 443
  }
}

# Allow HTTP traffic
resource "ibm_is_security_group_rule" "inbound_http" {
  group     = ibm_is_security_group.cluster_nodes.id
  direction = "inbound"
  remote    = "0.0.0.0/0"

  tcp {
    port_min = 80
    port_max = 80
  }
}

# Allow all outbound traffic
resource "ibm_is_security_group_rule" "outbound_all" {
  group     = ibm_is_security_group.cluster_nodes.id
  direction = "outbound"
  remote    = "0.0.0.0/0"
}

# Create subnet for each zone
resource "ibm_is_subnet" "iks_subnet" {
  count           = var.zones_count
  name            = "${local.name}-subnet-${count.index + 1}"
  vpc             = ibm_is_vpc.iks_vpc.id
  zone            = data.ibm_is_zones.zones.zones[count.index]
  ipv4_cidr_block = cidrsubnet(var.vpc_cidr, 4, count.index)
  public_gateway  = var.enable_public_gateway ? ibm_is_public_gateway.pgw[count.index].id : null
  resource_group  = data.ibm_resource_group.resource_group.id
  tags            = local.tags
}

# Create IKS cluster with enhanced security
resource "ibm_container_vpc_cluster" "iks_cluster" {
  name              = local.name
  vpc_id            = ibm_is_vpc.iks_vpc.id
  flavor            = var.system_worker_flavor
  worker_count      = var.system_workers_per_zone
  kube_version      = var.kubernetes_version
  resource_group_id = data.ibm_resource_group.resource_group.id
  wait_till         = "MasterNodeReady"

  dynamic "zones" {
    for_each = range(var.zones_count)
    content {
      subnet_id = ibm_is_subnet.iks_subnet[zones.value].id
      name      = data.ibm_is_zones.zones.zones[zones.value]
    }
  }

  # Private service endpoint by default
  disable_public_service_endpoint = var.disable_public_endpoint
  entitlement                     = var.entitlement

  # Encryption with Key Protect
  dynamic "kms_config" {
    for_each = var.enable_key_protect ? [1] : []
    content {
      instance_id      = ibm_resource_instance.key_protect[0].guid
      crk_id           = ibm_kms_key.root_key[0].key_id
      private_endpoint = var.key_protect_private_endpoint
    }
  }

  tags = local.tags

  # Set default worker pool labels
  labels = {
    "node.kubernetes.io/role" = "system"
    "workload-type"           = "platform"
  }

  # Taints for system pool (platform components only)
  dynamic "taints" {
    for_each = var.taint_system_pool ? [1] : []
    content {
      key    = "CriticalAddonsOnly"
      value  = "true"
      effect = "NoSchedule"
    }
  }
}

# Create workload worker pool (for application workloads)
resource "ibm_container_vpc_worker_pool" "workload_pool" {
  cluster           = ibm_container_vpc_cluster.iks_cluster.id
  worker_pool_name  = "${local.name}-workload-pool"
  flavor            = var.workload_worker_flavor
  vpc_id            = ibm_is_vpc.iks_vpc.id
  worker_count      = var.workload_workers_per_zone
  resource_group_id = data.ibm_resource_group.resource_group.id

  dynamic "zones" {
    for_each = range(var.zones_count)
    content {
      subnet_id = ibm_is_subnet.iks_subnet[zones.value].id
      name      = data.ibm_is_zones.zones.zones[zones.value]
    }
  }

  labels = {
    "node.kubernetes.io/role" = "workload"
    "workload-type"           = "application"
  }
}

# Optional: Create spot instances pool for cost savings
resource "ibm_container_vpc_worker_pool" "spot_pool" {
  count             = var.enable_spot_pool ? 1 : 0
  cluster           = ibm_container_vpc_cluster.iks_cluster.id
  worker_pool_name  = "${local.name}-spot-pool"
  flavor            = var.spot_worker_flavor
  vpc_id            = ibm_is_vpc.iks_vpc.id
  worker_count      = var.spot_workers_per_zone
  resource_group_id = data.ibm_resource_group.resource_group.id

  dynamic "zones" {
    for_each = range(var.zones_count)
    content {
      subnet_id = ibm_is_subnet.iks_subnet[zones.value].id
      name      = data.ibm_is_zones.zones.zones[zones.value]
    }
  }

  labels = {
    "node.kubernetes.io/role" = "workload"
    "workload-type"           = "application"
    "spot-instance"           = "true"
    "kubernetes.io/lifecycle" = "spot"
  }

  taints {
    key    = "spot-instance"
    value  = "true"
    effect = "NoSchedule"
  }
}

# COS bucket for backups (S3-compatible)
resource "ibm_resource_instance" "backup_cos" {
  count             = var.enable_backup_bucket ? 1 : 0
  name              = "${local.name}-backup-cos"
  service           = "cloud-object-storage"
  plan              = "standard"
  location          = "global"
  resource_group_id = data.ibm_resource_group.resource_group.id
  tags              = local.tags
}

resource "ibm_cos_bucket" "backups" {
  count                = var.enable_backup_bucket ? 1 : 0
  bucket_name          = "${replace(local.name, "-", "")}-backups"
  resource_instance_id = ibm_resource_instance.backup_cos[0].id
  region_location      = var.region
  storage_class        = "smart"

  expire_rule {
    rule_id = "delete-old-backups"
    enable  = true
    days    = var.backup_retention_days
  }

  # Enable versioning for backup protection
  object_versioning {
    enable = true
  }
}

# IBM Log Analysis instance for cluster logging
resource "ibm_resource_instance" "log_analysis" {
  count             = var.enable_log_analysis ? 1 : 0
  name              = "${local.name}-logs"
  service           = "logdna"
  plan              = var.log_analysis_plan
  location          = var.region
  resource_group_id = data.ibm_resource_group.resource_group.id
  tags              = local.tags
}

# IBM Cloud Monitoring instance
resource "ibm_resource_instance" "monitoring" {
  count             = var.enable_monitoring ? 1 : 0
  name              = "${local.name}-monitoring"
  service           = "sysdig-monitor"
  plan              = var.monitoring_plan
  location          = var.region
  resource_group_id = data.ibm_resource_group.resource_group.id
  tags              = local.tags
}

# Outputs
output "cluster_id" {
  description = "The ID of the IKS cluster"
  value       = ibm_container_vpc_cluster.iks_cluster.id
}

output "cluster_name" {
  description = "The name of the IKS cluster"
  value       = ibm_container_vpc_cluster.iks_cluster.name
}

output "ingress_hostname" {
  description = "The Ingress hostname for the cluster"
  value       = ibm_container_vpc_cluster.iks_cluster.ingress_hostname
}

output "ingress_secret" {
  description = "The Ingress secret name"
  value       = ibm_container_vpc_cluster.iks_cluster.ingress_secret
  sensitive   = true
}

output "master_url" {
  description = "The URL of the cluster master"
  value       = ibm_container_vpc_cluster.iks_cluster.master_url
}

output "vpc_id" {
  description = "The ID of the VPC"
  value       = ibm_is_vpc.iks_vpc.id
}

output "security_group_id" {
  description = "The ID of the cluster security group"
  value       = ibm_is_security_group.cluster_nodes.id
}

output "key_protect_instance_id" {
  description = "The GUID of the Key Protect instance"
  value       = var.enable_key_protect ? ibm_resource_instance.key_protect[0].guid : null
}

output "root_key_id" {
  description = "The ID of the root encryption key"
  value       = var.enable_key_protect ? ibm_kms_key.root_key[0].key_id : null
}

output "backup_bucket" {
  description = "The name of the backup bucket"
  value       = var.enable_backup_bucket ? ibm_cos_bucket.backups[0].bucket_name : null
}

output "backup_cos_instance_id" {
  description = "The ID of the backup COS instance"
  value       = var.enable_backup_bucket ? ibm_resource_instance.backup_cos[0].id : null
}

output "log_analysis_instance_id" {
  description = "The ID of the Log Analysis instance"
  value       = var.enable_log_analysis ? ibm_resource_instance.log_analysis[0].id : null
}

output "monitoring_instance_id" {
  description = "The ID of the Monitoring instance"
  value       = var.enable_monitoring ? ibm_resource_instance.monitoring[0].id : null
}

output "flow_logs_bucket" {
  description = "The name of the VPC flow logs bucket"
  value       = var.enable_flow_logs ? ibm_cos_bucket.flow_logs[0].bucket_name : null
}
