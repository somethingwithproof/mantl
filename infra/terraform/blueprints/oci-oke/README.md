<!-- SPDX-License-Identifier: Apache-2.0 -->
# Oracle Cloud Infrastructure (OCI) - OKE Blueprint

Terraform blueprint for deploying Oracle Kubernetes Engine (OKE) with Mantl platform.

## Prerequisites

1. **OCI Account**: Active Oracle Cloud account with a compartment
2. **OCI CLI**: Install from https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliinstall.htm
3. **API Keys**: Generate API signing keys
   ```bash
   mkdir -p ~/.oci
   openssl genrsa -out ~/.oci/oci_api_key.pem 2048
   openssl rsa -pubout -in ~/.oci/oci_api_key.pem -out ~/.oci/oci_api_key_public.pem
   ```
4. **Upload Public Key**: Add `oci_api_key_public.pem` to OCI Console → User Settings → API Keys

## Configuration

Create `terraform.tfvars`:

```hcl
tenancy_ocid     = "ocid1.tenancy.oc1..aaa..."
user_ocid        = "ocid1.user.oc1..aaa..."
fingerprint      = "aa:bb:cc:dd:ee:ff:..."
compartment_id   = "ocid1.compartment.oc1..aaa..."
region           = "us-ashburn-1"
cluster_name     = "mantl-oke"
kubernetes_version = "v1.31.1"
node_count       = 3
node_shape       = "VM.Standard.E4.Flex"
node_ocpus       = 2
node_memory_gb   = 16
```

## Available Regions

- `us-ashburn-1` (Ashburn, VA)
- `us-phoenix-1` (Phoenix, AZ)
- `eu-frankfurt-1` (Frankfurt, Germany)
- `eu-amsterdam-1` (Amsterdam, Netherlands)
- `uk-london-1` (London, UK)
- `ap-tokyo-1` (Tokyo, Japan)
- `ap-sydney-1` (Sydney, Australia)

## Node Shapes

**Flex Shapes** (recommended, customizable OCPUs/memory):
- `VM.Standard.E4.Flex` - AMD EPYC
- `VM.Standard.E5.Flex` - AMD EPYC Milan
- `VM.Standard.A1.Flex` - Ampere Altra (ARM)

**Fixed Shapes**:
- `VM.Standard2.4` - 4 OCPUs, 60 GB RAM
- `VM.Standard3.Flex` - Intel

## Deploy

```bash
terraform init
terraform plan
terraform apply
```

## Access Cluster

The cluster will automatically configure kubectl:

```bash
# Or manually:
oci ce cluster create-kubeconfig \
  --cluster-id <cluster-ocid> \
  --file $HOME/.kube/config \
  --region us-ashburn-1 \
  --token-version 2.0.0

# Verify
kubectl get nodes
```

## Workload Identity

OKE uses **Instance Principals** for pod authentication to OCI services.

### Enable Instance Principal for Pods

1. Create dynamic group for OKE nodes:
   ```
   ALL {instance.compartment.id = 'ocid1.compartment.oc1..aaa...'}
   ```

2. Create policy for ExternalDNS/cert-manager:
   ```
   Allow dynamic-group oke-nodes to manage dns in compartment mantl
   Allow dynamic-group oke-nodes to manage certificates in compartment mantl
   ```

3. Configure ExternalDNS/cert-manager to use instance principal authentication

## Network Architecture

- **VCN**: 10.0.0.0/16
- **API Subnet**: 10.0.0.0/28 (public)
- **Node Subnet**: 10.0.10.0/24 (public)
- **Load Balancer Subnet**: 10.0.20.0/24 (public)
- **Pod CIDR**: 10.244.0.0/16
- **Service CIDR**: 10.96.0.0/16

## Costs

Approximate monthly costs (us-ashburn-1):
- **OKE Control Plane**: Free (managed by Oracle)
- **3x VM.Standard.E4.Flex** (2 OCPU, 16 GB each): ~$75/month
- **Load Balancers**: ~$20-40/month per LB
- **Block Storage**: ~$0.05/GB/month

**Always Free Tier**: 2x VM.Standard.E2.1.Micro (1 OCPU, 1 GB RAM each)

## Destroy

```bash
terraform destroy
```

## Next Steps

Deploy the Mantl platform:

