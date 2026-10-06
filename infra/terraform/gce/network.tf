# SPDX-License-Identifier: Apache-2.0
# Custom-mode VPC network, subnetwork, and firewalls for the modern GCE schema.
# use_modern_gce_network is declared in main.tf alongside the instance schema.

variable "allowed_cidrs" {
  description = "Deprecated compatibility input. External source allowlists are unsupported; must remain empty."
  type        = list(string)
  default     = []

  validation {
    condition     = length(var.allowed_cidrs) == 0
    error_message = "External ingress is unsupported. Remove allowed_cidrs and use private connectivity or reviewed IAP SSH access."
  }
}

variable "enable_iap_ssh" {
  description = "Allow SSH from Google IAP to this module's tagged VMs. Requires separately reviewed IAP IAM and the managed network."
  type        = bool
  default     = false

  validation {
    condition     = !var.enable_iap_ssh || var.use_modern_gce_network
    error_message = "IAP SSH requires the managed network; externally supplied subnets need independent firewall review."
  }
}

# Custom-mode network with managed subnetwork
resource "google_compute_network" "mi_network_modern" {
  count                   = var.use_modern_gce_network ? 1 : 0
  name                    = var.long_name
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "mi_subnet_modern" {
  count         = var.use_modern_gce_network ? 1 : 0
  name          = "${var.short_name}-subnet"
  region        = var.region
  ip_cidr_range = var.network_ipv4
  network       = google_compute_network.mi_network_modern[0].self_link

  log_config {
    aggregation_interval = "INTERVAL_5_SEC"
    flow_sampling        = 1.0
    metadata             = "INCLUDE_ALL_METADATA"
  }
}

# This source range is Google's IAP TCP forwarding service, not public clients.
# Tunnel authorization and SSH authentication are separate deployment prerequisites.
resource "google_compute_firewall" "iap_ssh" {
  count         = var.enable_iap_ssh ? 1 : 0
  name          = "${var.short_name}-iap-ssh"
  network       = google_compute_network.mi_network_modern[0].name
  direction     = "INGRESS"
  source_ranges = ["35.235.240.0/20"]
  target_tags   = [var.short_name]

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  log_config {
    metadata = "INCLUDE_ALL_METADATA"
  }
}

resource "google_compute_firewall" "internal_modern" {
  count         = var.use_modern_gce_network ? 1 : 0
  name          = "${var.short_name}-firewall-internal-modern"
  network       = google_compute_network.mi_network_modern[0].name
  source_ranges = [var.network_ipv4]

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["1-65535"]
  }

  allow {
    protocol = "udp"
    ports    = ["1-65535"]
  }
}

output "modern_subnetwork_self_link" {
  value = var.use_modern_gce_network ? google_compute_subnetwork.mi_subnet_modern[0].self_link : null

  # Preserve a fail-closed compatibility contract for callers with legacy inputs.
  precondition {
    condition     = !var.gce_public_ip && length(var.allowed_cidrs) == 0
    error_message = "This subnet contract supports private VM interfaces only; legacy public-access inputs must be disabled."
  }
}
