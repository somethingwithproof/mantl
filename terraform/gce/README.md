# GCE Terraform (mantl/terraform/gce)

This module is being modernized incrementally for Terraform 1.x and the hashicorp/google provider v7.x.

Key flags and secure defaults
- use_modern_gce_schema (bool, default false): create control instances via modern google_compute_instance with shielded VMs, project SSH keys blocked, and no public IP by default.
- gce_public_ip (bool, default false): when true, attaches an external IP to modern instances.
- gce_boot_image_family (string) and gce_boot_image_project (string): resolve boot image from a family.
- gce_boot_kms_key (string, optional): KMS key self_link for CMEK boot disk encryption.
- allowed_cidrs (list(string), default ["0.0.0.0/0"]): source ranges for the external firewall; override to restrict.
- use_modern_gce_network (bool, default false): create a custom-mode VPC + subnetwork and modern firewalls (see network_modern.tf) — introduced in a separate PR.
- modern_subnetwork_self_link (string, optional): when set with use_modern_gce_schema, modern instances attach to this subnetwork instead of the legacy network.

Examples

Important: When enabling modern resources for a role, set the corresponding legacy count to 0 in the monolithic gce.tf to avoid creating duplicate nodes (e.g., set control_count = 0 or worker_count = 0 as appropriate).

Minimal (legacy-compatible):
```hcl
allowed_cidrs = ["203.0.113.0/24"]
```

Enable modern control instances with no public IP and attach to modern subnetwork:
```hcl
use_modern_gce_schema        = true
use_modern_gce_network       = true
modern_subnetwork_self_link  = "projects/your-proj/regions/us-central1/subnetworks/mantl-subnet"
allowed_cidrs                = ["203.0.113.0/24"]
```

Notes
- Legacy resources remain the default path to avoid breaking existing deployments. Opt-in to the modern paths using the flags above.
- The monolithic gce.tf will be retired in favor of split modules in future PRs.
