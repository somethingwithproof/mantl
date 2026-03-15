# Multi-Cluster Management

Manage and deploy applications across multiple Kubernetes clusters using ArgoCD ApplicationSets and cluster fleet patterns.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Management Cluster                         │
│                                                                │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────┐ │
│  │   ArgoCD    │  │  Prometheus  │  │  Thanos (Central)   │ │
│  │   Server    │  │   (Federated)│  │  Query/Store        │ │
│  └──────┬──────┘  └──────┬───────┘  └──────────┬──────────┘ │
│         │                │                     │              │
└─────────┼────────────────┼─────────────────────┼──────────────┘
          │                │                     │
    ┌─────┴─────┬──────────┴──────┬─────────────┴──────┐
    │           │                 │                     │
┌───▼────┐  ┌───▼────┐       ┌───▼────┐           ┌───▼────┐
│ Prod   │  │ Staging│       │  Dev   │           │  DR    │
│ Cluster│  │ Cluster│       │ Cluster│           │ Cluster│
└────────┘  └────────┘       └────────┘           └────────┘
```

## Components

### ArgoCD ApplicationSets
- **Cluster Generator**: Deploy to multiple clusters based on labels
- **Git Generator**: Deploy from multiple Git repositories
- **Matrix Generator**: Combine generators for complex patterns
- **Progressive Rollout**: Phased deployment across clusters

### Cluster Fleet Management
- **Cluster Registry**: Maintain inventory of all clusters
- **Cluster Classification**: Label clusters by environment, region, tier
- **Secret Management**: Cluster credentials via Sealed Secrets
- **Health Monitoring**: Aggregate cluster health metrics

### Multi-Cluster Observability
- **Thanos**: Global view of Prometheus metrics across clusters
- **Centralized Logging**: Aggregate logs from all clusters
- **Distributed Tracing**: Jaeger spans across cluster boundaries

## Quick Start

### Register Clusters with ArgoCD

```bash
# Production cluster
argocd cluster add prod-cluster \
  --name prod \
  --label environment=production \
  --label region=us-east-1 \
  --namespace argocd

# Staging cluster
argocd cluster add staging-cluster \
  --name staging \
  --label environment=staging \
  --label region=us-west-2 \
  --namespace argocd

# Development cluster
argocd cluster add dev-cluster \
  --name dev \
  --label environment=dev \
  --label region=eu-central-1 \
  --namespace argocd
```

### Deploy ApplicationSet

```bash
# Deploy to all production clusters
kubectl apply -f applicationsets/ecommerce-prod.yaml

# Deploy to all clusters in a region
kubectl apply -f applicationsets/monitoring-us-east.yaml

# Progressive rollout across clusters
kubectl apply -f applicationsets/progressive-rollout.yaml
```

### Verify Deployment

```bash
# List ApplicationSets
kubectl get applicationsets -n argocd

# Check generated Applications
kubectl get applications -n argocd

# View sync status across clusters
argocd appset list
```

## ApplicationSet Patterns

### Pattern 1: Environment-based Deployment

Deploy different configurations to dev/staging/prod:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: ecommerce
spec:
  generators:
  - clusters:
      selector:
        matchLabels:
          app: ecommerce
  template:
    metadata:
      name: 'ecommerce-{{name}}'
    spec:
      project: default
      source:
        repoURL: https://github.com/thomasvincent/mantl
        targetRevision: HEAD
        path: 'examples/ecommerce-microservices/overlays/{{metadata.labels.environment}}'
      destination:
        server: '{{server}}'
        namespace: ecommerce
```

### Pattern 2: Progressive Rollout

Deploy to clusters in phases:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: platform-rollout
spec:
  generators:
  - clusters:
      selector:
        matchLabels:
          rollout-phase: "1"  # Deploy to phase 1 first
  template:
    # ... application template
  strategy:
    type: RollingSync
    rollingSync:
      steps:
      - matchExpressions:
        - key: rollout-phase
          operator: In
          values: ["1"]
      - matchExpressions:
        - key: rollout-phase
          operator: In
          values: ["2"]
      - matchExpressions:
        - key: rollout-phase
          operator: In
          values: ["3"]
```

### Pattern 3: Multi-Repository

Deploy from multiple sources:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: multi-repo
spec:
  generators:
  - git:
      repoURL: https://github.com/thomasvincent/mantl
      directories:
      - path: 'applications/*'
  - clusters: {}
  template:
    metadata:
      name: '{{path.basename}}-{{name}}'
    spec:
      source:
        repoURL: '{{url}}'
        path: '{{path}}'
```

## Cluster Fleet Management

### Cluster Registration

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: prod-cluster-01
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: cluster
    environment: production
    region: us-east-1
    tier: gold
    rollout-phase: "1"
type: Opaque
stringData:
  name: prod-cluster-01
  server: https://prod-cluster-01.example.com
  config: |
    {
      "bearerToken": "<token>",
      "tlsClientConfig": {
        "insecure": false,
        "caData": "<base64-ca-cert>"
      }
    }
```

### Cluster Labels

Label clusters for targeted deployments:

- **environment**: dev, staging, production
- **region**: us-east-1, us-west-2, eu-central-1, ap-southeast-1
- **tier**: gold, silver, bronze (for different SLA levels)
- **rollout-phase**: "1", "2", "3" (for progressive rollouts)
- **workload-type**: web, database, ml, analytics

### Cluster Health Checks

```bash
# Check cluster connectivity
argocd cluster list

