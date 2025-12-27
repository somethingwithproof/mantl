terraform {
  required_version = ">= 1.5.0"
  required_providers {
    google = { source = "hashicorp/google", version = ">= 5.0" }
  }
}

variable "project" { type = string }
variable "gsa_name" { type = string }
variable "ksa_namespace" { type = string }
variable "ksa_name" { type = string }
variable "roles" { type = list(string) }

resource "google_service_account" "gsa" {
  account_id   = var.gsa_name
  display_name = "GSA for ${var.ksa_namespace}/${var.ksa_name}"
  project      = var.project
}

resource "google_service_account_iam_binding" "wi" {
  service_account_id = google_service_account.gsa.name
  role               = "roles/iam.workloadIdentityUser"
  members            = [
    "serviceAccount:${var.project}.svc.id.goog[${var.ksa_namespace}/${var.ksa_name}]"
  ]
}

resource "google_project_iam_member" "project_roles" {
  for_each = toset(var.roles)
  project  = var.project
  role     = each.value
  member   = "serviceAccount:${google_service_account.gsa.email}"
}

output "gsa_email" { value = google_service_account.gsa.email }
