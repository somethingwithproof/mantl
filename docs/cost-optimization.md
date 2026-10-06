<!-- SPDX-License-Identifier: Apache-2.0 -->
# Cost Optimization Guide

Comprehensive strategies for reducing cloud infrastructure costs while maintaining performance and reliability.

## Table of Contents

- [Multi-Cloud Cost Comparison](#multi-cloud-cost-comparison)
- [Compute Optimization](#compute-optimization)
- [Storage Optimization](#storage-optimization)
- [Network Cost Reduction](#network-cost-reduction)
- [Cost Monitoring](#cost-monitoring)
- [FinOps Best Practices](#finops-best-practices)
- [Quick Wins](#quick-wins)

## Multi-Cloud Cost Comparison

### Monthly Cost Estimates (Production Workload)

Baseline: 3-node cluster, 50GB storage, moderate traffic

| Provider | Service | Monthly Cost | Notes |
|----------|---------|--------------|-------|
| **AWS** | EKS | $73 (control plane) + ~$200 (nodes) | Add data transfer costs |
| **GCP** | GKE | $73 (control plane) + ~$180 (nodes) | Sustained use discounts |
| **Azure** | AKS | $0 (control plane) + ~$210 (nodes) | Free control plane |
| **DigitalOcean** | DOKS | $0 (control plane) + ~$120 (nodes) | Best for small workloads |
| **Oracle** | OKE | $0 (control plane) + ~$100 (nodes) | Generous free tier |
| **Linode** | LKE | $0 (control plane) + ~$90 (nodes) | Competitive pricing |

**Cost Breakdown Example (AWS EKS)**:
```
Control Plane:           $73/month
3x m6i.xlarge nodes:     $219/month
50GB gp3 storage:        $4/month
Application LB:          $23/month
NAT Gateway:             $33/month
Data Transfer (100GB):   $9/month
-----------------------------------
Total:                   ~$361/month
```

### Cost Optimization by Provider

#### AWS Cost Savings

1. **Savings Plans (30-50% off)**
   ```bash
   # Compute Savings Plans (1-year, no upfront)
   # - Flexible across instance types/regions
   # - Best for dynamic workloads

   # EC2 Instance Savings Plans (1-year, no upfront)
   # - Locked to instance family
   # - Higher discount rate
   ```

2. **Spot Instances (70-90% off)**
   ```yaml
   # Using Karpenter NodePool for spot
   apiVersion: karpenter.sh/v1beta1
   kind: NodePool
   spec:
     template:
       spec:
         requirements:
         - key: karpenter.sh/capacity-type
           operator: In
           values: ["spot", "on-demand"]
         - key: kubernetes.io/arch
           operator: In
           values: ["amd64"]
     disruption:
       consolidationPolicy: WhenUnderutilized
       expireAfter: 720h # 30 days
   ```

3. **Graviton Instances (20-40% cheaper)**
   ```yaml
   # ARM-based Graviton processors
   spec:
     requirements:
     - key: karpenter.sh/capacity-type
       operator: In
       values: ["spot"]
     - key: kubernetes.io/arch
       operator: In
       values: ["arm64"]  # Graviton
     - key: node.kubernetes.io/instance-type
       operator: In
       values: ["m7g.xlarge", "c7g.xlarge", "r7g.xlarge"]
   ```

4. **Storage Tier Optimization**
   ```yaml
   # Use gp3 instead of gp2 (20% cheaper, better performance)
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   metadata:
     name: gp3-optimized
   provisioner: ebs.csi.aws.com
   parameters:
     type: gp3
     iops: "3000"      # Configurable (vs gp2 fixed)
     throughput: "125" # Configurable
   allowVolumeExpansion: true
   ```

#### GCP Cost Savings

1. **Committed Use Discounts (37-57% off)**
   ```bash
   # 1-year commitment
   gcloud compute commitments create my-commitment \
     --resources vcpu=8,memory=32GB \
     --region us-central1 \
     --plan 12-month
   ```

2. **Sustained Use Discounts (automatic 30% off)**
   - No commitment required
   - Automatic based on usage
   - Stacks with committed use

3. **Preemptible VMs (80% off)**
   ```yaml
   # GKE node pool with preemptible nodes
   apiVersion: container.cnrm.cloud.google.com/v1beta1
   kind: ContainerNodePool
   spec:
     nodeConfig:
       preemptible: true
       machineType: n2-standard-4
   ```

#### Azure Cost Savings

1. **Reserved Instances (40-60% off)**
   ```bash
   # 1-year reserved VM
   az reservations reservation-order purchase \
     --applied-scope-type Shared \
     --billing-scope /subscriptions/xxx \
     --term P1Y \
     --sku Standard_D4s_v3
   ```

2. **Spot VMs (90% off)**
   ```yaml
   # AKS spot node pool
   apiVersion: containerservice.azure.com/v1
   kind: AgentPool
   spec:
     scaleSetPriority: Spot
     scaleSetEvictionPolicy: Delete
     spotMaxPrice: -1  # Pay up to regular price
   ```

3. **Azure Hybrid Benefit**
   - Use existing Windows Server licenses
   - Up to 40% savings on Windows nodes

## Compute Optimization

### Right-Sizing Strategy

**1. Identify Oversized Pods**

```bash
# Find pods using <50% of requested resources
kubectl top pods -A --containers | awk '{
  if (NR>1) {
    cpu_usage = substr($3, 1, length($3)-1)
    mem_usage = substr($4, 1, length($4)-2)
    if (cpu_usage < 50 || mem_usage < 50)
      print $1, $2, "CPU: " cpu_usage "% MEM: " mem_usage "%"
  }
}'
```

**2. Calculate Optimal Requests**

Use VPA (Vertical Pod Autoscaler) for recommendations:

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: myapp-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  updatePolicy:
    updateMode: "Off"  # Recommendations only
  resourcePolicy:
    containerPolicies:
    - containerName: '*'
      minAllowed:
        cpu: 50m
        memory: 64Mi
      maxAllowed:
        cpu: 2
        memory: 4Gi
```

```bash
# View VPA recommendations
kubectl describe vpa myapp-vpa
# Look for "Recommendation" section
```

**3. Implement Right-Sized Requests**

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: myapp
        resources:
          requests:
            cpu: 100m      # Based on VPA recommendation
            memory: 256Mi  # Based on VPA recommendation
          limits:
            cpu: 500m      # 5x requests for burst
            memory: 512Mi  # 2x requests for safety
```

### Spot Instance Best Practices

**1. Configure Karpenter for Spot**

```yaml
# NodePool prioritizing spot with on-demand fallback
apiVersion: karpenter.sh/v1beta1
kind: NodePool
metadata:
  name: default
spec:
  template:
    spec:
      requirements:
      - key: karpenter.sh/capacity-type
        operator: In
        values: ["spot", "on-demand"]
      - key: kubernetes.io/arch
        operator: In
        values: ["amd64", "arm64"]
  limits:
    cpu: 1000
  disruption:
    consolidationPolicy: WhenUnderutilized
    expireAfter: 720h
---
apiVersion: karpenter.k8s.aws/v1beta1
kind: EC2NodeClass
metadata:
  name: default
spec:
  amiFamily: AL2
  role: KarpenterNodeRole
  subnetSelectorTerms:
  - tags:
      karpenter.sh/discovery: my-cluster
  securityGroupSelectorTerms:
  - tags:
      karpenter.sh/discovery: my-cluster
```

**2. Handle Spot Interruptions**

```yaml
# Add pod disruption budget
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: myapp-pdb
spec:
  minAvailable: 2  # Keep at least 2 replicas running
  selector:
    matchLabels:
      app: myapp
```

**3. Spot Interruption Handler**

AWS Node Termination Handler (automatically installed with Karpenter):

```bash
# Verify node termination handler is running
kubectl get pods -n kube-system | grep node-termination
```

### Cluster Autoscaler Tuning

```yaml
# Cluster Autoscaler configuration
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-autoscaler
  namespace: kube-system
spec:
  template:
    spec:
      containers:
      - name: cluster-autoscaler
        image: registry.k8s.io/autoscaling/cluster-autoscaler:v1.28.2
        command:
        - ./cluster-autoscaler
        - --cloud-provider=aws
        - --skip-nodes-with-system-pods=false
        - --expander=least-waste  # Cost-optimized scaling
        - --scale-down-delay-after-add=5m
        - --scale-down-unneeded-time=5m
        - --scale-down-utilization-threshold=0.5
        - --balance-similar-node-groups
        - --skip-nodes-with-local-storage=false
```

## Storage Optimization

### Storage Class Comparison (AWS)

| Type | IOPS | Throughput | Cost/GB/month | Use Case |
|------|------|------------|---------------|----------|
| gp3 | 3,000-16,000 | 125-1,000 MB/s | $0.08 | General purpose (default) |
| gp2 | 100-16,000 (burst) | 250 MB/s | $0.10 | Legacy (avoid) |
| io2 | 100-64,000 | 1,000 MB/s | $0.125 + $0.065/IOPS | High-performance databases |
| st1 | 500 | 500 MB/s | $0.045 | Throughput-intensive (logs) |
| sc1 | 250 | 250 MB/s | $0.015 | Cold storage (archives) |

**Savings: gp3 vs gp2 = 20% cost reduction with better performance**

### Storage Optimization Strategies

**1. Use gp3 for All Workloads**

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp3
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  iops: "3000"
  throughput: "125"
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
```

**2. Lifecycle Policies for Snapshots**

```yaml
# Delete old EBS snapshots automatically
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: csi-aws-vsc
driver: ebs.csi.aws.com
deletionPolicy: Delete
parameters:
  tagSpecification_1: "Name=mantl-snapshot"
  tagSpecification_2: "Environment=production"
```

**3. Compress Logs Before Storage**

```yaml
# Loki configuration for compressed storage
apiVersion: v1
kind: ConfigMap
metadata:
  name: loki-config
data:
  loki.yaml: |
    compactor:
      working_directory: /data/loki/compactor
      shared_store: s3
      compaction_interval: 10m
      retention_enabled: true
      retention_delete_delay: 2h
      retention_delete_worker_count: 150

    limits_config:
      retention_period: 168h  # 7 days
      reject_old_samples: true
      reject_old_samples_max_age: 168h
```

**4. Use S3 Intelligent-Tiering**

```yaml
# S3 bucket for backups with auto-tiering
apiVersion: s3.aws.crossplane.io/v1beta1
kind: Bucket
metadata:
  name: mantl-backups
spec:
  forProvider:
    region: us-west-2
    lifecycleConfiguration:
      rules:
      - id: intelligent-tiering
        status: Enabled
        transitions:
        - days: 0
          storageClass: INTELLIGENT_TIERING
```

## Network Cost Reduction

### NAT Gateway Optimization

**Problem**: NAT Gateway costs $0.045/hour + $0.045/GB = ~$33/month + data transfer

**Solutions**:

**1. Use NAT Instances (70% cheaper)**

```bash
# Launch t4g.nano NAT instance ($3/month vs $33/month)
# See terraform/modules/vpc/nat-instance.tf for configuration
```

**2. Reduce Internet-Bound Traffic**

```yaml
# Use VPC Endpoints for AWS services (free data transfer)
apiVersion: ec2.aws.crossplane.io/v1beta1
kind: VPCEndpoint
metadata:
  name: s3-endpoint
spec:
  forProvider:
    vpcId: vpc-xxx
    serviceName: com.amazonaws.us-west-2.s3
    routeTableIds:
    - rtb-xxx
    vpcEndpointType: Gateway  # Free!
```

```yaml
# Interface endpoints for other services
apiVersion: ec2.aws.crossplane.io/v1beta1
kind: VPCEndpoint
metadata:
  name: ecr-api-endpoint
spec:
  forProvider:
    vpcId: vpc-xxx
    serviceName: com.amazonaws.us-west-2.ecr.api
    vpcEndpointType: Interface
    privateDnsEnabled: true
    subnetIds:
    - subnet-xxx
```

**3. Use CloudFront for Static Assets**

```yaml
# Reduce origin requests and data transfer
# See examples/static-website/k8s/cdn-cloudfront.yaml
# Savings: ~50% on data transfer costs
```

### Load Balancer Optimization

**Problem**: ALB costs $0.0225/hour + $0.008/LCU-hour = ~$16-50/month per ALB

**Solution**: Use Ingress Controller (Single ALB for Multiple Services)

```yaml
# One ALB for all ingresses
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: shared-ingress
  annotations:
    alb.ingress.kubernetes.io/group.name: shared  # Share ALB!
    alb.ingress.kubernetes.io/scheme: internet-facing
spec:
  ingressClassName: alb
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
```

**Savings**: $32-100/month (2-3 ALBs → 1 shared ALB)

## Cost Monitoring

### Install Kubecost (Free OpenSource)

```bash
# Add Kubecost Helm repo
helm repo add kubecost https://kubecost.github.io/cost-analyzer/
helm repo update

# Install Kubecost
helm install kubecost kubecost/cost-analyzer \
  --namespace kubecost \
  --create-namespace \
  --set kubecostToken="aGVsbUBrdWJlY29zdC5jb20=xm343yadf98" \
  --set prometheus.server.global.external_labels.cluster_id=mantl-prod
```

```bash
# Port-forward to access UI
kubectl port-forward -n kubecost svc/kubecost-cost-analyzer 9090:9090

# Open http://localhost:9090
```

### Key Kubecost Metrics

**1. Cost Allocation by Namespace**

```bash
# View namespace costs
curl http://localhost:9090/model/allocation \
  -d window=7d \
  -d aggregate=namespace \
  -G | jq '.data[] | {namespace: .name, totalCost: .totalCost}'
```

**2. Savings Recommendations**

Check the "Savings" tab in Kubecost UI:
- Abandoned workloads
- Right-sizing opportunities
- Cluster sizing recommendations

**3. Set Budget Alerts**

```yaml
# Alert when namespace exceeds budget
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubecost-alerts
  namespace: kubecost
data:
  alerts.json: |
    {
      "budgets": [
        {
          "name": "production-budget",
          "namespace": "production",
          "threshold": 1000,
          "window": "7d",
          "webhook": "https://hooks.slack.com/services/XXX"
        }
      ]
    }
```

### AWS Cost Explorer Integration

```bash
# Enable AWS Cost and Usage Reports
aws cur put-report-definition \
  --report-definition '{
    "ReportName": "mantl-cur",
    "TimeUnit": "HOURLY",
    "Format": "Parquet",
    "Compression": "Parquet",
    "S3Bucket": "mantl-billing-data",
    "S3Prefix": "cur/",
    "S3Region": "us-east-1",
    "ReportVersioning": "OVERWRITE_REPORT"
  }'

# Configure Kubecost to read CUR
kubectl create secret generic cloud-integration -n kubecost \
  --from-literal=cloud-integration.json='{
    "aws": {
      "serviceKeyName": "s3",
      "serviceKeySecret": "xxx",
      "bucketName": "mantl-billing-data",
      "region": "us-east-1",
      "curReportName": "mantl-cur"
    }
  }'
```

## FinOps Best Practices

### 1. Tagging Strategy

Tag all resources for cost allocation:

```yaml
# Deployment with standard tags
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  labels:
    app: myapp
    team: platform
    environment: production
    cost-center: engineering
    owner: platform-team@example.com
```

```bash
# Tag EC2 instances automatically (Karpenter)
kubectl patch ec2nodeclass default --type=merge -p '{
  "spec": {
    "tags": {
      "Team": "platform",
      "Environment": "production",
      "CostCenter": "engineering",
      "ManagedBy": "karpenter"
    }
  }
}'
```

### 2. Chargeback/Showback

```bash
# Generate monthly cost report per team
kubectl exec -n kubecost deploy/kubecost-cost-analyzer -- \
  /cost-model allocation \
  --window 30d \
  --aggregate namespace \
  --format csv > team-costs.csv
```

### 3. Cost Anomaly Detection

```yaml
# Prometheus alert for cost spikes
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cost-alerts
  namespace: monitoring
spec:
  groups:
  - name: cost
    interval: 1h
    rules:
    - alert: CostAnomalyDetected
      expr: |
        (
          sum by (namespace) (
            rate(kubecost_pod_cost_hourly[1h])
          ) >
          sum by (namespace) (
            rate(kubecost_pod_cost_hourly[1h] offset 7d)
          ) * 1.5
        )
      for: 2h
      annotations:
        summary: "Cost spike in {{ $labels.namespace }}"
        description: "Namespace {{ $labels.namespace }} cost increased >50% vs last week"
```

### 4. Reserved Capacity Planning

```bash
# Analyze 30-day usage patterns
kubectl exec -n kubecost deploy/kubecost-cost-analyzer -- \
  /cost-model savings/computeRecommendations \
  --window 30d

# Output shows RI/Savings Plan recommendations
```

## Quick Wins

### Immediate Cost Reductions (No Code Changes)

**1. Enable Cluster Autoscaler** (5-30% savings)
```bash
kubectl apply -f platform/base/cluster-autoscaler/
```

**2. Switch to gp3 Storage** (20% savings)
```bash
kubectl patch storageclass gp2 -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp3
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  iops: "3000"
  throughput: "125"
allowVolumeExpansion: true
EOF
```

**3. Enable Karpenter Consolidation** (10-40% savings)
```bash
kubectl apply -f infrastructure/karpenter/nodepool-default.yaml
# Karpenter will automatically replace nodes with smaller instances
```

**4. Delete Unused Load Balancers**
```bash
# Find unused services with LoadBalancer type
kubectl get svc -A -o json | jq -r '.items[] |
  select(.spec.type=="LoadBalancer") |
  select(.status.loadBalancer.ingress == null) |
  "\(.metadata.namespace)/\(.metadata.name)"'
```

**5. Set Up VPC Endpoints** (reduce NAT Gateway costs)
```bash
# See terraform/modules/vpc/endpoints.tf
terraform apply -target=module.vpc_endpoints
```

### Monthly Cost Checklist

- [ ] Review Kubecost recommendations weekly
- [ ] Check for unused resources (PVs, LBs, unused nodes)
- [ ] Analyze spot interruption rate (<5% is healthy)
- [ ] Review pod right-sizing recommendations from VPA
- [ ] Verify autoscaling is working (nodes scale down)
- [ ] Check CloudWatch/Stackdriver logs retention (default 30d is excessive)
- [ ] Review S3/GCS bucket sizes and lifecycle policies
- [ ] Audit IAM roles and remove unused credentials
- [ ] Check for forgotten dev/test clusters

## Cost Savings Summary

| Strategy | Effort | Savings | Timeframe |
|----------|--------|---------|-----------|
| Spot instances | Low | 60-90% | Immediate |
| Karpenter consolidation | Low | 10-40% | Hours |
| gp3 storage | Low | 20% | Immediate |
| Right-sizing pods | Medium | 20-50% | Days |
| Reserved instances | Low | 30-60% | 1-3 years |
| NAT instance vs Gateway | Medium | 70% | Days |
| Shared ALB | Low | $30-100/mo | Hours |
| VPC endpoints | Low | $10-50/mo | Hours |
| Graviton instances | Medium | 20-40% | Weeks |
| Multi-tenancy | High | 40-70% | Months |

**Expected Total Savings**: 40-70% of infrastructure costs with full implementation

## Additional Resources

- [AWS Cost Optimization](https://aws.amazon.com/pricing/cost-optimization/)
- [GCP Cost Management](https://cloud.google.com/cost-management)
- [Azure Cost Management](https://azure.microsoft.com/en-us/products/cost-management/)
- [Kubecost Documentation](https://docs.kubecost.com/)
- [FinOps Foundation](https://www.finops.org/)
- [CNCF FinOps for Kubernetes](https://www.cncf.io/blog/2021/06/29/finops-for-kubernetes/)
