terraform {
  required_version = ">= 1.10.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 7.0"
    }
  }
}

provider "google" {
  project = var.project
  region  = var.region
}

resource "google_storage_bucket" "spinnaker_front50" {
  name          = var.bucket_name
  location      = var.location
  force_destroy = var.force_destroy
  uniform_bucket_level_access = true
  versioning { enabled = true }
}

output "bucket_name" {
  value = google_storage_bucket.spinnaker_front50.name
}
