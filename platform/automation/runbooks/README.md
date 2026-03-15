# Automated Runbook Actions

Self-healing automation for common operational tasks.

## Features

- **Crashloop Detection**: Auto-restart failing pods
- **Auto-Scaling**: Adjust HPA based on time patterns
- **Certificate Renewal**: Proactive cert renewal
- **Log Cleanup**: Automated log rotation

## Deployed Automations

### 1. Crashloop Detector
**Runs**: Every 5 minutes
**Action**: Restarts pods with >5 restarts
**Alert**: Slack notification

```bash
kubectl apply -f platform/automation/runbooks/pod-restart/crashloop-detector.yaml
```

### 2. HPA Adjuster
**Runs**: Every hour
**Action**: Adjusts min/max replicas based on time
**Schedule**:
- Business hours (9am-6pm M-F): 5-20 replicas
- After hours: 2-10 replicas

```bash
kubectl apply -f platform/automation/runbooks/auto-scale/hpa-adjuster.yaml
```

### 3. Certificate Checker
**Runs**: Daily at midnight
**Action**: Force renewal if expires in <7 days
**Warning**: Alert if <14 days

```bash
kubectl apply -f platform/automation/runbooks/cert-renewal/cert-checker.yaml
```

### 4. Log Rotator
**Runs**: Continuously (DaemonSet)
**Action**:
- Delete logs >30 days old
- Compress logs >100MB
- Runs on all nodes

```bash
kubectl apply -f platform/automation/runbooks/log-cleanup/log-rotator.yaml
```

## Setup

### 1. Create Automation Namespace

```bash
kubectl create namespace automation
```

### 2. Create Service Account

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: automation-sa
  namespace: automation
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: automation-role
rules:
- apiGroups: ["*"]
  resources: ["pods", "certificates", "horizontalpodautoscalers"]
  verbs: ["get", "list", "delete", "patch", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: automation-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: automation-role
subjects:
- kind: ServiceAccount
  name: automation-sa
  namespace: automation
```

### 3. Configure Slack Webhook (Optional)

```bash
kubectl create secret generic slack-webhook -n automation \
  --from-literal=url=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

## Monitoring

View automation logs:
```bash
kubectl logs -n automation -l app=automation --tail=100 -f
```

Check CronJob history:
```bash
kubectl get jobs -n automation
```

## Customization

Edit ConfigMaps to adjust:
- Thresholds
- Schedules
- Retention periods
- Alert channels

```bash
kubectl edit configmap crashloop-detector -n automation
```

## Safety

- All automations run with limited RBAC
- Dry-run mode available
- Alerts sent before destructive actions
- Configurable thresholds

## Resources

- See individual automation READMEs for details
- Runbook documentation in docs/runbooks/
