# Kyverno Security Policies

Policy-as-code enforcement for Kubernetes security and compliance using Kyverno.

## Overview

This directory contains Kyverno policies that enforce security best practices and compliance requirements across the cluster. Policies are organized by category and can be deployed using the Helm release or directly with kubectl.

## Policy Categories

### Pod Security Standards
- `pss-restricted.yaml` - Enforce Pod Security Standards (Restricted profile)
- `require-non-root.yaml` - All containers must run as non-root
- `disallow-privilege-escalation.yaml` - Prevent privilege escalation
- `drop-all-capabilities.yaml` - Require dropping all Linux capabilities

### Resource Management
- `require-resource-limits.yaml` - All containers must have resource requests/limits
- `restrict-resource-limits.yaml` - Maximum resource limits per container
- `require-pod-requests-limits.yaml` - Pod-level resource requirements

### Image Security
- `restrict-image-registries.yaml` - Only allow approved registries
- `require-image-signature.yaml` - Require signed images (Sigstore/Cosign)
- `disallow-latest-tag.yaml` - Prevent :latest and mutable tags
- `require-image-digest.yaml` - Require images by digest (sha256)

### Network Security
- `require-network-policy.yaml` - Namespaces must have network policies
- `restrict-ingress-classes.yaml` - Only allow approved ingress classes
- `restrict-external-ips.yaml` - Prevent LoadBalancer with specific external IPs

### Labels & Metadata
- `require-labels.yaml` - Required labels for all resources
- `require-namespace-labels.yaml` - Required labels for namespaces
- `restrict-annotations.yaml` - Block certain annotations

### Compliance
- `disallow-default-namespace.yaml` - Prevent workloads in default namespace
- `disallow-automount-sa-token.yaml` - Disable automounting service account tokens
- `require-readonly-root-filesystem.yaml` - Enforce read-only root filesystems

## Deployment

### Using Helm

```bash
# Install Kyverno operator
helm repo add kyverno https://kyverno.github.io/kyverno/
helm repo update

helm install kyverno kyverno/kyverno \
  --namespace kyverno \
  --create-namespace \
  --set replicaCount=3

# Deploy policies
kubectl apply -k platform/security/kyverno/
```

### Verification

```bash
# Check policy status
kubectl get clusterpolicies

# View policy reports
kubectl get policyreports -A

# Check violations
kubectl get policyreport -n ecommerce -o yaml | \
  yq '.results[] | select(.result == "fail")'
```

## Policy Modes

### Enforce (Default)
Blocks resources that violate policies:

```yaml
spec:
  validationFailureAction: Enforce
```

### Audit Mode
Reports violations without blocking:

```yaml
spec:
  validationFailureAction: Audit
```

### Testing New Policies

Always test new policies in audit mode first:

```bash
# Deploy in audit mode
kubectl apply -f policy.yaml

# Check for violations
kubectl get policyreport -A

# Switch to enforce after validation
kubectl patch clusterpolicy <policy-name> \
  --type=merge \
  -p '{"spec":{"validationFailureAction":"Enforce"}}'
```

## Exemptions

### Namespace-level Exemptions

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: system-components
  labels:
    kyverno.io/exclude: "true"
```

### Resource-level Exemptions

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    policies.kyverno.io/exclude: "require-resource-limits,require-non-root"
```

### PolicyException Resources

```yaml
apiVersion: kyverno.io/v2alpha1
kind: PolicyException
metadata:
  name: allow-privileged-for-cni
  namespace: kube-system
spec:
  exceptions:
  - policyName: disallow-privilege-escalation
    ruleNames:
    - prevent-privilege-escalation
  match:
    any:
    - resources:
        kinds:
        - Pod
        namespaces:
        - kube-system
        names:
        - "cilium-*"
```

## CI/CD Integration

### Pre-commit Validation

Add to `.pre-commit-config.yaml`:

```yaml
- repo: https://github.com/kyverno/kyverno
  rev: v1.10.0
  hooks:
  - id: kyverno-test
    args: ['test', './platform/security/kyverno/']
```

### GitHub Actions

```yaml
- name: Validate Kyverno Policies
  uses: kyverno/action-validate@v1
  with:
    policies: platform/security/kyverno/
    manifests: manifests/
```

### ArgoCD Pre-sync Hook

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  annotations:
    argocd.argoproj.io/hook: PreSync
