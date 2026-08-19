# DigitalOcean Kubernetes (DOKS) Blueprint

Production-ready Kubernetes cluster on DigitalOcean with VPC isolation, auto-scaling, monitoring, and enterprise security features.

## Overview

This Terraform blueprint provisions a complete Kubernetes platform on DigitalOcean, including:

- **VPC Isolation**: Dedicated VPC with custom CIDR for network isolation
- **Multi-Node Pools**: System (platform), Workload (apps), and optional Spot instances
- **Auto-Scaling**: Automatic node scaling based on workload demands
- **Container Registry**: Integrated DigitalOcean Container Registry
- **Load Balancing**: Optional DigitalOcean Load Balancer for ingress
- **Managed Database**: Optional PostgreSQL cluster for platform services
- **Backup Storage**: S3-compatible Spaces bucket with lifecycle policies
- **Monitoring & Alerts**: CPU, memory, and disk alerts via DigitalOcean Monitoring
- **Security Features**: Firewall rules, Cilium CNI, Sealed Secrets, Falco runtime security

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    DigitalOcean Region                      │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                VPC (10.10.0.0/16)                     │ │
│  │                                                       │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │ │
│  │  │ System Pool  │  │ Workload Pool│  │ Spot Pool  │ │ │
│  │  │ (3 nodes)    │  │ (2-10 nodes) │  │ (optional) │ │ │
│  │  └──────────────┘  └──────────────┘  └────────────┘ │ │
│  │                                                       │ │
│  │  ┌──────────────────────────────────────────────────┐│ │
│  │  │         Kubernetes Cluster (DOKS)                ││ │
│  │  │  - Cilium CNI + Hubble (Flow Logs)              ││ │
│  │  │  - Sealed Secrets (KMS Alternative)             ││ │
│  │  │  - Falco (Runtime Security)                     ││ │
│  │  │  - cert-manager (TLS Automation)                ││ │
│  │  └──────────────────────────────────────────────────┘│ │
│  └───────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌───────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ Load Balancer │  │ PostgreSQL   │  │ Container       │ │
│  │ (optional)    │  │ (optional)   │  │ Registry        │ │
│  └───────────────┘  └──────────────┘  └─────────────────┘ │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Spaces Bucket (S3-compatible backups)               │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Prerequisites

1. **DigitalOcean Account** with API access
2. **DigitalOcean API Token** with read/write permissions
3. **Terraform** >= 1.10.0
4. **doctl** (DigitalOcean CLI) for post-deployment operations

### Install doctl

```bash
# macOS
brew install doctl

# Linux
cd ~
wget https://github.com/digitalocean/doctl/releases/download/v1.104.0/doctl-1.104.0-linux-amd64.tar.gz
tar xf doctl-1.104.0-linux-amd64.tar.gz
sudo mv doctl /usr/local/bin
```

### Authenticate doctl

```bash
doctl auth init
# Enter your DigitalOcean API token
```

## Quick Start

### 1. Set Environment Variables

```bash
export DIGITALOCEAN_TOKEN="your-api-token-here"
export TF_VAR_cluster_name="mantl-dev"
export TF_VAR_environment="dev"
```

### 2. Initialize Terraform

```bash
cd infra/terraform/blueprints/do-doks
terraform init
```

### 3. Create Configuration

```bash
# Copy example configuration
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
vim terraform.tfvars
```

### 4. Plan and Apply

```bash
# Review planned changes
terraform plan

# Apply configuration
terraform apply
```

### 5. Configure kubectl

```bash
# Download kubeconfig
doctl kubernetes cluster kubeconfig save $(terraform output -raw cluster_name)

# Verify access
kubectl get nodes
```

## Configuration

