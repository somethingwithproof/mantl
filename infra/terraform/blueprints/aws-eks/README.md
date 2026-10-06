<!-- SPDX-License-Identifier: Apache-2.0 -->
# Mantl AWS EKS Blueprint

Beta Amazon EKS blueprint with IRSA roles for Mantl platform components. Static validation does not establish production readiness; deployment requires cloud acceptance evidence.

## Features

- **VPC**: 3-AZ deployment with public, private, and intra subnets
- **EKS**: Managed Kubernetes with encrypted secrets (KMS)
- **Node Groups**: Separate system and workload node groups
- **IRSA Roles**: Pre-configured for External DNS, External Secrets, cert-manager, Cluster Autoscaler, EBS CSI, and AWS Load Balancer Controller
- **Logging**: CloudWatch logging for API, audit, authenticator, controller manager, and scheduler

## Prerequisites

- AWS CLI configured with appropriate credentials
- Terraform >= 1.6
- kubectl >= 1.28

## Upgrading existing deployments

This blueprint uses EKS module 21, IAM module 6, VPC module 6, KMS module 4,
and AWS provider 6. Kubernetes and Helm providers remain on their supported
major version 2. The public blueprint variables and output names are unchanged.

Before upgrading an existing deployment, back up Terraform state and inspect a
plan against that deployment. The IAM module consolidates its policy resources. All six IRSA modules retain
their role resource address (`aws_iam_role.this[0]`) and set
`use_name_prefix = false` to preserve their role names and ARNs. Review the EKS node AMI,
IMDS hop limit, OIDC issuer, monitoring, and add-on changes for your workloads.
IRSA trust remains limited to the existing namespace/service-account pairs.
No state migration or cloud apply is performed by local validation or CI.

### IRSA policy state migration

`migrations.tf` declares both policy and policy-attachment address moves for
all six modules. Terraform includes these moves in a reviewed plan; no separate
`terraform state mv` commands or role-state removal are required. For each row,
`module.<module>.aws_iam_policy.<old>[0]` moves to
`module.<module>.aws_iam_policy.this[0]`, and the matching
`aws_iam_role_policy_attachment.<old>[0]` moves to
`aws_iam_role_policy_attachment.this[0]`.

| Module | Old policy and attachment resource name |
| --- | --- |
| `external_dns_irsa` | `external_dns` |
| `external_secrets_irsa` | `external_secrets` |
| `cert_manager_irsa` | `cert_manager` |
| `cluster_autoscaler_irsa` | `cluster_autoscaler` |
| `ebs_csi_irsa` | `ebs_csi` |
| `load_balancer_controller_irsa` | `load_balancer_controller` |

Back up the state for the selected backend and workspace before planning. Do
not remove or import any of the six role resources: their addresses stay the
same. Stop if the plan replaces an IRSA role or changes a role name/ARN.
Managed policies move to unique, fixed cluster/component names, replacing the
old generated policy names; review their replacements, attachment ordering and
permission changes in a maintenance window. State-address moves alone do not
prevent policy replacements or establish uninterrupted access. Review the
EKS OIDC-provider changes with the corresponding trust-policy updates.

