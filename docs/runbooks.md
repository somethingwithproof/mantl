# Operational Runbooks

Standard operating procedures for maintaining the mantl platform.

## Table of Contents

- [Cluster Operations](#cluster-operations)
- [Security Operations](#security-operations)
- [Backup and Recovery](#backup-and-recovery)
- [Monitoring and Alerts](#monitoring-and-alerts)
- [Incident Response](#incident-response)

---

## Cluster Operations

### Cluster Upgrade

Upgrade Kubernetes version while minimizing downtime.

**Prerequisites:**
- Backup etcd and critical data
- Review release notes for breaking changes
- Test in staging environment first

**Procedure:**

1. **Update Terraform variable**:
   ```bash
   cd infra/terraform/terraform/blueprints/aws-eks
   # Edit variables.tf
   # Change kubernetes_version from "1.31" to "1.32"
   ```

2. **Plan and apply**:
   ```bash
   terraform plan -out=upgrade.tfplan
   # Review changes carefully
   terraform apply upgrade.tfplan
   ```

3. **Monitor control plane upgrade**:
   ```bash
   aws eks describe-cluster --name mantl --query 'cluster.version'
   kubectl get nodes
   ```

4. **Upgrade node groups** (automatic with managed node groups):
   ```bash
   # AWS EKS managed node groups auto-upgrade
   # Monitor progress
   kubectl get nodes -w
   ```

5. **Update add-ons**:
   ```bash
   # EKS will prompt for add-on updates
   aws eks list-addons --cluster-name mantl
   aws eks update-addon --cluster-name mantl --addon-name vpc-cni --addon-version <version>
   ```

6. **Verify cluster health**:
   ```bash
   kubectl get nodes
   kubectl get pods --all-namespaces | grep -v Running
   ```

**Rollback:**
- Restore from etcd backup if critical issues
- For AWS EKS, downgrade is not supported - must restore from backup

---

### Node Maintenance

Safely drain and update cluster nodes.

**Cordon and Drain Node:**

```bash
# Prevent new pods from scheduling
kubectl cordon <node-name>

# Evict existing pods (respects PodDisruptionBudgets)
kubectl drain <node-name> \
  --ignore-daemonsets \
  --delete-emptydir-data \
  --force \
  --grace-period=300

# Verify pods rescheduled
kubectl get pods -o wide --all-namespaces | grep <old-node>
```

**Perform Maintenance:**
- OS updates
- Security patches
- Kernel upgrades
- Hardware maintenance

**Uncordon Node:**

```bash
# Re-enable scheduling
kubectl uncordon <node-name>

# Verify node Ready
kubectl get nodes
```

---

### Scale Node Groups

Adjust capacity for workload demands.

**Manual Scaling:**

```bash
cd infra/terraform/terraform/blueprints/aws-eks

# Edit variables.tf
# Update desired_size for system or workload node group

terraform plan
terraform apply
```

**Auto-scaling:**

Cluster Autoscaler automatically scales based on pod resource requests:

```bash
# Check Cluster Autoscaler status
kubectl logs -n kube-system deployment/cluster-autoscaler

# View autoscaler decisions
kubectl get events -n kube-system | grep cluster-autoscaler
```

---

## Security Operations

### Rotate Credentials

Regularly rotate sensitive credentials and certificates.

**AWS IAM Service Account (IRSA) Rotation:**

IRSA uses temporary credentials that auto-rotate. Verify role trust policy:

```bash
aws iam get-role --role-name mantl-external-dns
```

**Image Registry Credentials:**

```bash
# Generate new GitHub token
# Update secret
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=$GITHUB_USER \
  --docker-password=$NEW_TOKEN \
  --dry-run=client -o yaml | kubectl apply -f -
```

**Certificate Rotation:**

cert-manager auto-renews certificates. Force renewal:

```bash
kubectl delete secret <tls-secret-name> -n <namespace>
# cert-manager will automatically recreate it
```

---

### Apply Security Patches

Keep platform components up-to-date with security patches.

**Check for Updates:**

```bash
# Check Renovate/Dependabot PRs
gh pr list --label dependencies

# Check for CVEs
safety check --json
```

**Update Platform Components:**

1. **Test in dev environment first**
2. **Review changelog for breaking changes**
3. **Apply updates via ArgoCD**:
   ```bash
   argocd app sync <app-name> --prune
   argocd app wait <app-name> --health
   ```

**Emergency Patching:**

For critical CVEs, bypass normal workflow:

```bash
# Update image tag directly (emergency only)
kubectl set image deployment/<name> <container>=<new-image> -n <namespace>

# Follow up with proper GitOps update
```

---

### Audit Policy Violations

Review and respond to Kyverno policy violations.

**Daily Audit:**

```bash
# Check policy reports
kubectl get policyreport --all-namespaces

# Failed policies
kubectl get policyreport --all-namespaces -o json | \
  jq '.items[] | select(.summary.fail > 0)'
```

**Investigation:**

```bash
# Detailed violation info
kubectl describe policyreport <report-name> -n <namespace>

# Check specific pod
kubectl get pod <pod-name> -n <namespace> -o yaml
```

**Remediation:**

1. **Fix non-compliant resources**
2. **Update policies if too restrictive**
3. **Document exemptions for special cases**

---

## Backup and Recovery

### Backup Cluster State

Regular backups of critical cluster state.

**etcd Backup (AWS EKS):**

AWS manages etcd backups automatically. For additional backup:

```bash
# Backup critical resources
kubectl get all --all-namespaces -o yaml > cluster-state-$(date +%Y%m%d).yaml

# Backup custom resources
kubectl get crd -o yaml > crds-$(date +%Y%m%d).yaml
```

**Backup Persistent Data:**

```bash
# List PVCs
kubectl get pvc --all-namespaces

# Use Velero for comprehensive backup
velero backup create daily-backup --include-namespaces='*'
```

**Backup Terraform State:**

Terraform state is stored in S3 with versioning enabled (recommended setup):

```bash
# List state versions
aws s3api list-object-versions \
  --bucket terraform-state-bucket \
  --prefix mantl/

# Download specific version
aws s3api get-object \
  --bucket terraform-state-bucket \
  --key mantl/terraform.tfstate \
  --version-id <version-id> \
  terraform.tfstate.backup
```

---

### Disaster Recovery

Restore cluster from catastrophic failure.

**Scenario 1: Cluster Deleted**

1. **Restore infrastructure**:
   ```bash
   cd infra/terraform/terraform/blueprints/aws-eks
   terraform init
   terraform apply
   ```

2. **Restore platform components**:
   ```bash
   # ArgoCD will auto-sync from Git
   kubectl apply -k platform/bootstrap
   ```

3. **Restore application data**:
   ```bash
   velero restore create --from-backup daily-backup
   ```

**Scenario 2: Data Corruption**

1. **Identify affected resources**
2. **Restore from most recent backup**:
   ```bash
   kubectl apply -f cluster-state-YYYYMMDD.yaml
   ```
3. **Verify data integrity**

**Recovery Time Objectives (RTO/RPO):**
- Infrastructure restore: 30 minutes
- Platform component restore: 15 minutes
- Application data restore: Varies by backup method
- Total RTO: < 1 hour
- RPO: < 24 hours (daily backups)

---

## Monitoring and Alerts

### Check System Health

Daily health check routine.

**Morning Health Check:**

```bash
#!/bin/bash
# Daily health check script

echo "=== Cluster Status ==="
kubectl get nodes
kubectl top nodes

echo "=== Failed Pods ==="
kubectl get pods --all-namespaces --field-selector=status.phase!=Running,status.phase!=Succeeded

echo "=== Resource Usage ==="
kubectl top pods --all-namespaces --sort-by=memory | head -20

echo "=== Recent Events ==="
kubectl get events --all-namespaces --sort-by=.metadata.creationTimestamp | tail -20

echo "=== Policy Violations ==="
kubectl get policyreport --all-namespaces -o json | \
  jq '.items[] | select(.summary.fail > 0) | {namespace: .metadata.namespace, failures: .summary.fail}'

echo "=== Certificate Expiry ==="
kubectl get certificate --all-namespaces
```

**Prometheus Alerts:**

```bash
# Check firing alerts
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Navigate to http://localhost:9090/alerts
```

---

### Investigate Performance Issues

Diagnose and resolve performance degradation.

**High CPU/Memory Usage:**

```bash
# Identify top consumers
kubectl top nodes
kubectl top pods --all-namespaces --sort-by=cpu
kubectl top pods --all-namespaces --sort-by=memory

# Check resource requests/limits
kubectl describe pod <pod-name> -n <namespace> | grep -A 5 "Limits\|Requests"
```

**Slow API Responses:**

```bash
# Check API server latency
kubectl get --raw /metrics | grep apiserver_request_duration

# Check etcd performance
kubectl logs -n kube-system etcd-<pod> | grep "slow"
```

**Network Latency:**

```bash
# Test pod-to-pod latency
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -- sh
# Inside pod: ping <service-name>

# Check Cilium metrics
kubectl port-forward -n kube-system svc/hubble-metrics 9091:9091
curl localhost:9091/metrics | grep cilium_forward_count_total
```

---

## Incident Response

### Security Incident

Response procedure for security breaches.

**Immediate Actions:**

1. **Contain**: Isolate affected workloads
   ```bash
   kubectl cordon <compromised-node>
   kubectl delete pod <compromised-pod> -n <namespace>
   ```

2. **Preserve evidence**:
   ```bash
   kubectl logs <pod-name> -n <namespace> --previous > incident-logs.txt
   kubectl describe pod <pod-name> -n <namespace> > incident-describe.txt
   ```

3. **Review access logs**:
   ```bash
   # CloudWatch Logs Insights query
   fields @timestamp, @message
   | filter @message like /unauthorized|forbidden|denied/
   | sort @timestamp desc
   ```

**Investigation:**

1. Check VPC Flow Logs for unauthorized traffic
2. Review Falco/Tetragon alerts for suspicious activity
3. Audit kubectl/API server logs
4. Check for policy violations

**Remediation:**

1. Patch vulnerability if identified
2. Rotate compromised credentials
3. Update security policies
4. Document incident and lessons learned

---

### Application Outage

Restore service availability quickly.

**Triage:**

```bash
# Check pod status
kubectl get pods -n <namespace>

# Check events
kubectl get events -n <namespace> --sort-by=.metadata.creationTimestamp

# Check logs
kubectl logs -n <namespace> <pod-name> --tail=100
```

**Common Causes & Fixes:**

1. **CrashLoopBackOff**:
   - Check logs: `kubectl logs <pod> --previous`
   - Verify configuration: ConfigMaps, Secrets
   - Check resource limits

2. **ImagePullBackOff**:
   - Verify image exists and tag is correct
   - Check image pull secrets
   - Review Kyverno policy violations

3. **Pending pods**:
   - Check node capacity: `kubectl describe nodes`
   - Review PVC status if using persistent storage
   - Check for pod anti-affinity conflicts

**Escalation Path:**

1. On-call engineer (immediate)
2. Platform team lead (15 minutes)
3. Architecture team (30 minutes)
4. Vendor support if vendor component (varies)

---

## Best Practices

### Change Management

1. **Always test in dev/staging first**
2. **Use GitOps**: Never `kubectl apply` directly in production
3. **Document changes**: Update runbooks after significant changes
4. **Communicate**: Notify team of planned maintenance
5. **Have rollback plan**: Know how to revert before applying changes

### Regular Maintenance Schedule

| Task | Frequency | Owner |
|------|-----------|-------|
| Health check script | Daily | On-call |
| Security patching | Weekly | Platform team |
| Backup verification | Weekly | Platform team |
| Policy audit | Weekly | Security team |
| Capacity review | Monthly | Platform team |
| Disaster recovery drill | Quarterly | All teams |
| Credential rotation | Quarterly | Security team |
| Cluster upgrade | As needed | Platform team |

### Monitoring Checklist

- ✅ All nodes in Ready state
- ✅ No pods in CrashLoopBackOff
- ✅ No policy violations in production
- ✅ Certificate expiry > 30 days
- ✅ Resource usage < 80%
- ✅ No firing critical alerts
- ✅ Backups completed successfully
- ✅ VPC Flow Logs collecting

---

## Contact Information

### Escalation Matrix

| Level | Contact | Response Time | Scope |
|-------|---------|---------------|-------|
| L1 | On-call engineer | Immediate | Platform issues |
| L2 | Platform team lead | 15 min | Complex technical issues |
| L3 | Architecture team | 30 min | Design decisions |
| L4 | Vendor support | Varies | Vendor product issues |

### Emergency Contacts

- **Slack**: #mantl-incidents
- **PagerDuty**: mantl-platform
- **Email**: platform-team@company.com
