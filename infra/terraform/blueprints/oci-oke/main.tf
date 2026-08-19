provider "oci" {
  tenancy_ocid     = var.tenancy_ocid
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
  region           = var.region
}

locals {
  name = var.cluster_name

  common_tags = merge(var.tags, {
    managed_by  = "terraform"
    platform    = "mantl"
    environment = var.environment
    blueprint   = "oci"
  })
}

# Get availability domains
data "oci_identity_availability_domains" "ads" {
  compartment_id = var.compartment_id
}

# Get latest Oracle Linux image for nodes
data "oci_core_images" "node_image" {
  compartment_id           = var.compartment_id
  operating_system         = "Oracle Linux"
  operating_system_version = "8"
  shape                    = var.system_node_shape

  filter {
    name   = "display_name"
    values = ["^.*Oracle-Linux-8.*-\\d{4}\\.\\d{2}\\.\\d{2}-\\d+$"]
    regex  = true
  }
}

# VCN for OKE cluster
resource "oci_core_vcn" "oke_vcn" {
  compartment_id = var.compartment_id
  display_name   = "${local.name}-vcn"
  cidr_blocks    = [var.vcn_cidr]
  dns_label      = replace(local.name, "-", "")

  freeform_tags = local.common_tags
}

# Internet Gateway
resource "oci_core_internet_gateway" "oke_igw" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${local.name}-igw"
  enabled        = true

  freeform_tags = local.common_tags
}

# NAT Gateway for private subnets
resource "oci_core_nat_gateway" "oke_nat" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${local.name}-nat"

  freeform_tags = local.common_tags
}

# Service Gateway for private access to Oracle services
resource "oci_core_service_gateway" "oke_sg" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${local.name}-sg"

  services {
    service_id = data.oci_core_services.all_services.services[0].id
  }

  freeform_tags = local.common_tags
}

data "oci_core_services" "all_services" {
}

# VCN Flow Logs for security monitoring
resource "oci_logging_log_group" "vcn_flow_logs" {
  compartment_id = var.compartment_id
  display_name   = "${local.name}-vcn-flow-logs"

  freeform_tags = local.common_tags
}

resource "oci_logging_log" "vcn_flow_log" {
  display_name = "${local.name}-vcn-flow"
  log_group_id = oci_logging_log_group.vcn_flow_logs.id
  log_type     = "SERVICE"

  configuration {
    source {
      category    = "all"
      resource    = oci_core_vcn.oke_vcn.id
      service     = "flowlogs"
      source_type = "OCISERVICE"
    }

    compartment_id = var.compartment_id
  }

  is_enabled         = var.enable_flow_logs
  retention_duration = var.flow_logs_retention_days

  freeform_tags = local.common_tags
}

# Route tables
resource "oci_core_route_table" "public_route_table" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${local.name}-public-rt"

  route_rules {
    network_entity_id = oci_core_internet_gateway.oke_igw.id
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
  }

  freeform_tags = local.common_tags
}

resource "oci_core_route_table" "private_route_table" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${local.name}-private-rt"

  route_rules {
    network_entity_id = oci_core_nat_gateway.oke_nat.id
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
  }

  route_rules {
    network_entity_id = oci_core_service_gateway.oke_sg.id
    destination       = data.oci_core_services.all_services.services[0].cidr_block
    destination_type  = "SERVICE_CIDR_BLOCK"
  }

  freeform_tags = local.common_tags
}

# Security lists with least privilege
resource "oci_core_security_list" "api_endpoint_seclist" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${local.name}-api-seclist"

  egress_security_rules {
    destination = "0.0.0.0/0"
    protocol    = "all"
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = var.enable_private_endpoint ? var.vcn_cidr : "0.0.0.0/0"
    tcp_options {
      min = 6443
      max = 6443
    }
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = var.vcn_cidr
    tcp_options {
      min = 12250
      max = 12250
    }
  }

  freeform_tags = local.common_tags
}

