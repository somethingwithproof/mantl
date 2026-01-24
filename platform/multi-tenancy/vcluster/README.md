# vCluster - Multi-Tenancy

Virtual clusters for team isolation.

## Features

- **True Isolation**: Each team gets their own "cluster"
- **Resource Quotas**: Enforce limits per team
- **Cost Allocation**: Accurate per-team billing
- **Self-Service**: Teams manage their own resources
- **Reduced Sprawl**: Multiple teams, one physical cluster

## Quick Start

### 1. Create Virtual Cluster

```bash
vcluster create team-a -n team-a \
  --values platform/multi-tenancy/vcluster/values.yaml
```

### 2. Connect

```bash
vcluster connect team-a -n team-a
```

### 3. Use Like Any Cluster

```bash
kubectl get namespaces  # Shows team-a's namespaces only
kubectl create namespace my-app
kubectl apply -f deployment.yaml
```

## Use Cases

### Development Teams
- Each team gets isolated environment
- Full kubectl access
- Can't affect other teams

### CI/CD Pipelines
- Ephemeral vclusters per test run
- Complete isolation
- Fast creation/deletion

### Multi-Tenancy
- Customer isolation
- Compliance boundaries
- Separate billing

## Resource Management

### Set Quotas

```yaml
resourceQuota:
  enabled: true
  quota:
    requests.cpu: "10"
    requests.memory: "20Gi"
    persistentvolumeclaims: "10"
```

### Monitor Usage

```bash
kubectl get resourcequota -n team-a
```

## Cost Allocation

OpenCost tracks vcluster costs:
```promql
sum by (vcluster) (
  rate(container_cpu_usage_seconds_total{namespace=~"vcluster-.*"}[1h])
  * on(node) group_left()
  node_cpu_hourly_cost
)
```

## Cleanup

```bash
vcluster delete team-a -n team-a
```

## Resources

- [vCluster Docs](https://www.vcluster.com/docs)
