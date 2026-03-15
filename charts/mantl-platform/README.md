# mantl Platform Helm Chart

Complete platform stack for production Kubernetes with monitoring, backup, security, cost management, and progressive delivery.

## Features

- ✅ **Monitoring**: Prometheus + Grafana with pre-configured dashboards
- ✅ **Backup & DR**: Velero with scheduled backups
- ✅ **Security**: Trivy vulnerability scanning + Falco runtime security
- ✅ **Database**: CloudNativePG operator for PostgreSQL
- ✅ **Progressive Delivery**: Flagger for canary deployments
- ✅ **Cost Management**: OpenCost for real-time cost visibility
- ✅ **Chaos Engineering**: Chaos Mesh (optional)
- ✅ **Developer Platform**: Backstage + vCluster (optional)

## Prerequisites

- Kubernetes 1.31+
- Helm 3.14+
- Storage class for persistent volumes
- Cloud provider credentials (for backup storage)

## Quick Start

### Install with Default Settings

```bash
# Add mantl Helm repository
helm repo add mantl https://charts.mantl.io
helm repo update

# Install with default settings (production profile)
helm install mantl-platform mantl/mantl-platform \
  --namespace mantl-system \
  --create-namespace
```

### Install with Custom Values

```bash
# Development profile
helm install mantl-platform mantl/mantl-platform \
  --namespace mantl-system \
  --create-namespace \
  --values values-dev.yaml

# Production profile with custom domain
helm install mantl-platform mantl/mantl-platform \
  --namespace mantl-system \
  --create-namespace \
  --values values-production.yaml \
  --set global.domain=mycompany.com
```

## Configuration

### Profiles

The chart includes three pre-configured profiles:

| Profile | Use Case | Resources | Retention | Cost |
|---------|----------|-----------|-----------|------|
| **dev** | Development/Testing | Minimal | 7 days | $ |
| **staging** | Pre-production | Medium | 10 days | $$ |
| **production** | Production | High | 30+ days | $$$ |

### Global Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.domain` | Base domain for ingress | `example.com` |
| `global.storageClass` | Storage class for PVCs | `""` (default) |
| `profile` | Configuration profile | `production` |

### Monitoring

```yaml
monitoring:
  enabled: true
  prometheus:
    retention: 15d
    storageSize: 50Gi
    resources:
      requests:
        cpu: 500m
        memory: 2Gi

  grafana:
    enabled: true
    ingress:
      enabled: true
      hosts:
        - grafana.example.com
```

### Backup

```yaml
backup:
  enabled: true
  velero:
    configuration:
      provider: aws  # or gcp, azure, do, linode
      backupStorageLocation:
        bucket: my-backups
        config:
          region: us-east-1

    schedules:
      daily:
        schedule: "0 2 * * *"
        template:
          ttl: 720h  # 30 days
```

### Security

```yaml
security:
  trivy:
    enabled: true
    scanJobsCronJob: "0 1 * * *"  # Daily scans
    compliance:
      enabled: true
      reports:
        - nsa
        - pci-dss

  falco:
    enabled: true
    priority: warning
    falcosidekick:
      enabled: true
      webui:
        enabled: true
```

## Installation Examples

### Minimal Installation (Monitoring Only)

```bash
helm install mantl-platform mantl/mantl-platform \
  --set backup.enabled=false \
  --set security.enabled=false \
  --set cost.enabled=false \
  --set progressiveDelivery.enabled=false
```

### Security-Focused Installation

```bash
helm install mantl-platform mantl/mantl-platform \
  --set monitoring.enabled=true \
  --set security.trivy.enabled=true \
  --set security.falco.enabled=true \
  --set backup.enabled=true
```

### Full Production Stack

```bash
helm install mantl-platform mantl/mantl-platform \
  --values values-production.yaml \
  --set global.domain=prod.example.com \
  --set monitoring.grafana.ingress.enabled=true \
  --set cost.opencost.ui.ingress.enabled=true \
  --set backup.velero.configuration.backupStorageLocation.bucket=prod-backups
```

## Accessing Components

### Grafana

```bash
# Get admin password
kubectl get secret --namespace mantl-system mantl-platform-grafana \
  -o jsonpath="{.data.admin-password}" | base64 --decode

# Port-forward (if ingress not enabled)
kubectl port-forward --namespace mantl-system \
  svc/mantl-platform-grafana 3000:80

# Open browser
open http://localhost:3000
```

### Prometheus

```bash
kubectl port-forward --namespace mantl-system \
  svc/mantl-platform-prometheus-server 9090:80

open http://localhost:9090
```

### OpenCost

```bash
kubectl port-forward --namespace mantl-system \
  svc/mantl-platform-opencost 9090:9090

open http://localhost:9090
```

### Falco UI

