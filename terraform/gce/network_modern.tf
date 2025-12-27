variable "use_modern_gce_network" {
  description = "If true, create a custom-mode VPC network + subnetwork and modern firewalls."
  default     = false
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
}

resource "google_compute_firewall" "external_modern" {
  count         = var.use_modern_gce_network ? 1 : 0
  name          = "${var.short_name}-firewall-external-modern"
  network       = google_compute_network.mi_network_modern[0].name
  source_ranges = var.allowed_cidrs

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["22", "80", "443", "4400", "5050", "8080", "8500"]
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
}