### Variables Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| **Core Configuration** |
| `cluster_name` | string | `"mantl-doks"` | Name of the Kubernetes cluster |
| `region` | string | `"nyc3"` | DigitalOcean region (nyc1, nyc3, sfo3, etc.) |
| `environment` | string | `"production"` | Environment name (production, staging, dev) |
| `kubernetes_version` | string | `"1.31.3-do.0"` | Kubernetes version |
| `tags` | list(string) | `["mantl", "kubernetes"]` | Additional resource tags |
| **Networking** |
| `vpc_cidr` | string | `"10.10.0.0/16"` | VPC CIDR block |
| `allowed_ssh_cidrs` | list(string) | `["10.0.0.0/8"]` | SSH allowed CIDR blocks |
| **System Node Pool** |
| `system_node_size` | string | `"s-4vcpu-8gb"` | System node droplet size |
| `system_node_count` | number | `3` | Number of system nodes (fixed) |
| **Workload Node Pool** |
| `workload_node_size` | string | `"s-4vcpu-8gb"` | Workload node droplet size |
| `workload_min_nodes` | number | `2` | Minimum workload nodes (autoscale) |
| `workload_max_nodes` | number | `10` | Maximum workload nodes (autoscale) |
| **Spot Pool (Optional)** |
| `enable_spot_pool` | bool | `false` | Enable spot instances for cost savings |
| `spot_node_size` | string | `"s-4vcpu-8gb"` | Spot node droplet size |
| `spot_min_nodes` | number | `0` | Minimum spot nodes |
| `spot_max_nodes` | number | `5` | Maximum spot nodes |
| **Cluster Management** |
| `enable_auto_upgrade` | bool | `false` | Auto-upgrade Kubernetes version |
| `maintenance_day` | string | `"sunday"` | Maintenance window day |
| `maintenance_start_time` | string | `"02:00"` | Maintenance start time (UTC) |
| **Container Registry** |
| `registry_tier` | string | `"basic"` | Registry tier (starter, basic, professional) |
| `registry_region` | string | `"nyc3"` | Registry region |
| **Load Balancer** |
| `enable_ingress_lb` | bool | `true` | Create load balancer for ingress |
| **Monitoring** |
| `enable_monitoring` | bool | `true` | Enable DigitalOcean monitoring |
| `alert_emails` | list(string) | `[]` | Email addresses for alerts |
| **Database (Optional)** |
| `enable_platform_db` | bool | `false` | Create PostgreSQL cluster |
| `postgres_version` | string | `"16"` | PostgreSQL version |
| `postgres_size` | string | `"db-s-2vcpu-4gb"` | Database cluster size |
| `postgres_node_count` | number | `2` | Database nodes (1=single, 2=HA) |
| **Backup Storage** |
| `enable_backup_bucket` | bool | `true` | Create Spaces backup bucket |
| `spaces_region` | string | `"nyc3"` | Spaces bucket region |
| `backup_retention_days` | number | `30` | Backup retention period |
| **Security Features** |
| `enable_cilium` | bool | `true` | Cilium CNI + Hubble (flow logs) |
| `enable_sealed_secrets` | bool | `true` | Sealed Secrets (KMS alternative) |
| `enable_falco` | bool | `true` | Falco runtime security |
| `enable_cert_manager` | bool | `true` | cert-manager for TLS |
| `enable_security_dashboard` | bool | `true` | Grafana security dashboard |

### Available Droplet Sizes

| Size Slug | vCPUs | Memory | Disk | Monthly Cost |
|-----------|-------|--------|------|--------------|
| `s-2vcpu-2gb` | 2 | 2 GB | 60 GB | $18 |
| `s-2vcpu-4gb` | 2 | 4 GB | 80 GB | $24 |
| `s-4vcpu-8gb` | 4 | 8 GB | 160 GB | $48 |
| `s-8vcpu-16gb` | 8 | 16 GB | 320 GB | $96 |

Check current sizes: `doctl compute size list`

### Available Regions

| Region | Location | Slug |
|--------|----------|------|
| New York 1 | USA | `nyc1` |
| New York 3 | USA | `nyc3` |
| San Francisco 3 | USA | `sfo3` |
| London 1 | UK | `lon1` |
| Frankfurt 1 | Germany | `fra1` |
| Singapore 1 | Singapore | `sgp1` |

Check current regions: `doctl compute region list`

## Outputs Reference

| Output | Description |
|--------|-------------|
| `cluster_id` | Kubernetes cluster ID |
| `cluster_name` | Kubernetes cluster name |
| `endpoint` | Kubernetes API server endpoint |
| `cluster_ca_certificate` | Base64-encoded cluster CA certificate (sensitive) |
| `vpc_id` | VPC ID |
| `vpc_cidr` | VPC CIDR block |
| `registry_endpoint` | Container registry endpoint |
| `loadbalancer_ip` | Ingress load balancer IP (if enabled) |
| `database_host` | Database host (if enabled, sensitive) |
| `database_port` | Database port (if enabled) |
| `database_uri` | Database connection URI (if enabled, sensitive) |
| `backup_bucket` | Backup bucket name (if enabled) |
| `backup_bucket_region` | Backup bucket region (if enabled) |
| `platform_security_enabled_features` | Summary of enabled security features |

## Post-Deployment

