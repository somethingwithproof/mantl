# Upgrade Procedures Runbook

Procedures for upgrading mantl platform components, applications, and Kubernetes itself.

## Overview

This runbook covers:
- Kubernetes cluster upgrades
- Platform component upgrades
- Application upgrades
- Database upgrades
- Rollback procedures

## Pre-Upgrade Checklist

Before any upgrade:

- [ ] Review release notes and breaking changes
- [ ] Create full backup with Velero
- [ ] Verify backup restoration in DR environment
- [ ] Notify stakeholders of maintenance window
- [ ] Ensure monitoring and alerting are active
- [ ] Have rollback plan ready
- [ ] Test upgrade in non-production environment first

## Kubernetes Cluster Upgrade

### EKS Cluster Upgrade

```bash
# Check current version
kubectl version --short

# List available versions
aws eks describe-addon-versions --kubernetes-version 1.28

# Upgrade control plane (managed by AWS)
aws eks update-cluster-version \
  --name mantl-prod \
  --kubernetes-version 1.29

# Monitor upgrade progress
aws eks describe-update \
  --name mantl-prod \
  --update-id <update-id>

# Wait for control plane upgrade to complete
aws eks wait cluster-active --name mantl-prod

# Verify control plane version
aws eks describe-cluster --name mantl-prod | jq -r '.cluster.version'
```

### Node Group Upgrade

```bash
# Create new node group with new version
aws eks create-nodegroup \
  --cluster-name mantl-prod \
  --nodegroup-name workers-1-29 \
  --subnets subnet-xxx subnet-yyy \
  --instance-types t3.xlarge \
  --scaling-config minSize=3,maxSize=20,desiredSize=5 \
  --kubernetes-version 1.29 \
  --node-role arn:aws:iam::ACCOUNT:role/EKSNodeRole

# Wait for new nodes to join
kubectl get nodes -l eks.amazonaws.com/nodegroup=workers-1-29 --watch

# Cordon old nodes (prevent new pods)
OLD_NODEGROUP="workers-1-28"
for node in $(kubectl get nodes -l eks.amazonaws.com/nodegroup=$OLD_NODEGROUP -o name); do
  kubectl cordon $node
done

# Drain nodes one by one (migrate pods)
for node in $(kubectl get nodes -l eks.amazonaws.com/nodegroup=$OLD_NODEGROUP -o name); do
  echo "Draining $node"
  kubectl drain $node \
    --ignore-daemonsets \
    --delete-emptydir-data \
    --grace-period=300 \
    --timeout=10m

  # Wait between nodes to avoid service disruption
  sleep 60

  # Verify pods are running
  kubectl get pods -A | grep -v Running
done

# Verify all pods are healthy
kubectl get pods -A -o wide

# Delete old node group
aws eks delete-nodegroup \
  --cluster-name mantl-prod \
  --nodegroup-name workers-1-28
```

### Upgrade Core Addons

```bash
# VPC CNI
aws eks update-addon \
  --cluster-name mantl-prod \
  --addon-name vpc-cni \
  --addon-version v1.15.1-eksbuild.1 \
  --resolve-conflicts OVERWRITE

# CoreDNS
aws eks update-addon \
  --cluster-name mantl-prod \
  --addon-name coredns \
  --addon-version v1.10.1-eksbuild.6 \
  --resolve-conflicts OVERWRITE

# kube-proxy
aws eks update-addon \
  --cluster-name mantl-prod \
  --addon-name kube-proxy \
  --addon-version v1.29.0-eksbuild.1 \
  --resolve-conflicts OVERWRITE

# Wait for addons to be active
aws eks describe-addon \
  --cluster-name mantl-prod \
  --addon-name vpc-cni | jq -r '.addon.status'
```

## Platform Components Upgrade

### Flux System Upgrade

```bash
# Check current version
flux version

# Check for available updates
flux check --pre

# Suspend reconciliation during upgrade
flux suspend kustomization --all

# Upgrade Flux CLI
curl -s https://fluxcd.io/install.sh | sudo bash

# Upgrade Flux controllers in cluster
flux install \
  --export > /tmp/flux-system-upgrade.yaml

kubectl apply -f /tmp/flux-system-upgrade.yaml

# Resume reconciliation
flux resume kustomization --all

# Verify upgrade
flux version
kubectl get pods -n flux-system
```

### Cert-Manager Upgrade

