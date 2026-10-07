<!-- SPDX-License-Identifier: Apache-2.0 -->
# openstack

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_openstack"></a> [openstack](#requirement\_openstack) | 3.4.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_openstack"></a> [openstack](#provider\_openstack) | 3.4.0 |

## Modules

No modules.

## Resources

| Name | Type |
| ---- | ---- |
| [openstack_containerinfra_cluster_v1.k8s_cluster](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/containerinfra_cluster_v1) | resource |
| [openstack_containerinfra_clustertemplate_v1.k8s_template](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/containerinfra_clustertemplate_v1) | resource |
| [openstack_identity_application_credential_v3.backup_credentials](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/identity_application_credential_v3) | resource |
| [openstack_networking_network_v2.cluster_network](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_network_v2) | resource |
| [openstack_networking_router_interface_v2.cluster_router_interface](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_router_interface_v2) | resource |
| [openstack_networking_router_v2.cluster_router](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_router_v2) | resource |
| [openstack_networking_secgroup_rule_v2.http](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_secgroup_rule_v2) | resource |
| [openstack_networking_secgroup_rule_v2.https](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_secgroup_rule_v2) | resource |
| [openstack_networking_secgroup_rule_v2.intra_cluster](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_secgroup_rule_v2) | resource |
| [openstack_networking_secgroup_rule_v2.kube_api](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_secgroup_rule_v2) | resource |
| [openstack_networking_secgroup_rule_v2.outbound](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_secgroup_rule_v2) | resource |
| [openstack_networking_secgroup_rule_v2.ssh](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_secgroup_rule_v2) | resource |
| [openstack_networking_secgroup_v2.cluster_nodes](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_secgroup_v2) | resource |
| [openstack_networking_subnet_v2.cluster_subnet](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/networking_subnet_v2) | resource |
| [openstack_objectstorage_container_v1.backups](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/resources/objectstorage_container_v1) | resource |
| [openstack_containerinfra_clustertemplate_v1.k8s_template](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/data-sources/containerinfra_clustertemplate_v1) | data source |
| [openstack_networking_network_v2.external](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/3.4.0/docs/data-sources/networking_network_v2) | data source |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_allowed_ssh_cidrs"></a> [allowed\_ssh\_cidrs](#input\_allowed\_ssh\_cidrs) | CIDR blocks allowed to SSH to nodes | `list(string)` | `[]` | no |
| <a name="input_api_server_access_cidr"></a> [api\_server\_access\_cidr](#input\_api\_server\_access\_cidr) | CIDR block allowed to access Kubernetes API server | `string` | `"0.0.0.0/0"` | no |
| <a name="input_cluster_cidr"></a> [cluster\_cidr](#input\_cluster\_cidr) | CIDR block for cluster network | `string` | `"10.0.0.0/24"` | no |
| <a name="input_cluster_create_timeout"></a> [cluster\_create\_timeout](#input\_cluster\_create\_timeout) | Timeout for cluster creation in minutes | `number` | `60` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the Kubernetes cluster | `string` | `"mantl-openstack"` | no |
| <a name="input_dns_nameserver"></a> [dns\_nameserver](#input\_dns\_nameserver) | DNS nameserver for cluster nodes | `string` | `"8.8.8.8"` | no |
| <a name="input_docker_volume_size"></a> [docker\_volume\_size](#input\_docker\_volume\_size) | Size of Docker storage volume in GB | `number` | `50` | no |
| <a name="input_enable_auto_healing"></a> [enable\_auto\_healing](#input\_enable\_auto\_healing) | Enable auto-healing for failed nodes | `bool` | `true` | no |
| <a name="input_enable_auto_scaling"></a> [enable\_auto\_scaling](#input\_enable\_auto\_scaling) | Enable cluster auto-scaling | `bool` | `true` | no |
| <a name="input_enable_backup_container"></a> [enable\_backup\_container](#input\_enable\_backup\_container) | Create Swift container for backups | `bool` | `true` | no |
| <a name="input_enable_floating_ip"></a> [enable\_floating\_ip](#input\_enable\_floating\_ip) | Assign floating IP to master nodes | `bool` | `true` | no |
| <a name="input_enable_master_lb"></a> [enable\_master\_lb](#input\_enable\_master\_lb) | Enable load balancer for master nodes | `bool` | `true` | no |
| <a name="input_enable_registry"></a> [enable\_registry](#input\_enable\_registry) | Enable integrated Docker registry | `bool` | `false` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name (e.g., production, staging, development) | `string` | `"production"` | no |
| <a name="input_existing_template_name"></a> [existing\_template\_name](#input\_existing\_template\_name) | Name of existing cluster template (if use\_existing\_template is true) | `string` | `""` | no |
| <a name="input_external_network_name"></a> [external\_network\_name](#input\_external\_network\_name) | Name of the external network for floating IPs | `string` | n/a | yes |
| <a name="input_http_proxy"></a> [http\_proxy](#input\_http\_proxy) | HTTP proxy URL (optional) | `string` | `""` | no |
| <a name="input_https_proxy"></a> [https\_proxy](#input\_https\_proxy) | HTTPS proxy URL (optional) | `string` | `""` | no |
| <a name="input_initial_node_count"></a> [initial\_node\_count](#input\_initial\_node\_count) | Initial number of worker nodes | `number` | `3` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version tag (e.g., v1.31.1) | `string` | `"v1.31.1"` | no |
| <a name="input_master_count"></a> [master\_count](#input\_master\_count) | Number of master nodes (1 or 3 recommended for HA) | `number` | `3` | no |
| <a name="input_master_flavor"></a> [master\_flavor](#input\_master\_flavor) | OpenStack flavor for master nodes | `string` | `"m1.medium"` | no |
| <a name="input_max_node_count"></a> [max\_node\_count](#input\_max\_node\_count) | Maximum number of worker nodes (for autoscaling) | `number` | `10` | no |
| <a name="input_min_node_count"></a> [min\_node\_count](#input\_min\_node\_count) | Minimum number of worker nodes (for autoscaling) | `number` | `2` | no |
| <a name="input_network_driver"></a> [network\_driver](#input\_network\_driver) | Network driver for Kubernetes (flannel or calico) | `string` | `"flannel"` | no |
| <a name="input_no_proxy"></a> [no\_proxy](#input\_no\_proxy) | No proxy list (comma-separated, optional) | `string` | `""` | no |
| <a name="input_node_image_name"></a> [node\_image\_name](#input\_node\_image\_name) | Name of the Glance image for Kubernetes nodes (Fedora CoreOS recommended) | `string` | `"fedora-coreos-latest"` | no |
| <a name="input_ssh_keypair_name"></a> [ssh\_keypair\_name](#input\_ssh\_keypair\_name) | Name of OpenStack SSH keypair for node access | `string` | n/a | yes |
| <a name="input_template_labels"></a> [template\_labels](#input\_template\_labels) | Additional labels for cluster template | `map(string)` | `{}` | no |
| <a name="input_use_existing_template"></a> [use\_existing\_template](#input\_use\_existing\_template) | Use an existing cluster template instead of creating new one | `bool` | `false` | no |
| <a name="input_worker_flavor"></a> [worker\_flavor](#input\_worker\_flavor) | OpenStack flavor for worker nodes | `string` | `"m1.large"` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_api_address"></a> [api\_address](#output\_api\_address) | The API server address |
| <a name="output_backup_container"></a> [backup\_container](#output\_backup\_container) | The name of the backup Swift container |
| <a name="output_backup_credentials_id"></a> [backup\_credentials\_id](#output\_backup\_credentials\_id) | The ID of backup application credentials |
| <a name="output_backup_credentials_secret"></a> [backup\_credentials\_secret](#output\_backup\_credentials\_secret) | The secret for backup application credentials |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | The ID of the Kubernetes cluster |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | The name of the Kubernetes cluster |
| <a name="output_kubeconfig"></a> [kubeconfig](#output\_kubeconfig) | The kubeconfig for accessing the cluster |
| <a name="output_master_addresses"></a> [master\_addresses](#output\_master\_addresses) | The IP addresses of master nodes |
| <a name="output_network_id"></a> [network\_id](#output\_network\_id) | The ID of the cluster network |
| <a name="output_node_addresses"></a> [node\_addresses](#output\_node\_addresses) | The IP addresses of worker nodes |
| <a name="output_security_group_id"></a> [security\_group\_id](#output\_security\_group\_id) | The ID of the cluster security group |
| <a name="output_subnet_id"></a> [subnet\_id](#output\_subnet\_id) | The ID of the cluster subnet |
<!-- END_TF_DOCS -->
