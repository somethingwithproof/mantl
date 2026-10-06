# Scaling Operations Runbook

> Historical component-stack procedure examples. These commands and objectives
> have not established acceptance for the current beta runtime. Use the maintained
> [operations guide](../runbooks.md) and [maturity table](../../README.md#capabilities-and-maturity)
> first. Any recovery objectives here are proposed targets, not measured guarantees.


Procedures for scaling mantl platform components and applications.

## Overview

This runbook covers:
- Horizontal Pod Autoscaling (HPA)
- Vertical Pod Autoscaling (VPA)
- Cluster autoscaling
- Database scaling
- Manual scaling procedures
- Performance optimization

## Horizontal Pod Autoscaling

### Check Current HPA Status

```bash
# List all HPAs
kubectl get hpa -A

# Describe specific HPA
kubectl describe hpa <hpa-name> -n <namespace>

# Watch HPA in real-time
kubectl get hpa -n <namespace> --watch
```

### Scale Application Deployment

```bash
# Create HPA
kubectl autoscale deployment <deployment-name> \
  --namespace <namespace> \
  --cpu-percent=70 \
  --min=3 \
  --max=10

# Or apply manifest:
kubectl apply -f - <<EOF
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: <app-name>
  namespace: <namespace>
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: <deployment-name>
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "100"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 4
        periodSeconds: 30
      selectPolicy: Max
EOF
```

### Custom Metrics Scaling

```bash
# Scale based on custom Prometheus metrics
kubectl apply -f - <<EOF
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: custom-metrics-hpa
  namespace: ecommerce
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: products-api
  minReplicas: 3
  maxReplicas: 50
  metrics:
  # Custom metric from Prometheus
  - type: Pods
    pods:
      metric:
        name: http_request_duration_p99
      target:
        type: AverageValue
        averageValue: "500m"  # 500ms
  - type: Object
    object:
      metric:
        name: requests_per_second
      describedObject:
        apiVersion: v1
        kind: Service
        name: products-api
      target:
        type: AverageValue
        averageValue: "1000"
EOF
```

## Cluster Autoscaling

### Check Cluster Autoscaler Status

```bash
# Get autoscaler deployment
kubectl get deployment cluster-autoscaler -n kube-system

# Check logs
kubectl logs -f deployment/cluster-autoscaler -n kube-system

# View current node groups
kubectl get nodes -L node.kubernetes.io/instance-type
```

### Configure Node Groups (AWS EKS Example)

```bash
# Update node group min/max
aws eks update-nodegroup-config \
  --cluster-name mantl-prod \
  --nodegroup-name standard-workers \
  --scaling-config minSize=3,maxSize=50,desiredSize=5

# Create new node group for specific workload
aws eks create-nodegroup \
  --cluster-name mantl-prod \
  --nodegroup-name high-memory-workers \
  --subnets subnet-xxx subnet-yyy \
  --instance-types r5.2xlarge r5.4xlarge \
  --scaling-config minSize=0,maxSize=20,desiredSize=0 \
  --labels workload-type=memory-intensive

# Tag pods to use specific node group
kubectl label pod <pod-name> workload-type=memory-intensive
```

## Database Scaling

### PostgreSQL (CloudNativePG)

#### Vertical Scaling (Resources)

```bash
# Update cluster resources
kubectl patch cluster <cluster-name> -n <namespace> --type=merge -p '
spec:
  resources:
    requests:
      cpu: 2000m
      memory: 8Gi
    limits:
      cpu: 4000m
      memory: 16Gi
'

# Operator will rolling restart pods with new resources
kubectl get pods -n <namespace> -l cnpg.io/cluster=<cluster-name> --watch
```

#### Horizontal Scaling (Read Replicas)

```bash
# Increase replica count
kubectl patch cluster <cluster-name> -n <namespace> --type=merge -p '
spec:
  instances: 5
'

# New replicas will be added automatically
# Verify
kubectl get pods -n <namespace> -l cnpg.io/cluster=<cluster-name>

# Check replication lag
kubectl exec -n <namespace> <cluster-name>-1 -- \
  psql -c "SELECT client_addr, state, sync_state, replay_lag FROM pg_stat_replication;"
```

#### Storage Scaling

```bash
# Expand PVC (requires storage class with allowVolumeExpansion: true)
kubectl patch pvc <pvc-name> -n <namespace> -p '
spec:
  resources:
    requests:
      storage: 100Gi
'

# Verify expansion
kubectl get pvc -n <namespace> <pvc-name>

# If PVC is in Pending, check events
kubectl describe pvc <pvc-name> -n <namespace>
```

## Manual Scaling

### Scale Deployment

```bash
# Scale to specific replica count
kubectl scale deployment <deployment-name> \
  --namespace <namespace> \
  --replicas=10

# Verify scaling
kubectl get deployment <deployment-name> -n <namespace>
kubectl get pods -n <namespace> -l app=<app-label> --watch
```

### Scale StatefulSet

```bash
# Scale statefulset
kubectl scale statefulset <statefulset-name> \
  --namespace <namespace> \
  --replicas=5

# StatefulSets scale in order (0, 1, 2, 3, 4)
kubectl get pods -n <namespace> -l app=<app-label> --watch

# When scaling down, pods terminate in reverse order
```

## Performance Optimization

### Resource Right-Sizing

```bash
# Analyze actual resource usage
kubectl top pods -n <namespace>

# Get historical usage (requires metrics-server)
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/namespaces/<namespace>/pods" | jq '.'

# Compare requests vs usage
kubectl get pods -n <namespace> -o json | jq -r '
  .items[] |
  "\(.metadata.name),\(.spec.containers[0].resources.requests.cpu),\(.spec.containers[0].resources.limits.cpu)"
' > resource-requests.csv

# Recommend VPA for auto-adjustment
kubectl apply -f - <<EOF
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: <app-name>-vpa
  namespace: <namespace>
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: <deployment-name>
  updatePolicy:
    updateMode: "Auto"  # or "Recreate" or "Initial" or "Off"
  resourcePolicy:
    containerPolicies:
    - containerName: '*'
      minAllowed:
        cpu: 100m
        memory: 128Mi
      maxAllowed:
        cpu: 4000m
        memory: 16Gi
EOF
```

### Connection Pool Tuning

```bash
# For PostgreSQL (CloudNativePG)
kubectl patch cluster <cluster-name> -n <namespace> --type=merge -p '
spec:
  postgresql:
    parameters:
      max_connections: "500"
      shared_buffers: "2GB"
      effective_cache_size: "6GB"
      maintenance_work_mem: "512MB"
      checkpoint_completion_target: "0.9"
      wal_buffers: "16MB"
      default_statistics_target: "100"
      random_page_cost: "1.1"
      effective_io_concurrency: "200"
      work_mem: "10MB"
      min_wal_size: "2GB"
      max_wal_size: "8GB"
'

# Add PgBouncer for connection pooling
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Pooler
metadata:
  name: <cluster-name>-pooler
  namespace: <namespace>
spec:
  cluster:
    name: <cluster-name>
  instances: 3
  type: rw
  pgbouncer:
    poolMode: transaction
    parameters:
      max_client_conn: "1000"
      default_pool_size: "25"
      max_db_connections: "100"
      max_user_connections: "100"
EOF
```

## Load Testing

### Before Scaling

```bash
# Run load test to determine scaling needs
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: load-test
  namespace: <namespace>
spec:
  template:
    spec:
      containers:
      - name: k6
        image: grafana/k6:latest
        args:
        - run
        - --vus=100
        - --duration=5m
        - /scripts/load-test.js
        volumeMounts:
        - name: scripts
          mountPath: /scripts
      restartPolicy: Never
      volumes:
      - name: scripts
        configMap:
          name: load-test-scripts
EOF

# Monitor during load test
watch kubectl top pods -n <namespace>
kubectl get hpa -n <namespace> --watch
```

## Scaling Best Practices

### 1. Set Resource Requests and Limits

```yaml
resources:
  requests:
    cpu: "100m"      # Minimum needed
    memory: "128Mi"
  limits:
    cpu: "1000m"     # Maximum allowed
    memory: "512Mi"
```

### 2. Configure HPA Behavior

```yaml
behavior:
  scaleDown:
    stabilizationWindowSeconds: 300  # Wait 5min before scale down
    policies:
    - type: Percent
      value: 50     # Scale down max 50% at a time
      periodSeconds: 60
  scaleUp:
    stabilizationWindowSeconds: 0   # Scale up immediately
    policies:
    - type: Percent
      value: 100    # Double capacity if needed
      periodSeconds: 30
```

### 3. Use PodDisruptionBudgets

```bash
kubectl apply -f - <<EOF
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: <app-name>-pdb
  namespace: <namespace>
spec:
  minAvailable: 2  # or maxUnavailable: 1
  selector:
    matchLabels:
      app: <app-name>
EOF
```

### 4. Monitor Key Metrics

```bash
# CPU utilization
kubectl top nodes
kubectl top pods -n <namespace>

# Memory pressure
kubectl describe nodes | grep -A 5 "Allocated resources"

# Pending pods (need more capacity)
kubectl get pods -A --field-selector=status.phase=Pending

# HPA status
kubectl get hpa -A
```

## Troubleshooting

### HPA Not Scaling

```bash
# Check metrics-server
kubectl get deployment metrics-server -n kube-system

# Check HPA conditions
kubectl describe hpa <hpa-name> -n <namespace>

# Verify metrics available
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/namespaces/<namespace>/pods"

# Check resource requests are set
kubectl get pods -n <namespace> -o yaml | grep -A 5 resources
```

### Cluster Autoscaler Not Adding Nodes

```bash
# Check autoscaler logs
kubectl logs -f deployment/cluster-autoscaler -n kube-system | grep -i "scale"

# Check for unschedulable pods
kubectl get pods -A --field-selector=status.phase=Pending

# Verify node group limits
aws eks describe-nodegroup \
  --cluster-name mantl-prod \
  --nodegroup-name standard-workers
```

## Emergency Scale-Down

```bash
# If needed to reduce costs or during incident:

# Scale down non-critical workloads
for ns in dev staging; do
  kubectl scale deployment --all --replicas=1 -n $ns
done

# Suspend CronJobs
kubectl patch cronjob <cronjob-name> -n <namespace> -p '{"spec":{"suspend":true}}'

# Reduce HPA minimums
kubectl patch hpa <hpa-name> -n <namespace> --type=merge -p '{"spec":{"minReplicas":1}}'
```

## Additional Resources

- [Kubernetes HPA Documentation](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [Cluster Autoscaler Documentation](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler)
- [VPA Documentation](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler)
