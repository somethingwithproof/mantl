# Troubleshooting Guide

Common issues and solutions for the mantl platform.

## Table of Contents

- [Cluster Issues](#cluster-issues)
- [Networking Issues](#networking-issues)
- [Security Policy Issues](#security-policy-issues)
- [Application Deployment Issues](#application-deployment-issues)
- [Observability Issues](#observability-issues)
- [Infrastructure Issues](#infrastructure-issues)

---

## Cluster Issues

### Cluster Not Accessible

**Symptoms:**
- `kubectl` commands timeout
- Unable to connect to cluster API

**Diagnosis:**
```bash
# Check cluster status
kubectl cluster-info
kubectl get nodes

# Check VPN/bastion connection if using private endpoint
# AWS EKS private endpoint is default for security
```

**Solutions:**

1. **Private endpoint (production)**: Verify VPN or bastion host connection
   ```bash
   # Connect through bastion
   ssh -L 6443:eks-endpoint:443 bastion-host
   export KUBECONFIG=~/.kube/config
   ```

2. **Enable public endpoint (dev/testing only)**:
   - In `infra/terraform/terraform/blueprints/aws-eks/variables.tf`
   - Set `cluster_endpoint_public_access = true`
   - Apply Terraform: `terraform apply`

3. **AWS credentials**: Verify AWS CLI configuration
   ```bash
   aws sts get-caller-identity
   aws eks update-kubeconfig --name mantl --region us-east-1
   ```

### Nodes Not Ready

**Symptoms:**
- `kubectl get nodes` shows NotReady status
- Pods stuck in Pending state

**Diagnosis:**
```bash
kubectl get nodes -o wide
kubectl describe node <node-name>

# Check node system pods
kubectl get pods -n kube-system -o wide
```

**Common Causes:**

1. **VPC CNI issues**: Check vpc-cni daemonset
   ```bash
   kubectl logs -n kube-system -l k8s-app=aws-node
   kubectl rollout restart daemonset -n kube-system aws-node
   ```

2. **Insufficient resources**: Check node capacity
   ```bash
   kubectl describe node <node-name> | grep -A 5 "Allocated resources"
   ```

3. **Network connectivity**: Verify security groups allow node-to-node communication

---

## Networking Issues

### Pods Cannot Communicate

**Symptoms:**
- Pods cannot reach services
- Network timeouts between pods

**Diagnosis:**
```bash
# Test pod-to-pod connectivity
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -- sh
# Inside pod: ping <service-name>.<namespace>.svc.cluster.local

# Check Cilium status
kubectl -n kube-system exec -it cilium-<pod> -- cilium status
```

**Solutions:**

1. **Cilium issues**: Restart Cilium pods
   ```bash
   kubectl rollout restart daemonset -n kube-system cilium
   kubectl rollout restart deployment -n kube-system cilium-operator
   ```

2. **Network policies**: Check if policies are blocking traffic
   ```bash
   kubectl get networkpolicies --all-namespaces
   kubectl describe networkpolicy <policy-name> -n <namespace>
   ```

### External DNS Not Working

**Symptoms:**
- DNS records not created in Route53
- Services not accessible via domain name

**Diagnosis:**
```bash
# Check External DNS logs
kubectl logs -n external-dns deployment/external-dns

# Verify IRSA permissions
kubectl describe sa external-dns -n external-dns
```

**Solutions:**

1. **IRSA permissions**: Verify IAM role has Route53 access
   ```bash
   # Check role ARN annotation
   kubectl get sa external-dns -n external-dns -o yaml | grep eks.amazonaws.com/role-arn
   ```

2. **Route53 zone**: Verify zone ARNs in Terraform variables
   - Check `var.route53_zone_arns` in `infra/terraform/terraform/blueprints/aws-eks/variables.tf`

---

## Security Policy Issues

### Pods Failing Admission (Kyverno)

**Symptoms:**
- Pod creation fails with policy violation
- "admission webhook denied the request"

**Diagnosis:**
```bash
# Check Kyverno policy reports
kubectl get policyreport --all-namespaces

# View specific policy
kubectl describe clusterpolicy <policy-name>

# Check Kyverno logs
kubectl logs -n kyverno deployment/kyverno -c kyverno
```

**Common Policy Violations:**

1. **Image registry not allowed**:
   ```
   Error: Images must come from approved registries
   ```
   **Solution**: Use approved registries (ghcr.io/thomasvincent/*, docker.io/library/*, quay.io/prometheus/*)

2. **Image not signed**:
   ```
   Error: image verification failed
   ```
   **Solution**:
   - Sign images with Cosign in CI
   - Or add namespace to policy exclusions (dev/test only)
   ```bash
   # Temporarily exempt namespace (dev only)
   kubectl label namespace <namespace> policy.kyverno.io/exclude=true
   ```

3. **Missing required labels**:
   ```
   Error: required labels missing
   ```
   **Solution**: Add required labels to pod spec
   ```yaml
   metadata:
     labels:
       app: myapp
       version: v1.0.0
   ```

### Image Pull Failures

**Symptoms:**
- Pods stuck in ImagePullBackOff
- "Failed to pull image" errors

**Diagnosis:**
```bash
kubectl describe pod <pod-name>
kubectl get events --sort-by=.metadata.creationTimestamp
```

**Solutions:**

1. **Registry authentication**: Create image pull secret
   ```bash
   kubectl create secret docker-registry ghcr-secret \
     --docker-server=ghcr.io \
     --docker-username=$GITHUB_USER \
     --docker-password=$GITHUB_TOKEN \
     -n <namespace>
   ```

2. **IRSA for ECR**: Verify service account has ECR permissions
   ```bash
   kubectl describe sa <service-account> -n <namespace>
   ```

---

## Application Deployment Issues

### Helm Chart Installation Fails

**Symptoms:**
- `helm install` fails
- Chart templates don't render

**Diagnosis:**
```bash
# Dry-run to see rendered templates
helm install <release> <chart> --dry-run --debug

# Check Helm releases
helm list --all-namespaces
```

**Common Issues:**

1. **Missing values**: Provide required values
   ```bash
   helm install <release> <chart> -f values.yaml
   ```

2. **Namespace doesn't exist**: Create namespace first
   ```bash
   kubectl create namespace <namespace>
   helm install <release> <chart> -n <namespace>
   ```

### Kustomize Build Fails

**Symptoms:**
- `kustomize build` errors
- ArgoCD sync fails

**Diagnosis:**
```bash
# Test kustomize build locally
kustomize build <overlay-dir> --enable-helm

# Validate all builds
./ci/validate-kustomize.sh
```

**Solutions:**

1. **Missing resources**: Check all resource paths in kustomization.yaml
2. **Helm chart issues**: Verify Helm chart repo and version
3. **Circular dependencies**: Check for resource dependencies

---

## Observability Issues

### Metrics Not Showing in Prometheus

**Symptoms:**
- Dashboards empty
- Metrics queries return no data

**Diagnosis:**
```bash
# Check Prometheus targets
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Open http://localhost:9090/targets

# Check ServiceMonitor resources
kubectl get servicemonitors --all-namespaces
```

**Solutions:**

1. **ServiceMonitor not discovered**: Add prometheus.io labels
   ```yaml
   metadata:
     labels:
       prometheus: kube-prometheus
   ```

2. **Metrics endpoint not accessible**: Verify service and pod labels match
   ```bash
   kubectl get svc <service-name> -o yaml
   kubectl get pods -l app=<app-name> -o wide
   ```

### Logs Not Appearing

**Symptoms:**
- Logs not in Loki/aggregator
- Fluent Bit pods in error

**Diagnosis:**
```bash
# Check Fluent Bit status
kubectl logs -n logging daemonset/fluent-bit

# Check Loki status
kubectl logs -n monitoring deployment/loki
```

**Solutions:**

1. **Parser issues**: Check Fluent Bit config
   ```bash
   kubectl get configmap -n logging fluent-bit -o yaml
   ```

2. **Storage issues**: Verify Loki has sufficient storage
   ```bash
   kubectl get pvc -n monitoring
   ```

---

## Infrastructure Issues

### Terraform Apply Fails

**Symptoms:**
- Terraform plan/apply errors
- Infrastructure changes not applied

**Diagnosis:**
```bash
cd infra/terraform/terraform/blueprints/aws-eks
terraform init
terraform plan
```

**Common Issues:**

1. **State lock**: Wait for other operations or force-unlock
   ```bash
   terraform force-unlock <lock-id>
   ```

2. **Insufficient permissions**: Verify AWS credentials
   ```bash
   aws sts get-caller-identity
   # Should show account with sufficient permissions
   ```

3. **Resource quota exceeded**: Check AWS service quotas
   ```bash
   aws service-quotas get-service-quota \
     --service-code ec2 \
     --quota-code L-1216C47A
   ```

### VPC Flow Logs Not Collecting

**Symptoms:**
- No flow logs in CloudWatch
- Network monitoring gaps

**Diagnosis:**
```bash
# Check Flow Log status
aws ec2 describe-flow-logs --filter Name=resource-id,Values=<vpc-id>

# Check CloudWatch log group
aws logs describe-log-groups --log-group-name-prefix /aws/vpc/
```

**Solutions:**

1. **IAM role missing**: Verify Flow Log IAM role in Terraform
   - Check `aws_iam_role.vpc_flow_logs` in main.tf

2. **Log group not created**: Ensure CloudWatch log group exists
   - Check `aws_cloudwatch_log_group.vpc_flow_logs` in main.tf

---

## Getting Help

### Diagnostic Commands

Collect comprehensive diagnostics:

```bash
#!/bin/bash
# Save this as diagnose.sh

echo "=== Cluster Info ==="
kubectl cluster-info
kubectl version

echo "=== Node Status ==="
kubectl get nodes -o wide

echo "=== System Pods ==="
kubectl get pods -n kube-system

echo "=== Recent Events ==="
kubectl get events --all-namespaces --sort-by=.metadata.creationTimestamp | tail -20

echo "=== Policy Reports ==="
kubectl get policyreport --all-namespaces

echo "=== Cilium Status ==="
kubectl -n kube-system exec -it cilium-<pod> -- cilium status
```

### Support Resources

- **GitHub Issues**: https://github.com/thomasvincent/mantl/issues
- **CNCF Slack**: Join relevant project channels (Kyverno, Cilium, etc.)
- **Documentation**: Check `docs/` directory for component-specific guides

### Before Opening an Issue

Include:
1. Output of diagnostic commands above
2. Relevant pod/deployment YAML
3. Recent events: `kubectl get events --all-namespaces`
4. Component versions: `kubectl version`, `terraform version`, etc.
5. Steps to reproduce the issue