```bash
cd ../../../
kubectl apply -f clusters/production/platform-apps.yaml
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_oci"></a> [oci](#requirement\_oci) | 9.8.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_oci"></a> [oci](#provider\_oci) | 9.8.0 |

## Modules

No modules.

## Resources

| Name | Type |
| ---- | ---- |
| [oci_containerengine_cluster.oke_cluster](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/containerengine_cluster) | resource |
| [oci_containerengine_node_pool.system_pool](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/containerengine_node_pool) | resource |
| [oci_containerengine_node_pool.workload_pool](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/containerengine_node_pool) | resource |
| [oci_core_internet_gateway.oke_igw](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_internet_gateway) | resource |
| [oci_core_nat_gateway.oke_nat](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_nat_gateway) | resource |
| [oci_core_route_table.private_route_table](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_route_table) | resource |
| [oci_core_route_table.public_route_table](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_route_table) | resource |
| [oci_core_security_list.api_endpoint_seclist](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_security_list) | resource |
| [oci_core_security_list.lb_seclist](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_security_list) | resource |
| [oci_core_security_list.node_seclist](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_security_list) | resource |
| [oci_core_service_gateway.oke_sg](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_service_gateway) | resource |
| [oci_core_subnet.api_endpoint_subnet](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_subnet) | resource |
| [oci_core_subnet.lb_subnet](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_subnet) | resource |
| [oci_core_subnet.node_subnet](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_subnet) | resource |
| [oci_core_vcn.oke_vcn](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/core_vcn) | resource |
| [oci_kms_key.oke_encryption_key](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/kms_key) | resource |
| [oci_kms_vault.oke_vault](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/kms_vault) | resource |
| [oci_logging_log.cluster_audit_log](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/logging_log) | resource |
| [oci_logging_log.vcn_flow_log](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/logging_log) | resource |
| [oci_logging_log_group.vcn_flow_logs](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/logging_log_group) | resource |
| [oci_objectstorage_bucket.backups](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/resources/objectstorage_bucket) | resource |
| [oci_core_images.node_image](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/data-sources/core_images) | data source |
| [oci_core_services.all_services](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/data-sources/core_services) | data source |
| [oci_identity_availability_domains.ads](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/data-sources/identity_availability_domains) | data source |
| [oci_objectstorage_namespace.ns](https://registry.terraform.io/providers/oracle/oci/9.8.0/docs/data-sources/objectstorage_namespace) | data source |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_compartment_id"></a> [compartment\_id](#input\_compartment\_id) | OCI compartment OCID for resources | `string` | n/a | yes |
| <a name="input_fingerprint"></a> [fingerprint](#input\_fingerprint) | OCI API key fingerprint | `string` | n/a | yes |
| <a name="input_tenancy_ocid"></a> [tenancy\_ocid](#input\_tenancy\_ocid) | OCI tenancy OCID | `string` | n/a | yes |
| <a name="input_user_ocid"></a> [user\_ocid](#input\_user\_ocid) | OCI user OCID | `string` | n/a | yes |
| <a name="input_allowed_ssh_cidrs"></a> [allowed\_ssh\_cidrs](#input\_allowed\_ssh\_cidrs) | CIDR blocks allowed to SSH to nodes (empty list = no SSH access) | `list(string)` | `[]` | no |
| <a name="input_audit_log_retention_days"></a> [audit\_log\_retention\_days](#input\_audit\_log\_retention\_days) | Number of days to retain audit logs | `number` | `90` | no |
| <a name="input_backup_retention_days"></a> [backup\_retention\_days](#input\_backup\_retention\_days) | Number of days to retain backups | `number` | `30` | no |
| <a name="input_backup_retention_lock_time"></a> [backup\_retention\_lock\_time](#input\_backup\_retention\_lock\_time) | Optional stable RFC3339 timestamp after which the backup retention rule becomes immutable | `string` | `null` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the OKE cluster | `string` | `"mantl-oke"` | no |
| <a name="input_enable_audit_logs"></a> [enable\_audit\_logs](#input\_enable\_audit\_logs) | Enable OKE cluster audit logs | `bool` | `true` | no |
| <a name="input_enable_backup_bucket"></a> [enable\_backup\_bucket](#input\_enable\_backup\_bucket) | Create Object Storage bucket for backups | `bool` | `true` | no |
| <a name="input_enable_flow_logs"></a> [enable\_flow\_logs](#input\_enable\_flow\_logs) | Enable VCN Flow Logs for security monitoring | `bool` | `true` | no |
| <a name="input_enable_image_policy"></a> [enable\_image\_policy](#input\_enable\_image\_policy) | Enable image signature verification | `bool` | `false` | no |
| <a name="input_enable_object_events"></a> [enable\_object\_events](#input\_enable\_object\_events) | Enable object storage events for backup bucket | `bool` | `false` | no |
| <a name="input_enable_private_endpoint"></a> [enable\_private\_endpoint](#input\_enable\_private\_endpoint) | Make API endpoint private (accessible only within VCN) | `bool` | `true` | no |
| <a name="input_enable_vault"></a> [enable\_vault](#input\_enable\_vault) | Enable OCI Vault for secrets encryption | `bool` | `true` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name (e.g., production, staging, development) | `string` | `"production"` | no |
| <a name="input_flow_logs_retention_days"></a> [flow\_logs\_retention\_days](#input\_flow\_logs\_retention\_days) | Number of days to retain flow logs | `number` | `90` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version for OKE cluster | `string` | `"v1.31.1"` | no |
| <a name="input_node_image_id"></a> [node\_image\_id](#input\_node\_image\_id) | OCID of the node image (Oracle Linux 8 recommended). Leave empty to use latest. | `string` | `""` | no |
| <a name="input_pod_cidr"></a> [pod\_cidr](#input\_pod\_cidr) | CIDR block for Kubernetes pods | `string` | `"10.244.0.0/16"` | no |
| <a name="input_private_key_path"></a> [private\_key\_path](#input\_private\_key\_path) | Path to OCI API private key | `string` | `"~/.oci/oci_api_key.pem"` | no |
| <a name="input_region"></a> [region](#input\_region) | OCI region (e.g., us-ashburn-1, us-phoenix-1, eu-frankfurt-1, ap-tokyo-1) | `string` | `"us-ashburn-1"` | no |
| <a name="input_service_cidr"></a> [service\_cidr](#input\_service\_cidr) | CIDR block for Kubernetes services | `string` | `"10.96.0.0/16"` | no |
| <a name="input_system_node_count"></a> [system\_node\_count](#input\_system\_node\_count) | Number of system nodes (across all availability domains) | `number` | `3` | no |
| <a name="input_system_node_memory_gb"></a> [system\_node\_memory\_gb](#input\_system\_node\_memory\_gb) | Memory in GB per system node (for Flex shapes) | `number` | `16` | no |
| <a name="input_system_node_ocpus"></a> [system\_node\_ocpus](#input\_system\_node\_ocpus) | Number of OCPUs per system node (for Flex shapes) | `number` | `4` | no |
| <a name="input_system_node_shape"></a> [system\_node\_shape](#input\_system\_node\_shape) | OCI compute shape for system nodes (e.g., VM.Standard.E4.Flex) | `string` | `"VM.Standard.E4.Flex"` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Additional tags for all resources | `map(string)` | `{}` | no |
| <a name="input_use_flannel_cni"></a> [use\_flannel\_cni](#input\_use\_flannel\_cni) | Use Flannel CNI (overlay) instead of OCI VCN-native pod networking | `bool` | `false` | no |
| <a name="input_vcn_cidr"></a> [vcn\_cidr](#input\_vcn\_cidr) | CIDR block for the VCN | `string` | `"10.0.0.0/16"` | no |
| <a name="input_workload_initial_node_count"></a> [workload\_initial\_node\_count](#input\_workload\_initial\_node\_count) | Initial number of workload nodes (across all availability domains) | `number` | `3` | no |
| <a name="input_workload_node_memory_gb"></a> [workload\_node\_memory\_gb](#input\_workload\_node\_memory\_gb) | Memory in GB per workload node (for Flex shapes) | `number` | `16` | no |
| <a name="input_workload_node_ocpus"></a> [workload\_node\_ocpus](#input\_workload\_node\_ocpus) | Number of OCPUs per workload node (for Flex shapes) | `number` | `4` | no |
| <a name="input_workload_node_shape"></a> [workload\_node\_shape](#input\_workload\_node\_shape) | OCI compute shape for workload nodes | `string` | `"VM.Standard.E4.Flex"` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_api_endpoint"></a> [api\_endpoint](#output\_api\_endpoint) | The endpoint of the Kubernetes API server |
| <a name="output_backup_bucket"></a> [backup\_bucket](#output\_backup\_bucket) | The name of the backup bucket |
| <a name="output_backup_bucket_namespace"></a> [backup\_bucket\_namespace](#output\_backup\_bucket\_namespace) | The namespace of the backup bucket |
| <a name="output_cluster_endpoint"></a> [cluster\_endpoint](#output\_cluster\_endpoint) | Kubernetes API endpoint |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | The ID of the OKE cluster |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | The name of the OKE cluster |
| <a name="output_encryption_key_id"></a> [encryption\_key\_id](#output\_encryption\_key\_id) | The ID of the encryption key |
| <a name="output_flow_log_group_id"></a> [flow\_log\_group\_id](#output\_flow\_log\_group\_id) | The ID of the VCN flow logs log group |
| <a name="output_kubeconfig_command"></a> [kubeconfig\_command](#output\_kubeconfig\_command) | Command to generate kubeconfig |
| <a name="output_kubernetes_version"></a> [kubernetes\_version](#output\_kubernetes\_version) | The Kubernetes version of the cluster |
| <a name="output_private_endpoint"></a> [private\_endpoint](#output\_private\_endpoint) | The private endpoint of the Kubernetes API server |
| <a name="output_region"></a> [region](#output\_region) | OCI region |
| <a name="output_system_node_pool_id"></a> [system\_node\_pool\_id](#output\_system\_node\_pool\_id) | OCID of the system node pool |
| <a name="output_vault_id"></a> [vault\_id](#output\_vault\_id) | The ID of the OCI Vault |
| <a name="output_vcn_id"></a> [vcn\_id](#output\_vcn\_id) | The ID of the VCN |
| <a name="output_workload_node_pool_id"></a> [workload\_node\_pool\_id](#output\_workload\_node\_pool\_id) | OCID of the workload node pool |
<!-- END_TF_DOCS -->
