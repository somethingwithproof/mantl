# Platform Security Quick Start Guide

## 🚀 Deploy a Secure Cluster in 10 Minutes

This guide shows you how to deploy a production-ready Kubernetes cluster with enterprise security on DigitalOcean or Linode in minutes.

## Prerequisites

- Terraform 1.10.0+
- Cloud provider account (DigitalOcean or Linode)
- Cloud provider API token
- `kubectl` installed

## Option 1: DigitalOcean DOKS (Recommended for Cost)

### Step 1: Create terraform.tfvars

```hcl
# terraform.tfvars
cluster_name = "my-secure-cluster"
region       = "nyc3"

# Kubernetes version
kubernetes_version = "1.31.3-do.0"

# Node configuration
system_node_count = 3
system_node_size  = "s-2vcpu-4gb"

workload_initial_count = 2
workload_min_count     = 2
workload_max_count     = 10
workload_node_size     = "s-4vcpu-8gb"

# Security: All enabled by default ✅
enable_cilium         = true  # Flow logs
enable_sealed_secrets = true  # Secrets encryption
enable_falco          = true  # Runtime security
enable_cert_manager   = true  # TLS automation

# Optional: Private access
enable_tailscale = false  # Set to true if you want VPN access
# tailscale_client_id = "your-client-id"
# tailscale_client_secret = "your-client-secret"

# Backup storage
enable_backup_bucket    = true
enable_flow_logs_export = true  # Export Cilium flows to Spaces
```

### Step 2: Set Environment Variables

```bash
export DIGITALOCEAN_TOKEN="your-do-api-token"
```

### Step 3: Deploy

```bash
cd infra/terraform/terraform/blueprints/do-doks

# Initialize Terraform
terraform init

# Review plan
terraform plan

# Deploy (takes ~8-10 minutes)
terraform apply
```

### Step 4: Configure kubectl

```bash
# Get kubeconfig
export DIGITALOCEAN_TOKEN="your-do-api-token"
export CLUSTER_ID=$(terraform output -raw cluster_id)

doctl kubernetes cluster kubeconfig save $CLUSTER_ID

# Verify cluster
kubectl get nodes
kubectl get pods --all-namespaces
```

### Step 5: Verify Platform Security

```bash
# Check Cilium (Flow Logs)
kubectl get pods -n cilium
kubectl exec -n cilium ds/cilium -- cilium status

# Check Sealed Secrets (KMS)
kubectl get pods -n sealed-secrets

# Check Falco (Runtime Security)
kubectl get pods -n falco

# Check Cert-Manager (TLS)
kubectl get pods -n cert-manager
```

**🎉 You now have a secure production cluster!**

**Monthly Cost:** ~$60-80 (vs $200-400 on AWS/GCP/Azure)

## Option 2: Linode LKE (Lowest Cost)

### Step 1: Create terraform.tfvars

```hcl
# terraform.tfvars
cluster_name = "my-secure-cluster"
region       = "us-east"

# Kubernetes version
kubernetes_version = "1.31"

# High availability control plane
enable_ha_control_plane = true

# Node configuration
system_node_count = 3
system_node_type  = "g6-standard-2"  # 2 vCPU, 4GB RAM

# Workload pools
workload_pools = [
  {
    name       = "apps"
    node_type  = "g6-standard-4"  # 4 vCPU, 8GB RAM
    node_count = 2
    min_nodes  = 2
    max_nodes  = 10
  }
]

# Security: All enabled by default ✅
enable_cilium         = true  # Flow logs
enable_sealed_secrets = true  # Secrets encryption
enable_falco          = true  # Runtime security
enable_cert_manager   = true  # TLS automation

# Optional: Private access
enable_tailscale = false  # Set to true if you want VPN access

# Backup storage
enable_backup_bucket    = true
enable_flow_logs_export = true  # Export Cilium flows to Object Storage
```

### Step 2: Set Environment Variables

```bash
export LINODE_TOKEN="your-linode-api-token"
```

### Step 3: Deploy

```bash
cd infra/terraform/terraform/blueprints/linode-lke

# Initialize Terraform
terraform init

# Review plan
terraform plan

# Deploy (takes ~8-10 minutes)
terraform apply
```

### Step 4: Configure kubectl

```bash
# Get kubeconfig
export KUBE_VAR=$(terraform output -raw kubeconfig | base64 -d)
echo "$KUBE_VAR" > ~/.kube/config-linode
export KUBECONFIG=~/.kube/config-linode

# Verify cluster
kubectl get nodes
kubectl get pods --all-namespaces
```

