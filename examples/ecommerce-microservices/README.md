# E-Commerce Microservices Example

Production-ready microservices application demonstrating mantl platform capabilities.

## Architecture

```
                   ┌──────────────┐
                   │   Ingress    │
                   └──────┬───────┘
                          │
            ┌─────────────┼─────────────┐
            │             │             │
      ┌─────▼────┐  ┌────▼─────┐  ┌───▼──────┐
      │ Frontend │  │ Products │  │  Orders  │
      │  (React) │  │   API    │  │   API    │
      └──────────┘  └────┬─────┘  └────┬─────┘
                         │             │
                    ┌────▼─────────────▼────┐
                    │   PostgreSQL Cluster  │
                    │   (CloudNativePG)     │
                    └───────────────────────┘
```

## Components

### Frontend Service
- React SPA served by Nginx
- Prometheus metrics via nginx-prometheus-exporter
- Progressive delivery with Flagger (canary deployments)
- CDN-ready with proper cache headers

### Products API
- Python FastAPI service
- REST API for product catalog
- PostgreSQL database backend
- Redis caching layer
- Prometheus metrics endpoint
- OpenTelemetry tracing

### Orders API
- Python FastAPI service
- REST API for order management
- PostgreSQL database backend
- Event publishing for order status changes
- Prometheus metrics endpoint
- OpenTelemetry tracing

### Database
- PostgreSQL 15 via CloudNativePG operator
- 3-node cluster with automatic failover
- Continuous backup to S3/GCS/Azure
- Point-in-time recovery (PITR)
- Connection pooling with PgBouncer

## Platform Integration

### Monitoring
- Prometheus ServiceMonitor for each service
- Custom metrics dashboards in Grafana
- Alert rules for SLIs (latency, error rate, saturation)

### Progressive Delivery
- Flagger canary deployments with automated rollback
- Metrics-driven promotion (success rate, latency)
- A/B testing support
- Blue-green deployment examples

### Security
- Trivy vulnerability scanning on all images
- Falco runtime security rules
- Network policies for zero-trust networking
- Pod Security Standards (restricted)
- Secret management via Sealed Secrets

### Cost Management
- OpenCost labels for namespace cost allocation
- Resource requests/limits properly configured
- Cost-optimized storage classes

### Backup & DR
- Velero backup annotations for stateful workloads
- Database continuous backup
- Disaster recovery runbooks

## Quick Start

### Prerequisites

```bash
# Deploy mantl platform first
helm install mantl-platform mantl/mantl-platform \
  --namespace mantl-system \
  --create-namespace \
  --values values-production.yaml
```

### Deploy Application

```bash
# Deploy database cluster
kubectl apply -f database/postgres-cluster.yaml

# Wait for database to be ready
kubectl wait --for=condition=Ready cluster/ecommerce-db --timeout=5m

# Deploy application services
kubectl apply -k overlays/production/

# Wait for canary rollout
kubectl wait --for=condition=Promoted canary/frontend --timeout=5m
kubectl wait --for=condition=Promoted canary/products-api --timeout=5m
kubectl wait --for=condition=Promoted canary/orders-api --timeout=5m
```

### Access Application

```bash
# Get ingress URL
kubectl get ingress ecommerce-frontend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'

# Or port-forward for testing
kubectl port-forward svc/frontend 8080:80
open http://localhost:8080
```

## Deployment Patterns

### Canary Deployment

```bash
# Update image version
kubectl set image deployment/products-api \
  products-api=products-api:v2.0.0

# Watch Flagger analysis
kubectl describe canary products-api

# Flagger will:
# 1. Deploy canary pods (10% traffic)
# 2. Monitor metrics (success rate, latency)
# 3. Gradually increase traffic (20%, 30%, 40%, 50%)
# 4. Promote or rollback based on metrics
```

### A/B Testing

```bash
# Deploy A/B test configuration
kubectl apply -f flagger/ab-testing-example.yaml

# Monitor user behavior metrics
kubectl logs -f deploy/products-api | grep "variant=B"
```

### Blue-Green Deployment

```bash
# Deploy blue-green configuration
kubectl apply -f flagger/blue-green-example.yaml

# Switch traffic instantly after validation
kubectl annotate canary/orders-api \
  flagger.app/promote=true
```

## Monitoring & Observability

### View Metrics

```bash
# Access Grafana dashboard
kubectl port-forward -n mantl-system svc/mantl-platform-grafana 3000:80
open http://localhost:3000

# Navigate to: Dashboards → E-Commerce Application
```

### View Traces

```bash
# Access Jaeger UI (if enabled)
kubectl port-forward -n mantl-system svc/jaeger-query 16686:16686
open http://localhost:16686
```

### View Logs