```bash
kubectl port-forward --namespace mantl-system \
  svc/mantl-platform-falcosidekick-ui 2802:2802

open http://localhost:2802
```

## Cloud Provider-Specific Configuration

### AWS

```yaml
backup:
  velero:
    configuration:
      provider: aws
      backupStorageLocation:
        bucket: my-backups
        config:
          region: us-east-1
    credentials:
      useSecret: true
      secretContents:
        cloud: |
          [default]
          aws_access_key_id=<ACCESS_KEY>
          aws_secret_access_key=<SECRET_KEY>
```

### Google Cloud

```yaml
backup:
  velero:
    configuration:
      provider: gcp
      backupStorageLocation:
        bucket: my-backups
        config:
          project: my-project
    credentials:
      useSecret: true
      secretContents:
        cloud: <GCP_SERVICE_ACCOUNT_JSON>
```

### Azure

```yaml
backup:
  velero:
    configuration:
      provider: azure
      backupStorageLocation:
        bucket: my-backups
        config:
          resourceGroup: my-rg
          storageAccount: myaccount
```

### DigitalOcean

```yaml
backup:
  velero:
    configuration:
      provider: aws  # S3-compatible
      backupStorageLocation:
        bucket: my-backups
        config:
          region: nyc3
          s3ForcePathStyle: "true"
          s3Url: https://nyc3.digitaloceanspaces.com
```

## Upgrade

```bash
# Update Helm repository
helm repo update

# Upgrade installation
helm upgrade mantl-platform mantl/mantl-platform \
  --namespace mantl-system \
  --values my-values.yaml
```

## Uninstall

```bash
helm uninstall mantl-platform --namespace mantl-system
```

**Warning:** This will delete all platform components. Backup data in Velero will remain in object storage.

## Monitoring & Alerts

### Pre-configured Dashboards

The chart includes these Grafana dashboards:

- Kubernetes Cluster Overview
- Node Metrics
- Pod Resources
- Security Monitoring (Falco/Trivy)
- Cost Analysis (OpenCost)
- Backup Status (Velero)

### Alert Rules

Default alert rules included:

- High CPU/Memory usage
- Pod crash loops
- Persistent volume issues
- Backup failures
- Security vulnerabilities (Critical/High)
- Cost anomalies

### Alert Routing

Configure Alertmanager:

```yaml
monitoring:
  alertmanager:
    config:
      receivers:
        - name: 'slack'
          slack_configs:
            - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK'
              channel: '#alerts'
      route:
        receiver: 'slack'
        group_by: ['alertname', 'cluster']
        group_wait: 10s
        group_interval: 10s
        repeat_interval: 12h
```

## Backup & Restore

### Create Manual Backup

```bash
velero backup create manual-backup \
  --include-namespaces default,production \
  --wait
```

### Restore from Backup

```bash
velero restore create --from-backup daily-20240124
```

### List Backups

```bash
velero backup get
```

## Security Scanning

### View Vulnerability Reports

```bash
# List all vulnerability reports
kubectl get vulnerabilityreports -A

# View specific report
kubectl get vulnerabilityreport <name> -o yaml
```

### View Falco Alerts

```bash
# Real-time alerts
kubectl logs -n mantl-system daemonset/mantl-platform-falco -f

# Filter by priority
kubectl logs -n mantl-system daemonset/mantl-platform-falco | \
  grep "Priority: Warning"
```

## Cost Management

### View Cost by Namespace

```bash
kubectl port-forward -n mantl-system svc/mantl-platform-opencost 9090:9090

# Query API
curl http://localhost:9090/allocation/compute \
  ?window=7d \
  &aggregate=namespace
```

### Export Cost Reports

Configure CSV export:

```yaml
cost:
  opencost:
    exporter:
      enabled: true
      csv:
        endpoint: s3://my-bucket/cost-reports
```

## Progressive Delivery

### Create Canary Deployment

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: myapp
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  service:
    port: 80
  analysis:
    interval: 1m
    threshold: 10
    maxWeight: 50
    stepWeight: 10
    metrics:
    - name: request-success-rate
      thresholdRange:
        min: 99
      interval: 1m
```

## Troubleshooting

### Prometheus Not Starting

```bash
# Check PVC
kubectl get pvc -n mantl-system

# Check pod logs
kubectl logs -n mantl-system -l app=prometheus-server
```

### Velero Backup Failing

```bash
# Check Velero logs
kubectl logs -n mantl-system deploy/mantl-platform-velero

# Verify credentials
kubectl get secret -n mantl-system velero-credentials
```

### Falco High CPU

```bash
# Check rules
kubectl get configmap -n mantl-system mantl-platform-falco

# Adjust priority threshold
helm upgrade mantl-platform mantl/mantl-platform \
  --set security.falco.priority=error
```

## Values Reference

Full values reference: [values.yaml](./values.yaml)

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md)

## License

Apache 2.0
