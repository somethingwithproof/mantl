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