```bash
# Check current version
kubectl get deployment cert-manager -n cert-manager -o yaml | grep image:

# Suspend issuance during upgrade
kubectl annotate clusterissuer letsencrypt-prod \
  cert-manager.io/issue-temporary-certificate-

# Upgrade via Helm
helm repo update
helm upgrade cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --version v1.14.0 \
  --reuse-values

# Verify upgrade
kubectl rollout status deployment/cert-manager -n cert-manager
kubectl get pods -n cert-manager

# Resume issuance
kubectl annotate clusterissuer letsencrypt-prod \
  cert-manager.io/issue-temporary-certificate--
```

### CloudNativePG Operator Upgrade

```bash
# Check current version
kubectl get deployment cnpg-controller-manager -n cnpg-system -o yaml | grep image:

# CloudNativePG supports in-place upgrades
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.22/releases/cnpg-1.22.0.yaml

# Verify operator upgrade
kubectl get pods -n cnpg-system
kubectl logs -n cnpg-system deploy/cnpg-controller-manager | tail -20

# Database clusters will be automatically reconciled
# Monitor cluster health
kubectl get cluster -A
```

### Prometheus Stack Upgrade

```bash
# Check current version
helm list -n mantl-system | grep prometheus-stack

# Download new values
helm show values prometheus-community/kube-prometheus-stack > /tmp/new-values.yaml

# Compare with current values
helm get values prometheus-stack -n mantl-system > /tmp/current-values.yaml
diff /tmp/current-values.yaml /tmp/new-values.yaml

# Upgrade (CRDs must be updated separately)
kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.71.0/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml

helm upgrade prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace mantl-system \
  --version 56.0.0 \
  --reuse-values

# Verify upgrade
kubectl get pods -n mantl-system | grep prometheus
kubectl get prometheus -n mantl-system
```

### Kyverno Upgrade

```bash
# Kyverno upgrades require careful planning due to webhook validation
# Suspend webhook temporarily during upgrade
kubectl annotate validatingwebhookconfigurations kyverno-resource-validating-webhook-cfg \
  kyverno.io/background-scan=disable

# Upgrade via Helm
helm upgrade kyverno kyverno/kyverno \
  --namespace kyverno \
  --version 3.1.0 \
  --reuse-values

# Wait for new pods
kubectl rollout status deployment/kyverno -n kyverno

# Re-enable webhook
kubectl annotate validatingwebhookconfigurations kyverno-resource-validating-webhook-cfg \
  kyverno.io/background-scan-

# Verify policies still work
kubectl run test-nginx --image=nginx:latest --dry-run=server
```

## Application Upgrades

### Blue-Green Deployment Strategy

```bash
# Tag current production as "blue"
kubectl label deployment myapp -n production version=blue --overwrite

# Deploy new version as "green" (separate deployment)
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-green
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
      version: green
  template:
    metadata:
      labels:
        app: myapp
        version: green
    spec:
      containers:
      - name: myapp
        image: myregistry/myapp:v2.0.0
        ports:
        - containerPort: 8080
EOF

# Wait for green deployment to be ready
kubectl wait --for=condition=Available deployment/myapp-green -n production --timeout=5m

# Run smoke tests against green deployment
kubectl port-forward -n production deployment/myapp-green 8080:8080 &
curl http://localhost:8080/health

# Switch service to green
kubectl patch service myapp -n production -p '{"spec":{"selector":{"version":"green"}}}'

# Monitor for 10 minutes
watch kubectl get pods -n production -l app=myapp

# If successful, scale down blue
kubectl scale deployment myapp -n production --replicas=0

# If issues, rollback to blue
kubectl patch service myapp -n production -p '{"spec":{"selector":{"version":"blue"}}}'
kubectl scale deployment myapp -n production --replicas=3
```

### Canary Deployment with Flagger

