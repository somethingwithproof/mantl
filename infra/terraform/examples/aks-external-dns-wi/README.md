# aks-external-dns-wi

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.10.0 |
| <a name="requirement_azurerm"></a> [azurerm](#requirement\_azurerm) | >= 4.57 |

## Providers

No providers.

## Modules

| Name | Source | Version |
| ---- | ------ | ------- |
| <a name="module_wi"></a> [wi](#module\_wi) | ../../modules/iam/aks-workload-identity | n/a |

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_dns_zone_id"></a> [dns\_zone\_id](#input\_dns\_zone\_id) | n/a | `string` | n/a | yes |
| <a name="input_issuer_url"></a> [issuer\_url](#input\_issuer\_url) | n/a | `string` | n/a | yes |
| <a name="input_ksa_name"></a> [ksa\_name](#input\_ksa\_name) | n/a | `string` | `"external-dns"` | no |
| <a name="input_ksa_namespace"></a> [ksa\_namespace](#input\_ksa\_namespace) | n/a | `string` | `"external-dns"` | no |
| <a name="input_location"></a> [location](#input\_location) | n/a | `string` | n/a | yes |
| <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name) | n/a | `string` | n/a | yes |
| <a name="input_uami_name"></a> [uami\_name](#input\_uami\_name) | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_client_id"></a> [client\_id](#output\_client\_id) | n/a |
<!-- END_TF_DOCS -->
