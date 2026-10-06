<!-- SPDX-License-Identifier: Apache-2.0 -->
# Core platform inputs

Experimental passthrough module for shared platform inputs. It creates no cloud
resources and is not wired into the current compiler or provider blueprints.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |

## Providers

No providers.

## Modules

No modules.

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the Kubernetes cluster | `string` | n/a | yes |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name (dev, staging, prod) | `string` | n/a | yes |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version | `string` | `"1.31"` | no |
| <a name="input_region"></a> [region](#input\_region) | Cloud region to deploy into | `string` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Resource tags | `map(string)` | <pre>{<br/>  "ManagedBy": "Terraform",<br/>  "Project": "Mantl"<br/>}</pre> | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | Name of the cluster |
| <a name="output_environment"></a> [environment](#output\_environment) | Deployment environment |
| <a name="output_kubernetes_version"></a> [kubernetes\_version](#output\_kubernetes\_version) | Version of Kubernetes |
| <a name="output_region"></a> [region](#output\_region) | Configured cloud region |
| <a name="output_tags"></a> [tags](#output\_tags) | Configured resource tags |
<!-- END_TF_DOCS -->