```bash
# Initialize canary
kubectl apply -f - <<EOF
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: myapp
  namespace: production
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  service:
    port: 8080
  analysis:
    interval: 1m
    threshold: 10
    maxWeight: 50
    stepWeight: 5
    metrics:
    - name: request-success-rate
      thresholdRange:
        min: 99
      interval: 1m
    - name: request-duration
      thresholdRange:
        max: 500
      interval: 1m
  webhooks:
  - name: load-test
    url: http://flagger-loadtester/
    timeout: 5s
    metadata:
      cmd: "hey -z 1m -q 10 -c 2 http://myapp-canary:8080/"
EOF

# Update deployment image (Flagger will manage rollout)
kubectl set image deployment/myapp myapp=myregistry/myapp:v2.0.0 -n production

# Monitor canary progress
watch kubectl get canary -n production

# Flagger will:
# 1. Create canary deployment
# 2. Gradually shift 5% -> 10% -> 15% ... -> 50% traffic
# 3. Run automated tests at each step
# 4. Promote to 100% if metrics pass
# 5. Rollback automatically if metrics fail
```

## Database Upgrades

### PostgreSQL Minor Version Upgrade (CloudNativePG)

```bash
# CloudNativePG handles minor version upgrades automatically
# Update image tag in cluster spec
kubectl patch cluster myapp-db -n production --type=merge -p '
spec:
  imageName: ghcr.io/cloudnative-pg/postgresql:16.2
'

# Operator will rolling restart pods with new version
kubectl get pods -n production -l cnpg.io/cluster=myapp-db --watch

# Verify new version
kubectl exec -n production myapp-db-1 -- psql -c "SELECT version();"
```

### PostgreSQL Major Version Upgrade

```bash
# Major upgrades require pg_upgrade
# 1. Create backup
kubectl cnpg backup myapp-db -n production

# 2. Create new cluster with new version
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: myapp-db-v16
  namespace: production
spec:
  instances: 3
  imageName: ghcr.io/cloudnative-pg/postgresql:16.2
  bootstrap:
    initdb:
      database: myapp
      owner: myapp
  storage:
    size: 100Gi
EOF

# 3. Wait for new cluster
kubectl wait --for=condition=Ready cluster/myapp-db-v16 -n production --timeout=10m

# 4. Dump from old cluster
kubectl exec -n production myapp-db-1 -- \
  pg_dump -Fc myapp > /tmp/myapp-db-dump.sql

# 5. Restore to new cluster
kubectl exec -i -n production myapp-db-v16-1 -- \
  pg_restore -d myapp < /tmp/myapp-db-dump.sql

# 6. Verify data
kubectl exec -n production myapp-db-v16-1 -- \
  psql -d myapp -c "SELECT count(*) FROM important_table;"

# 7. Update application to point to new cluster
kubectl set env deployment/myapp -n production \
  DATABASE_URL="postgresql://myapp:password@myapp-db-v16-rw:5432/myapp"

# 8. Monitor for issues
kubectl logs -n production deployment/myapp -f

# 9. If successful, delete old cluster
kubectl delete cluster myapp-db -n production
```

## Rollback Procedures

### Application Rollback (Deployment)

```bash
# View rollout history
kubectl rollout history deployment/myapp -n production

# Rollback to previous version
kubectl rollout undo deployment/myapp -n production

# Rollback to specific revision
kubectl rollout undo deployment/myapp -n production --to-revision=3

# Verify rollback
kubectl rollout status deployment/myapp -n production
kubectl get pods -n production -l app=myapp
```

### Helm Rollback

```bash
# View release history
helm history myapp -n production

# Rollback to previous release
helm rollback myapp -n production

# Rollback to specific revision
helm rollback myapp 5 -n production

# Verify rollback
helm status myapp -n production
kubectl get pods -n production
```

### Cluster Rollback (Snapshot Restore)

```bash
# If cluster upgrade fails catastrophically
# Restore from pre-upgrade backup

# List backups
velero backup get

# Restore entire cluster
velero restore create cluster-rollback \
  --from-backup pre-upgrade-1-29 \
  --wait

# Verify restoration
kubectl get nodes
kubectl get pods -A
```

## Post-Upgrade Verification

### Verification Checklist

- [ ] All nodes showing Ready
- [ ] All pods in Running state
- [ ] No CrashLoopBackOff or ImagePullBackOff
- [ ] All PVCs bound
- [ ] Ingress routing working
- [ ] Certificates valid and not expiring soon
- [ ] Monitoring collecting metrics
- [ ] Logs flowing to central logging
- [ ] Backups completing successfully
- [ ] No critical alerts firing
- [ ] Application smoke tests passing
- [ ] Performance metrics within normal ranges

### Automated Verification Script

