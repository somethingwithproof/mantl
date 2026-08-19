# Mantl Platform Quickstart

**Last Updated**: 2025-12-27

Deploy a production-ready Kubernetes platform with observability, security, and GitOps in minutes.

## Quick Start (Recommended)

The fastest way to get started is using the built-in automation:

### Local Development (kind cluster)

```bash
git clone https://github.com/thomasvincent/mantl.git
cd mantl
make install-dev
```

This single command:
- Creates a kind cluster with proper configuration
- Installs ArgoCD and all platform components
- Deploys example applications
- Sets up the observability stack

### Interactive Setup (Custom Configuration)

For more control over the installation:

```bash
make wizard
```

The wizard guides you through:
- Environment selection (dev, staging, production)
- Cloud provider choice
- Resource profile (small, medium, full)
- Component selection
- Domain configuration

### CLI Tool (Advanced Users)

```bash
./bin/mantl init           # Initialize configuration
./bin/mantl status         # Check platform health
./bin/mantl dashboards     # Open dashboards
./bin/mantl deploy <app>   # Deploy applications
```

## Manual Deployment

If you prefer manual control or need to customize the deployment:

### Prerequisites

- Kubernetes cluster 1.27+ (EKS, GKE, AKS, DOKS, LKE, or kind for local)
- `kubectl` configured for your cluster
- Git repository access (GitHub, GitLab, etc.)

### Manual Setup Steps

### 1. Install ArgoCD

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

Wait for ArgoCD to be ready:
```bash
kubectl wait --for=condition=available --timeout=300s deployment/argocd-server -n argocd
```

### 2. Deploy Platform Stack

Deploy all platform components using the app-of-apps pattern:

```bash
kubectl apply -f clusters/production/platform-apps.yaml
```

This deploys in dependency order (sync waves):
- **Wave 1**: Cilium (CNI + service mesh)
- **Wave 2**: cert-manager (TLS automation)
- **Wave 3**: External Secrets, Kyverno, Crossplane (secrets + policy + IaC)
- **Wave 4**: Prometheus, Loki, Tempo (observability backends)
- **Wave 5**: OTel Collector, Falco (telemetry + security)
- **Wave 6**: Harbor, External DNS (registry + DNS)

### 3. Monitor Deployment

**Watch ArgoCD sync status:**
```bash
kubectl get applications -n argocd -w
```

**Access ArgoCD UI:**
```bash
# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port forward to UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Open https://localhost:8080
# Username: admin
# Password: (from command above)
```

### 4. Access Observability Stack

**Grafana (metrics + dashboards):**
```bash
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
# Open http://localhost:3000
# Default credentials: admin/prom-operator
```

**Tempo (distributed tracing):**
```bash
kubectl port-forward -n observability svc/tempo 3100:3100
# Query via Grafana or directly at http://localhost:3100
```

**Prometheus (metrics):**
```bash
kubectl port-forward -n monitoring svc/prometheus-prometheus 9090:9090
# Open http://localhost:9090
```

## Optional Configuration

### DNS Automation (Cloud Providers)

Choose one ExternalDNS configuration for your provider:

```bash
# AWS Route53
kubectl apply -f clusters/production/apps/external-dns-aws.yaml

# Google Cloud DNS
kubectl apply -f clusters/production/apps/external-dns-gcp.yaml

# Azure DNS
kubectl apply -f clusters/production/apps/external-dns-azure.yaml

# DigitalOcean
kubectl apply -f clusters/production/apps/external-dns-digitalocean.yaml

# Linode
kubectl apply -f clusters/production/apps/external-dns-linode.yaml
```

### TLS Certificate Issuers

Choose one DNS-01 ACME issuer for automated TLS:

```bash
# Cloudflare
kubectl apply -f clusters/production/apps/cert-manager-issuer-dns-cloudflare.yaml

# Route53
kubectl apply -f clusters/production/apps/cert-manager-issuer-dns-route53.yaml

# Google Cloud DNS
kubectl apply -f clusters/production/apps/cert-manager-issuer-dns-gcp.yaml
```

### Secrets Management

Configure External Secrets to sync from your cloud provider:

1. **AWS Secrets Manager**: Set up IRSA for the external-secrets pod
2. **GCP Secret Manager**: Configure Workload Identity
3. **Azure Key Vault**: Set up Managed Identity
4. **HashiCorp Vault**: Configure Vault authentication

See `docs/secrets-and-issuers.md` for detailed configuration.

## Verification

### Check All Platform Components

```bash
# All applications should show "Healthy" and "Synced"
kubectl get applications -n argocd

# Check all platform pods are running
kubectl get pods -A | grep -E "argocd|monitoring|observability|kube-system|cert-manager|kyverno|falco"
```

### Test Observability Stack

**1. Generate test metrics:**
```bash
kubectl run test-pod --image=nginx --restart=Never
kubectl delete pod test-pod
```

**2. Send test traces:**
```bash
# Port forward OTel Collector
kubectl port-forward -n observability svc/otel-collector 4317:4317

# Use any OTLP-compatible client to send traces to localhost:4317
```

**3. View in Grafana:**
- Navigate to http://localhost:3000
- Explore → Select "Prometheus" data source
- Query: `up{job="kubernetes-pods"}`

**4. View traces in Grafana:**
- Navigate to http://localhost:3000
- Explore → Select "Tempo" data source
- Query for traces by service name or trace ID

## Next Steps

1. **Deploy your first application**: See `applications/templates/web-service/base/`
2. **Configure ingress**: Deploy Gateway API resources in `platform/ingress/gateway-api/`
3. **Set up CI/CD**: Configure ArgoCD webhooks or Image Updater for automated deployments
4. **Review policies**: Check Kyverno policies in `policies/kyverno/`
5. **Enable runtime security**: Review Falco alerts in Grafana

## Troubleshooting

**ArgoCD sync failures:**
```bash
# Get detailed sync status
kubectl describe application <app-name> -n argocd

# View application logs
kubectl logs -n argocd deployment/argocd-application-controller
```

**Pod failures:**
```bash
# Check pod status
kubectl get pods -A | grep -v Running

# View pod logs
kubectl logs -n <namespace> <pod-name>

# Describe pod for events
kubectl describe pod -n <namespace> <pod-name>
```

## Documentation

- Platform Components: `docs/platform-apps.rst`
- Architecture Decisions: `docs/ara/README.md`
- Observability Stack: `docs/ara/ARA-0005-observability-stack.md`
- Security Baseline: `docs/security.md`
- Multi-Cloud Setup: `docs/multicloud.rst`