spec:
  template:
    spec:
      containers:
      - name: policy-validation
        image: ghcr.io/kyverno/kyverno-cli:latest
        command:
        - kyverno
        - apply
        - /policies
        - --resource
        - /manifests
```

## Policy Testing

### Unit Tests

Create test files in `tests/`:

```yaml
# tests/require-labels-test.yaml
name: require-labels
policies:
- ../../policies/require-labels.yaml
resources:
- resource.yaml
results:
- policy: require-labels
  rule: check-required-labels
  resource: test-deployment
  kind: Deployment
  result: pass
```

Run tests:

```bash
kyverno test tests/
```

### Integration Tests

```bash
# Create test namespace
kubectl create namespace policy-test

# Try to create non-compliant resource
kubectl run test-pod \
  --image=nginx:latest \
  --namespace=policy-test
# Should be blocked by policies

# Create compliant resource
kubectl apply -f manifests/compliant-pod.yaml
# Should succeed
```

## Monitoring & Alerts

### Prometheus Metrics

Kyverno exports Prometheus metrics:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: kyverno-svc-metrics
  namespace: kyverno
spec:
  ports:
  - name: metrics
    port: 8000
    targetPort: 8000
  selector:
    app.kubernetes.io/name: kyverno
```

### Grafana Dashboard

Import Kyverno dashboard (ID: 13995):

```bash
kubectl port-forward -n mantl-system svc/grafana 3000:80
# Navigate to: Dashboards → Import → 13995
```

### Alerting

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: kyverno-alerts
  namespace: kyverno
spec:
  groups:
  - name: kyverno
    interval: 30s
    rules:
    - alert: KyvernoPolicyViolations
      expr: |
        increase(kyverno_policy_results_total{
          status="fail"
        }[5m]) > 10
      annotations:
        summary: "High number of policy violations detected"
```

## Policy Reports

### View Violations by Namespace

```bash
kubectl get policyreport -n ecommerce -o json | \
  jq '.results[] | select(.result == "fail") |
      {policy: .policy, rule: .rule, resource: .resource}'
```

### Generate Compliance Report

```bash
# Export all policy reports
kubectl get policyreport -A -o yaml > compliance-report.yaml

# Or use Kyverno CLI
kyverno jp query -i compliance-report.yaml \
  -q 'results[?result==`fail`].{policy: policy, namespace: metadata.namespace}'
```

## Best Practices

1. **Start with Audit Mode**: Test policies without enforcement first
2. **Use Exemptions Sparingly**: Document why exemptions are needed
3. **Test Thoroughly**: Use kyverno test with comprehensive test cases
4. **Monitor Violations**: Set up alerts for policy violations
5. **Version Control**: Keep policies in Git with proper review process
6. **Documentation**: Document policy intent and exemption criteria
7. **Regular Reviews**: Review and update policies quarterly

## Policy Development

### Creating New Policies

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: my-policy
  annotations:
    policies.kyverno.io/title: "Policy Title"
    policies.kyverno.io/category: "Security"
    policies.kyverno.io/severity: "high"
    policies.kyverno.io/description: |
      Detailed description of what this policy does and why.
spec:
  validationFailureAction: Audit  # Start with Audit
  background: true
  rules:
  - name: rule-name
    match:
      any:
      - resources:
          kinds:
          - Pod
    validate:
      message: "Validation failed message"
      pattern:
        spec:
          # validation logic
```

### Policy Validation

```bash
# Validate policy syntax
kyverno validate policy.yaml

# Test against resources
kyverno apply policy.yaml --resource test-resource.yaml

# Check for conflicts
kubectl get clusterpolicies -o yaml | \
  yq '.items[].spec.rules[].match'
```

## Troubleshooting

### Policy Not Applied

```bash
# Check policy status
kubectl describe clusterpolicy <policy-name>

# Check Kyverno logs
kubectl logs -n kyverno deploy/kyverno -f

# Verify webhook configuration
kubectl get validatingwebhookconfigurations | grep kyverno
```

### False Positives

```bash
# Add namespace exclusion
kubectl label namespace kube-system \
  kyverno.io/exclude=true

# Or use PolicyException for specific cases
```

### Performance Issues

```bash
# Check webhook latency
kubectl get --raw /metrics | grep webhook_latency

# Reduce background scans
helm upgrade kyverno kyverno/kyverno \
  --set backgroundController.enabled=false
```

## References

- [Kyverno Documentation](https://kyverno.io/docs/)
- [Policy Library](https://kyverno.io/policies/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