resource "oci_core_security_list" "node_seclist" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${local.name}-node-seclist"

  egress_security_rules {
    destination = "0.0.0.0/0"
    protocol    = "all"
  }

  # Allow worker nodes to communicate with control plane
  ingress_security_rules {
    protocol = "all"
    source   = var.vcn_cidr
  }

  # SSH from bastion only (if specified)
  dynamic "ingress_security_rules" {
    for_each = length(var.allowed_ssh_cidrs) > 0 ? var.allowed_ssh_cidrs : []
    content {
      protocol = "6" # TCP
      source   = ingress_security_rules.value
      tcp_options {
        min = 22
        max = 22
      }
    }
  }

  freeform_tags = local.common_tags
}

resource "oci_core_security_list" "lb_seclist" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.oke_vcn.id
  display_name   = "${local.name}-lb-seclist"

  egress_security_rules {
    destination = "0.0.0.0/0"
    protocol    = "all"
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    tcp_options {
      min = 443
      max = 443
    }
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    tcp_options {
      min = 80
      max = 80
    }
  }

  freeform_tags = local.common_tags
}

# Subnets
resource "oci_core_subnet" "api_endpoint_subnet" {
  compartment_id    = var.compartment_id
  vcn_id            = oci_core_vcn.oke_vcn.id
  cidr_block        = cidrsubnet(var.vcn_cidr, 8, 0)
  display_name      = "${local.name}-api-subnet"
  route_table_id    = var.enable_private_endpoint ? oci_core_route_table.private_route_table.id : oci_core_route_table.public_route_table.id
  security_list_ids = [oci_core_security_list.api_endpoint_seclist.id]
  dns_label         = "api"

  prohibit_public_ip_on_vnic = var.enable_private_endpoint

  freeform_tags = local.common_tags
}

resource "oci_core_subnet" "node_subnet" {
  compartment_id    = var.compartment_id
  vcn_id            = oci_core_vcn.oke_vcn.id
  cidr_block        = cidrsubnet(var.vcn_cidr, 4, 1)
  display_name      = "${local.name}-node-subnet"
  route_table_id    = oci_core_route_table.private_route_table.id
  security_list_ids = [oci_core_security_list.node_seclist.id]
  dns_label         = "nodes"

  prohibit_public_ip_on_vnic = true # Always private for nodes

  freeform_tags = local.common_tags
}

resource "oci_core_subnet" "lb_subnet" {
  compartment_id    = var.compartment_id
  vcn_id            = oci_core_vcn.oke_vcn.id
  cidr_block        = cidrsubnet(var.vcn_cidr, 8, 2)
  display_name      = "${local.name}-lb-subnet"
  route_table_id    = oci_core_route_table.public_route_table.id
  security_list_ids = [oci_core_security_list.lb_seclist.id]
  dns_label         = "loadbalancers"

  freeform_tags = local.common_tags
}

# OCI Vault for secrets encryption
resource "oci_kms_vault" "oke_vault" {
  count          = var.enable_vault ? 1 : 0
  compartment_id = var.compartment_id
  display_name   = "${local.name}-vault"
  vault_type     = "DEFAULT"

  freeform_tags = local.common_tags
}

resource "oci_kms_key" "oke_encryption_key" {
  count               = var.enable_vault ? 1 : 0
  compartment_id      = var.compartment_id
  display_name        = "${local.name}-encryption-key"
  management_endpoint = oci_kms_vault.oke_vault[0].management_endpoint

  key_shape {
    algorithm = "AES"
    length    = 32 # 256-bit
  }

  freeform_tags = local.common_tags
}

