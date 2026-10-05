# Experimental GCE virtual machines

This standalone Terraform module creates optional shielded GCE virtual machines,
data disks and a custom-mode VPC. It does not bootstrap Kubernetes or Mantl. It
has no recorded cloud acceptance and is not part of the supported AWS/GCP/Azure
static parity contract. Use `infra/terraform/blueprints/gcp-gke/` for that contract.

Terraform 1.9 or later is required; the Google provider is pinned in `versions.tf`.
Select Terraform through the repository's `mise` configuration. Provide the GCP
project through the provider's normal runtime configuration, with scoped credentials;
never place credentials in a variables file.

All instance counts default to zero and `use_modern_gce_schema` defaults to false.
There is no remaining monolithic legacy deployment path. To enable instances,
provide names, zone, region, a public SSH key file and explicit role counts. Select
`use_modern_gce_network = true` or supply `modern_subnetwork_self_link` for an
existing subnet. The default image is Debian 12, resolved from its image family;
this is not an immutable-image production deployment contract. Data disks and
boot disks accept an optional `gce_boot_kms_key` for CMEK.

All VM roles are private-only. There are no external-IP attachment blocks or
external service-ingress firewall rules. The deprecated `gce_public_ip` input
accepts only `false`, and `allowed_cidrs` accepts only an empty list; incompatible
legacy inputs fail validation rather than retaining public access.