# Test cluster access
kubectl --context prod-cluster-01 get nodes

# Verify ArgoCD agent
kubectl --context prod-cluster-01 get pods -n argocd
```

## Multi-Cluster Observability

### Thanos Setup

Central Prometheus with Thanos:

```yaml
# Management cluster - Thanos Query
apiVersion: monitoring.coreos.com/v1
kind: ThanosRuler
metadata:
  name: thanos-query
  namespace: mantl-system
spec:
  replicas: 2
  queryEndpoints:
  - https://prod-cluster-01:10901
  - https://prod-cluster-02:10901
  - https://staging-cluster:10901
```

### Viewing Multi-Cluster Metrics

```bash
# Access Thanos Query UI
kubectl port-forward -n mantl-system svc/thanos-query 9090:9090

# Query across all clusters
# Open http://localhost:9090 and query:
#   sum(rate(http_requests_total[5m])) by (cluster)
```

### Centralized Logging

```yaml
# Fluentd forwarder on each cluster
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
spec:
  template:
    spec:
      containers:
      - name: fluentd
        image: fluent/fluentd-kubernetes-daemonset:v1-debian-elasticsearch
        env:
        - name: FLUENT_ELASTICSEARCH_HOST
          value: "elasticsearch.mantl-system.svc"
        - name: FLUENT_ELASTICSEARCH_PORT
          value: "9200"
        - name: CLUSTER_NAME
          value: "prod-cluster-01"
```

## Disaster Recovery

### Cross-Cluster Backup

```yaml
apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  name: cross-cluster
  namespace: velero
spec:
  provider: aws
  objectStorage:
    bucket: mantl-dr-backups
    prefix: cross-cluster
  config:
    region: us-east-1
```

### DR Cluster Promotion

```bash
# In DR scenario, promote DR cluster to production

# 1. Restore latest backup
velero restore create --from-backup prod-daily-20240124

# 2. Update DNS to point to DR cluster
# (handled externally)

# 3. Update ArgoCD to deploy to DR cluster
kubectl patch applicationset ecommerce-prod \
  --type=merge \
  -p '{"spec":{"generators":[{"clusters":{"selector":{"matchLabels":{"name":"dr-cluster"}}}}]}}'

# 4. Verify applications are running
argocd app list --selector app=ecommerce
```

## Best Practices

### Cluster Naming

Use consistent naming convention:
- `<environment>-<region>-<number>`
- Examples: prod-us-east-01, staging-eu-west-01, dev-ap-south-01

### Deployment Strategy

1. **Dev First**: Always deploy to dev clusters first
2. **Canary Clusters**: Use canary clusters for staging
3. **Progressive Production**: Roll out to production in phases
4. **Automated Rollback**: Monitor and auto-rollback on errors

### Secret Management

```bash
# Seal cluster credentials
kubeseal --format=yaml < cluster-secret.yaml > sealed-cluster-secret.yaml

# Store in Git
git add sealed-cluster-secret.yaml
git commit -m "Add production cluster credentials"
```

### Monitoring

- Alert on cluster health degradation
- Monitor ApplicationSet sync status
- Track deployment success rate per cluster
- Set up PagerDuty/Slack integration

## Troubleshooting

### ApplicationSet Not Creating Applications

```bash
# Check ApplicationSet status
kubectl describe applicationset <name> -n argocd

# Check cluster registration
argocd cluster list

# Verify cluster labels match selectors
kubectl get secret -n argocd -l argocd.argoproj.io/secret-type=cluster \
  -o jsonpath='{.items[*].metadata.labels}'
```

### Application Stuck in Progressing

```bash
# Check application status
argocd app get <app-name>

# View detailed sync status
argocd app sync <app-name> --dry-run

# Check target cluster
kubectl --context <cluster> get pods -n <namespace>
```

### Cluster Connectivity Issues

```bash
# Test cluster access
kubectl --context <cluster> get nodes

# Refresh cluster info in ArgoCD
argocd cluster get <cluster-url> --refresh

# Check ArgoCD application controller logs
kubectl logs -n argocd deploy/argocd-application-controller -f
```

## Advanced Patterns

### Blue-Green Cluster Deployment

```yaml
# Deploy to blue clusters
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: app-blue
spec:
  generators:
  - clusters:
      selector:
        matchLabels:
          deployment-slot: blue

# Switch traffic by updating label
kubectl label secret prod-cluster-01 deployment-slot=green --overwrite
```

### A/B Testing Across Clusters

```yaml
# Deploy variant A to 50% of clusters
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: ab-test
spec:
  generators:
  - clusters:
      selector:
        matchExpressions:
        - key: ab-variant
          operator: In
          values: ["A"]
  template:
    spec:
      source:
        path: overlays/variant-a
```

### Geo-Distributed Deployment

```yaml
# Deploy to specific regions
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: geo-distributed
spec:
  generators:
  - clusters:
      selector:
        matchLabels:
          region: us-east-1
  - clusters:
      selector:
        matchLabels:
          region: eu-west-1
  - clusters:
      selector:
        matchLabels:
          region: ap-south-1
```

## References

- [ArgoCD ApplicationSets](https://argo-cd.readthedocs.io/en/stable/user-guide/application-set/)
- [Thanos Documentation](https://thanos.io/tip/thanos/getting-started.md/)
- [Velero Multi-Cluster](https://velero.io/docs/main/multi-cluster/)
