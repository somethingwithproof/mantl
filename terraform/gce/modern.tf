# Modern GCE instances (feature-flagged)

variable "use_modern_gce_schema" {
  description = "If true, use modern google_compute_instance schema alongside legacy resources"
  type        = bool
  default     = false
}

# Image data (adjust to your OS family as needed)
variable "boot_image_family" {
  description = "GCE image family for boot disks"
  type        = string
  default     = "ubuntu-2204-lts"
}

variable "boot_image_project" {
  description = "GCE image project for boot disks"
  type        = string
  default     = "ubuntu-os-cloud"
}

data "google_compute_image" "boot" {
  family  = var.boot_image_family
  project = var.boot_image_project
}

# Control nodes (modern)
resource "google_compute_instance" "control_modern" {
  count        = var.use_modern_gce_schema ? var.control_count : 0
  name         = format("%s-control-%02d", var.short_name, count.index + 1)
  machine_type = var.control_type
  zone         = var.zone
  tags         = [var.short_name, "control"]

  boot_disk {
    initialize_params {
      image = data.google_compute_image.boot.self_link
      size  = var.control_volume_size
      type  = "pd-ssd"
    }
  }

  network_interface {
    network       = google_compute_network.mi-network.name
    access_config {}
  }

  metadata = {
    dc       = var.datacenter
    role     = "control"
    ssh-user = var.ssh_user
  }
}

# Worker nodes (modern)
resource "google_compute_instance" "worker_modern" {
  count        = var.use_modern_gce_schema ? var.worker_count : 0
  name         = format("%s-worker-%03d", var.short_name, count.index + 1)
  machine_type = var.worker_type
  zone         = var.zone
  tags         = [var.short_name, "worker"]

  boot_disk {
    initialize_params {
      image = data.google_compute_image.boot.self_link
      size  = var.worker_volume_size
      type  = "pd-ssd"
    }
  }

  network_interface {
    network       = google_compute_network.mi-network.name
    access_config {}
  }

  metadata = {
    dc       = var.datacenter
    role     = "worker"
    ssh-user = var.ssh_user
  }
}