**🎉 You now have a secure production cluster!**

**Monthly Cost:** ~$50-70 (lowest cost option)

## Platform Security Features

### 1. Network Observability (Cilium + Hubble)

**View Flow Logs:**
```bash
# Port-forward Hubble UI
kubectl port-forward -n cilium svc/hubble-ui 8080:80

# Open browser
open http://localhost:8080

# Or use CLI
kubectl exec -n cilium deploy/hubble-relay -- hubble observe --follow
```

**Export to Storage:**
Flow logs are automatically exported to your backup bucket (Spaces/Object Storage) if `enable_flow_logs_export = true`.

**Query flows:**
```bash
# View DNS requests
hubble observe --type l7 --protocol dns

# View HTTP requests
hubble observe --type l7 --protocol http

# View drops/errors
hubble observe --verdict DROPPED
```

### 2. Secrets Encryption (Sealed Secrets)

**Install kubeseal CLI:**
```bash
# macOS
brew install kubeseal

# Linux
wget https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.25.0/kubeseal-0.25.0-linux-amd64.tar.gz
tar xfz kubeseal-0.25.0-linux-amd64.tar.gz
sudo install -m 755 kubeseal /usr/local/bin/kubeseal
```

**Encrypt a secret:**
```bash
# Create secret (don't apply yet!)
kubectl create secret generic my-database-password \
  --from-literal=password='super-secret-password' \
  --dry-run=client -o yaml > secret.yaml

# Encrypt it
kubeseal -f secret.yaml -w sealed-secret.yaml

# Safe to commit to Git!
git add sealed-secret.yaml
git commit -m "Add database password (encrypted)"

# Apply to cluster (only cluster can decrypt)
kubectl apply -f sealed-secret.yaml

# Verify
kubectl get secret my-database-password -o yaml
```

**Backup encryption keys:**
```bash
# Backup keys to secure location
kubectl get secret -n sealed-secrets \
  -l sealedsecrets.bitnami.com/sealed-secrets-key \
  -o yaml > sealed-secrets-keys-backup.yaml

# Store securely (1Password, vault, etc.)
```

### 3. Runtime Security (Falco)

**View alerts:**
```bash
# Watch real-time alerts
kubectl logs -n falco daemonset/falco -f

# Filter by priority
kubectl logs -n falco daemonset/falco | grep "Priority: Warning"
kubectl logs -n falco daemonset/falco | grep "Priority: Critical"
```

**Common alerts you'll see:**
```
Warning: Shell spawned in container (container=nginx-123 command=/bin/sh)
Warning: Sensitive file opened (file=/etc/shadow container=app-456)
Critical: Unexpected network connection (destination=suspicious-ip.com)
```

**Configure webhook alerts:**
```hcl
# In terraform.tfvars
enable_falco      = true
falco_webhook_url = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
```

### 4. TLS Certificates (Cert-Manager)

**Create Let's Encrypt issuer:**
```yaml
# letsencrypt-prod.yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

```bash
kubectl apply -f letsencrypt-prod.yaml
```

**Request certificate:**
```yaml
# certificate.yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-com
  namespace: default
spec:
  secretName: example-com-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - example.com
  - www.example.com
```

```bash
kubectl apply -f certificate.yaml

# Check status
kubectl get certificate
kubectl describe certificate example-com
```

**Auto-TLS for ingress:**
```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - example.com
    secretName: example-com-tls  # Cert-manager creates this
  rules:
  - host: example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web
            port:
              number: 80
```

### 5. Private Access (Tailscale - Optional)

**Setup:**
1. Create OAuth client at https://login.tailscale.com/admin/settings/oauth
2. Enable in terraform.tfvars:
   ```hcl
   enable_tailscale        = true
   tailscale_client_id     = "tskey-api-..."
   tailscale_client_secret = "tskey-..."
   ```
3. Apply Terraform changes

**Connect:**
```bash
# Install Tailscale on your machine
# macOS
brew install tailscale

# Start Tailscale
tailscale up

# Access cluster via Tailscale network
kubectl get nodes
```

## Monitoring

### Grafana Dashboard

Access the security monitoring dashboard:

```bash
# Port-forward Grafana
kubectl port-forward -n monitoring svc/grafana 3000:80

# Open browser
open http://localhost:3000

# Default credentials (change immediately!)
username: admin
password: <get from secret>
kubectl get secret -n monitoring grafana -o jsonpath='{.data.admin-password}' | base64 -d
```

Navigate to: **Dashboards → Security Monitoring**

Panels:
- Cilium Flow Rate
- Falco Alert Rate
- Sealed Secrets Operations
- Certificate Expiration
- Tailscale Connection Status

### Prometheus Alerts

View active alerts:
```bash
# Port-forward Prometheus
kubectl port-forward -n monitoring svc/prometheus-server 9090:80

