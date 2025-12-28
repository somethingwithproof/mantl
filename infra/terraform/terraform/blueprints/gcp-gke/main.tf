terraform {
  required_version = ">= 1.10.0"
  required_providers {
    google = { source = "hashicorp/google", version = ">= 7.0" }
  }
}

provider "google" {
  project = var.project
  region  = var.region
}

module "gke" {
  source            = "terraform-google-modules/kubernetes-engine/google//modules/beta-private-cluster"
  version           = "~> 42.0"
  project_id        = var.project
  name              = var.cluster_name
  region            = var.region
  network           = var.network
  subnetwork        = var.subnetwork
  ip_range_pods     = var.ip_range_pods
  ip_range_services = var.ip_range_services
  release_channel   = "REGULAR"
  enable_private_endpoint = false
  enable_private_nodes    = true
  remove_default_node_pool = true
  node_pools = [{
    name         = "default"
    machine_type = "e2-standard-4"
    min_count    = 1
    max_count    = 4
    local_ssd_count = 0
    spot         = false
    disk_size_gb = 100
    disk_type    = "pd-balanced"
    image_type   = "COS_CONTAINERD"
  }]
}

output "cluster_name" { value = module.gke.name }
output "endpoint" { value = module.gke.endpoint }
output "ca_certificate" { value = module.gke.ca_certificate }
