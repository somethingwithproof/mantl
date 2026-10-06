<!-- SPDX-License-Identifier: Apache-2.0 -->
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

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_digitalocean"></a> [digitalocean](#requirement\_digitalocean) | >= 2.70, < 3.0 |
| <a name="requirement_helm"></a> [helm](#requirement\_helm) | 3.3.0 |
| <a name="requirement_kubernetes"></a> [kubernetes](#requirement\_kubernetes) | 3.3.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_digitalocean"></a> [digitalocean](#provider\_digitalocean) | 2.104.0 |

## Modules

| Name | Source | Version |
| ---- | ------ | ------- |
| <a name="module_platform_security"></a> [platform\_security](#module\_platform\_security) | ../../modules/platform-security | n/a |

## Resources

| Name | Type |
| ---- | ---- |
| [digitalocean_container_registry.main](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/container_registry) | resource |
| [digitalocean_database_cluster.platform](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/database_cluster) | resource |
| [digitalocean_database_firewall.platform](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/database_firewall) | resource |
| [digitalocean_firewall.kubernetes_nodes](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/firewall) | resource |
| [digitalocean_kubernetes_cluster.main](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/kubernetes_cluster) | resource |
| [digitalocean_kubernetes_node_pool.spot](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/kubernetes_node_pool) | resource |
| [digitalocean_kubernetes_node_pool.workload](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/kubernetes_node_pool) | resource |
| [digitalocean_loadbalancer.ingress](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/loadbalancer) | resource |
| [digitalocean_monitor_alert.cluster_cpu_high](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/monitor_alert) | resource |
| [digitalocean_monitor_alert.cluster_disk_high](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/monitor_alert) | resource |
| [digitalocean_monitor_alert.cluster_memory_high](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/monitor_alert) | resource |
| [digitalocean_project.main](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/project) | resource |
| [digitalocean_spaces_bucket.backups](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/spaces_bucket) | resource |
| [digitalocean_vpc.main](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/vpc) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_alert_emails"></a> [alert\_emails](#input\_alert\_emails) | Email addresses for monitoring alerts | `list(string)` | `[]` | no |
| <a name="input_allowed_ssh_cidrs"></a> [allowed\_ssh\_cidrs](#input\_allowed\_ssh\_cidrs) | CIDR blocks allowed to SSH to nodes | `list(string)` | <pre>[<br/>  "10.0.0.0/8"<br/>]</pre> | no |
| <a name="input_backup_retention_days"></a> [backup\_retention\_days](#input\_backup\_retention\_days) | Number of days to retain backups | `number` | `30` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the Kubernetes cluster | `string` | `"mantl-doks"` | no |
| <a name="input_db_maintenance_day"></a> [db\_maintenance\_day](#input\_db\_maintenance\_day) | Day of week for database maintenance | `string` | `"sunday"` | no |
| <a name="input_db_maintenance_hour"></a> [db\_maintenance\_hour](#input\_db\_maintenance\_hour) | Hour of day for database maintenance (00-23 UTC) | `string` | `"03:00"` | no |
| <a name="input_destroy_all_resources"></a> [destroy\_all\_resources](#input\_destroy\_all\_resources) | Destroy all associated resources when cluster is destroyed | `bool` | `false` | no |
| <a name="input_enable_auto_upgrade"></a> [enable\_auto\_upgrade](#input\_enable\_auto\_upgrade) | Enable automatic Kubernetes version upgrades | `bool` | `false` | no |
| <a name="input_enable_backup_bucket"></a> [enable\_backup\_bucket](#input\_enable\_backup\_bucket) | Create a Spaces bucket for backups (S3-compatible) | `bool` | `true` | no |
| <a name="input_enable_cert_manager"></a> [enable\_cert\_manager](#input\_enable\_cert\_manager) | Enable cert-manager for automated TLS certificate management | `bool` | `true` | no |
| <a name="input_enable_cilium"></a> [enable\_cilium](#input\_enable\_cilium) | Enable Cilium CNI with Hubble for network observability (replaces missing VPC Flow Logs) | `bool` | `true` | no |
| <a name="input_enable_external_secrets"></a> [enable\_external\_secrets](#input\_enable\_external\_secrets) | Enable External Secrets Operator for external secrets management | `bool` | `false` | no |
| <a name="input_enable_falco"></a> [enable\_falco](#input\_enable\_falco) | Enable Falco for runtime security monitoring (enhanced security) | `bool` | `true` | no |
| <a name="input_enable_flow_logs_export"></a> [enable\_flow\_logs\_export](#input\_enable\_flow\_logs\_export) | Export Cilium flow logs to Spaces bucket | `bool` | `true` | no |
| <a name="input_enable_ingress_lb"></a> [enable\_ingress\_lb](#input\_enable\_ingress\_lb) | Create a load balancer for ingress traffic | `bool` | `true` | no |
| <a name="input_enable_monitoring"></a> [enable\_monitoring](#input\_enable\_monitoring) | Enable DigitalOcean monitoring and alerts | `bool` | `true` | no |
| <a name="input_enable_platform_db"></a> [enable\_platform\_db](#input\_enable\_platform\_db) | Create a managed PostgreSQL database for platform services | `bool` | `false` | no |
| <a name="input_enable_platform_security"></a> [enable\_platform\_security](#input\_enable\_platform\_security) | Deploy Kubernetes platform security components after the cluster exists; enable on a second apply | `bool` | `false` | no |
| <a name="input_enable_sealed_secrets"></a> [enable\_sealed\_secrets](#input\_enable\_sealed\_secrets) | Enable Sealed Secrets for encrypted secrets in Git (replaces missing KMS) | `bool` | `true` | no |
| <a name="input_enable_security_dashboard"></a> [enable\_security\_dashboard](#input\_enable\_security\_dashboard) | Enable Grafana dashboard for security monitoring | `bool` | `true` | no |
| <a name="input_enable_spot_pool"></a> [enable\_spot\_pool](#input\_enable\_spot\_pool) | Enable spot instances node pool for cost savings | `bool` | `false` | no |
| <a name="input_enable_tailscale"></a> [enable\_tailscale](#input\_enable\_tailscale) | Enable Tailscale VPN for secure cluster access (replaces missing private endpoint) | `bool` | `false` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name used for tags; unsupported DigitalOcean project values map to development | `string` | `"production"` | no |
| <a name="input_falco_ui_enabled"></a> [falco\_ui\_enabled](#input\_falco\_ui\_enabled) | Enable Falco UI for alert visualization | `bool` | `false` | no |
| <a name="input_falco_webhook_url"></a> [falco\_webhook\_url](#input\_falco\_webhook\_url) | Webhook URL for Falco alerts (e.g., Slack, PagerDuty) | `string` | `""` | no |
| <a name="input_hubble_ui_hosts"></a> [hubble\_ui\_hosts](#input\_hubble\_ui\_hosts) | Hostnames for Hubble UI ingress | `list(string)` | `[]` | no |
| <a name="input_hubble_ui_ingress_enabled"></a> [hubble\_ui\_ingress\_enabled](#input\_hubble\_ui\_ingress\_enabled) | Enable ingress for Hubble UI | `bool` | `false` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version for the cluster | `string` | `"1.31.3-do.0"` | no |
| <a name="input_maintenance_day"></a> [maintenance\_day](#input\_maintenance\_day) | Day of week for maintenance (monday, tuesday, etc.) | `string` | `"sunday"` | no |
| <a name="input_maintenance_start_time"></a> [maintenance\_start\_time](#input\_maintenance\_start\_time) | Start time for maintenance window (HH:MM in UTC) | `string` | `"02:00"` | no |
| <a name="input_postgres_node_count"></a> [postgres\_node\_count](#input\_postgres\_node\_count) | Number of database nodes (1 for single, 2 for HA) | `number` | `2` | no |
| <a name="input_postgres_size"></a> [postgres\_size](#input\_postgres\_size) | Database cluster size slug | `string` | `"db-s-2vcpu-4gb"` | no |
| <a name="input_postgres_version"></a> [postgres\_version](#input\_postgres\_version) | PostgreSQL version | `string` | `"16"` | no |
| <a name="input_region"></a> [region](#input\_region) | DigitalOcean region (e.g., nyc1, nyc3, sfo3, sgp1, lon1, fra1) | `string` | `"nyc3"` | no |
| <a name="input_registry_region"></a> [registry\_region](#input\_registry\_region) | Region for container registry | `string` | `"nyc3"` | no |
| <a name="input_registry_tier"></a> [registry\_tier](#input\_registry\_tier) | Container registry tier (starter, basic, professional) | `string` | `"basic"` | no |
| <a name="input_spaces_region"></a> [spaces\_region](#input\_spaces\_region) | Region for Spaces bucket (nyc3, sfo3, sgp1, fra1, ams3) | `string` | `"nyc3"` | no |
| <a name="input_spot_max_nodes"></a> [spot\_max\_nodes](#input\_spot\_max\_nodes) | Maximum number of spot nodes | `number` | `5` | no |
| <a name="input_spot_min_nodes"></a> [spot\_min\_nodes](#input\_spot\_min\_nodes) | Minimum number of spot nodes | `number` | `0` | no |
| <a name="input_spot_node_size"></a> [spot\_node\_size](#input\_spot\_node\_size) | Droplet size for spot nodes | `string` | `"s-4vcpu-8gb"` | no |
| <a name="input_system_node_count"></a> [system\_node\_count](#input\_system\_node\_count) | Number of system nodes (non-autoscaling) | `number` | `3` | no |
| <a name="input_system_node_size"></a> [system\_node\_size](#input\_system\_node\_size) | Droplet size for system nodes (platform components) | `string` | `"s-4vcpu-8gb"` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Additional tags to apply to resources | `list(string)` | <pre>[<br/>  "mantl",<br/>  "kubernetes"<br/>]</pre> | no |
| <a name="input_tailscale_client_id"></a> [tailscale\_client\_id](#input\_tailscale\_client\_id) | Tailscale OAuth client ID | `string` | `""` | no |
| <a name="input_tailscale_client_secret"></a> [tailscale\_client\_secret](#input\_tailscale\_client\_secret) | Tailscale OAuth client secret | `string` | `""` | no |
| <a name="input_vpc_cidr"></a> [vpc\_cidr](#input\_vpc\_cidr) | CIDR block for the VPC | `string` | `"10.10.0.0/16"` | no |
| <a name="input_workload_max_nodes"></a> [workload\_max\_nodes](#input\_workload\_max\_nodes) | Maximum number of workload nodes (autoscaling) | `number` | `10` | no |
| <a name="input_workload_min_nodes"></a> [workload\_min\_nodes](#input\_workload\_min\_nodes) | Minimum number of workload nodes (autoscaling) | `number` | `2` | no |
| <a name="input_workload_node_size"></a> [workload\_node\_size](#input\_workload\_node\_size) | Droplet size for workload nodes (application workloads) | `string` | `"s-4vcpu-8gb"` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_backup_bucket"></a> [backup\_bucket](#output\_backup\_bucket) | The name of the backup bucket |
| <a name="output_backup_bucket_region"></a> [backup\_bucket\_region](#output\_backup\_bucket\_region) | The region of the backup bucket |
| <a name="output_cluster_ca_certificate"></a> [cluster\_ca\_certificate](#output\_cluster\_ca\_certificate) | The base64 encoded cluster CA certificate |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | The ID of the Kubernetes cluster |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | The name of the Kubernetes cluster |
| <a name="output_database_host"></a> [database\_host](#output\_database\_host) | The host of the platform database |
| <a name="output_database_port"></a> [database\_port](#output\_database\_port) | The port of the platform database |
| <a name="output_database_uri"></a> [database\_uri](#output\_database\_uri) | The connection URI of the platform database |
| <a name="output_endpoint"></a> [endpoint](#output\_endpoint) | The endpoint of the Kubernetes API server |
| <a name="output_loadbalancer_ip"></a> [loadbalancer\_ip](#output\_loadbalancer\_ip) | The IP address of the ingress load balancer |
| <a name="output_platform_security_enabled_features"></a> [platform\_security\_enabled\_features](#output\_platform\_security\_enabled\_features) | Summary of enabled platform security features |
| <a name="output_registry_endpoint"></a> [registry\_endpoint](#output\_registry\_endpoint) | The endpoint of the container registry |
| <a name="output_vpc_cidr"></a> [vpc\_cidr](#output\_vpc\_cidr) | The CIDR block of the VPC |
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | The ID of the VPC |
<!-- END_TF_DOCS -->
