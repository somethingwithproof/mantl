mock_provider "google" {}

variables {
  long_name  = "mantl-local-test"
  short_name = "mantl-test"
  zone       = "us-central1-a"
  region     = "us-central1"
  ssh_key    = "./tests/fixture.pub"
}

run "all_roles_are_private" {
  command = plan
  variables {
    use_modern_gce_schema  = true
    use_modern_gce_network = true
    control_count         = 1
    worker_count          = 1
    kubeworker_count       = 1
  }
  assert {
    condition = alltrue([
      for instance in concat(google_compute_instance.control_modern, google_compute_instance.worker_modern, google_compute_instance.kubeworker_modern) : length(instance.network_interface[0].access_config) == 0
    ])
    error_message = "Every VM role must remain private, including control-plane nodes."
  }
  assert {
    condition     = length(google_compute_firewall.iap_ssh) == 0
    error_message = "Administrative ingress must be disabled by default."
  }
  assert {
    condition     = google_compute_subnetwork.mi_subnet_modern[0].log_config[0].flow_sampling == 1.0
    error_message = "Managed subnet must retain full-sampling Flow Logs."
  }
}

run "reject_public_ip_even_with_managed_network" {
  command = plan
  variables {
    use_modern_gce_schema  = true
    use_modern_gce_network = true
    control_count         = 1
    gce_public_ip         = true
  }
  expect_failures = [var.gce_public_ip]
}

run "reject_legacy_external_allowlist" {
  command = plan
  variables {
    allowed_cidrs = ["203.0.113.0/24"]
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

run "private_existing_subnet" {
  command = plan
  variables {
    use_modern_gce_schema       = true
    control_count              = 1
    worker_count               = 1
    kubeworker_count            = 1
    modern_subnetwork_self_link = "projects/test/regions/us-central1/subnetworks/external"
  }
  assert {
    condition = alltrue([
      for instance in concat(google_compute_instance.control_modern, google_compute_instance.worker_modern, google_compute_instance.kubeworker_modern) : length(instance.network_interface[0].access_config) == 0
    ])
    error_message = "An externally supplied subnet must not re-enable public IPs."
  }
}

run "reject_iap_without_managed_network" {
  command = plan
  variables {
    enable_iap_ssh              = true
    modern_subnetwork_self_link = "projects/test/regions/us-central1/subnetworks/external"
  }
  expect_failures = [var.enable_iap_ssh]
}

run "iap_ssh_is_scoped_and_logged" {
  command = plan
  variables {
    use_modern_gce_schema  = true
    use_modern_gce_network = true
    control_count         = 1
    enable_iap_ssh         = true
  }
  assert {
    condition = (
      google_compute_firewall.iap_ssh[0].source_ranges == toset(["35.235.240.0/20"]) &&
      google_compute_firewall.iap_ssh[0].target_tags == toset([var.short_name]) &&
      length(google_compute_firewall.iap_ssh[0].allow) == 1 &&
      one(google_compute_firewall.iap_ssh[0].allow).protocol == "tcp" &&
      one(google_compute_firewall.iap_ssh[0].allow).ports == tolist(["22"]) &&
      google_compute_firewall.iap_ssh[0].log_config[0].metadata == "INCLUDE_ALL_METADATA"
    )
    error_message = "IAP ingress must allow only logged SSH from the IAP service to this module's VMs."
  }
  assert {
    condition     = length(google_compute_instance.control_modern[0].network_interface[0].access_config) == 0
    error_message = "Enabling IAP must never attach a public IP."
  }
}
