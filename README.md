# Mantl 2026 — Cloud-Native Platform Engineering Toolkit

[![CI](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml/badge.svg)](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CNCF](https://img.shields.io/badge/CNCF-Graduated%20%26%20Incubating-blue)](https://www.cncf.io/projects/)

Mantl is a **Kubernetes-native, GitOps-driven, secure-by-default** platform engineering toolkit for building production-ready, multi-cloud infrastructure. Built entirely on CNCF Graduated and Incubating projects.

## 🌟 What You Get

Deploy a complete, production-ready Kubernetes platform in minutes:

| Layer | Component | CNCF Status |
|-------|-----------|-------------|
| **GitOps** | ArgoCD | Graduated |
| **Networking** | Cilium (eBPF + Gateway API) | Graduated |
| **TLS** | cert-manager | Graduated |
| **Policy** | Kyverno | Graduated |
| **Secrets** | External Secrets Operator | Graduated |
| **Metrics** | Prometheus + Grafana | Graduated |
| **Logs** | Grafana Loki | — |
| **Traces** | Grafana Tempo | — |
| **Telemetry** | OpenTelemetry Collector | Graduated |
| **Security** | Falco (runtime) | Graduated |
| **Registry** | Harbor (on-prem) | Graduated |
| **DNS** | ExternalDNS | Incubating |

## 🚀 Quickstart

### Prerequisites

- `kubectl` 1.28+
- `terraform` (or `tofu`) 1.5+
- Cloud provider credentials (AWS/GCP/Azure/DO/Linode)

### Deploy

```bash
# 1. Clone repository
git clone https://github.com/thomasvincent/mantl.git
cd mantl

# 2. Create Kubernetes cluster (pick one)
terraform -chdir=terraform/blueprints/aws-eks init && terraform -chdir=terraform/blueprints/aws-eks apply
# OR: gcp-gke, azure-aks, do-doks, linode-lke

# 3. Bootstrap platform via GitOps
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl apply -f clusters/production/app-of-apps.yaml

# 4. Access ArgoCD
kubectl port-forward svc/argocd-server -n argocd 8080:443
# Get password: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

### Local Development (kind)

```bash
# Create local cluster
kind create cluster --name mantl

# Deploy platform
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl apply -f clusters/dev/app-of-apps.yaml
```

## 📁 Project Structure

```
mantl/
├── terraform/blueprints/       # Infrastructure as Code
│   ├── aws-eks/               # Amazon EKS
│   ├── gcp-gke/               # Google GKE
│   ├── azure-aks/             # Azure AKS
│   ├── do-doks/               # DigitalOcean DOKS
│   └── linode-lke/            # Linode LKE
│
├── platform/                   # Kubernetes platform components
│   ├── base/
│   │   ├── argocd/            # GitOps engine
│   │   ├── cilium/            # CNI + Gateway API + Network Policy
│   │   └── cert-manager/      # TLS automation
│   ├── observability/
│   │   ├── prometheus/        # Metrics + Alertmanager + Grafana
│   │   ├── loki/              # Log aggregation
│   │   ├── tempo/             # Distributed tracing
│   │   └── otel-collector/    # Unified telemetry ingestion
│   ├── security/
│   │   └── falco/             # Runtime security
│   ├── secrets/
│   │   └── external-secrets/  # Cloud secrets integration
│   ├── dns/
│   │   └── external-dns/      # Automatic DNS management
│   ├── registry/
│   │   └── harbor/            # Container registry (on-prem)
│   └── ingress/
│       └── gateway-api/       # Gateway API resources
│
├── policies/kyverno/           # Policy as Code
│   ├── policies/
│   │   ├── require-labels.yaml
│   │   ├── restrict-registries.yaml
│   │   └── verify-image-signatures.yaml
│   └── values.yaml
│
├── clusters/                   # GitOps cluster configs
│   ├── dev/                   # Development environment
│   ├── staging/               # Staging environment
│   └── production/            # Production environment
│       ├── app-of-apps.yaml   # Root ArgoCD application
│       └── platform-apps.yaml # Platform components
│
├── applications/               # Sample workloads
│   └── examples/
│       └── hello/             # Hello world app
│
└── scripts/                    # Helper scripts
    ├── bootstrap-cluster.sh   # Initial setup
    ├── set-cosign-key.sh      # Image signing
    └── set-overlay.sh         # Kustomize helper
```

## 🔧 Platform Components

### Networking (Cilium)

- **eBPF-based CNI** — High-performance packet processing
- **Gateway API** — Modern ingress (HTTPRoute, GRPCRoute, TLSRoute)
- **Network Policies** — Zero-trust pod-to-pod security
- **Hubble** — Network observability UI
- **WireGuard encryption** — Node-to-node encryption
- **Load Balancer IPAM** — L2 announcements for bare metal

### Observability Stack

```
┌─────────────────────────────────────────────────────────────┐
│                        Grafana                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Prometheus  │  │    Loki     │  │   Tempo     │          │
│  │  (metrics)  │  │   (logs)    │  │  (traces)   │          │
│  └──────▲──────┘  └──────▲──────┘  └──────▲──────┘          │
└─────────┼────────────────┼────────────────┼─────────────────┘
          │                │                │
     ┌────┴────────────────┴────────────────┴────┐
     │         OpenTelemetry Collector            │
     │  (OTLP, Jaeger, Zipkin, Prometheus)       │
     └────────────────────▲───────────────────────┘
                          │
              ┌───────────┴───────────┐
              │   Your Applications   │
              └───────────────────────┘
```

### Security

| Component | Purpose |
|-----------|---------|
| **Kyverno** | Admission policies (require labels, restrict registries, verify signatures) |
| **Falco** | Runtime threat detection |
| **External Secrets** | Sync secrets from AWS/GCP/Azure/Vault |
| **cert-manager** | Automated TLS from Let's Encrypt or cloud CA |
| **Cosign** | Container image signing and verification |

### Secrets Providers

Choose one provider in `platform/secrets/external-secrets/kustomization.yaml`:

```yaml
resources:
  - namespace.yaml
  # Uncomment ONE:
  - providers/aws-secrets-manager.yaml
  # - providers/gcp-secret-manager.yaml
  # - providers/azure-key-vault.yaml
  # - providers/hashicorp-vault.yaml
```

### DNS Providers

Configure in `platform/dns/external-dns/values.yaml`:

```yaml
provider: cloudflare  # or: aws, google, azure, digitalocean
domainFilters:
  - example.com
```

## ☁️ Supported Cloud Providers

| Provider | Blueprint | Workload Identity | DNS | Status |
|----------|-----------|-------------------|-----|--------|
| AWS | `aws-eks` | ✅ IRSA | Route53 | ✅ Ready |
| GCP | `gcp-gke` | ✅ Workload Identity | Cloud DNS | ✅ Ready |
| Azure | `azure-aks` | ✅ Managed Identity | Azure DNS | ✅ Ready |
| DigitalOcean | `do-doks` | API Token | DO DNS | ✅ Ready |
| Linode | `linode-lke` | API Token | Linode DNS | ✅ Ready |
| Oracle Cloud | — | Instance Principal | OCI DNS | 🚧 Planned |
| OpenStack | — | — | Designate | 🚧 Planned |
| Bare Metal | k3s | — | RFC2136 | 🚧 Planned |

## 🔐 Policy Enforcement

Kyverno policies are deployed in **Audit** mode by default. To enforce:

```yaml
# policies/kyverno/policies/require-labels.yaml
spec:
  validationFailureAction: Enforce  # Change from Audit
```

### Included Policies

1. **require-labels** — All pods must have `app.kubernetes.io/name` and `app.kubernetes.io/managed-by`
2. **restrict-registries** — Only allow images from approved registries (ghcr.io, gcr.io, quay.io, etc.)
3. **verify-image-signatures** — Verify Sigstore/Cosign signatures on container images

## 📊 Deployment Order (Sync Waves)

ArgoCD deploys components in order:

| Wave | Components |
|------|------------|
| 1 | Cilium (CNI must be first) |
| 2 | cert-manager (TLS prerequisite) |
| 3 | External Secrets, Kyverno |
| 4 | Prometheus, Loki, Tempo |
| 5 | OTel Collector, Falco |
| 6 | Harbor, External DNS |

## 🛠️ Development

### Prerequisites

```bash
pip install -r requirements.txt
pip install -r requirements-test.txt
pre-commit install
```

### Testing

```bash
# Validate Terraform
make terraform-validate

# Test Kyverno policies
make policy-test

# Run all checks
make test
```

### Local Cluster Testing

```bash
# Create kind cluster
kind create cluster --name mantl-test

# Install platform
kubectl apply -k clusters/dev

# Run integration tests
./scripts/test-platform.sh
```

## 📚 Documentation

- [Quickstart Guide](docs/quickstart-platform.md)
- [Security Baseline](docs/security.md)
- [Secrets Management](docs/secrets-and-issuers.md)
- [Multi-Cloud Patterns](docs/multicloud.rst)

### Architecture Decision Records

- [ARA-0001: Kubernetes First](docs/ara/ARA-0001-kubernetes-first.md)
- [ARA-0002: GitOps and Delivery](docs/ara/ARA-0002-gitops-and-delivery.md)
- [ARA-0003: Policy as Code](docs/ara/ARA-0003-policy-as-code.md)
- [ARA-0004: Secrets Strategy](docs/ara/ARA-0004-secrets-strategy.md)

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, code style, and PR process.

## 📜 License

Apache License 2.0 — see [LICENSE](LICENSE).

## 🙏 Acknowledgments

Mantl was originally created by Cisco Cloud in 2015 for Mesos/Marathon orchestration. This 2026 modernization rebuilds the platform on Kubernetes-native, CNCF-backed technologies while maintaining the "batteries-included" philosophy.

---

**Built with ❤️ for Platform Engineers**
