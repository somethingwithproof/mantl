# OpenCost - Cost Management

Real-time cost allocation and optimization for the Mantl platform.

## Features

- **Real-time Cost Visibility**: Per-namespace, pod, and workload cost tracking
- **Idle Resource Detection**: Identify and eliminate waste (typically 20-30% savings)
- **Budget Alerts**: Get notified before overspending
- **Multi-Cloud Support**: AWS, GCP, Azure cost reconciliation
- **Cost Allocation**: Fair cost sharing across teams and projects

## Quick Start

### 1. Deploy OpenCost

```bash
kubectl apply -k platform/cost/opencost/
```

### 2. Access UI

```bash
# Port forward to access locally
kubectl port-forward -n opencost svc/opencost 9090:9090

# Or visit configured ingress
open https://opencost.mantl.example.com
```

### 3. View Costs

**By Namespace:**
```bash
curl http://localhost:9090/allocation/compute?window=7d&aggregate=namespace
```

**By Pod:**
```bash
curl http://localhost:9090/allocation/compute?window=24h&aggregate=pod
```

## Cloud Provider Integration

### AWS

```yaml
# Configure via External Secrets
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: opencost-aws
  namespace: opencost
spec:
  secretStoreRef:
    name: aws-secretsmanager
    kind: SecretStore
  target:
    name: opencost-cloud-credentials
  data:
    - secretKey: AWS_ACCESS_KEY_ID
      remoteRef:
        key: opencost/aws
        property: access_key_id
    - secretKey: AWS_SECRET_ACCESS_KEY
      remoteRef:
        key: opencost/aws
        property: secret_access_key
```

Enable Spot Data Feed in S3 for accurate spot instance pricing.

### GCP

```yaml
# Use Workload Identity
apiVersion: v1
kind: ServiceAccount
metadata:
  name: opencost
  namespace: opencost
  annotations:
    iam.gke.io/gcp-service-account: opencost@PROJECT_ID.iam.gserviceaccount.com
```

Grant BigQuery read access to billing export dataset.

### Azure

```yaml
# Configure via Managed Identity
# Grant Cost Management Reader role to AKS managed identity
```

## Cost Optimization Queries

### Find Idle Resources

```bash
# Pods with <10% CPU utilization
curl 'http://localhost:9090/allocation/compute?window=7d&idle=true'
```

### Most Expensive Namespaces

```bash
curl 'http://localhost:9090/allocation/compute?window=30d&aggregate=namespace' \
  | jq -r '.data[] | [.name, .totalCost] | @tsv' \
  | sort -k2 -rn \
  | head -10
```

### Cost Trend Analysis

```bash
# Compare this month vs last month
curl 'http://localhost:9090/allocation/compute?window=month'
curl 'http://localhost:9090/allocation/compute?window=month&offset=month'
```

## Grafana Integration

OpenCost exports metrics to Prometheus that can be visualized in Grafana:

```promql
# Total cluster hourly cost
sum(node_total_hourly_cost)

# Cost per namespace
sum by (namespace) (
  rate(container_cpu_usage_seconds_total[1h]) *
  on(node) group_left()
  node_cpu_hourly_cost
)

# Monthly projected cost
sum(node_total_hourly_cost) * 730
```

## Budget Enforcement

### Set Namespace Budget

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: namespace-budgets
  namespace: opencost
data:
  production: "5000"  # $5000/month
  staging: "1000"     # $1000/month
  development: "500"  # $500/month
```

### Alert Rules

Alerts are automatically configured in `values.yaml`:
- `HighMonthlyCostIncrease`: Costs up >20% month-over-month
- `NamespaceOverBudget`: Namespace exceeds monthly budget

## Best Practices

1. **Enable Spot/Preemptible Instances**: 60-90% cost savings for workload pools
2. **Right-size Requests**: Set accurate CPU/memory requests to avoid over-provisioning
3. **Use Node Auto-scaling**: Scale down during off-hours
4. **Review Idle Resources**: Weekly review of <10% utilization resources
5. **Tag Resources**: Use labels for accurate cost allocation

## API Reference

Full API documentation: https://www.opencost.io/docs/api

### Common Endpoints

- `/allocation/compute` - Cost allocation data
- `/allocation/assets` - Asset cost data (volumes, load balancers)
- `/allocation/summary` - High-level cost summary

## Troubleshooting

### Missing Cloud Provider Costs

Check cloud credentials:
```bash
kubectl logs -n opencost deployment/opencost -c opencost
```

### Inaccurate Costs

Ensure Prometheus is scraping node metrics:
```bash
kubectl get servicemonitor -n opencost
```

### UI Not Accessible

Check ingress configuration:
```bash
kubectl get ingress -n opencost
kubectl describe ingress opencost -n opencost
```

## Resources

- [OpenCost Documentation](https://www.opencost.io/docs/)
- [Cost Optimization Guide](https://www.opencost.io/docs/optimization)
- [API Reference](https://www.opencost.io/docs/api)
