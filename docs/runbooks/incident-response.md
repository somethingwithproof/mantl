# Incident Response Playbooks

> Historical component-stack procedure examples. These commands and objectives
> have not established acceptance for the current beta runtime. Use the maintained
> [operations guide](../runbooks.md) and [maturity table](../../README.md#capabilities-and-maturity)
> first. Any recovery objectives here are proposed targets, not measured guarantees.


Standardized procedures for responding to production incidents on the mantl platform.

## Overview

This runbook covers:
- Incident classification and severity levels
- Response procedures by incident type
- Communication protocols
- Post-incident review process

## Incident Severity Levels

| Severity | Definition | Response Time | Examples |
|----------|-----------|---------------|----------|
| **SEV1 - Critical** | Complete service outage affecting all users | 5 minutes | Cluster down, database unavailable, security breach |
| **SEV2 - High** | Major functionality degraded for most users | 15 minutes | High error rates, significant performance degradation |
| **SEV3 - Medium** | Partial functionality degraded for some users | 1 hour | Intermittent errors, minor performance issues |
| **SEV4 - Low** | Minor issue with minimal user impact | Next business day | UI glitch, non-critical feature unavailable |

## Incident Response Process

```mermaid
graph TD
    A[Incident Detected] --> B[Assess Severity]
    B --> C{Severity Level}
    C -->|SEV1/2| D[Page On-Call]
    C -->|SEV3/4| E[Create Ticket]
    D --> F[Incident Commander Assigned]
    F --> G[Assemble Response Team]
    G --> H[Execute Playbook]
    H --> I[Provide Status Updates]
    I --> J[Resolve & Verify]
    J --> K[Post-Incident Review]
    E --> H
```

## Communication Protocols

### Initial Response (within 5 minutes)

1. **Acknowledge the incident**:
   - In PagerDuty: Acknowledge the alert
   - In Slack: Join #incidents channel and post: "Investigating [incident]"

2. **Assess severity**:
   - Check monitoring dashboards
   - Verify user impact
   - Classify severity level

3. **Escalate if needed**:
   ```bash
   # SEV1: Page Platform Lead immediately
   # SEV2: Notify team in Slack
   # SEV3/4: Create ticket, handle during business hours
   ```

### Status Updates

**Update frequency**:
- SEV1: Every 15 minutes
- SEV2: Every 30 minutes
- SEV3: Every 2 hours
- SEV4: Daily

**Status page template**:
```
Status: [Investigating | Identified | Monitoring | Resolved]
Severity: SEV[1-4]
Impact: [Description of user impact]
Last Update: [Timestamp]

[Current status description]

Next Update: [Estimated time]
```

### Resolution Communication

```
RESOLVED: [Incident Title]

Duration: [Start time] - [End time] ([X] minutes total)
Root Cause: [Brief description]
Impact: [Number of users/services affected]

Actions Taken:
- [Action 1]
- [Action 2]

Preventive Measures:
- [Measure 1]
- [Measure 2]

Post-Incident Review: [Date scheduled]
```

## SEV1: Complete Service Outage

### Initial Response Checklist

- [ ] Page on-call engineer (PagerDuty)
- [ ] Start incident call (Zoom/Meet)
- [ ] Post to #incidents channel
- [ ] Update status page to "Major Outage"
- [ ] Assign Incident Commander

### Investigation Phase

```bash
# 1. Check cluster API accessibility
kubectl cluster-info

# 2. Check node status
kubectl get nodes

# 3. Check critical pods
kubectl get pods -A | grep -v Running

# 4. Check recent events
kubectl get events -A --sort-by='.lastTimestamp' | tail -50

# 5. Check control plane components (if self-managed)
kubectl get componentstatuses

# 6. Check cloud provider status
# AWS: https://status.aws.amazon.com/
# GCP: https://status.cloud.google.com/
# Azure: https://status.azure.com/
```

### Recovery Procedures

**Scenario: Control Plane Unavailable**

```bash
# If managed Kubernetes (EKS/GKE/AKS):
# 1. Check cloud provider console for control plane health
# 2. Contact cloud provider support (Priority 1 ticket)
# 3. Activate DR cluster if recovery > 30 minutes

# Activate DR cluster
kubectl config use-context dr-cluster
velero restore create sev1-failover \
  --from-backup latest-full-backup \
  --include-namespaces production,mantl-system \
  --wait

# Update DNS to DR cluster
# (See disaster-recovery.md for detailed DNS failover)
```

**Scenario: Complete Node Failure**

```bash
# Check node group status
aws eks describe-nodegroup \
  --cluster-name mantl-prod \
  --nodegroup-name workers | jq '.nodegroup.health'

# Force node group recreation
aws eks update-nodegroup-config \
  --cluster-name mantl-prod \
  --nodegroup-name workers \
  --scaling-config minSize=5,maxSize=20,desiredSize=5

# Monitor node recovery
watch kubectl get nodes
```

**Scenario: Database Cluster Down**

```bash
# Check cluster status
kubectl get cluster -n production

# Check operator logs
kubectl logs -n cnpg-system deploy/cnpg-controller-manager --tail=100

# Force failover if primary down
kubectl cnpg promote myapp-db myapp-db-2 -n production

# If complete cluster loss, restore from backup
# (See disaster-recovery.md for database recovery)
```

## SEV2: High Error Rate

### Investigation

```bash
# 1. Check error rate in Prometheus
kubectl port-forward -n mantl-system svc/prometheus 9090:9090
# Query: rate(http_requests_total{status=~"5.."}[5m])

# 2. Check application logs
kubectl logs -n production -l app=myapp --tail=100 | grep ERROR

# 3. Check recent deployments
kubectl rollout history deployment/myapp -n production

# 4. Check resource utilization
kubectl top pods -n production

# 5. Check database health
kubectl exec -n production myapp-db-1 -- \
  psql -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';"
```

### Recovery Procedures

**Scenario: Bad Deployment**

```bash
# Rollback to previous version
kubectl rollout undo deployment/myapp -n production

# Verify rollback
kubectl rollout status deployment/myapp -n production

# Monitor error rate
watch kubectl logs -n production -l app=myapp --tail=5 | grep ERROR
```

**Scenario: Resource Exhaustion**

```bash
# Identify resource bottleneck
kubectl top pods -n production --sort-by=memory

# Scale horizontally
kubectl scale deployment/myapp -n production --replicas=10

# Increase resource limits
kubectl patch deployment myapp -n production --type=json -p='[
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/resources/limits/memory",
    "value": "4Gi"
  }
]'
```

**Scenario: External Dependency Failure**

```bash
# Check connectivity to dependency
kubectl exec -it -n production deploy/myapp -- curl -v http://external-api.example.com/health

# Enable circuit breaker (if using service mesh)
kubectl apply -f - <<EOF
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: external-api-circuit-breaker
  namespace: production
spec:
  host: external-api.example.com
  trafficPolicy:
    outlierDetection:
      consecutive5xxErrors: 5
      interval: 30s
      baseEjectionTime: 30s
EOF

# Implement fallback/cache if available
kubectl set env deployment/myapp -n production ENABLE_CACHE=true
```

## SEV3: Performance Degradation

### Investigation

```bash
# 1. Identify slow endpoints
kubectl logs -n production -l app=myapp --tail=1000 | \
  grep "duration" | awk '{print $NF}' | sort -n | tail -20

# 2. Check database performance
kubectl exec -n production myapp-db-1 -- \
  psql -c "SELECT query, calls, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# 3. Check cache hit ratio
kubectl exec -n production redis-0 -- redis-cli INFO stats | grep cache_hit_rate

# 4. Profile application (if profiling enabled)
kubectl port-forward -n production deploy/myapp 6060:6060
go tool pprof http://localhost:6060/debug/pprof/profile
```

### Mitigation

```bash
# Scale up to handle load
kubectl scale deployment/myapp -n production --replicas=15

# Add caching layer
kubectl apply -f manifests/redis-cache.yaml

# Optimize database queries (coordinate with dev team)
# - Add indexes
# - Optimize slow queries
# - Enable query caching

# Consider adding CDN for static assets
```

## SEV4: Minor Issues

### Standard Response

```bash
# Create tracking ticket
# Investigate during business hours
# Follow standard debugging procedures (see troubleshooting-guide.md)
# Deploy fix via normal CI/CD process
```

## Security Incidents

### Potential Security Breach

**IMMEDIATE ACTIONS**:

```bash
# 1. ISOLATE affected resources
kubectl label namespace compromised-ns security-incident=true
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
  namespace: compromised-ns
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
EOF

# 2. PRESERVE evidence
kubectl get pods -n compromised-ns -o yaml > incident-pods.yaml
kubectl logs -n compromised-ns --all-containers=true --timestamps=true > incident-logs.txt
kubectl get events -n compromised-ns > incident-events.txt

# 3. NOTIFY security team immediately
# Email: security@example.com
# Slack: #security-incidents

# 4. ROTATE credentials
kubectl delete secret -n compromised-ns --all
# Regenerate and deploy new secrets

# 5. SCAN for indicators of compromise
kubectl run kube-hunter --image=aquasec/kube-hunter -it --rm --restart=Never -- --pod

# 6. REVIEW audit logs
kubectl get events -A --sort-by='.lastTimestamp' | grep -i "delete\|create\|update"
```

### Data Leak Detection

```bash
# 1. Identify leaked data
# Check logs, monitoring, external reports

# 2. Revoke exposed credentials
aws iam update-access-key --access-key-id AKIAIOSFODNN7EXAMPLE --status Inactive

# 3. Rotate secrets
kubectl delete secret sensitive-creds -n production
kubectl create secret generic sensitive-creds --from-literal=key=new-value -n production
kubectl rollout restart deployment/myapp -n production

# 4. Notify affected parties per data breach protocol

# 5. Implement additional controls
kubectl apply -f manifests/security/secrets-encryption.yaml
```

## Post-Incident Review (PIR)

### Conduct Within 48 Hours

**Agenda**:
1. Timeline of events
2. Root cause analysis
3. What went well
4. What could be improved
5. Action items

**Template**:

```markdown
# Post-Incident Review: [Incident Title]

**Date**: [Date]
**Severity**: SEV[X]
**Duration**: [Duration]
**Participants**: [Names]

## Summary
[Brief description of incident]

## Impact
- Users affected: [Number/Percentage]
- Services affected: [List]
- Revenue impact: [If applicable]
- Downtime: [Duration]

## Timeline
| Time | Event |
|------|-------|
| 14:23 | Alert fired: High error rate |
| 14:25 | On-call engineer acknowledged |
| 14:30 | Root cause identified: Bad deployment |
| 14:32 | Rollback initiated |
| 14:35 | Service restored |
| 14:40 | Monitoring confirmed resolution |

## Root Cause
[Detailed explanation]

## Detection
- How was the incident detected? [Alert/User report/Monitoring]
- Was detection timely? [Yes/No + explanation]

## Response
### What Went Well
- [Item 1]
- [Item 2]

### What Could Be Improved
- [Item 1]
- [Item 2]

## Action Items
| Action | Owner | Due Date | Priority |
|--------|-------|----------|----------|
| Add pre-deployment validation | DevOps | 2024-04-01 | High |
| Improve alerting threshold | SRE | 2024-04-05 | Medium |

## Lessons Learned
- [Lesson 1]
- [Lesson 2]
```

## Incident Metrics

Track and review monthly:

```bash
# Generate incident report
cat <<EOF > incident-report.md
# Incident Report: $(date +%B\ %Y)

## Summary
- Total incidents: X
- SEV1: X (target: <2)
- SEV2: X (target: <5)
- SEV3: X (target: <15)
- SEV4: X

## MTTR (Mean Time To Recovery)
- SEV1: X minutes (target: <30min)
- SEV2: X minutes (target: <2h)
- Average: X minutes

## MTTD (Mean Time To Detect)
- Average: X minutes (target: <5min)

## Top Incident Causes
1. [Cause 1] - X incidents
2. [Cause 2] - X incidents
3. [Cause 3] - X incidents

## Action Items Status
- Completed: X
- In Progress: X
- Not Started: X
EOF
```

## Runbook Maintenance

- Review and update runbooks quarterly
- Update after each major incident
- Test procedures in DR drills
- Incorporate feedback from PIRs

## Emergency Contacts

| Role | Contact | Escalation |
|------|---------|------------|
| On-Call Engineer | PagerDuty | Automatic |
| Platform Lead | [Phone] | After 15min (SEV1) |
| Security Team | security@example.com | Immediate (security incidents) |
| Cloud Provider Support | [Account Link] | As needed |
| Database Vendor | [Support Portal] | Critical issues |

## Additional Resources

- [Incident Management Guide](https://response.pagerduty.com/)
- [Google SRE Book - Incident Response](https://sre.google/sre-book/effective-troubleshooting/)
- [Post-Incident Review Template](https://postmortems.pagerduty.com/)
