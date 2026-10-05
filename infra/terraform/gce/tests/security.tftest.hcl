mock_provider "google" {}

variables {
  long_name  = "mantl-local-test"
  short_name = "mantl-test"
  zone       = "us-central1-a"
  region     = "us-central1"
  ssh_key    = "./tests/fixture.pub"
}

run "private_network_by_default" {
  command = plan
  variables {
    use_modern_gce_schema  = true
    use_modern_gce_network = true
    control_count          = 1
  }
  assert {
    condition     = length(google_compute_instance.control_modern[0].network_interface[0].access_config) == 0
    error_message = "Private defaults must never attach an external IP."
  }
  assert {
    condition     = google_compute_subnetwork.mi_subnet_modern[0].log_config[0].flow_sampling == 1.0
    error_message = "Managed subnet must retain full-sampling Flow Logs."
  }
}

run "public_ip_requires_allowlist" {
  command = plan
  variables {
    gce_public_ip = true
  }
  expect_failures = [var.gce_public_ip]
}

run "reject_global_ipv4" {
  command = plan
  variables {
    allowed_cidrs = ["0.0.0.0/0"]
  }
  expect_failures = [var.allowed_cidrs]
}

run "reject_global_ipv6" {
  command = plan
  variables {
    allowed_cidrs = ["::/0"]
  }
  expect_failures = [var.allowed_cidrs]
}

run "reject_malformed_cidr" {
  command = plan
  variables {
    allowed_cidrs = ["not-a-cidr"]
  }
  expect_failures = [var.allowed_cidrs]
}

run "reject_missing_network" {
  command = plan
  variables {
    use_modern_gce_schema = true
  }
  expect_failures = [var.modern_subnetwork_self_link]
}

run "public_ip_requires_managed_firewall" {
  command = plan
  variables {
    gce_public_ip               = true
    allowed_cidrs               = ["203.0.113.0/24"]
    modern_subnetwork_self_link = "projects/test/regions/us-central1/subnetworks/external"
  }
  expect_failures = [var.gce_public_ip]
}

run "explicit_restricted_public_ip" {
  command = plan
  variables {
    use_modern_gce_schema  = true
    use_modern_gce_network = true
    control_count          = 1
    gce_public_ip          = true
    allowed_cidrs          = ["203.0.113.0/24"]
  }
  assert {
    condition     = length(google_compute_instance.control_modern[0].network_interface[0].access_config) == 1 && google_compute_firewall.external_modern[0].source_ranges == toset(["203.0.113.0/24"])
    error_message = "Public IP opt-in must install its exact restricted source firewall."
  }
}
