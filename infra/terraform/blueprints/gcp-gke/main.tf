# SPDX-License-Identifier: Apache-2.0
################################################################################
# Mantl GCP GKE Blueprint
# Production-ready GKE cluster with security hardening
################################################################################


locals {
  name   = var.cluster_name
  region = var.region

  common_tags = merge(var.tags, {
    managed_by  = "terraform"
    platform    = "mantl"
    environment = var.environment
    blueprint   = "gcp"
  })
}

provider "google" {
  project = var.project
  region  = local.region
}

################################################################################
# VPC Flow Logs
################################################################################

resource "google_compute_subnetwork" "gke_subnet" {
  name          = "${local.name}-gke-subnet"
  project       = var.project
  region        = local.region
  network       = var.network
  ip_cidr_range = var.subnet_cidr

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = var.ip_range_pods
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = var.ip_range_services
  }

  # Enable VPC Flow Logs for security monitoring
  log_config {
    aggregation_interval = "INTERVAL_5_SEC"
    flow_sampling        = 0.5
    metadata             = "INCLUDE_ALL_METADATA"
  }

  private_ip_google_access = true
}

################################################################################
# GKE Cluster
################################################################################

module "gke" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/beta-private-cluster"
  version = "~> 45.0"

  project_id = var.project
  name       = local.name
  region     = local.region

  # Network configuration
  network           = var.network
  subnetwork        = google_compute_subnetwork.gke_subnet.name
  ip_range_pods     = "pods"
  ip_range_services = "services"

  # Kubernetes version - latest stable
  kubernetes_version       = var.kubernetes_version
  release_channel          = "REGULAR"
  remove_default_node_pool = true

  # Security: Private cluster configuration
  enable_private_endpoint = var.enable_private_endpoint # Default: true (production)
  enable_private_nodes    = true
  master_ipv4_cidr_block  = "172.16.0.0/28"

  # Master authorized networks (restrict API access)
  master_authorized_networks = var.master_authorized_networks

  # Security: Application-layer secrets encryption
  database_encryption = [{
    state    = "ENCRYPTED"
    key_name = google_kms_crypto_key.gke.id
  }]

  # Security: Enable Workload Identity (set via identity namespace)
  identity_namespace = "${var.project}.svc.id.goog"

  # Security: Enable Shielded Nodes
  enable_shielded_nodes = true

  # Security: Binary Authorization (optional)
  enable_binary_authorization = var.enable_binary_authorization

  # Logging and monitoring
  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"

  # Cluster add-ons
  network_policy             = true
  horizontal_pod_autoscaling = true
  http_load_balancing        = true
  gke_backup_agent_config    = true

  # Security: Enable GKE Security Posture
  security_posture_mode               = "BASIC"
  security_posture_vulnerability_mode = "VULNERABILITY_ENTERPRISE"

  # Node pools
  node_pools = [
    {
      name               = "system"
      machine_type       = var.system_node_machine_type
      min_count          = var.system_node_min_count
      max_count          = var.system_node_max_count
      local_ssd_count    = 0
      spot               = false
      disk_size_gb       = 100
      disk_type          = "pd-balanced"
      image_type         = "COS_CONTAINERD"
      enable_gcfs        = false
      enable_gvnic       = true
      auto_repair        = true
      auto_upgrade       = true
      preemptible        = false
      initial_node_count = var.system_node_min_count
    },
    {
      name               = "workload"
      machine_type       = var.workload_node_machine_type
      min_count          = var.workload_node_min_count
      max_count          = var.workload_node_max_count
      local_ssd_count    = 0
      spot               = var.workload_node_spot
      disk_size_gb       = 100
      disk_type          = "pd-balanced"
      image_type         = "COS_CONTAINERD"
      enable_gcfs        = false
      enable_gvnic       = true
      auto_repair        = true
      auto_upgrade       = true
      preemptible        = var.workload_node_spot
      initial_node_count = var.workload_node_min_count
    }
  ]

  node_pools_oauth_scopes = {
    all = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }

  node_pools_labels = {
    all = local.common_tags
    system = {
      role = "system"
    }
    workload = {
      role = "workload"
    }
  }

  node_pools_metadata = {
    all = {}
  }

  node_pools_taints = {
    all = []
  }

  node_pools_tags = {
    all = [local.name]
  }
}

################################################################################
# KMS Encryption Key for Secrets
################################################################################

resource "google_kms_key_ring" "gke" {
  name     = "${local.name}-keyring"
  project  = var.project
  location = local.region
}

resource "google_kms_crypto_key" "gke" {
  name            = "${local.name}-key"
  key_ring        = google_kms_key_ring.gke.id
  purpose         = "ENCRYPT_DECRYPT"
  rotation_period = "7776000s"

  lifecycle {
    prevent_destroy = true
  }
}

# Grant GKE service account access to KMS key
resource "google_kms_crypto_key_iam_member" "gke" {
  crypto_key_id = google_kms_crypto_key.gke.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.project.number}@container-engine-robot.iam.gserviceaccount.com"
}

data "google_project" "project" {
  project_id = var.project
}

################################################################################
# Workload Identity for Platform Components
################################################################################

# External DNS
module "external_dns_workload_identity" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/workload-identity"
  version = "~> 45.0"

  project_id          = var.project
  name                = "${local.name}-external-dns"
  namespace           = "external-dns"
  use_existing_k8s_sa = true

  roles = ["roles/dns.admin"]
}

# cert-manager
module "cert_manager_workload_identity" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/workload-identity"
  version = "~> 45.0"

  project_id          = var.project
  name                = "${local.name}-cert-manager"
  namespace           = "cert-manager"
  use_existing_k8s_sa = true

  roles = ["roles/dns.admin"]
}

# External Secrets
module "external_secrets_workload_identity" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/workload-identity"
  version = "~> 45.0"

  project_id          = var.project
  name                = "${local.name}-external-secrets"
  namespace           = "external-secrets"
  use_existing_k8s_sa = true

  roles = ["roles/secretmanager.secretAccessor"]
}

################################################################################
# Outputs
################################################################################

output "cluster_name" {
  description = "GKE cluster name"
  value       = module.gke.name
}

output "endpoint" {
  description = "GKE cluster endpoint"
  value       = module.gke.endpoint
  sensitive   = true
}

output "ca_certificate" {
  description = "GKE cluster CA certificate"
  value       = module.gke.ca_certificate
  sensitive   = true
}

output "cluster_id" {
  description = "GKE cluster ID"
  value       = module.gke.cluster_id
}

output "region" {
  description = "GKE cluster region"
  value       = local.region
}

output "workload_identity_config" {
  description = "Workload Identity configuration"
  value = {
    external_dns_sa     = module.external_dns_workload_identity.gcp_service_account_email
    cert_manager_sa     = module.cert_manager_workload_identity.gcp_service_account_email
    external_secrets_sa = module.external_secrets_workload_identity.gcp_service_account_email
  }
}

resource "google_project_iam_audit_config" "kubernetes" {
  project = var.project
  service = "container.googleapis.com"
  audit_log_config {
    log_type = "ADMIN_READ"
  }
  audit_log_config {
    log_type = "DATA_READ"
  }
  audit_log_config {
    log_type = "DATA_WRITE"
  }
}
