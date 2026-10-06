# Troubleshooting Guide

> Historical component-stack procedure examples. These commands and objectives
> have not established acceptance for the current beta runtime. Use the maintained
> [operations guide](../runbooks.md) and [maturity table](../../README.md#capabilities-and-maturity)
> first. Any recovery objectives here are proposed targets, not measured guarantees.


Comprehensive troubleshooting procedures for common mantl platform issues.

## Overview

This guide covers:
- Pod issues (CrashLoopBackOff, ImagePullBackOff, OOMKilled)
- Networking issues (DNS, Ingress, NetworkPolicy)
- Storage issues (PVC binding, disk pressure)
- Performance issues (CPU throttling, memory pressure)
- Application debugging techniques

## Quick Diagnostic Commands

```bash
# Cluster health overview
kubectl get nodes
kubectl get pods -A | grep -v Running
kubectl top nodes
kubectl top pods -A --sort-by=memory

# Recent events (all namespaces)
kubectl get events -A --sort-by='.lastTimestamp' | tail -50

# Resource pressure
kubectl describe nodes | grep -A 5 "Allocated resources"

# Certificate status
kubectl get certificate -A

# Service mesh status (if using)
kubectl get pods -n istio-system
```

## Pod Issues

### CrashLoopBackOff

**Symptom**: Pod repeatedly crashes and restarts

**Diagnosis**:

```bash
# Get pod details
kubectl describe pod <pod-name> -n <namespace>

# Check current logs
kubectl logs <pod-name> -n <namespace>

# Check previous crash logs
kubectl logs <pod-name> -n <namespace> --previous

# Check for init container failures
kubectl logs <pod-name> -n <namespace> -c <init-container-name>

# Get events for this pod
kubectl get events -n <namespace> --field-selector involvedObject.name=<pod-name>
```

**Common Causes & Solutions**:

1. **Application crash on startup**:
```bash
# Check application configuration
kubectl get configmap -n <namespace>
kubectl get secret -n <namespace>

# Verify environment variables
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 20 env:

# Check for missing dependencies
kubectl exec -it <pod-name> -n <namespace> -- curl http://dependency-service:8080/health
```

2. **Database connection failure**:
```bash
# Verify database is accessible
kubectl exec -it <pod-name> -n <namespace> -- \
  nc -zv postgres-service 5432

# Check database credentials
kubectl get secret db-credentials -n <namespace> -o yaml | \
  grep password | awk '{print $2}' | base64 -d

# Test database connection
kubectl exec -it <pod-name> -n <namespace> -- \
  psql postgresql://user:pass@host:5432/dbname -c "SELECT 1"
```

3. **Health check failing too quickly**:
```bash
# Adjust initialDelaySeconds
kubectl patch deployment <deployment-name> -n <namespace> --type=json -p='[
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/livenessProbe/initialDelaySeconds",
    "value": 60
  }
]'
```

### ImagePullBackOff

**Symptom**: Cannot pull container image

**Diagnosis**:

```bash
# Check image pull errors
kubectl describe pod <pod-name> -n <namespace> | grep -A 10 "Events:"

# Verify image exists
docker pull <image-name>:<tag>

# Check imagePullSecrets
kubectl get pod <pod-name> -n <namespace> -o yaml | grep imagePullSecrets -A 5
```

**Common Causes & Solutions**:

1. **Image doesn't exist or wrong tag**:
```bash
# List available tags
aws ecr describe-images --repository-name myapp --query 'imageDetails[*].imageTags' --output table

# Update deployment with correct tag
kubectl set image deployment/<deployment-name> <container-name>=<image>:<correct-tag> -n <namespace>
```

2. **Missing credentials**:
```bash
# Create docker-registry secret
kubectl create secret docker-registry ecr-credentials \
  --docker-server=123456789.dkr.ecr.us-east-1.amazonaws.com \
  --docker-username=AWS \
  --docker-password=$(aws ecr get-login-password --region us-east-1) \
  -n <namespace>

# Patch serviceaccount to use secret
kubectl patch serviceaccount default -n <namespace> -p '{"imagePullSecrets":[{"name":"ecr-credentials"}]}'
```

3. **Rate limiting (Docker Hub)**:
```bash
# Use authenticated pulls
kubectl create secret docker-registry dockerhub \
  --docker-server=docker.io \
  --docker-username=<username> \
  --docker-password=<password> \
  -n <namespace>

# Or migrate to different registry
kubectl set image deployment/<deployment-name> \
  <container>=ghcr.io/<org>/<image>:<tag> -n <namespace>
```

### OOMKilled (Out of Memory)

**Symptom**: Pod killed due to memory limit exceeded

**Diagnosis**:

```bash
# Check if pod was OOMKilled
kubectl describe pod <pod-name> -n <namespace> | grep -A 5 "Last State"

# Check current memory usage
kubectl top pod <pod-name> -n <namespace>

# Check memory limits
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 10 resources:
```

**Solutions**:

```bash
# Increase memory limits
kubectl patch deployment <deployment-name> -n <namespace> --type=json -p='[
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/resources/limits/memory",
    "value": "2Gi"
  },
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/resources/requests/memory",
    "value": "1Gi"
  }
]'

# Enable VPA for automatic right-sizing
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
    updateMode: "Auto"
EOF

# Investigate memory leak (if applicable)
kubectl exec -it <pod-name> -n <namespace> -- curl http://localhost:8080/debug/pprof/heap > heap.prof
```

### Pending (Cannot be scheduled)

**Symptom**: Pod stuck in Pending state

**Diagnosis**:

```bash
# Check why pod is pending
kubectl describe pod <pod-name> -n <namespace>

# Check node resources
kubectl describe nodes | grep -A 5 "Allocated resources"

# Check if waiting for PVC
kubectl get pvc -n <namespace>
```

**Common Causes & Solutions**:

1. **Insufficient resources**:
```bash
# Enable cluster autoscaler or add nodes manually
aws eks update-nodegroup-config \
  --cluster-name mantl-prod \
  --nodegroup-name workers \
  --scaling-config desiredSize=10

# Or reduce resource requests
kubectl patch deployment <deployment-name> -n <namespace> --type=json -p='[
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/resources/requests/cpu",
    "value": "100m"
  }
]'
```

2. **Node affinity not satisfied**:
```bash
# Check node labels
kubectl get nodes --show-labels

# Remove or adjust node affinity
kubectl patch deployment <deployment-name> -n <namespace> --type=json -p='[
  {
    "op": "remove",
    "path": "/spec/template/spec/affinity"
  }
]'
```

3. **Waiting for PVC**:
```bash
# Check PVC status
kubectl describe pvc <pvc-name> -n <namespace>

# Check storage class
kubectl get storageclass

# Create PV manually if needed (for static provisioning)
kubectl apply -f pv.yaml
```

## Networking Issues

### DNS Resolution Failing

**Symptom**: Services cannot resolve each other by name

**Diagnosis**:

```bash
# Test DNS from within cluster
kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup kubernetes.default

# Check CoreDNS pods
kubectl get pods -n kube-system -l k8s-app=kube-dns

# Check CoreDNS logs
kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50

# Check DNS configuration
kubectl get configmap coredns -n kube-system -o yaml
```

**Solutions**:

```bash
# Restart CoreDNS
kubectl rollout restart deployment coredns -n kube-system

# Increase CoreDNS replicas if overloaded
kubectl scale deployment coredns -n kube-system --replicas=4

# Update CoreDNS config to increase cache
kubectl patch configmap coredns -n kube-system --type=merge -p='
data:
  Corefile: |
    .:53 {
        cache 300  # Increased cache TTL
        errors
        health
        kubernetes cluster.local in-addr.arpa ip6.arpa {
          pods insecure
          fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        reload
        loadbalance
    }
'
```

### Ingress Not Working

**Symptom**: Cannot access application via ingress URL

**Diagnosis**:

```bash
# Check ingress resource
kubectl get ingress -n <namespace>
kubectl describe ingress <ingress-name> -n <namespace>

# Check ingress controller
kubectl get pods -n ingress-nginx
kubectl logs -n ingress-nginx deploy/ingress-nginx-controller

# Verify service and endpoints
kubectl get svc -n <namespace>
kubectl get endpoints <service-name> -n <namespace>

# Check certificate (if using TLS)
kubectl get certificate -n <namespace>
kubectl describe certificate <cert-name> -n <namespace>
```

**Solutions**:

1. **Service has no endpoints**:
```bash
# Check pod labels match service selector
kubectl get pods -n <namespace> --show-labels
kubectl get svc <service-name> -n <namespace> -o yaml | grep -A 5 selector:

# Fix label mismatch
kubectl label pods -n <namespace> -l app=myapp version=v1
```

2. **Certificate issues**:
```bash
# Force certificate renewal
kubectl delete certificate <cert-name> -n <namespace>

# Check cert-manager logs
kubectl logs -n cert-manager deploy/cert-manager

# Manually create certificate (if cert-manager failing)
kubectl create secret tls <tls-secret> \
  --cert=path/to/cert.crt \
  --key=path/to/cert.key \
  -n <namespace>
```

3. **Ingress controller not routing**:
```bash
# Check ingress class
kubectl get ingressclass

# Update ingress to use correct class
kubectl annotate ingress <ingress-name> -n <namespace> \
  kubernetes.io/ingress.class=nginx --overwrite

# Or use spec.ingressClassName (newer API)
kubectl patch ingress <ingress-name> -n <namespace> --type=merge -p '
spec:
  ingressClassName: nginx
'
```

### NetworkPolicy Blocking Traffic

**Symptom**: Application cannot connect to dependencies

**Diagnosis**:

```bash
# List network policies
kubectl get networkpolicy -n <namespace>

# Describe specific policy
kubectl describe networkpolicy <policy-name> -n <namespace>

# Test connectivity
kubectl exec -it <pod-name> -n <namespace> -- curl http://target-service:8080

# Check if policy is blocking
kubectl auth can-i create pods --as=system:serviceaccount:<namespace>:<serviceaccount> -n <namespace>
```

**Solutions**:

```bash
# Temporarily disable network policy (for testing only!)
kubectl delete networkpolicy <policy-name> -n <namespace>

# Add egress rule for specific service
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-to-database
  namespace: <namespace>
spec:
  podSelector:
    matchLabels:
      app: myapp
  policyTypes:
  - Egress
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: database
    ports:
    - protocol: TCP
      port: 5432
EOF

# Allow all egress (less secure, but useful for debugging)
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-all-egress
  namespace: <namespace>
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - {}
EOF
```

## Storage Issues

### PVC Not Binding

**Symptom**: PersistentVolumeClaim stuck in Pending

**Diagnosis**:

```bash
# Check PVC status
kubectl describe pvc <pvc-name> -n <namespace>

# Check available PVs
kubectl get pv

# Check storage class
kubectl get storageclass
kubectl describe storageclass <storage-class-name>

# Check provisioner logs (CSI driver)
kubectl logs -n kube-system -l app=ebs-csi-controller
```

**Solutions**:

1. **No storage class**:
```bash
# Create storage class
kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp3
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  encrypted: "true"
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
EOF

# Update PVC to use storage class
kubectl patch pvc <pvc-name> -n <namespace> -p '{"spec":{"storageClassName":"gp3"}}'
```

2. **Insufficient permissions**:
```bash
# Check IAM role for CSI driver
aws iam get-role --role-name AmazonEKS_EBS_CSI_DriverRole

# Attach required policy
aws iam attach-role-policy \
  --role-name AmazonEKS_EBS_CSI_DriverRole \
  --policy-arn arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy
```

3. **Volume size mismatch**:
```bash
# Increase PVC size (requires allowVolumeExpansion: true)
kubectl patch pvc <pvc-name> -n <namespace> -p '{"spec":{"resources":{"requests":{"storage":"200Gi"}}}}'
```

### Disk Pressure on Node

**Symptom**: Node marked with DiskPressure condition

**Diagnosis**:

```bash
# Check node conditions
kubectl describe node <node-name> | grep -A 10 Conditions:

# Check disk usage on node
kubectl debug node/<node-name> -it --image=ubuntu
df -h
du -sh /var/lib/containerd/* | sort -h

# Check pod disk usage
kubectl top pods -n <namespace> --containers
```

**Solutions**:

```bash
# Clean up old images
kubectl debug node/<node-name> -it --image=ubuntu
crictl rmi --prune

# Clean up old containers
crictl rm $(crictl ps -a -q)

# Increase node disk size (AWS example)
aws ec2 modify-volume \
  --volume-id vol-xxx \
  --size 200

# Expand filesystem (on node)
kubectl debug node/<node-name> -it --image=ubuntu
growpart /dev/nvme0n1 1
resize2fs /dev/nvme0n1p1

# Set image garbage collection limits (kubelet config)
kubectl patch kubeletconfig -n kube-system -p '
imageGCHighThresholdPercent: 85
imageGCLowThresholdPercent: 80
'
```

## Performance Issues

### CPU Throttling

**Symptom**: Application slow despite low CPU usage

**Diagnosis**:

```bash
# Check CPU throttling
kubectl top pods -n <namespace>

# Check CPU limits vs usage
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 10 resources:

# Check for throttling in metrics
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/namespaces/<namespace>/pods/<pod-name>" | jq .
```

**Solutions**:

```bash
# Increase CPU limits
kubectl patch deployment <deployment-name> -n <namespace> --type=json -p='[
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/resources/limits/cpu",
    "value": "2000m"
  }
]'

# Remove CPU limits entirely (allow bursting)
kubectl patch deployment <deployment-name> -n <namespace> --type=json -p='[
  {
    "op": "remove",
    "path": "/spec/template/spec/containers/0/resources/limits/cpu"
  }
]'

# Use VPA to automatically adjust
kubectl apply -f vpa.yaml
```

### Database Connection Pool Exhaustion

**Symptom**: "Too many connections" or connection timeouts

**Diagnosis**:

```bash
# Check active connections
kubectl exec -n <namespace> <postgres-pod> -- \
  psql -c "SELECT count(*) FROM pg_stat_activity;"

# Check connection limit
kubectl exec -n <namespace> <postgres-pod> -- \
  psql -c "SHOW max_connections;"

# Check application connection pool config
kubectl get deployment <app-deployment> -n <namespace> -o yaml | grep -i pool
```

**Solutions**:

```bash
# Increase PostgreSQL max_connections
kubectl patch cluster <cluster-name> -n <namespace> --type=merge -p '
spec:
  postgresql:
    parameters:
      max_connections: "500"
'

# Add PgBouncer connection pooler
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
EOF

# Update application to use pooler
kubectl set env deployment/<app-name> -n <namespace> \
  DATABASE_URL="postgresql://user:pass@<cluster-name>-pooler-rw:5432/db"
```

## Application Debugging

### Enable Debug Logging

```bash
# Update environment variable
kubectl set env deployment/<deployment-name> -n <namespace> LOG_LEVEL=debug

# Watch logs with filtering
kubectl logs -n <namespace> -l app=myapp -f | grep ERROR

# Save logs for analysis
kubectl logs -n <namespace> <pod-name> --since=1h > debug.log
```

### Interactive Debugging

```bash
# Exec into running container
kubectl exec -it <pod-name> -n <namespace> -- /bin/bash

# Run debug sidecar
kubectl debug <pod-name> -n <namespace> -it --image=ubuntu --target=<container-name>

# Copy files from pod
kubectl cp <namespace>/<pod-name>:/app/debug.log ./debug.log

# Port forward for local debugging
kubectl port-forward -n <namespace> <pod-name> 8080:8080
```

### Performance Profiling

```bash
# CPU profiling (Go apps with pprof)
kubectl port-forward -n <namespace> <pod-name> 6060:6060
go tool pprof http://localhost:6060/debug/pprof/profile

# Heap profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Python profiling (with py-spy)
kubectl exec -it <pod-name> -n <namespace> -- \
  py-spy record -o profile.svg -- python app.py

# Java profiling (with async-profiler)
kubectl exec -it <pod-name> -n <namespace> -- \
  /opt/async-profiler/profiler.sh -d 30 -f /tmp/profile.html <pid>
```

## Getting Help

If you've tried these troubleshooting steps and still need help:

1. **Gather diagnostic information**:
```bash
# Create diagnostics bundle
kubectl cluster-info dump --output-directory=/tmp/cluster-diagnostics

# Export pod logs
kubectl logs <pod-name> -n <namespace> --previous > /tmp/pod-logs.txt

# Export events
kubectl get events -A --sort-by='.lastTimestamp' > /tmp/events.txt
```

2. **Contact support**:
   - Platform team: platform-team@example.com
   - PagerDuty: Escalate to on-call engineer
   - Vendor support: Include cluster diagnostics bundle

3. **Check documentation**:
   - Internal wiki: https://wiki.example.com/mantl
   - Kubernetes docs: https://kubernetes.io/docs/
   - Component-specific docs (cert-manager, CloudNativePG, etc.)

## Additional Resources

- [Kubernetes Debugging Documentation](https://kubernetes.io/docs/tasks/debug/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [CloudNativePG Troubleshooting](https://cloudnative-pg.io/documentation/current/troubleshooting/)