```bash
# Stream logs from all services
kubectl logs -f -l app.kubernetes.io/part-of=ecommerce --all-containers=true

# View specific service logs
kubectl logs -f deploy/products-api
kubectl logs -f deploy/orders-api
kubectl logs -f deploy/frontend
```

## Cost Analysis

### View Namespace Costs

```bash
# Access OpenCost UI
kubectl port-forward -n mantl-system svc/mantl-platform-opencost 9090:9090
open http://localhost:9090

# Filter by namespace: ecommerce
```

### Optimize Costs

```bash
# Review resource utilization
kubectl top pods -n ecommerce

# Adjust resource requests if overprovisioned
kubectl set resources deployment/products-api \
  --requests=cpu=100m,memory=256Mi \
  --limits=cpu=500m,memory=512Mi
```

## Security

### Scan for Vulnerabilities

```bash
# View vulnerability reports
kubectl get vulnerabilityreports -n ecommerce

# View critical vulnerabilities
kubectl get vulnerabilityreports -n ecommerce \
  -o json | jq '.items[] | select(.report.summary.criticalCount > 0)'
```

### View Security Alerts

```bash
# View Falco alerts
kubectl logs -n mantl-system daemonset/mantl-platform-falco | \
  grep "ecommerce"

# Access Falco UI
kubectl port-forward -n mantl-system svc/mantl-platform-falcosidekick-ui 2802:2802
open http://localhost:2802
```

### Test Network Policies

```bash
# Verify isolation
kubectl exec -it deploy/frontend -- curl http://products-api/health
# Should succeed (allowed)

kubectl exec -it deploy/frontend -- curl http://orders-api:8000/admin
# Should fail (blocked by network policy)
```

## Backup & Recovery

### Create Manual Backup

```bash
# Backup entire namespace
velero backup create ecommerce-backup \
  --include-namespaces ecommerce \
  --wait

# Verify backup
velero backup describe ecommerce-backup
```

### Restore from Backup

```bash
# Restore to new namespace
velero restore create --from-backup ecommerce-backup \
  --namespace-mappings ecommerce:ecommerce-restored

# Verify restore
kubectl get all -n ecommerce-restored
```

### Database PITR

```bash
# Restore database to specific point in time
kubectl apply -f database/pitr-restore.yaml

# Monitor restore progress
kubectl describe cluster ecommerce-db-restored
```

## Load Testing

### Run Load Test

```bash
# Install k6
brew install k6  # or download from k6.io

# Run load test
k6 run loadtest/stress-test.js

# Watch metrics in Grafana during test
```

### Chaos Testing

```bash
# Deploy chaos experiments
kubectl apply -f chaos/pod-failure.yaml
kubectl apply -f chaos/network-delay.yaml

# Monitor application resilience
kubectl describe chaosexperiment pod-failure

# Verify metrics don't degrade significantly
```

## Directory Structure

```
ecommerce-microservices/
├── README.md                    # This file
├── base/                        # Kustomize base configurations
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── frontend/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── canary.yaml
│   │   └── servicemonitor.yaml
│   ├── products-api/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── canary.yaml
│   │   └── servicemonitor.yaml
│   └── orders-api/
│       ├── deployment.yaml
│       ├── service.yaml
│       ├── canary.yaml
│       └── servicemonitor.yaml
├── overlays/
│   ├── dev/                     # Development environment
│   │   ├── kustomization.yaml
│   │   └── patches/
│   ├── staging/                 # Staging environment
│   │   ├── kustomization.yaml
│   │   └── patches/
│   └── production/              # Production environment
│       ├── kustomization.yaml
│       └── patches/
├── database/                    # Database configurations
│   ├── postgres-cluster.yaml
│   ├── pitr-restore.yaml
│   └── pooler.yaml
├── flagger/                     # Progressive delivery examples
│   ├── canary-basic.yaml
│   ├── ab-testing-example.yaml
│   └── blue-green-example.yaml
├── monitoring/                  # Monitoring configurations
│   ├── dashboards/
│   │   └── ecommerce-overview.json
│   ├── alerts/
│   │   └── sli-alerts.yaml
│   └── servicemonitors/
├── security/                    # Security policies
│   ├── network-policies.yaml
│   ├── pod-security.yaml
│   └── sealed-secrets/
├── backup/                      # Backup configurations
│   ├── schedule.yaml
│   └── restore-procedures.md
├── chaos/                       # Chaos experiments
│   ├── pod-failure.yaml
│   ├── network-delay.yaml
│   └── stress-test.yaml
├── loadtest/                    # Load testing scripts
│   ├── stress-test.js
│   └── smoke-test.js
└── src/                         # Application source code
    ├── frontend/
    ├── products-api/
    └── orders-api/
```

## Contributing

See main [CONTRIBUTING.md](../../CONTRIBUTING.md)

## License

Apache 2.0
