# SPDX-License-Identifier: Apache-2.0
# Mantl GCE instances using the modern google_compute_instance schema:
# shielded VMs, optional CMEK boot disks, and private-only network interfaces.

variable "use_modern_gce_schema" {
  description = "If true, create control nodes with modern google_compute_instance schema."
  type        = bool
  default     = false
}

variable "gce_public_ip" {
  description = "Deprecated compatibility input. Public IPs are unsupported; only false is accepted."
  type        = bool
  default     = false

  validation {
    condition     = !var.gce_public_ip
    error_message = "Public IPs are unsupported. Use private connectivity or reviewed IAP SSH access."
  }
}

variable "gce_boot_image_family" {
  description = "Boot image family for control instances."
  default     = "debian-12"
}

variable "gce_boot_image_project" {
  description = "GCP project hosting the image family."
  default     = "debian-cloud"
}

variable "gce_boot_kms_key" {
  description = "Optional self_link of the KMS crypto key for boot disk encryption."
  default     = ""
}

variable "modern_subnetwork_self_link" {
  description = "Optional self_link of a modern subnetwork to attach (use with use_modern_gce_network)."
  default     = ""

  validation {
    condition     = !var.use_modern_gce_schema || var.use_modern_gce_network || var.modern_subnetwork_self_link != ""
    error_message = "Modern instances require either a managed modern network or an explicit existing subnetwork."
  }
}

variable "use_modern_gce_network" {
  description = "If true, expect a modern subnetwork; set modern_subnetwork_self_link accordingly."
  type        = bool
  default     = false
}

# Resolve the boot image from family
data "google_compute_image" "boot" {
  family  = var.gce_boot_image_family
  project = var.gce_boot_image_project
}

# Modern control instances using shielded VMs and secure defaults
resource "google_compute_instance" "control_modern" {
  count        = var.use_modern_gce_schema ? var.control_count : 0
  name         = "${var.short_name}-control-${format("%02d", count.index + 1)}"
  description  = "${var.long_name} control node (modern) #${format("%02d", count.index + 1)}"
  machine_type = var.control_type
  zone         = var.zone
  tags         = [var.short_name, "control"]

  # Boot disk
  # Boot like you mean it: pinned image family + optional CMEK.
  boot_disk {
    initialize_params {
      image = data.google_compute_image.boot.self_link
      size  = var.control_volume_size
      type  = "pd-balanced"
    }
    kms_key_self_link = var.gce_boot_kms_key != "" ? var.gce_boot_kms_key : null
  }

  # Attached LVM/data disk
  # Because someone always needs more IOPS later: attach the LVM/data disk.
  attached_disk {
    source      = google_compute_disk.mi-control-lvm[count.index].self_link
    device_name = "lvm"
    mode        = "READ_WRITE"
  }

  # Network
  # Private interfaces only; no external IP attachment is supported.
  network_interface {
    subnetwork = var.use_modern_gce_network ? google_compute_subnetwork.mi_subnet_modern[0].self_link : var.modern_subnetwork_self_link
  }

  # Shielded VM
  # Shielded VMs: future you will thank you; attacker you will curse you.
  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  # Metadata: block project SSH keys, inject only provided SSH key
  metadata = {
    "block-project-ssh-keys" = "TRUE"
    "ssh-keys"               = "${var.ssh_user}:${file(var.ssh_key)} ${var.ssh_user}"
    "dc"                     = var.datacenter
    "role"                   = "control"
    "ssh_user"               = var.ssh_user
  }
}

# Modern worker instances (opt-in). When enabling, set worker_count = 0 on legacy path to avoid duplicates.
resource "google_compute_instance" "worker_modern" {
  count        = var.use_modern_gce_schema ? var.worker_count : 0
  name         = "${var.short_name}-worker-${format("%03d", count.index + 1)}"
  description  = "${var.long_name} worker node (modern) #${format("%03d", count.index + 1)}"
  machine_type = var.worker_type
  zone         = var.zone
  tags         = [var.short_name, "worker"]

  boot_disk {
    initialize_params {
      image = data.google_compute_image.boot.self_link
      size  = var.worker_volume_size
      type  = "pd-balanced"
    }
    kms_key_self_link = var.gce_boot_kms_key != "" ? var.gce_boot_kms_key : null
  }

  attached_disk {
    source      = google_compute_disk.mi-worker-lvm[count.index].self_link
    device_name = "lvm"
    mode        = "READ_WRITE"
  }

  network_interface {
    subnetwork = var.use_modern_gce_network ? google_compute_subnetwork.mi_subnet_modern[0].self_link : var.modern_subnetwork_self_link
  }

  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  metadata = {
    "block-project-ssh-keys" = "TRUE"
    "ssh-keys"               = "${var.ssh_user}:${file(var.ssh_key)} ${var.ssh_user}"
    "dc"                     = var.datacenter
    "role"                   = "worker"
    "ssh_user"               = var.ssh_user
  }
}

# Modern kubeworker instances (opt-in). When enabling, set kubeworker_count = 0 on legacy path to avoid duplicates.
resource "google_compute_instance" "kubeworker_modern" {
  count        = var.use_modern_gce_schema ? var.kubeworker_count : 0
  name         = "${var.short_name}-kubeworker-${format("%03d", count.index + 1)}"
  description  = "${var.long_name} kube worker node (modern) #${format("%03d", count.index + 1)}"
  machine_type = var.worker_type
  zone         = var.zone
  tags         = [var.short_name, "kubeworker"]

  boot_disk {
    initialize_params {
      image = data.google_compute_image.boot.self_link
      size  = var.worker_volume_size
      type  = "pd-balanced"
    }
    kms_key_self_link = var.gce_boot_kms_key != "" ? var.gce_boot_kms_key : null
  }

  attached_disk {
    source      = google_compute_disk.mi-kubeworker-lvm[count.index].self_link
    device_name = "lvm"
    mode        = "READ_WRITE"
  }

  network_interface {
    subnetwork = var.use_modern_gce_network ? google_compute_subnetwork.mi_subnet_modern[0].self_link : var.modern_subnetwork_self_link
  }

  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  metadata = {
    "block-project-ssh-keys" = "TRUE"
    "ssh-keys"               = "${var.ssh_user}:${file(var.ssh_key)} ${var.ssh_user}"
    "dc"                     = var.datacenter
    "role"                   = "kubeworker"
    "ssh_user"               = var.ssh_user
  }
}