# OKE Cluster
resource "oci_containerengine_cluster" "oke_cluster" {
  compartment_id     = var.compartment_id
  kubernetes_version = var.kubernetes_version
  name               = local.name
  vcn_id             = oci_core_vcn.oke_vcn.id

  endpoint_config {
    is_public_ip_enabled = false
    subnet_id            = oci_core_subnet.api_endpoint_subnet.id
  }

  options {
    service_lb_subnet_ids = [oci_core_subnet.lb_subnet.id]

    add_ons {
      is_kubernetes_dashboard_enabled = false # Security best practice
      is_tiller_enabled               = false # Deprecated
    }

    kubernetes_network_config {
      pods_cidr     = var.pod_cidr
      services_cidr = var.service_cidr
    }

    persistent_volume_config {
      freeform_tags = local.common_tags
    }

    service_lb_config {
      freeform_tags = local.common_tags
    }
  }

  # Enable OCI Logging for cluster
  cluster_pod_network_options {
    cni_type = var.use_flannel_cni ? "FLANNEL_OVERLAY" : "OCI_VCN_IP_NATIVE"
  }

  dynamic "image_policy_config" {
    for_each = var.enable_image_policy ? [1] : []
    content {
      is_policy_enabled = true
      key_details {
        kms_key_id = var.enable_vault ? oci_kms_key.oke_encryption_key[0].id : null
      }
    }
  }

  freeform_tags = local.common_tags

  lifecycle {
    precondition {
      condition     = !var.enable_image_policy || var.enable_vault
      error_message = "enable_image_policy requires enable_vault so a KMS key is available."
    }
  }
}

# System Node Pool (for platform components)
resource "oci_containerengine_node_pool" "system_pool" {
  cluster_id         = oci_containerengine_cluster.oke_cluster.id
  compartment_id     = var.compartment_id
  kubernetes_version = var.kubernetes_version
  name               = "${local.name}-system-pool"

  node_config_details {
    dynamic "placement_configs" {
      for_each = slice(data.oci_identity_availability_domains.ads.availability_domains, 0, min(3, length(data.oci_identity_availability_domains.ads.availability_domains)))
      content {
        availability_domain = placement_configs.value.name
        subnet_id           = oci_core_subnet.node_subnet.id
      }
    }
    size = var.system_node_count

    dynamic "node_pool_pod_network_option_details" {
      for_each = var.use_flannel_cni ? [] : [1]
      content {
        cni_type       = "OCI_VCN_IP_NATIVE"
        pod_subnet_ids = [oci_core_subnet.node_subnet.id]
      }
    }

    freeform_tags = merge(local.common_tags, {
      "NodePool"     = "system"
      "WorkloadType" = "platform"
    })
  }

  node_shape = var.system_node_shape

  dynamic "node_shape_config" {
    for_each = can(regex("Flex$", var.system_node_shape)) ? [1] : []
    content {
      memory_in_gbs = var.system_node_memory_gb
      ocpus         = var.system_node_ocpus
    }
  }

  node_source_details {
    image_id    = coalesce(var.node_image_id, data.oci_core_images.node_image.images[0].id)
    source_type = "IMAGE"
  }

  initial_node_labels {
    key   = "node.kubernetes.io/role"
    value = "system"
  }

  initial_node_labels {
    key   = "workload-type"
    value = "platform"
  }

  freeform_tags = merge(local.common_tags, {
    "NodePool" = "system"
  })
}

# Workload Node Pool (for application workloads)
resource "oci_containerengine_node_pool" "workload_pool" {
  cluster_id         = oci_containerengine_cluster.oke_cluster.id
  compartment_id     = var.compartment_id
  kubernetes_version = var.kubernetes_version
  name               = "${local.name}-workload-pool"

  node_config_details {
    dynamic "placement_configs" {
      for_each = slice(data.oci_identity_availability_domains.ads.availability_domains, 0, min(3, length(data.oci_identity_availability_domains.ads.availability_domains)))
      content {
        availability_domain = placement_configs.value.name
        subnet_id           = oci_core_subnet.node_subnet.id
      }
    }
    size = var.workload_initial_node_count

    dynamic "node_pool_pod_network_option_details" {
      for_each = var.use_flannel_cni ? [] : [1]
      content {
        cni_type       = "OCI_VCN_IP_NATIVE"
        pod_subnet_ids = [oci_core_subnet.node_subnet.id]
      }
    }

    freeform_tags = merge(local.common_tags, {
      "NodePool"     = "workload"
      "WorkloadType" = "application"
    })
  }

  node_shape = var.workload_node_shape

  dynamic "node_shape_config" {
    for_each = can(regex("Flex$", var.workload_node_shape)) ? [1] : []
    content {
      memory_in_gbs = var.workload_node_memory_gb
      ocpus         = var.workload_node_ocpus
    }
  }

  node_source_details {
    image_id    = coalesce(var.node_image_id, data.oci_core_images.node_image.images[0].id)
    source_type = "IMAGE"
  }

  initial_node_labels {
    key   = "node.kubernetes.io/role"
    value = "workload"
  }

  initial_node_labels {
    key   = "workload-type"
    value = "application"
  }

  freeform_tags = merge(local.common_tags, {
    "NodePool" = "workload"
  })
}

