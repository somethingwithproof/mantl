# IBM Cloud - IKS (IBM Kubernetes Service) Blueprint

Terraform blueprint for deploying IBM Kubernetes Service with Mantl platform.

## Prerequisites

1. **IBM Cloud Account**: Active IBM Cloud account
2. **IBM Cloud CLI**: Install from https://cloud.ibm.com/docs/cli
3. **API Key**: Create an IBM Cloud API key
   ```bash
   ibmcloud login
   ibmcloud iam api-key-create mantl-key -d "Mantl platform key" --file ~/ibm-api-key.json
   ```
4. **Kubectl Plugin**: Install IKS plugin
   ```bash
   ibmcloud plugin install kubernetes-service
   ```

## Configuration

Create `terraform.tfvars`:

```hcl
ibmcloud_api_key   = "your-api-key-here"
region             = "us-south"
resource_group     = "default"
cluster_name       = "mantl-iks"
kubernetes_version = "1.31"
zones_count        = 3
worker_flavor      = "bx2.4x16"
workers_per_zone   = 1
tags               = ["mantl", "production"]
```

## Available Regions

- `us-south` (Dallas, TX)
- `us-east` (Washington, DC)
- `eu-de` (Frankfurt, Germany)
- `eu-gb` (London, UK)
- `jp-tok` (Tokyo, Japan)
- `jp-osa` (Osaka, Japan)
- `au-syd` (Sydney, Australia)
- `ca-tor` (Toronto, Canada)
- `br-sao` (São Paulo, Brazil)

## Worker Node Flavors

### Balanced (bx2) - **Recommended**
- `bx2.2x8` - 2 vCPU, 8 GB RAM
- `bx2.4x16` - 4 vCPU, 16 GB RAM (default)
- `bx2.8x32` - 8 vCPU, 32 GB RAM
- `bx2.16x64` - 16 vCPU, 64 GB RAM
- `bx2.32x128` - 32 vCPU, 128 GB RAM

### Compute (cx2) - CPU optimized
- `cx2.2x4` - 2 vCPU, 4 GB RAM
- `cx2.4x8` - 4 vCPU, 8 GB RAM
- `cx2.8x16` - 8 vCPU, 16 GB RAM

### Memory (mx2) - Memory optimized
- `mx2.2x16` - 2 vCPU, 16 GB RAM
- `mx2.4x32` - 4 vCPU, 32 GB RAM
- `mx2.8x64` - 8 vCPU, 64 GB RAM

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
ibmcloud login --apikey <your-key>
ibmcloud ks cluster config --cluster <cluster-name> --admin

