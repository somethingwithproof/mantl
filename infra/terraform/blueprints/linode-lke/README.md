<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_helm"></a> [helm](#requirement\_helm) | 3.3.0 |
| <a name="requirement_kubernetes"></a> [kubernetes](#requirement\_kubernetes) | 3.3.0 |
| <a name="requirement_linode"></a> [linode](#requirement\_linode) | >= 3.0, < 4.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_linode"></a> [linode](#provider\_linode) | 3.14.1 |

## Modules

| Name | Source | Version |
| ---- | ------ | ------- |
| <a name="module_platform_security"></a> [platform\_security](#module\_platform\_security) | ../../modules/platform-security | n/a |

## Resources

| Name | Type |
| ---- | ---- |
| [linode_firewall.kubernetes](https://registry.terraform.io/providers/linode/linode/latest/docs/resources/firewall) | resource |
| [linode_lke_cluster.main](https://registry.terraform.io/providers/linode/linode/latest/docs/resources/lke_cluster) | resource |
| [linode_nodebalancer.ingress](https://registry.terraform.io/providers/linode/linode/latest/docs/resources/nodebalancer) | resource |
| [linode_nodebalancer_config.http](https://registry.terraform.io/providers/linode/linode/latest/docs/resources/nodebalancer_config) | resource |
| [linode_nodebalancer_config.https](https://registry.terraform.io/providers/linode/linode/latest/docs/resources/nodebalancer_config) | resource |
| [linode_object_storage_bucket.backups](https://registry.terraform.io/providers/linode/linode/latest/docs/resources/object_storage_bucket) | resource |
| [linode_object_storage_key.backups](https://registry.terraform.io/providers/linode/linode/latest/docs/resources/object_storage_key) | resource |
| [linode_volume.platform_data](https://registry.terraform.io/providers/linode/linode/latest/docs/resources/volume) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_allowed_nodeport_cidrs"></a> [allowed\_nodeport\_cidrs](#input\_allowed\_nodeport\_cidrs) | CIDR blocks allowed to access NodePort services | `list(string)` | <pre>[<br/>  "0.0.0.0/0"<br/>]</pre> | no |
| <a name="input_allowed_ssh_cidrs"></a> [allowed\_ssh\_cidrs](#input\_allowed\_ssh\_cidrs) | CIDR blocks allowed to SSH to nodes | `list(string)` | <pre>[<br/>  "10.0.0.0/8"<br/>]</pre> | no |
| <a name="input_backup_retention_days"></a> [backup\_retention\_days](#input\_backup\_retention\_days) | Number of days to retain backups | `number` | `30` | no |
| <a name="input_cluster_cidr"></a> [cluster\_cidr](#input\_cluster\_cidr) | CIDR block for Kubernetes cluster internal networking | `string` | `"10.2.0.0/16"` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the LKE cluster | `string` | `"mantl-lke"` | no |
| <a name="input_enable_backup_bucket"></a> [enable\_backup\_bucket](#input\_enable\_backup\_bucket) | Create Object Storage bucket for backups (S3-compatible) | `bool` | `true` | no |
| <a name="input_enable_cert_manager"></a> [enable\_cert\_manager](#input\_enable\_cert\_manager) | Enable cert-manager for automated TLS certificate management | `bool` | `true` | no |
| <a name="input_enable_cilium"></a> [enable\_cilium](#input\_enable\_cilium) | Enable Cilium CNI with Hubble for network observability (replaces missing VPC Flow Logs) | `bool` | `true` | no |
| <a name="input_enable_external_secrets"></a> [enable\_external\_secrets](#input\_enable\_external\_secrets) | Enable External Secrets Operator for external secrets management | `bool` | `false` | no |
| <a name="input_enable_falco"></a> [enable\_falco](#input\_enable\_falco) | Enable Falco for runtime security monitoring (enhanced security) | `bool` | `true` | no |
| <a name="input_enable_flow_logs_export"></a> [enable\_flow\_logs\_export](#input\_enable\_flow\_logs\_export) | Export Cilium flow logs to Object Storage bucket | `bool` | `true` | no |
| <a name="input_enable_ha_control_plane"></a> [enable\_ha\_control\_plane](#input\_enable\_ha\_control\_plane) | Enable high availability control plane | `bool` | `true` | no |
| <a name="input_enable_ingress_lb"></a> [enable\_ingress\_lb](#input\_enable\_ingress\_lb) | Create a NodeBalancer for ingress traffic | `bool` | `true` | no |
| <a name="input_enable_platform_volume"></a> [enable\_platform\_volume](#input\_enable\_platform\_volume) | Create a Block Storage volume for platform data | `bool` | `false` | no |
| <a name="input_enable_sealed_secrets"></a> [enable\_sealed\_secrets](#input\_enable\_sealed\_secrets) | Enable Sealed Secrets for encrypted secrets in Git (replaces missing KMS) | `bool` | `true` | no |
| <a name="input_enable_security_dashboard"></a> [enable\_security\_dashboard](#input\_enable\_security\_dashboard) | Enable Grafana dashboard for security monitoring | `bool` | `true` | no |
| <a name="input_enable_tailscale"></a> [enable\_tailscale](#input\_enable\_tailscale) | Enable Tailscale VPN for secure cluster access (no private endpoint available) | `bool` | `false` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name (e.g., production, staging, development) | `string` | `"production"` | no |
| <a name="input_falco_ui_enabled"></a> [falco\_ui\_enabled](#input\_falco\_ui\_enabled) | Enable Falco UI for alert visualization | `bool` | `false` | no |
| <a name="input_falco_webhook_url"></a> [falco\_webhook\_url](#input\_falco\_webhook\_url) | Webhook URL for Falco alerts (e.g., Slack, PagerDuty) | `string` | `""` | no |
| <a name="input_hubble_ui_hosts"></a> [hubble\_ui\_hosts](#input\_hubble\_ui\_hosts) | Hostnames for Hubble UI ingress | `list(string)` | `[]` | no |
| <a name="input_hubble_ui_ingress_enabled"></a> [hubble\_ui\_ingress\_enabled](#input\_hubble\_ui\_ingress\_enabled) | Enable ingress for Hubble UI | `bool` | `false` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version for the cluster | `string` | `"1.31"` | no |
| <a name="input_lb_throttle_connections"></a> [lb\_throttle\_connections](#input\_lb\_throttle\_connections) | Connection throttle for NodeBalancer (connections per second) | `number` | `0` | no |
| <a name="input_object_storage_region"></a> [object\_storage\_region](#input\_object\_storage\_region) | Region for Object Storage bucket (us-east-1, eu-central-1, ap-south-1) | `string` | `"us-east-1"` | no |
| <a name="input_platform_volume_size"></a> [platform\_volume\_size](#input\_platform\_volume\_size) | Size of platform volume in GB | `number` | `100` | no |
| <a name="input_region"></a> [region](#input\_region) | Linode region (e.g., us-east, us-west, eu-west, ap-south) | `string` | `"us-east"` | no |
| <a name="input_system_node_count"></a> [system\_node\_count](#input\_system\_node\_count) | Number of system nodes (non-autoscaling) | `number` | `3` | no |
| <a name="input_system_node_type"></a> [system\_node\_type](#input\_system\_node\_type) | Linode type for system nodes (g6-standard-4 = 4 vCPU, 8 GB RAM) | `string` | `"g6-standard-4"` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Additional tags to apply to resources | `list(string)` | <pre>[<br/>  "mantl",<br/>  "kubernetes"<br/>]</pre> | no |
| <a name="input_tailscale_client_id"></a> [tailscale\_client\_id](#input\_tailscale\_client\_id) | Tailscale OAuth client ID | `string` | `""` | no |
| <a name="input_tailscale_client_secret"></a> [tailscale\_client\_secret](#input\_tailscale\_client\_secret) | Tailscale OAuth client secret | `string` | `""` | no |
| <a name="input_workload_pools"></a> [workload\_pools](#input\_workload\_pools) | List of workload node pools with autoscaling | <pre>list(object({<br/>    name       = string<br/>    node_type  = string<br/>    node_count = number<br/>    min_nodes  = number<br/>    max_nodes  = number<br/>  }))</pre> | <pre>[<br/>  {<br/>    "max_nodes": 10,<br/>    "min_nodes": 2,<br/>    "name": "workload",<br/>    "node_count": 2,<br/>    "node_type": "g6-standard-4"<br/>  }<br/>]</pre> | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_api_endpoints"></a> [api\_endpoints](#output\_api\_endpoints) | The API endpoints of the LKE cluster |
| <a name="output_backup_access_key"></a> [backup\_access\_key](#output\_backup\_access\_key) | The access key for the backup bucket |
| <a name="output_backup_bucket"></a> [backup\_bucket](#output\_backup\_bucket) | The name of the backup bucket |
| <a name="output_backup_bucket_region"></a> [backup\_bucket\_region](#output\_backup\_bucket\_region) | The region of the backup bucket |
| <a name="output_backup_secret_key"></a> [backup\_secret\_key](#output\_backup\_secret\_key) | The secret key for the backup bucket |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | The ID of the LKE cluster |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | The name of the LKE cluster |
| <a name="output_firewall_id"></a> [firewall\_id](#output\_firewall\_id) | The ID of the cluster firewall |
| <a name="output_kubeconfig"></a> [kubeconfig](#output\_kubeconfig) | The kubeconfig for the LKE cluster (base64 encoded) |
| <a name="output_nodebalancer_ip"></a> [nodebalancer\_ip](#output\_nodebalancer\_ip) | The IP address of the ingress NodeBalancer |
| <a name="output_platform_security_enabled_features"></a> [platform\_security\_enabled\_features](#output\_platform\_security\_enabled\_features) | Summary of enabled platform security features |
| <a name="output_platform_volume_id"></a> [platform\_volume\_id](#output\_platform\_volume\_id) | The ID of the platform data volume |
| <a name="output_status"></a> [status](#output\_status) | The status of the LKE cluster |
<!-- END_TF_DOCS -->