# Object Storage bucket for backups
resource "oci_objectstorage_bucket" "backups" {
  count          = var.enable_backup_bucket ? 1 : 0
  compartment_id = var.compartment_id
  name           = "${local.name}-backups"
  namespace      = data.oci_objectstorage_namespace.ns.namespace

  access_type           = "NoPublicAccess"
  storage_tier          = "Standard"
  versioning            = "Enabled"
  object_events_enabled = var.enable_object_events

  retention_rules {
    display_name = "delete-old-backups"
    duration {
      time_amount = var.backup_retention_days
      time_unit   = "DAYS"
    }
    time_rule_locked = var.backup_retention_lock_time
  }

  freeform_tags = local.common_tags
}

data "oci_objectstorage_namespace" "ns" {
  compartment_id = var.compartment_id
}

# OCI Logging for cluster audit logs
resource "oci_logging_log" "cluster_audit_log" {
  count        = var.enable_audit_logs ? 1 : 0
  display_name = "${local.name}-audit-log"
  log_group_id = oci_logging_log_group.vcn_flow_logs.id
  log_type     = "SERVICE"

  configuration {
    source {
      category    = "all"
      resource    = oci_containerengine_cluster.oke_cluster.id
      service     = "oke"
      source_type = "OCISERVICE"
    }

    compartment_id = var.compartment_id
  }

  is_enabled         = true
  retention_duration = var.audit_log_retention_days

  freeform_tags = local.common_tags
}

# Outputs
output "cluster_id" {
  description = "The ID of the OKE cluster"
  value       = oci_containerengine_cluster.oke_cluster.id
}

output "cluster_name" {
  description = "The name of the OKE cluster"
  value       = oci_containerengine_cluster.oke_cluster.name
}

output "kubernetes_version" {
  description = "The Kubernetes version of the cluster"
  value       = oci_containerengine_cluster.oke_cluster.kubernetes_version
}

output "vcn_id" {
  description = "The ID of the VCN"
  value       = oci_core_vcn.oke_vcn.id
}

output "api_endpoint" {
  description = "The endpoint of the Kubernetes API server"
  value       = oci_containerengine_cluster.oke_cluster.endpoints[0].kubernetes
}

output "private_endpoint" {
  description = "The private endpoint of the Kubernetes API server"
  value       = var.enable_private_endpoint ? oci_containerengine_cluster.oke_cluster.endpoints[0].private_endpoint : null
}

output "vault_id" {
  description = "The ID of the OCI Vault"
  value       = var.enable_vault ? oci_kms_vault.oke_vault[0].id : null
}

output "encryption_key_id" {
  description = "The ID of the encryption key"
  value       = var.enable_vault ? oci_kms_key.oke_encryption_key[0].id : null
}

output "backup_bucket" {
  description = "The name of the backup bucket"
  value       = var.enable_backup_bucket ? oci_objectstorage_bucket.backups[0].name : null
}

output "backup_bucket_namespace" {
  description = "The namespace of the backup bucket"
  value       = var.enable_backup_bucket ? data.oci_objectstorage_namespace.ns.namespace : null
}

output "flow_log_group_id" {
  description = "The ID of the VCN flow logs log group"
  value       = var.enable_flow_logs ? oci_logging_log_group.vcn_flow_logs.id : null
}
