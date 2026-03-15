# Disaster Recovery Runbook

Complete disaster recovery procedures for mantl platform.

## Prerequisites

- Access to management cluster
- Velero CLI installed
- Backup storage credentials
- DR cluster available (if applicable)
- PagerDuty/Slack access for notifications

## RTO/RPO Objectives

| Component | RTO (Recovery Time Objective) | RPO (Recovery Point Objective) |
|-----------|-------------------------------|-------------------------------|
| Platform Services | 15 minutes | 1 hour |
| Application Data | 30 minutes | 15 minutes (continuous backup) |
| Monitoring Data | 1 hour | 5 minutes (federated) |
| Complete Cluster | 2 hours | 1 hour |

## Scenario 1: Single Pod Failure

**Detection**: Pod in CrashLoopBackOff or Error state

**Recovery**:

```bash
# Identify failing pod
kubectl get pods -A | grep -v Running

# Get pod details
kubectl describe pod <pod-name> -n <namespace>

# Check logs
kubectl logs <pod-name> -n <namespace> --previous

# Common fixes:
# 1. Resource constraints
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 5 resources

# 2. ConfigMap/Secret missing
kubectl get configmaps,secrets -n <namespace>

# 3. Image pull issues
kubectl get events -n <namespace> | grep <pod-name>

# Force pod restart
kubectl delete pod <pod-name> -n <namespace>

# Verify recovery
kubectl wait --for=condition=Ready pod/<pod-name> -n <namespace> --timeout=5m
```

**Escalation**: If pod continues to fail after 3 restart attempts, escalate to on-call engineer.

## Scenario 2: Node Failure

**Detection**: Node in NotReady state

**Recovery**:

```bash
# Identify failed node
kubectl get nodes

# Describe node to see conditions
kubectl describe node <node-name>

# Check node events
kubectl get events --field-selector involvedObject.name=<node-name>

# Drain node (if recoverable)
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data

# Pods will be rescheduled to healthy nodes automatically

# If node is unrecoverable:
# 1. Delete node from cluster
kubectl delete node <node-name>

# 2. Provision new node (cloud provider specific)
# AWS example:
aws ec2 run-instances \
  --launch-template LaunchTemplateName=eks-worker \
  --count 1

# 3. New node will auto-join cluster (if using managed node groups)

# Verify cluster health
kubectl get nodes
kubectl get pods -A -o wide | grep <node-name>
```

**Prevention**: Set up node auto-scaling and use managed node groups.

## Scenario 3: Namespace Deletion (Accidental)

**Detection**: Namespace missing, applications down

**Recovery**:

```bash
# List available backups
velero backup get

# Find recent backup containing the namespace
velero backup describe <backup-name> | grep <namespace>

# Restore entire namespace
velero restore create --from-backup <backup-name> \
  --include-namespaces <namespace> \
  --wait

# Verify restoration
kubectl get all -n <namespace>

# Check application health
kubectl get pods -n <namespace>
kubectl logs -n <namespace> deploy/<deployment-name>
```

**Post-Recovery**:
- Implement RBAC to prevent accidental deletions
- Add finalizers to critical namespaces
- Enable audit logging

## Scenario 4: Database Failure

**Detection**: Database pods down, connectivity errors

**Recovery**:

### PostgreSQL (CloudNativePG)

```bash
# Check cluster status
kubectl get cluster -A

# Describe cluster
kubectl describe cluster <cluster-name> -n <namespace>

# Check failover status
kubectl get pods -n <namespace> -l cnpg.io/cluster=<cluster-name>

# If primary failed, operator should auto-failover
# Wait 2 minutes for automatic recovery

# If no automatic recovery:
# 1. Identify current primary
kubectl get pods -n <namespace> \
  -l cnpg.io/cluster=<cluster-name>,role=primary

# 2. Force failover (if needed)
kubectl cnpg promote <cluster-name> <new-primary-pod>

# 3. Verify new topology
kubectl get cluster <cluster-name> -n <namespace> -o yaml

# If complete cluster loss, restore from backup:
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: <cluster-name>-restored
  namespace: <namespace>
spec:
  instances: 3
  bootstrap:
    recovery:
      backup:
        name: <backup-name>
      recoveryTarget:
        targetTime: "<timestamp>"
EOF

# Wait for restore
kubectl wait --for=condition=Ready cluster/<cluster-name>-restored \
  -n <namespace> --timeout=10m
```

## Scenario 5: Complete Cluster Failure

**Detection**: Cluster API unreachable, all services down

**Recovery Steps**:

### Phase 1: Assessment (0-5 minutes)

```bash
# Check cluster API
kubectl cluster-info

# If unreachable, check control plane
# For managed Kubernetes (EKS/GKE/AKS):
# - Check cloud provider console
# - Verify control plane health

# Check monitoring (if accessible)
# - Grafana dashboards
# - Prometheus federated view
# - Cloud provider metrics
```

### Phase 2: Activate DR Cluster (5-15 minutes)