# Open browser
open http://localhost:9090/alerts
```

## Maintenance

### Update Cluster

```bash
# Pull latest changes
git pull origin main

# Review changes
terraform plan

# Apply updates
terraform apply
```

### Rotate Sealed Secrets Keys

Keys rotate automatically every 30 days. To manually rotate:

```bash
# Delete old keys (new ones auto-generated)
kubectl delete secret -n sealed-secrets \
  -l sealedsecrets.bitnami.com/sealed-secrets-key

# Wait for controller to generate new keys
sleep 30

# Re-seal all secrets with new keys
for file in secrets/*.yaml; do
  kubeseal -f "$file" -w "sealed-${file}"
done
```

### Backup

Velero backups are configured by default:

```bash
# Check backup schedules
kubectl get schedules -n velero

# Trigger manual backup
velero backup create manual-backup --include-namespaces default,production

# Check backup status
velero backup describe manual-backup
```

## Cost Optimization

### Spot Instances (DigitalOcean only)

Add spot instance pool for non-critical workloads:

```hcl
# terraform.tfvars
enable_spot_pool = true
spot_node_count  = 3
spot_node_size   = "s-4vcpu-8gb"
spot_max_price   = 0.05  # 50% discount
```

**Savings:** 50-60% on workload nodes

### Auto-scaling

Workload nodes auto-scale by default:

```hcl
# terraform.tfvars
workload_min_count = 2    # Minimum nodes (nights/weekends)
workload_max_count = 10   # Maximum nodes (peak traffic)
```

**Savings:** Pay only for what you use

### Resource Right-sizing

Monitor resource usage:

```bash
# Install metrics-server (if not already)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# View node resource usage
kubectl top nodes

# View pod resource usage
kubectl top pods --all-namespaces
```

Adjust node sizes based on actual usage.

## Troubleshooting

### Cilium Not Starting

```bash
# Check status
kubectl get pods -n cilium

# View logs
kubectl logs -n cilium daemonset/cilium

# Common issue: Conflicting CNI
# Solution: Remove old CNI before deploying Cilium
```

### Sealed Secrets Decryption Fails

```bash
# Check controller
kubectl logs -n sealed-secrets deploy/sealed-secrets-controller

# Verify keys exist
kubectl get secrets -n sealed-secrets

# Solution: Restore from backup
kubectl apply -f sealed-secrets-keys-backup.yaml
```

### Falco High CPU

```bash
# Check rules
kubectl logs -n falco daemonset/falco | tail -100

# Solution: Tune rules in terraform.tfvars
# Increase priority threshold or disable noisy rules
```

## Next Steps

1. ✅ **Deploy your applications**
2. ✅ **Configure domain and TLS** (cert-manager + Let's Encrypt)
3. ✅ **Set up monitoring alerts** (Prometheus → Slack/PagerDuty)
4. ✅ **Enable Velero scheduled backups**
5. ✅ **Test disaster recovery**
6. ✅ **Configure Falco webhook** for security alerts
7. ✅ **Review Hubble flow logs** for network anomalies

## Support

- **Platform Security Module:** Review `/infra/terraform/terraform/modules/platform-security/`
- **Implementation Guide:** See `FEATURE_PARITY_IMPLEMENTATION.md`
- **Provider Comparison:** See `UPDATED_CLOUD_COMPARISON.md`
- **E2E Tests:** `pytest tests/e2e/platform/`

## Example Application Deployment

```yaml
# app.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      # Use sealed secret
      env:
      - name: DATABASE_PASSWORD
        valueFrom:
          secretKeyRef:
            name: my-database-password
            key: password
      containers:
      - name: web
        image: nginx:1.27-alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: web
spec:
  selector:
    app: web
  ports:
  - port: 80
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - myapp.example.com
    secretName: myapp-tls
  rules:
  - host: myapp.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web
            port:
              number: 80
```

```bash
kubectl apply -f app.yaml

# Watch deployment
kubectl get pods -w

# Check certificate
kubectl get certificate

# Check Falco for any security alerts
kubectl logs -n falco daemonset/falco | grep "web-app"

# View in Hubble
kubectl port-forward -n cilium svc/hubble-ui 8080:80
```

**Congratulations! You have a production-ready, secure Kubernetes cluster with enterprise features at 60-80% cost savings!** 🎉