- [EKS 21 migration guide](https://github.com/terraform-aws-modules/terraform-aws-eks/blob/v21.26.0/docs/UPGRADE-21.0.md)
- [IAM 6 migration guide](https://github.com/terraform-aws-modules/terraform-aws-iam/blob/v6.8.0/docs/UPGRADE-6.0.md)

## Quick Start

```bash
# Copy example configuration
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
vim terraform.tfvars

# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply
terraform apply

# Configure kubectl
$(terraform output -raw configure_kubectl)
```

## Deployment Time

- Initial deployment: ~20-25 minutes
- Destroy: ~15-20 minutes

## Outputs

After deployment, use these outputs to configure the Mantl platform:

```bash
# Get IRSA role ARNs for platform components
terraform output platform_irsa_annotations

# Get cluster endpoint
terraform output cluster_endpoint

# Configure kubectl
$(terraform output -raw configure_kubectl)
```

## Platform Integration

After deploying the cluster, configure the platform overlays with the IRSA role ARNs:

### 1. External DNS

Create `clusters/production/overlays/aws/external-dns-patch.yaml`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns
  namespace: external-dns
  annotations:
    eks.amazonaws.com/role-arn: <external_dns_role_arn from terraform output>
```

### 2. External Secrets

Create `clusters/production/overlays/aws/external-secrets-patch.yaml`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-secrets
  namespace: external-secrets
  annotations:
    eks.amazonaws.com/role-arn: <external_secrets_role_arn from terraform output>
```

### 3. cert-manager

Create `clusters/production/overlays/aws/cert-manager-patch.yaml`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cert-manager
  namespace: cert-manager
  annotations:
    eks.amazonaws.com/role-arn: <cert_manager_role_arn from terraform output>
```

## Cost Optimization

For non-production environments:

```hcl
# Use single NAT Gateway
single_nat_gateway = true

# Use Spot instances for workloads
workload_node_capacity_type = "SPOT"

# Reduce node counts
system_node_desired_size   = 1
workload_node_desired_size = 2
```

## Security Considerations

- Secrets are encrypted at rest with KMS
- Node groups run in private subnets
- IRSA provides pod-level IAM permissions (no node-level access)
- Control plane logging enabled for audit trail

## Cleanup

```bash
# Destroy all resources
terraform destroy
```

**Warning**: This will delete all resources including the VPC. Ensure no other resources depend on this VPC.

## Troubleshooting

### Nodes not joining cluster

Check VPC CNI addon status:

```bash
kubectl get pods -n kube-system -l k8s-app=aws-node
```

### IRSA not working

Verify ServiceAccount annotations:

```bash
kubectl get sa external-dns -n external-dns -o yaml | grep annotations -A5
```

Check OIDC provider:

```bash
aws eks describe-cluster --name mantl --query cluster.identity.oidc.issuer
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 6.67.0 |
| <a name="requirement_helm"></a> [helm](#requirement\_helm) | ~> 3.3.0 |
| <a name="requirement_kubernetes"></a> [kubernetes](#requirement\_kubernetes) | ~> 3.3.0 |
| <a name="requirement_tls"></a> [tls](#requirement\_tls) | ~> 4.4.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_aws"></a> [aws](#provider\_aws) | 6.67.0 |

## Modules

| Name | Source | Version |
| ---- | ------ | ------- |
| <a name="module_cert_manager_irsa"></a> [cert\_manager\_irsa](#module\_cert\_manager\_irsa) | terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts | ~> 6.8 |
| <a name="module_cluster_autoscaler_irsa"></a> [cluster\_autoscaler\_irsa](#module\_cluster\_autoscaler\_irsa) | terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts | ~> 6.8 |
| <a name="module_ebs_csi_irsa"></a> [ebs\_csi\_irsa](#module\_ebs\_csi\_irsa) | terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts | ~> 6.8 |
| <a name="module_eks"></a> [eks](#module\_eks) | terraform-aws-modules/eks/aws | ~> 21.26 |
| <a name="module_external_dns_irsa"></a> [external\_dns\_irsa](#module\_external\_dns\_irsa) | terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts | ~> 6.8 |
| <a name="module_external_secrets_irsa"></a> [external\_secrets\_irsa](#module\_external\_secrets\_irsa) | terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts | ~> 6.8 |
| <a name="module_kms"></a> [kms](#module\_kms) | terraform-aws-modules/kms/aws | ~> 4.2 |
| <a name="module_load_balancer_controller_irsa"></a> [load\_balancer\_controller\_irsa](#module\_load\_balancer\_controller\_irsa) | terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts | ~> 6.8 |
| <a name="module_vpc"></a> [vpc](#module\_vpc) | terraform-aws-modules/vpc/aws | ~> 6.7 |

## Resources

| Name | Type |
| ---- | ---- |
| [aws_cloudwatch_log_group.vpc_flow_logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group) | resource |
| [aws_flow_log.vpc](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/flow_log) | resource |
| [aws_iam_role.vpc_flow_logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy.vpc_flow_logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_kms_key.vpc_flow_logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_availability_zones.available](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/availability_zones) | data source |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_cluster_endpoint_public_access"></a> [cluster\_endpoint\_public\_access](#input\_cluster\_endpoint\_public\_access) | Enable public access to cluster API endpoint (NOT recommended for production) | `bool` | `false` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the EKS cluster | `string` | `"mantl"` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name (e.g., production, staging, development) | `string` | `"production"` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version for EKS cluster | `string` | `"1.31"` | no |
| <a name="input_region"></a> [region](#input\_region) | AWS region | `string` | `"us-east-1"` | no |
| <a name="input_route53_zone_arns"></a> [route53\_zone\_arns](#input\_route53\_zone\_arns) | ARNs of Route53 hosted zones for External DNS and cert-manager | `list(string)` | `[]` | no |
| <a name="input_secrets_manager_arns"></a> [secrets\_manager\_arns](#input\_secrets\_manager\_arns) | ARNs of Secrets Manager secrets for External Secrets | `list(string)` | `[]` | no |
| <a name="input_single_nat_gateway"></a> [single\_nat\_gateway](#input\_single\_nat\_gateway) | Use a single NAT Gateway (cost savings for non-prod) | `bool` | `false` | no |
| <a name="input_ssm_parameter_arns"></a> [ssm\_parameter\_arns](#input\_ssm\_parameter\_arns) | ARNs of SSM parameters for External Secrets | `list(string)` | `[]` | no |
| <a name="input_system_node_desired_size"></a> [system\_node\_desired\_size](#input\_system\_node\_desired\_size) | Desired number of system nodes | `number` | `2` | no |
| <a name="input_system_node_instance_types"></a> [system\_node\_instance\_types](#input\_system\_node\_instance\_types) | Instance types for system node group | `list(string)` | <pre>[<br/>  "m6i.large",<br/>  "m5.large"<br/>]</pre> | no |
| <a name="input_system_node_max_size"></a> [system\_node\_max\_size](#input\_system\_node\_max\_size) | Maximum number of system nodes | `number` | `4` | no |
| <a name="input_system_node_min_size"></a> [system\_node\_min\_size](#input\_system\_node\_min\_size) | Minimum number of system nodes | `number` | `2` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Additional tags for all resources | `map(string)` | `{}` | no |
| <a name="input_vpc_cidr"></a> [vpc\_cidr](#input\_vpc\_cidr) | CIDR block for VPC | `string` | `"10.0.0.0/16"` | no |
| <a name="input_workload_node_capacity_type"></a> [workload\_node\_capacity\_type](#input\_workload\_node\_capacity\_type) | Capacity type for workload nodes (ON\_DEMAND or SPOT) | `string` | `"ON_DEMAND"` | no |
| <a name="input_workload_node_desired_size"></a> [workload\_node\_desired\_size](#input\_workload\_node\_desired\_size) | Desired number of workload nodes | `number` | `3` | no |
| <a name="input_workload_node_instance_types"></a> [workload\_node\_instance\_types](#input\_workload\_node\_instance\_types) | Instance types for workload node group | `list(string)` | <pre>[<br/>  "m6i.xlarge",<br/>  "m5.xlarge"<br/>]</pre> | no |
| <a name="input_workload_node_max_size"></a> [workload\_node\_max\_size](#input\_workload\_node\_max\_size) | Maximum number of workload nodes | `number` | `10` | no |
| <a name="input_workload_node_min_size"></a> [workload\_node\_min\_size](#input\_workload\_node\_min\_size) | Minimum number of workload nodes | `number` | `2` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_cert_manager_role_arn"></a> [cert\_manager\_role\_arn](#output\_cert\_manager\_role\_arn) | IAM role ARN for cert-manager |
| <a name="output_cluster_autoscaler_role_arn"></a> [cluster\_autoscaler\_role\_arn](#output\_cluster\_autoscaler\_role\_arn) | IAM role ARN for Cluster Autoscaler |
| <a name="output_cluster_certificate_authority_data"></a> [cluster\_certificate\_authority\_data](#output\_cluster\_certificate\_authority\_data) | Base64 encoded cluster CA certificate |
| <a name="output_cluster_endpoint"></a> [cluster\_endpoint](#output\_cluster\_endpoint) | EKS cluster API endpoint |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | EKS cluster ID |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | EKS cluster name |
| <a name="output_cluster_oidc_issuer_url"></a> [cluster\_oidc\_issuer\_url](#output\_cluster\_oidc\_issuer\_url) | OIDC issuer URL for the cluster |
| <a name="output_cluster_oidc_provider_arn"></a> [cluster\_oidc\_provider\_arn](#output\_cluster\_oidc\_provider\_arn) | OIDC provider ARN for IRSA |
| <a name="output_cluster_version"></a> [cluster\_version](#output\_cluster\_version) | Kubernetes version of the cluster |
| <a name="output_configure_kubectl"></a> [configure\_kubectl](#output\_configure\_kubectl) | Command to configure kubectl |
| <a name="output_ebs_csi_role_arn"></a> [ebs\_csi\_role\_arn](#output\_ebs\_csi\_role\_arn) | IAM role ARN for EBS CSI Driver |
| <a name="output_external_dns_role_arn"></a> [external\_dns\_role\_arn](#output\_external\_dns\_role\_arn) | IAM role ARN for External DNS |
| <a name="output_external_secrets_role_arn"></a> [external\_secrets\_role\_arn](#output\_external\_secrets\_role\_arn) | IAM role ARN for External Secrets |
| <a name="output_load_balancer_controller_role_arn"></a> [load\_balancer\_controller\_role\_arn](#output\_load\_balancer\_controller\_role\_arn) | IAM role ARN for AWS Load Balancer Controller |
| <a name="output_platform_contract"></a> [platform\_contract](#output\_platform\_contract) | Provider-neutral platform identity; deployment verification is separate |
| <a name="output_platform_irsa_annotations"></a> [platform\_irsa\_annotations](#output\_platform\_irsa\_annotations) | ServiceAccount annotations for platform components |
| <a name="output_private_subnet_ids"></a> [private\_subnet\_ids](#output\_private\_subnet\_ids) | Private subnet IDs |
| <a name="output_public_subnet_ids"></a> [public\_subnet\_ids](#output\_public\_subnet\_ids) | Public subnet IDs |
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | VPC ID |
<!-- END_TF_DOCS -->