```bash
#!/bin/bash
set -e

echo "=== Post-Upgrade Verification ==="

# Check nodes
echo "Checking nodes..."
NOTREADY=$(kubectl get nodes --no-headers | grep -cv " Ready ")
if [ "$NOTREADY" -gt 0 ]; then
  echo "ERROR: $NOTREADY nodes not ready"
  exit 1
fi
echo "✓ All nodes ready"

# Check pod status
echo "Checking pod status..."
FAILED=$(kubectl get pods -A --no-headers | grep -cv " Running ")
if [ "$FAILED" -gt 0 ]; then
  echo "WARNING: $FAILED pods not running"
  kubectl get pods -A | grep -v " Running "
fi

# Check PVCs
echo "Checking PVCs..."
UNBOUND=$(kubectl get pvc -A --no-headers | grep -cv " Bound ")
if [ "$UNBOUND" -gt 0 ]; then
  echo "ERROR: $UNBOUND PVCs not bound"
  exit 1
fi
echo "✓ All PVCs bound"

# Check certificates
echo "Checking certificates..."
NOTREADY=$(kubectl get certificate -A --no-headers | grep -cv " True ")
if [ "$NOTREADY" -gt 0 ]; then
  echo "WARNING: $NOTREADY certificates not ready"
  kubectl get certificate -A | grep -v " True "
fi

# Check critical services
echo "Checking critical services..."
for ns in mantl-system kube-system; do
  NOTREADY=$(kubectl get pods -n $ns --no-headers | grep -cv " Running ")
  if [ "$NOTREADY" -gt 0 ]; then
    echo "ERROR: $NOTREADY pods not running in $ns"
    exit 1
  fi
done
echo "✓ Critical services running"

# Run smoke tests
echo "Running smoke tests..."
kubectl apply -f tests/smoke-tests.yaml
kubectl wait --for=condition=Complete job/smoke-test --timeout=5m
RESULT=$(kubectl logs job/smoke-test | grep -c "PASS")
if [ "$RESULT" -lt 1 ]; then
  echo "ERROR: Smoke tests failed"
  kubectl logs job/smoke-test
  exit 1
fi
echo "✓ Smoke tests passed"

echo "=== Verification Complete ==="
```

## Upgrade Schedule

### Recommended Schedule

- **Security patches**: Within 7 days of release
- **Minor versions**: Quarterly (after testing in dev/staging)
- **Major versions**: Biannually (with extensive testing)
- **Emergency patches**: Immediate (with expedited testing)

### Maintenance Windows

- **Production**: Sunday 02:00-06:00 UTC (low traffic period)
- **Staging**: Anytime (notify team first)
- **Development**: Anytime

## Communication Plan

### Before Upgrade

```
Subject: Scheduled Maintenance - Kubernetes Cluster Upgrade

Date: Sunday, March 15, 2026
Time: 02:00-06:00 UTC
Expected Downtime: < 5 minutes (rolling updates)

Changes:
- Kubernetes version: 1.28 → 1.29
- Node group upgrade with zero-downtime migration
- Core addon updates (VPC CNI, CoreDNS, kube-proxy)

Impact:
- Brief pod restarts during node migration
- No expected user-facing downtime

Rollback Plan:
- Pre-upgrade backup created
- Old node group retained until verification complete
- Can restore from backup within 30 minutes if needed

Contact: platform-team@example.com
Status Page: https://status.example.com
```

### After Upgrade

```
Subject: Maintenance Complete - Kubernetes Cluster Upgrade

Status: ✅ Complete
Duration: 3h 42m (under 4h window)
Downtime: 0 minutes

Changes Applied:
- ✅ Kubernetes upgraded: 1.28 → 1.29
- ✅ All nodes migrated to new version
- ✅ All addons updated
- ✅ All verification tests passed

Verification:
- All services operational
- Performance metrics normal
- No alerts firing

Next Steps:
- Monitor for 24 hours
- Old node group retained for 7 days
- Full post-mortem scheduled for March 17

Thank you for your patience.
```

## Additional Resources

- [Kubernetes Upgrade Documentation](https://kubernetes.io/docs/tasks/administer-cluster/cluster-upgrade/)
- [EKS Upgrade Guide](https://docs.aws.amazon.com/eks/latest/userguide/update-cluster.html)
- [CloudNativePG Upgrade Procedures](https://cloudnative-pg.io/documentation/current/upgrade/)
- [Flux Upgrade Guide](https://fluxcd.io/flux/installation/upgrade/)