Administrative ingress is disabled by default. Optional `enable_iap_ssh = true`
requires the managed network and creates a logged TCP/22-only firewall rule from
Google IAP's `35.235.240.0/20` service range, targeted to this module's VM tag. This
firewall does not grant tunnel access or authenticate SSH. Before use, provision
and review least-privilege IAP tunnel IAM (scoped to the intended instances and
port 22), SSH authentication and audit-log retention. Network tags scope firewall
traffic, not IAP IAM authorization. See
[Google's IAP prerequisites](https://docs.cloud.google.com/iap/docs/using-tcp-forwarding).
Externally supplied subnets require independent firewall and logging review;
this module cannot ensure they lack broader pre-existing rules.

The managed subnet retains full-sampling VPC Flow Logs. Configure retention and
log access separately. Internal subnet traffic remains broadly permitted and
requires a workload-specific segmentation review before production. Private
outbound connectivity, such as reviewed Cloud NAT or Private Google Access, must
be provisioned independently when needed; this module adds neither a NAT gateway
nor a bastion. IAP does not provide outbound internet connectivity.

For existing deployments, removing external addresses and the old firewall can
interrupt access. Review the Terraform plan and state, prove a private or IAP
administration path, and record a rollback procedure before applying. No cloud
resources are changed by the local checks. See
[ADR 010](../../../docs/adr/010-security-exception-boundaries.md).

Local checks do not use cloud credentials or provision resources:

```sh
mise exec -- terraform -chdir=infra/terraform/gce init -backend=false
mise exec -- terraform -chdir=infra/terraform/gce validate
mise exec -- terraform -chdir=infra/terraform/gce test
```

The test suite uses a mocked provider and plan-only runs. Its public-key fixture
is an inert string for metadata assertions, not a usable deployment key. Tests
verify private interfaces for all VM roles on managed and supplied subnets,
rejection of legacy public-access inputs, no default administrative ingress, and
the IAP rule's exact source, target, SSH port and logging configuration. These
checks do not establish deployment success, IAP authorization, SSH access or log
delivery. This module remains experimental and is not approved for production.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.9 |
| <a name="requirement_google"></a> [google](#requirement\_google) | 7.45.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_google"></a> [google](#provider\_google) | 7.45.0 |

## Modules

No modules.

## Resources

| Name | Type |
| ---- | ---- |
| [google_compute_disk.mi-control-lvm](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_disk) | resource |
| [google_compute_disk.mi-kubeworker-lvm](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_disk) | resource |
| [google_compute_disk.mi-worker-lvm](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_disk) | resource |
| [google_compute_firewall.iap_ssh](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_firewall) | resource |
| [google_compute_firewall.internal_modern](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_firewall) | resource |
| [google_compute_instance.control_modern](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_instance) | resource |
| [google_compute_instance.kubeworker_modern](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_instance) | resource |
| [google_compute_instance.worker_modern](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_instance) | resource |
| [google_compute_network.mi_network_modern](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_network) | resource |
| [google_compute_subnetwork.mi_subnet_modern](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/resources/compute_subnetwork) | resource |
| [google_compute_image.boot](https://registry.terraform.io/providers/hashicorp/google/7.45.0/docs/data-sources/compute_image) | data source |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_allowed_cidrs"></a> [allowed\_cidrs](#input\_allowed\_cidrs) | Deprecated compatibility input. External source allowlists are unsupported; must remain empty. | `list(string)` | `[]` | no |
| <a name="input_control_count"></a> [control\_count](#input\_control\_count) | n/a | `number` | `0` | no |
| <a name="input_control_type"></a> [control\_type](#input\_control\_type) | n/a | `string` | `"e2-standard-2"` | no |
| <a name="input_control_volume_size"></a> [control\_volume\_size](#input\_control\_volume\_size) | n/a | `number` | `50` | no |
| <a name="input_datacenter"></a> [datacenter](#input\_datacenter) | n/a | `string` | `"gce"` | no |
| <a name="input_enable_iap_ssh"></a> [enable\_iap\_ssh](#input\_enable\_iap\_ssh) | Allow SSH from Google IAP to this module's tagged VMs. Requires separately reviewed IAP IAM and the managed network. | `bool` | `false` | no |
| <a name="input_gce_boot_image_family"></a> [gce\_boot\_image\_family](#input\_gce\_boot\_image\_family) | Boot image family for control instances. | `string` | `"debian-12"` | no |
| <a name="input_gce_boot_image_project"></a> [gce\_boot\_image\_project](#input\_gce\_boot\_image\_project) | GCP project hosting the image family. | `string` | `"debian-cloud"` | no |
| <a name="input_gce_boot_kms_key"></a> [gce\_boot\_kms\_key](#input\_gce\_boot\_kms\_key) | Optional self\_link of the KMS crypto key for boot disk encryption. | `string` | `""` | no |
| <a name="input_gce_public_ip"></a> [gce\_public\_ip](#input\_gce\_public\_ip) | Deprecated compatibility input. Public IPs are unsupported; only false is accepted. | `bool` | `false` | no |
| <a name="input_kubeworker_count"></a> [kubeworker\_count](#input\_kubeworker\_count) | n/a | `number` | `0` | no |
| <a name="input_long_name"></a> [long\_name](#input\_long\_name) | Inputs for the experimental standalone VM module. No cluster bootstrap is performed. | `string` | n/a | yes |
| <a name="input_modern_subnetwork_self_link"></a> [modern\_subnetwork\_self\_link](#input\_modern\_subnetwork\_self\_link) | Optional self\_link of a modern subnetwork to attach (use with use\_modern\_gce\_network). | `string` | `""` | no |
| <a name="input_network_ipv4"></a> [network\_ipv4](#input\_network\_ipv4) | n/a | `string` | `"10.20.0.0/24"` | no |
| <a name="input_region"></a> [region](#input\_region) | n/a | `string` | n/a | yes |
| <a name="input_short_name"></a> [short\_name](#input\_short\_name) | n/a | `string` | n/a | yes |
| <a name="input_ssh_key"></a> [ssh\_key](#input\_ssh\_key) | n/a | `string` | n/a | yes |
| <a name="input_ssh_user"></a> [ssh\_user](#input\_ssh\_user) | n/a | `string` | `"mantl"` | no |
| <a name="input_use_modern_gce_network"></a> [use\_modern\_gce\_network](#input\_use\_modern\_gce\_network) | If true, expect a modern subnetwork; set modern\_subnetwork\_self\_link accordingly. | `bool` | `false` | no |
| <a name="input_use_modern_gce_schema"></a> [use\_modern\_gce\_schema](#input\_use\_modern\_gce\_schema) | If true, create control nodes with modern google\_compute\_instance schema. | `bool` | `false` | no |
| <a name="input_worker_count"></a> [worker\_count](#input\_worker\_count) | n/a | `number` | `0` | no |
| <a name="input_worker_type"></a> [worker\_type](#input\_worker\_type) | n/a | `string` | `"e2-standard-2"` | no |
| <a name="input_worker_volume_size"></a> [worker\_volume\_size](#input\_worker\_volume\_size) | n/a | `number` | `50` | no |
| <a name="input_zone"></a> [zone](#input\_zone) | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_modern_subnetwork_self_link"></a> [modern\_subnetwork\_self\_link](#output\_modern\_subnetwork\_self\_link) | n/a |
<!-- END_TF_DOCS -->
