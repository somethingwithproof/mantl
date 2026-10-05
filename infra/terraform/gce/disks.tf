resource "google_compute_disk" "mi-control-lvm" {
  count = var.use_modern_gce_schema ? var.control_count : 0
  name  = "${var.short_name}-control-data-${count.index + 1}"
  zone  = var.zone
  type  = "pd-balanced"
  size  = var.control_volume_size
  dynamic "disk_encryption_key" {
    for_each = var.gce_boot_kms_key != "" ? [1] : []
    content {
      kms_key_self_link = var.gce_boot_kms_key
    }
  }
}

resource "google_compute_disk" "mi-worker-lvm" {
  count = var.use_modern_gce_schema ? var.worker_count : 0
  name  = "${var.short_name}-worker-data-${count.index + 1}"
  zone  = var.zone
  type  = "pd-balanced"
  size  = var.worker_volume_size
  dynamic "disk_encryption_key" {
    for_each = var.gce_boot_kms_key != "" ? [1] : []
    content {
      kms_key_self_link = var.gce_boot_kms_key
    }
  }
}

resource "google_compute_disk" "mi-kubeworker-lvm" {
  count = var.use_modern_gce_schema ? var.kubeworker_count : 0
  name  = "${var.short_name}-kubeworker-data-${count.index + 1}"
  zone  = var.zone
  type  = "pd-balanced"
  size  = var.worker_volume_size
  dynamic "disk_encryption_key" {
    for_each = var.gce_boot_kms_key != "" ? [1] : []
    content {
      kms_key_self_link = var.gce_boot_kms_key
    }
  }
}
