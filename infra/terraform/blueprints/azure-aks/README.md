# azure-aks

This blueprint is beta. Static validation does not establish cloud readiness.

## AzureRM 4 to 5 migration

Provider 5.8.0 requires service-endpoint blocks, ID-based private DNS links and
federated credentials, explicit manual node provisioning, and a separate queue
properties resource. The blueprint registers only the Azure resource-provider
namespaces it uses; the deployment principal must have registration permission,
or these namespaces must be registered before deployment.

Key Vault now uses RBAC. The AKS identity receives **Key Vault Crypto Service
Encryption User** on the versionless ARM ID of the AKS key; External Secrets
receives **Key Vault Secrets User** on this vault. Existing access-policy resource
addresses cannot be moved into role assignments because their resource types and
identities differ. Before changing an existing vault's authorization mode, create
and verify these two assignments externally, allow for RBAC propagation, and
import them into `azurerm_role_assignment.aks_key` and
`azurerm_role_assignment.external_secrets` using their ARM role-assignment IDs.
Review a saved plan in a maintenance window: the old
`azurerm_key_vault_access_policy.aks` and
`azurerm_key_vault_access_policy.external_secrets` resources are removed, and
vault authorization changes to RBAC. Do not apply unless both workloads will
retain their required permissions through that transition.

The Terraform deployment identity also needs pre-existing **Key Vault Crypto
Officer** permission on the vault to manage the key and its rotation policy, plus
permission to manage role assignments. It must reach the private vault endpoint;
public access remains disabled. New environments need staged vault/network and
identity provisioning before key creation. Do not grant subscription-wide roles
to bypass these prerequisites.

Import existing queue-service diagnostics into
`azurerm_storage_account_queue_properties.flow_logs` with the storage account ARM
ID followed by `/queueServices/default`. Review the resulting logging settings
before apply. Existing NSG flow-log deployments retain their current target;
new deployments must review Azure's NSG-flow-log retirement and migrate to VNet
flow logs separately. No cloud plan, state mutation or apply has been performed.