```bash
# Switch kubectl context to DR cluster
kubectl config use-context dr-cluster

# Verify DR cluster health
kubectl get nodes
kubectl get ns

# List available backups
velero backup get

# Identify latest full backup
LATEST_BACKUP=$(velero backup get -o json | \
  jq -r '.items | sort_by(.status.startTimestamp) | last | .metadata.name')

echo "Latest backup: $LATEST_BACKUP"

# Restore critical namespaces first
CRITICAL_NS="mantl-system ecommerce production"

for ns in $CRITICAL_NS; do
  echo "Restoring namespace: $ns"
  velero restore create restore-$ns-$(date +%s) \
    --from-backup $LATEST_BACKUP \
    --include-namespaces $ns \
    --wait
done

# Verify restorations
velero restore get
kubectl get pods -A
```

### Phase 3: DNS Failover (15-25 minutes)

```bash
# Update DNS to point to DR cluster
# Method depends on DNS provider

# Route53 example:
aws route53 change-resource-record-sets \
  --hosted-zone-id Z1234567890ABC \
  --change-batch file://dns-failover.json

# dns-failover.json:
{
  "Changes": [{
    "Action": "UPSERT",
    "ResourceRecordSet": {
      "Name": "app.example.com",
      "Type": "A",
      "TTL": 60,
      "ResourceRecords": [
        {"Value": "<dr-cluster-ip>"}
      ]
    }
  }]
}

# Verify DNS propagation
dig app.example.com +short

# Test application access
curl -I https://app.example.com/health
```

### Phase 4: Verification (25-45 minutes)

```bash
# Check all pods are running
kubectl get pods -A | grep -v Running

# Verify services
kubectl get svc -A

# Check ingress
kubectl get ingress -A

# Verify database connections
for deploy in $(kubectl get deploy -n ecommerce -o name); do
  echo "Checking $deploy"
  kubectl exec -n ecommerce $deploy -- \
    psql $DATABASE_URL -c "SELECT 1" || echo "FAILED: $deploy"
done

# Run smoke tests
kubectl apply -f tests/smoke-tests.yaml
kubectl wait --for=condition=Complete job/smoke-test --timeout=5m
kubectl logs job/smoke-test
```

### Phase 5: Monitoring (Ongoing)

```bash
# Monitor application metrics
kubectl port-forward -n mantl-system svc/grafana 3000:80

# Check error rates
# Check latency
# Check resource utilization

# Set up alerts for anomalies
kubectl apply -f monitoring/dr-alerts.yaml
```

## Scenario 6: Backup Storage Failure

**Detection**: Velero backup failures, storage unavailable

**Recovery**:

```bash
# Check Velero status
kubectl get backupstoragelocation -n velero

# Check Velero logs
kubectl logs -n velero deploy/velero -f

# If primary storage unavailable:
# 1. Create new backup storage location
kubectl apply -f - <<EOF
apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  name: secondary
  namespace: velero
spec:
  provider: aws
  objectStorage:
    bucket: mantl-backups-secondary
    prefix: velero
  config:
    region: us-west-2
EOF

# 2. Set as default
kubectl patch backupstoragelocation default \
  -n velero \
  -p '{"spec":{"default":false}}'

kubectl patch backupstoragelocation secondary \
  -n velero \
  -p '{"spec":{"default":true}}'

# 3. Verify new location
velero backup-location get

# 4. Trigger manual backup
velero backup create manual-failover-backup --wait
```

## Scenario 7: Certificate Expiration

**Detection**: Certificate expiration warnings, TLS errors

**Recovery**:

```bash
# Check certificate expiration
kubectl get certificate -A

# Check cert-manager logs
kubectl logs -n cert-manager deploy/cert-manager -f

# Force certificate renewal
kubectl delete certificate <cert-name> -n <namespace>

# Cert-manager will recreate automatically

# If cert-manager is down:
# 1. Restart cert-manager
kubectl rollout restart deploy/cert-manager -n cert-manager

# 2. Check CertificateRequest status
kubectl get certificaterequest -A

# 3. Manually approve if needed
kubectl certificate approve <cert-request>

# Verify new certificate
kubectl get secret <tls-secret> -n <namespace> -o yaml | \
  grep tls.crt | awk '{print $2}' | base64 -d | \
  openssl x509 -noout -dates
```

## Recovery Validation Checklist

After any recovery procedure:

- [ ] All pods in Running state
- [ ] All services responding to health checks
- [ ] Database connectivity verified
- [ ] Ingress routing working
- [ ] Monitoring collecting metrics
- [ ] Logs flowing to central logging
- [ ] Backups resuming successfully
- [ ] No critical alerts firing
- [ ] Application smoke tests passing
- [ ] User-facing services accessible

## Post-Recovery Actions

1. **Root Cause Analysis**:
   ```bash
   # Collect logs
   kubectl logs -n <namespace> <pod> --previous > incident-logs.txt

   # Export events
   kubectl get events -A --sort-by='.lastTimestamp' > incident-events.txt

   # Document timeline
   # Create incident report
   ```

2. **Update Runbooks**: Document any deviations or improvements

3. **Improve Monitoring**: Add alerts for early detection

4. **Test Recovery**: Schedule DR drills quarterly

## Emergency Contacts

- **On-Call Engineer**: PagerDuty escalation policy
- **Platform Team Lead**: [Contact Info]
- **Cloud Provider Support**: [Account Number]
- **Backup Storage Support**: [Contact Info]

## Additional Resources

- [Velero Documentation](https://velero.io/docs/)
- [CloudNativePG Disaster Recovery](https://cloudnative-pg.io/documentation/)
- [Kubernetes Disaster Recovery](https://kubernetes.io/docs/tasks/administer-cluster/)
