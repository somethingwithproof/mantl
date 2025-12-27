terraform {
  required_version = ">= 1.5.0"
  required_providers {
    google = { source = "hashicorp/google", version = ">= 5.0" }
  }
}

provider "google" {
  project = var.project
  region  = var.region
}

module "wi" {
  source        = "../../modules/iam/gke-workload-identity"
  project       = var.project
  gsa_name      = var.gsa_name
  ksa_namespace = var.ksa_namespace
  ksa_name      = var.ksa_name
  roles         = ["roles/dns.admin"]
}

output "gsa_email" { value = module.wi.gsa_email }
