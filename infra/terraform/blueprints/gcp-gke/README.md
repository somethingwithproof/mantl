# gcp-gke

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_google"></a> [google](#requirement\_google) | >= 7.0, < 8.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_google"></a> [google](#provider\_google) | 7.46.1 |

## Modules

| Name | Source | Version |
| ---- | ------ | ------- |
| <a name="module_cert_manager_workload_identity"></a> [cert\_manager\_workload\_identity](#module\_cert\_manager\_workload\_identity) | terraform-google-modules/kubernetes-engine/google//modules/workload-identity | ~> 45.0 |
| <a name="module_external_dns_workload_identity"></a> [external\_dns\_workload\_identity](#module\_external\_dns\_workload\_identity) | terraform-google-modules/kubernetes-engine/google//modules/workload-identity | ~> 45.0 |
| <a name="module_external_secrets_workload_identity"></a> [external\_secrets\_workload\_identity](#module\_external\_secrets\_workload\_identity) | terraform-google-modules/kubernetes-engine/google//modules/workload-identity | ~> 45.0 |
| <a name="module_gke"></a> [gke](#module\_gke) | terraform-google-modules/kubernetes-engine/google//modules/beta-private-cluster | ~> 45.0 |

## Resources

| Name | Type |
| ---- | ---- |
| [google_compute_subnetwork.gke_subnet](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_subnetwork) | resource |
| [google_kms_crypto_key.gke](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/kms_crypto_key) | resource |
| [google_kms_crypto_key_iam_member.gke](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/kms_crypto_key_iam_member) | resource |
| [google_kms_key_ring.gke](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/kms_key_ring) | resource |
| [google_project_iam_audit_config.kubernetes](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/project_iam_audit_config) | resource |
| [google_project.project](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/project) | data source |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the GKE cluster | `string` | `"mantl"` | no |
| <a name="input_enable_binary_authorization"></a> [enable\_binary\_authorization](#input\_enable\_binary\_authorization) | Enable Binary Authorization for deployment policy enforcement | `bool` | `false` | no |
| <a name="input_enable_private_endpoint"></a> [enable\_private\_endpoint](#input\_enable\_private\_endpoint) | Enable private endpoint (NOT recommended for development) | `bool` | `true` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name (e.g., production, staging, development) | `string` | `"production"` | no |
| <a name="input_ip_range_pods"></a> [ip\_range\_pods](#input\_ip\_range\_pods) | CIDR block for pods | `string` | `"10.4.0.0/14"` | no |
| <a name="input_ip_range_services"></a> [ip\_range\_services](#input\_ip\_range\_services) | CIDR block for services | `string` | `"10.8.0.0/20"` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version for GKE cluster | `string` | `"1.31"` | no |
| <a name="input_master_authorized_networks"></a> [master\_authorized\_networks](#input\_master\_authorized\_networks) | List of CIDR blocks that can access the Kubernetes master | <pre>list(object({<br/>    cidr_block   = string<br/>    display_name = string<br/>  }))</pre> | `[]` | no |
| <a name="input_network"></a> [network](#input\_network) | VPC network name | `string` | `"default"` | no |
| <a name="input_project"></a> [project](#input\_project) | GCP project ID | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | GCP region | `string` | `"us-central1"` | no |
| <a name="input_subnet_cidr"></a> [subnet\_cidr](#input\_subnet\_cidr) | CIDR block for GKE subnet | `string` | `"10.0.0.0/20"` | no |
| <a name="input_system_node_machine_type"></a> [system\_node\_machine\_type](#input\_system\_node\_machine\_type) | Machine type for system node pool | `string` | `"e2-standard-4"` | no |
| <a name="input_system_node_max_count"></a> [system\_node\_max\_count](#input\_system\_node\_max\_count) | Maximum number of system nodes | `number` | `4` | no |
| <a name="input_system_node_min_count"></a> [system\_node\_min\_count](#input\_system\_node\_min\_count) | Minimum number of system nodes | `number` | `2` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Additional tags for all resources | `map(string)` | `{}` | no |
| <a name="input_workload_node_machine_type"></a> [workload\_node\_machine\_type](#input\_workload\_node\_machine\_type) | Machine type for workload node pool | `string` | `"e2-standard-4"` | no |
| <a name="input_workload_node_max_count"></a> [workload\_node\_max\_count](#input\_workload\_node\_max\_count) | Maximum number of workload nodes | `number` | `10` | no |
| <a name="input_workload_node_min_count"></a> [workload\_node\_min\_count](#input\_workload\_node\_min\_count) | Minimum number of workload nodes | `number` | `2` | no |
| <a name="input_workload_node_spot"></a> [workload\_node\_spot](#input\_workload\_node\_spot) | Use spot instances for workload nodes (cost savings) | `bool` | `false` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_ca_certificate"></a> [ca\_certificate](#output\_ca\_certificate) | GKE cluster CA certificate |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | GKE cluster ID |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | GKE cluster name |
| <a name="output_endpoint"></a> [endpoint](#output\_endpoint) | GKE cluster endpoint |
| <a name="output_platform_contract"></a> [platform\_contract](#output\_platform\_contract) | Provider-neutral platform identity; deployment verification is separate |
| <a name="output_region"></a> [region](#output\_region) | GKE cluster region |
| <a name="output_workload_identity_config"></a> [workload\_identity\_config](#output\_workload\_identity\_config) | Workload Identity configuration |
<!-- END_TF_DOCS -->
