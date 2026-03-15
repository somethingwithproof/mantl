# Mantl AWS EKS Blueprint

Production-ready Amazon EKS cluster with IRSA roles for Mantl platform components.

## Features

- **VPC**: 3-AZ deployment with public, private, and intra subnets
- **EKS**: Managed Kubernetes with encrypted secrets (KMS)
- **Node Groups**: Separate system and workload node groups
- **IRSA Roles**: Pre-configured for External DNS, External Secrets, cert-manager, Cluster Autoscaler, EBS CSI, and AWS Load Balancer Controller
- **Logging**: CloudWatch logging for API, audit, authenticator, controller manager, and scheduler

## Prerequisites

- AWS CLI configured with appropriate credentials
- Terraform >= 1.5.0
- kubectl >= 1.28

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