References: [AzureRM 5 upgrade guide](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/guides/5.0-upgrade-guide),
[Key Vault RBAC migration](https://learn.microsoft.com/azure/key-vault/general/rbac-guide),
[queue diagnostics resource](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/storage_account_queue_properties).

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_azurerm"></a> [azurerm](#requirement\_azurerm) | 5.8.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_azurerm"></a> [azurerm](#provider\_azurerm) | 5.8.0 |

## Modules

No modules.

## Resources

| Name | Type |
| ---- | ---- |
| [azurerm_federated_identity_credential.cert_manager](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/federated_identity_credential) | resource |
| [azurerm_federated_identity_credential.external_dns](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/federated_identity_credential) | resource |
| [azurerm_federated_identity_credential.external_secrets](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/federated_identity_credential) | resource |
| [azurerm_key_vault.main](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/key_vault) | resource |
| [azurerm_key_vault_key.aks](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/key_vault_key) | resource |
| [azurerm_kubernetes_cluster.main](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/kubernetes_cluster) | resource |
| [azurerm_kubernetes_cluster_node_pool.workload](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/kubernetes_cluster_node_pool) | resource |
| [azurerm_log_analytics_workspace.main](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/log_analytics_workspace) | resource |
| [azurerm_monitor_diagnostic_setting.aks_audit](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/monitor_diagnostic_setting) | resource |
| [azurerm_network_security_group.aks_nodes](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/network_security_group) | resource |
| [azurerm_network_watcher.main](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/network_watcher) | resource |
| [azurerm_network_watcher_flow_log.aks_nodes](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/network_watcher_flow_log) | resource |
| [azurerm_private_dns_zone.key_vault](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/private_dns_zone) | resource |
| [azurerm_private_dns_zone_virtual_network_link.key_vault](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/private_dns_zone_virtual_network_link) | resource |
| [azurerm_private_endpoint.key_vault](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/private_endpoint) | resource |
| [azurerm_resource_group.main](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/resource_group) | resource |
| [azurerm_role_assignment.aks_key](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/role_assignment) | resource |
| [azurerm_role_assignment.external_secrets](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/role_assignment) | resource |
| [azurerm_storage_account.flow_logs](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/storage_account) | resource |
| [azurerm_storage_account_queue_properties.flow_logs](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/storage_account_queue_properties) | resource |
| [azurerm_subnet.aks_nodes](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/subnet) | resource |
| [azurerm_subnet_network_security_group_association.aks_nodes](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/subnet_network_security_group_association) | resource |
| [azurerm_user_assigned_identity.aks](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/user_assigned_identity) | resource |
| [azurerm_user_assigned_identity.cert_manager](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/user_assigned_identity) | resource |
| [azurerm_user_assigned_identity.external_dns](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/user_assigned_identity) | resource |
| [azurerm_user_assigned_identity.external_secrets](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/user_assigned_identity) | resource |
| [azurerm_virtual_network.main](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/resources/virtual_network) | resource |
| [azurerm_client_config.current](https://registry.terraform.io/providers/hashicorp/azurerm/5.8.0/docs/data-sources/client_config) | data source |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_admin_group_object_ids"></a> [admin\_group\_object\_ids](#input\_admin\_group\_object\_ids) | Azure AD group object IDs for cluster admin access | `list(string)` | `[]` | no |
| <a name="input_aks_subnet_cidr"></a> [aks\_subnet\_cidr](#input\_aks\_subnet\_cidr) | CIDR block for AKS subnet | `string` | `"10.0.0.0/20"` | no |
| <a name="input_authorized_ip_ranges"></a> [authorized\_ip\_ranges](#input\_authorized\_ip\_ranges) | List of authorized IP ranges that can access the Kubernetes API server (when not using private endpoint) | `list(string)` | `[]` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the AKS cluster | `string` | `"mantl"` | no |
| <a name="input_dns_service_ip"></a> [dns\_service\_ip](#input\_dns\_service\_ip) | IP address for Kubernetes DNS service (must be within service\_cidr) | `string` | `"10.1.0.10"` | no |
| <a name="input_enable_private_endpoint"></a> [enable\_private\_endpoint](#input\_enable\_private\_endpoint) | Enable private endpoint (NOT recommended for development) | `bool` | `true` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name (e.g., production, staging, development) | `string` | `"production"` | no |
| <a name="input_flow_logs_retention_days"></a> [flow\_logs\_retention\_days](#input\_flow\_logs\_retention\_days) | Flow logs retention in days | `number` | `7` | no |
| <a name="input_key_vault_key_expiration_date"></a> [key\_vault\_key\_expiration\_date](#input\_key\_vault\_key\_expiration\_date) | RFC 3339 expiration date for the AKS Key Vault encryption key | `string` | `null` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version for AKS cluster | `string` | `"1.31"` | no |
| <a name="input_location"></a> [location](#input\_location) | Azure region | `string` | `"East US"` | no |
| <a name="input_log_retention_days"></a> [log\_retention\_days](#input\_log\_retention\_days) | Log retention in days for Log Analytics workspace | `number` | `30` | no |
| <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name) | Name of the Azure Resource Group | `string` | n/a | yes |
| <a name="input_service_cidr"></a> [service\_cidr](#input\_service\_cidr) | CIDR block for Kubernetes services | `string` | `"10.1.0.0/16"` | no |
| <a name="input_system_node_count"></a> [system\_node\_count](#input\_system\_node\_count) | Initial number of system nodes | `number` | `2` | no |
| <a name="input_system_node_max_count"></a> [system\_node\_max\_count](#input\_system\_node\_max\_count) | Maximum number of system nodes (auto-scaling) | `number` | `4` | no |
| <a name="input_system_node_min_count"></a> [system\_node\_min\_count](#input\_system\_node\_min\_count) | Minimum number of system nodes (auto-scaling) | `number` | `2` | no |
| <a name="input_system_node_vm_size"></a> [system\_node\_vm\_size](#input\_system\_node\_vm\_size) | VM size for system node pool | `string` | `"Standard_D4s_v5"` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Additional tags for all resources | `map(string)` | `{}` | no |
| <a name="input_vnet_cidr"></a> [vnet\_cidr](#input\_vnet\_cidr) | CIDR block for VNet | `string` | `"10.0.0.0/16"` | no |
| <a name="input_workload_node_max_count"></a> [workload\_node\_max\_count](#input\_workload\_node\_max\_count) | Maximum number of workload nodes (auto-scaling) | `number` | `10` | no |
| <a name="input_workload_node_min_count"></a> [workload\_node\_min\_count](#input\_workload\_node\_min\_count) | Minimum number of workload nodes (auto-scaling) | `number` | `2` | no |
| <a name="input_workload_node_spot"></a> [workload\_node\_spot](#input\_workload\_node\_spot) | Use spot instances for workload nodes (cost savings) | `bool` | `false` | no |
| <a name="input_workload_node_vm_size"></a> [workload\_node\_vm\_size](#input\_workload\_node\_vm\_size) | VM size for workload node pool | `string` | `"Standard_D4s_v5"` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_cert_manager_identity_client_id"></a> [cert\_manager\_identity\_client\_id](#output\_cert\_manager\_identity\_client\_id) | Client ID for cert-manager managed identity |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | AKS cluster ID |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | AKS cluster name |
| <a name="output_external_dns_identity_client_id"></a> [external\_dns\_identity\_client\_id](#output\_external\_dns\_identity\_client\_id) | Client ID for External DNS managed identity |
| <a name="output_external_secrets_identity_client_id"></a> [external\_secrets\_identity\_client\_id](#output\_external\_secrets\_identity\_client\_id) | Client ID for External Secrets managed identity |
| <a name="output_host"></a> [host](#output\_host) | Kubernetes API server endpoint |
| <a name="output_key_vault_id"></a> [key\_vault\_id](#output\_key\_vault\_id) | Azure Key Vault ID for secrets |
| <a name="output_kube_config"></a> [kube\_config](#output\_kube\_config) | Kubeconfig for the cluster |
| <a name="output_log_analytics_workspace_id"></a> [log\_analytics\_workspace\_id](#output\_log\_analytics\_workspace\_id) | Log Analytics workspace ID |
| <a name="output_platform_contract"></a> [platform\_contract](#output\_platform\_contract) | Provider-neutral platform identity; deployment verification is separate |
<!-- END_TF_DOCS -->
