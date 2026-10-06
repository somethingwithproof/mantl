# gke-external-dns-wi

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.10.0 |
| <a name="requirement_google"></a> [google](#requirement\_google) | >= 7.0 |

## Providers

No providers.

## Modules

| Name | Source | Version |
| ---- | ------ | ------- |
| <a name="module_wi"></a> [wi](#module\_wi) | ../../modules/iam/gke-workload-identity | n/a |

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_gsa_name"></a> [gsa\_name](#input\_gsa\_name) | n/a | `string` | n/a | yes |
| <a name="input_ksa_name"></a> [ksa\_name](#input\_ksa\_name) | n/a | `string` | `"external-dns"` | no |
| <a name="input_ksa_namespace"></a> [ksa\_namespace](#input\_ksa\_namespace) | n/a | `string` | `"external-dns"` | no |
| <a name="input_project"></a> [project](#input\_project) | SPDX-License-Identifier: Apache-2.0 | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_gsa_email"></a> [gsa\_email](#output\_gsa\_email) | n/a |
<!-- END_TF_DOCS -->
