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