# Verify
kubectl get nodes
```

## Workload Identity

IKS uses **IBM Cloud IAM** for service-to-service authentication.

### Enable IAM for ExternalDNS

1. Create a service ID:
   ```bash
   ibmcloud iam service-id-create external-dns-sa --description "ExternalDNS service account"
   ```

2. Create API key for service ID:
   ```bash
   ibmcloud iam service-api-key-create external-dns-key external-dns-sa
   ```

3. Assign DNS permissions:
   ```bash
   ibmcloud iam service-policy-create external-dns-sa \
     --roles Manager --service-name internet-svcs
   ```

4. Create Kubernetes secret:
   ```bash
   kubectl create secret generic external-dns-secret \
     --from-literal=api-key=<service-id-api-key> \
     -n external-dns
   ```

## Network Architecture

- **VPC**: Automatically created
- **Subnets**: One per zone (256 IPs each)
- **Public Gateways**: One per zone for outbound internet access
- **Pod CIDR**: 172.30.0.0/16 (default)
- **Service CIDR**: 172.21.0.0/16 (default)

## High Availability

Deploy across 3 zones for HA:

```hcl
zones_count      = 3
workers_per_zone = 2  # 6 total workers
```

## Costs

Approximate monthly costs (us-south):
- **Free Tier**: 1 free cluster (single zone, 2 vCPU, 4 GB RAM)
- **3x bx2.4x16 workers**: ~$280/month
- **VPC networking**: ~$10-20/month
- **Load Balancers**: ~$50-100/month per LB

**Cost Optimization**:
- Use `bx2.2x8` for dev/test: ~$70/month per worker
- Single zone deployment reduces costs by 66%
- Use IBM Cloud DNS (free tier available)

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

## IBM Cloud Kubernetes Features

- **Built-in Ingress ALB**: IBM Cloud Application Load Balancer (free)
- **Managed Add-ons**: Istio, Knative, Log Analysis
- **Cluster Autoscaler**: Automatic node scaling
- **VPC Integration**: Native VPC Gen 2 support
- **Portworx**: Enterprise storage solution (optional)

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_ibm"></a> [ibm](#requirement\_ibm) | 2.6.2 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_ibm"></a> [ibm](#provider\_ibm) | 2.6.2 |

## Modules

No modules.

## Resources

| Name | Type |
| ---- | ---- |
| [ibm_container_vpc_cluster.iks_cluster](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/container_vpc_cluster) | resource |
| [ibm_container_vpc_worker_pool.spot_pool](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/container_vpc_worker_pool) | resource |
| [ibm_container_vpc_worker_pool.workload_pool](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/container_vpc_worker_pool) | resource |
| [ibm_cos_bucket.backups](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/cos_bucket) | resource |
| [ibm_cos_bucket.flow_logs](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/cos_bucket) | resource |
| [ibm_is_flow_log.vpc_flow_logs](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_flow_log) | resource |
| [ibm_is_public_gateway.pgw](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_public_gateway) | resource |
| [ibm_is_security_group.cluster_nodes](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_security_group) | resource |
| [ibm_is_security_group_rule.inbound_http](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_security_group_rule) | resource |
| [ibm_is_security_group_rule.inbound_https](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_security_group_rule) | resource |
| [ibm_is_security_group_rule.inbound_ssh](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_security_group_rule) | resource |
| [ibm_is_security_group_rule.inbound_vpc](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_security_group_rule) | resource |
| [ibm_is_security_group_rule.outbound_all](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_security_group_rule) | resource |
| [ibm_is_subnet.iks_subnet](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_subnet) | resource |
| [ibm_is_vpc.iks_vpc](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_vpc) | resource |
| [ibm_is_vpc_address_prefix.zone_prefix](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/is_vpc_address_prefix) | resource |
| [ibm_kms_key.root_key](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/kms_key) | resource |
| [ibm_resource_instance.backup_cos](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/resource_instance) | resource |
| [ibm_resource_instance.cos_instance](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/resource_instance) | resource |
| [ibm_resource_instance.key_protect](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/resource_instance) | resource |
| [ibm_resource_instance.log_analysis](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/resource_instance) | resource |
| [ibm_resource_instance.monitoring](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/resources/resource_instance) | resource |
| [ibm_is_zones.zones](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/data-sources/is_zones) | data source |
| [ibm_resource_group.resource_group](https://registry.terraform.io/providers/IBM-Cloud/ibm/2.6.2/docs/data-sources/resource_group) | data source |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_allowed_ssh_cidrs"></a> [allowed\_ssh\_cidrs](#input\_allowed\_ssh\_cidrs) | CIDR blocks allowed to SSH to nodes | `list(string)` | `[]` | no |
| <a name="input_backup_retention_days"></a> [backup\_retention\_days](#input\_backup\_retention\_days) | Number of days to retain backups | `number` | `30` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the IKS cluster | `string` | `"mantl-iks"` | no |
| <a name="input_disable_public_endpoint"></a> [disable\_public\_endpoint](#input\_disable\_public\_endpoint) | Disable public service endpoint (private only) | `bool` | `true` | no |
| <a name="input_enable_backup_bucket"></a> [enable\_backup\_bucket](#input\_enable\_backup\_bucket) | Create Cloud Object Storage bucket for backups (S3-compatible) | `bool` | `true` | no |
| <a name="input_enable_flow_logs"></a> [enable\_flow\_logs](#input\_enable\_flow\_logs) | Enable VPC Flow Logs for security monitoring | `bool` | `true` | no |
| <a name="input_enable_key_protect"></a> [enable\_key\_protect](#input\_enable\_key\_protect) | Enable IBM Key Protect for secrets encryption | `bool` | `true` | no |
| <a name="input_enable_log_analysis"></a> [enable\_log\_analysis](#input\_enable\_log\_analysis) | Create IBM Log Analysis instance for cluster logging | `bool` | `true` | no |
| <a name="input_enable_monitoring"></a> [enable\_monitoring](#input\_enable\_monitoring) | Create IBM Cloud Monitoring instance (Sysdig) | `bool` | `true` | no |
| <a name="input_enable_public_gateway"></a> [enable\_public\_gateway](#input\_enable\_public\_gateway) | Create public gateways for subnets (for internet access) | `bool` | `true` | no |
| <a name="input_enable_spot_pool"></a> [enable\_spot\_pool](#input\_enable\_spot\_pool) | Enable spot instances worker pool for cost savings | `bool` | `false` | no |
| <a name="input_entitlement"></a> [entitlement](#input\_entitlement) | Entitlement for OCP on IBM Cloud (cloud\_pak for OCP) | `string` | `null` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name (e.g., production, staging, development) | `string` | `"production"` | no |
| <a name="input_flow_logs_retention_days"></a> [flow\_logs\_retention\_days](#input\_flow\_logs\_retention\_days) | Number of days to retain flow logs | `number` | `90` | no |
| <a name="input_ibmcloud_api_key"></a> [ibmcloud\_api\_key](#input\_ibmcloud\_api\_key) | IBM Cloud API key | `string` | n/a | yes |
| <a name="input_key_protect_force_delete"></a> [key\_protect\_force\_delete](#input\_key\_protect\_force\_delete) | Force delete Key Protect keys on destroy | `bool` | `false` | no |
| <a name="input_key_protect_private_endpoint"></a> [key\_protect\_private\_endpoint](#input\_key\_protect\_private\_endpoint) | Use private endpoint for Key Protect | `bool` | `true` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version for IKS cluster | `string` | `"1.31"` | no |
| <a name="input_log_analysis_plan"></a> [log\_analysis\_plan](#input\_log\_analysis\_plan) | IBM Log Analysis plan (7-day, 14-day, 30-day, lite) | `string` | `"7-day"` | no |
| <a name="input_monitoring_plan"></a> [monitoring\_plan](#input\_monitoring\_plan) | IBM Cloud Monitoring plan (graduated-tier, lite) | `string` | `"graduated-tier"` | no |
| <a name="input_region"></a> [region](#input\_region) | IBM Cloud region (e.g., us-south, us-east, eu-de, eu-gb, jp-tok, au-syd) | `string` | `"us-south"` | no |
| <a name="input_resource_group"></a> [resource\_group](#input\_resource\_group) | IBM Cloud resource group name | `string` | `"default"` | no |
| <a name="input_spot_worker_flavor"></a> [spot\_worker\_flavor](#input\_spot\_worker\_flavor) | Worker node flavor for spot pool | `string` | `"bx2.4x16"` | no |
| <a name="input_spot_workers_per_zone"></a> [spot\_workers\_per\_zone](#input\_spot\_workers\_per\_zone) | Number of spot workers per zone | `number` | `1` | no |
| <a name="input_system_worker_flavor"></a> [system\_worker\_flavor](#input\_system\_worker\_flavor) | Worker node flavor for system pool (e.g., bx2.4x16 = 4 vCPU, 16 GB RAM) | `string` | `"bx2.4x16"` | no |
| <a name="input_system_workers_per_zone"></a> [system\_workers\_per\_zone](#input\_system\_workers\_per\_zone) | Number of system workers per zone | `number` | `1` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags to apply to resources | `list(string)` | <pre>[<br/>  "mantl",<br/>  "kubernetes"<br/>]</pre> | no |
| <a name="input_taint_system_pool"></a> [taint\_system\_pool](#input\_taint\_system\_pool) | Apply CriticalAddonsOnly taint to system pool | `bool` | `true` | no |
| <a name="input_vpc_cidr"></a> [vpc\_cidr](#input\_vpc\_cidr) | CIDR block for the VPC | `string` | `"10.240.0.0/16"` | no |
| <a name="input_workload_worker_flavor"></a> [workload\_worker\_flavor](#input\_workload\_worker\_flavor) | Worker node flavor for workload pool | `string` | `"bx2.4x16"` | no |
| <a name="input_workload_workers_per_zone"></a> [workload\_workers\_per\_zone](#input\_workload\_workers\_per\_zone) | Number of workload workers per zone | `number` | `1` | no |
| <a name="input_zones_count"></a> [zones\_count](#input\_zones\_count) | Number of zones to deploy to (1-3) | `number` | `3` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_backup_bucket"></a> [backup\_bucket](#output\_backup\_bucket) | The name of the backup bucket |
| <a name="output_backup_cos_instance_id"></a> [backup\_cos\_instance\_id](#output\_backup\_cos\_instance\_id) | The ID of the backup COS instance |
| <a name="output_cluster_endpoint"></a> [cluster\_endpoint](#output\_cluster\_endpoint) | Public service endpoint for the cluster |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | The ID of the IKS cluster |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | The name of the IKS cluster |
| <a name="output_flow_logs_bucket"></a> [flow\_logs\_bucket](#output\_flow\_logs\_bucket) | The name of the VPC flow logs bucket |
| <a name="output_ingress_hostname"></a> [ingress\_hostname](#output\_ingress\_hostname) | The Ingress hostname for the cluster |
| <a name="output_ingress_secret"></a> [ingress\_secret](#output\_ingress\_secret) | The Ingress secret name |
| <a name="output_key_protect_instance_id"></a> [key\_protect\_instance\_id](#output\_key\_protect\_instance\_id) | The GUID of the Key Protect instance |
| <a name="output_kube_version"></a> [kube\_version](#output\_kube\_version) | Kubernetes version |
| <a name="output_kubeconfig_command"></a> [kubeconfig\_command](#output\_kubeconfig\_command) | Command to configure kubectl |
| <a name="output_log_analysis_instance_id"></a> [log\_analysis\_instance\_id](#output\_log\_analysis\_instance\_id) | The ID of the Log Analysis instance |
| <a name="output_master_url"></a> [master\_url](#output\_master\_url) | The URL of the cluster master |
| <a name="output_monitoring_instance_id"></a> [monitoring\_instance\_id](#output\_monitoring\_instance\_id) | The ID of the Monitoring instance |
| <a name="output_resource_group_id"></a> [resource\_group\_id](#output\_resource\_group\_id) | Resource group ID |
| <a name="output_root_key_id"></a> [root\_key\_id](#output\_root\_key\_id) | The ID of the root encryption key |
| <a name="output_security_group_id"></a> [security\_group\_id](#output\_security\_group\_id) | The ID of the cluster security group |
| <a name="output_subnet_ids"></a> [subnet\_ids](#output\_subnet\_ids) | IDs of created subnets |
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | The ID of the VPC |
<!-- END_TF_DOCS -->