### Access Cluster

```bash
# Get kubeconfig
doctl kubernetes cluster kubeconfig save $(terraform output -raw cluster_name)

# Verify cluster
kubectl get nodes
kubectl get pods -A
```

### Container Registry

```bash
# Login to registry
doctl registry login

# Tag and push image
docker tag my-app:latest $(terraform output -raw registry_endpoint)/my-app:latest
docker push $(terraform output -raw registry_endpoint)/my-app:latest
```

### Database Access

```bash
# Get connection details
terraform output database_uri

# Connect via psql
doctl databases connection $(terraform output -raw cluster_name)-db --username doadmin
```

### Monitoring

```bash
# View cluster metrics
doctl kubernetes cluster get $(terraform output -raw cluster_name)

# View node metrics
doctl compute droplet list --tag-name $(terraform output -raw cluster_name)
```

## Security Features

### Cilium + Hubble (Flow Logs Alternative)

DigitalOcean doesn't provide native VPC Flow Logs. Cilium + Hubble provides:
- Network policy enforcement
- Layer 7 visibility
- Flow logs export to Spaces
- Web UI for flow visualization

```bash
# Access Hubble UI (if enabled)
kubectl port-forward -n kube-system svc/hubble-ui 12000:80
# Open http://localhost:12000
```

### Sealed Secrets (KMS Alternative)

DigitalOcean doesn't provide managed KMS. Sealed Secrets allows:
- Encrypted secrets in Git
- GitOps-friendly secret management
- Public key encryption

```bash
# Install kubeseal CLI
brew install kubeseal

# Create sealed secret
kubeseal --fetch-cert > pub-cert.pem
kubectl create secret generic my-secret --dry-run=client -o yaml | \
  kubeseal --cert pub-cert.pem -o yaml > sealed-secret.yaml
kubectl apply -f sealed-secret.yaml
```

### Falco (Runtime Security)

Monitors runtime activity and detects anomalous behavior:
- Process execution monitoring
- File system activity
- Network connections
- Container drift detection

```bash
# View Falco alerts
kubectl logs -n falco -l app=falco
```

## Cost Optimization

### Development Environment

```hcl
# terraform.tfvars
environment         = "dev"
system_node_count   = 1
workload_min_nodes  = 1
workload_max_nodes  = 3
enable_platform_db  = false
enable_ingress_lb   = false
```

**Estimated Monthly Cost**: ~$50-100

### Staging Environment

```hcl
# terraform.tfvars
environment         = "staging"
system_node_count   = 2
workload_min_nodes  = 2
workload_max_nodes  = 5
enable_platform_db  = true
postgres_node_count = 1
```

**Estimated Monthly Cost**: ~$150-250

### Production Environment

```hcl
# terraform.tfvars
environment         = "production"
system_node_count   = 3
workload_min_nodes  = 3
workload_max_nodes  = 10
enable_platform_db  = true
postgres_node_count = 2  # HA
```

**Estimated Monthly Cost**: ~$300-600

## Troubleshooting

### Cluster Creation Failed

```bash
# Check Terraform state
terraform show

# View DigitalOcean API errors
doctl kubernetes cluster list

# Force cluster deletion
doctl kubernetes cluster delete <cluster-id> --dangerous
```

### Node Pool Not Scaling

```bash
# Check node pool configuration
kubectl get nodes -o wide

# View DigitalOcean autoscaler logs
doctl kubernetes cluster node-pool list <cluster-id>
```

### Registry Authentication Failed

```bash
# Refresh registry credentials
doctl registry login

# Regenerate Kubernetes secret
doctl registry kubernetes-manifest | kubectl apply -f -
```

## Cleanup

```bash
# Destroy all resources
terraform destroy

# Verify cleanup
doctl kubernetes cluster list
doctl compute droplet list
doctl vpcs list
```

## References

- [DigitalOcean Kubernetes Documentation](https://docs.digitalocean.com/products/kubernetes/)
- [Terraform DigitalOcean Provider](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs)
- [doctl CLI Reference](https://docs.digitalocean.com/reference/doctl/)
- [Cilium Documentation](https://docs.cilium.io/)
- [Sealed Secrets Documentation](https://sealed-secrets.netlify.app/)
- [Falco Documentation](https://falco.org/docs/)

## Support

For issues and questions:
- GitHub Issues: [mantl/issues](https://github.com/thomasvincent/mantl/issues)
- DigitalOcean Support: [DigitalOcean Support](https://www.digitalocean.com/support/)
