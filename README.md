![image](./docs/_static/mantl-logo.png)

# Mantl — Modern Platform Engineering Toolkit

[![CI](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml/badge.svg)](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Mantl is a **Kubernetes-first, GitOps-driven, secure-by-default** platform engineering toolkit for building production-ready, multi-cloud distributed services. Originally created for Mesos/Marathon orchestration, Mantl has been modernized to embrace cloud-native ecosystems while maintaining its "batteries-included" philosophy.

## 🌟 Overview

Mantl provides everything you need to build, deploy, and operate production Kubernetes platforms across multiple cloud providers:

- **Multi-Cloud Native**: Pre-configured blueprints for AWS (EKS), GCP (GKE), Azure (AKS), DigitalOcean (DOKS), and Linode (LKE)
- **GitOps-First**: ArgoCD-based continuous delivery with app-of-apps pattern
- **Security Built-In**: SBOM generation, image signing (Cosign), policy enforcement (Kyverno/OPA)
- **Modern Networking**: Gateway API support with Cilium or Envoy Gateway
- **Zero-Touch DNS & TLS**: ExternalDNS and cert-manager with provider-specific configurations
- **Spinnaker Integration**: Multi-cloud continuous deployment pipelines
- **Developer Experience**: Templates, examples, and best practices

## 🎯 Key Features

### Infrastructure as Code
- **Terraform Blueprints**: Production-ready infrastructure modules for all major cloud providers
- **Hardened Defaults**: IMDSv2, encryption, least-privilege IAM, private networking
- **Global DNS**: Route53 failover, health-checked multi-region routing
- **Workload Identity**: Native cloud provider identity integration (AWS IRSA, GCP Workload Identity, Azure Managed Identity)

### GitOps & Delivery
- **ArgoCD Platform**: Declarative, Git-driven application deployment
- **Environment Overlays**: Dev, staging, production with Kustomize
- **Spinnaker Pipelines**: Advanced deployment strategies (blue/green, canary)
- **Automated Syncing**: Continuous reconciliation of desired state

### Security & Compliance
- **Policy as Code**: Kyverno admission policies and OPA Gatekeeper rules
- **Image Scanning**: Integration with vulnerability scanners
- **Secrets Management**: External Secrets Operator with cloud provider integration
- **SBOM & Provenance**: Supply chain security with Cosign and in-toto
- **Network Policies**: Zero-trust networking with Cilium

### Observability
- **Metrics**: Prometheus & Grafana (via platform addons)
- **Logging**: Structured logging with cloud-native backends
- **Tracing**: Distributed tracing support
- **Service Mesh**: Optional Istio/Linkerd integration

## 📋 Prerequisites

- **Kubernetes**: 1.27+ (EKS, GKE, AKS, DOKS, LKE, or kind for local)
- **Tools**:
  - `kubectl` 1.27+
  - `terraform` 1.5+
  - `helm` 3.12+
  - `argocd` CLI (optional)
  - `kind` (for local development)
- **Cloud Access**: Valid credentials for your chosen provider(s)

## 🚀 Quickstart

### Option 1: Local Development (kind)

Perfect for testing and development:

```bash
# Create local Kubernetes cluster
kind create cluster --name mantl

# Install ArgoCD and platform apps
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Deploy the platform
kubectl apply -k clusters/production

# Access ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

See [docs/quickstart-platform.md](docs/quickstart-platform.md) for detailed instructions.

### Option 2: AWS (EKS)

```bash
# Navigate to EKS blueprint
cd terraform/blueprints/aws/eks

# Configure
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your settings

# Deploy infrastructure
terraform init
terraform plan
terraform apply

# Configure kubectl
aws eks update-kubeconfig --name mantl-eks --region us-east-1

# Deploy platform
kubectl apply -k clusters/production/overlays/aws
```

### Option 3: Google Cloud (GKE)

```bash
# Navigate to GKE blueprint
cd terraform/blueprints/gcp/gke

# Configure
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars

# Deploy
terraform init && terraform apply

# Get credentials
gcloud container clusters get-credentials mantl-gke --region us-central1

# Deploy platform
kubectl apply -k clusters/production/overlays/gke-wi
```

### Option 4: Azure (AKS)

```bash
cd terraform/blueprints/azure/aks
cp terraform.tfvars.example terraform.tfvars
terraform init && terraform apply

az aks get-credentials --resource-group mantl --name mantl-aks

kubectl apply -k clusters/production/overlays/aks-wi
```

## 📁 Project Structure

```
mantl/
├── terraform/
│   ├── blueprints/          # Cloud provider infrastructure
│   │   ├── aws/             # EKS, VPC, IAM
│   │   ├── gcp/             # GKE, VPC, Service Accounts
│   │   ├── azure/           # AKS, VNet, Managed Identity
│   │   ├── digitalocean/    # DOKS clusters
│   │   └── linode/          # LKE clusters
│   └── modules/             # Reusable Terraform modules
│       ├── global-dns/      # Multi-region DNS failover
│       └── kubernetes-platform/ # Shared K8s resources
├── clusters/
│   ├── production/          # Production ArgoCD apps
│   │   ├── apps/           # Platform component definitions
│   │   └── overlays/       # Cloud-specific configurations
│   ├── staging/            # Staging environment
│   └── dev/                # Development environment
├── platform/
│   ├── base/               # Base platform configurations
│   │   ├── argocd/         # ArgoCD setup
│   │   ├── cert-manager/   # TLS certificate management
│   │   └── external-dns/   # DNS automation
│   ├── ingress/            # Gateway API configurations
│   └── secrets/            # External Secrets templates
├── addons/
│   ├── spinnaker/          # Spinnaker deployment configs
│   └── examples/           # Sample applications
├── applications/
│   ├── examples/           # Example workloads
│   └── templates/          # Application templates
├── policies/
│   ├── kyverno/            # Kyverno admission policies
│   └── opa/                # OPA Rego policies
└── docs/                   # Documentation
```

## 🔧 Core Platform Components

### GitOps & Deployment
- **ArgoCD**: Declarative continuous delivery
- **Spinnaker**: Advanced deployment strategies and multi-cloud orchestration
- **App-of-Apps**: Centralized platform management

### Networking & Ingress
- **Gateway API**: Modern Kubernetes ingress with Cilium or Envoy Gateway
- **ExternalDNS**: Automatic DNS record management (Route53, Cloud DNS, Azure DNS, Cloudflare, Linode)
- **cert-manager**: Automated TLS certificate provisioning (Let's Encrypt, cloud CA)

### Security
- **Kyverno**: Kubernetes-native policy engine
- **OPA Gatekeeper**: Policy-based admission control
- **External Secrets**: Secure secrets injection from cloud providers
- **Cosign**: Container image signing and verification
- **Network Policies**: Pod-to-pod security with Cilium

### Multi-Cloud Support

| Provider | Infrastructure | Workload Identity | DNS | Cert Manager |
|----------|---------------|-------------------|-----|--------------|
| AWS | ✅ EKS | ✅ IRSA | ✅ Route53 | ✅ ACM |
| GCP | ✅ GKE | ✅ Workload Identity | ✅ Cloud DNS | ✅ Google CA |
| Azure | ✅ AKS | ✅ Managed Identity | ✅ Azure DNS | ✅ Azure CA |
| DigitalOcean | ✅ DOKS | ⚠️ API Token | ✅ DO DNS | ✅ ACME |
| Linode | ✅ LKE | ⚠️ API Token | ✅ Linode DNS | ✅ ACME Webhook |
| Local | ✅ kind | N/A | ✅ Manual | ✅ Self-signed |

## 🔐 Security Features

### Supply Chain Security
```bash
# SBOM generation
make sbom

# Image signing with Cosign
cosign sign --key cosign.key your-image:tag

# Policy verification
kubectl apply -f policies/kyverno/verify-image-signatures.yaml
```

### Policy Enforcement
- **Require Labels**: Enforce mandatory labels on all resources
- **Image Verification**: Only allow signed images
- **Registry Restriction**: Limit to approved container registries
- **Resource Quotas**: Prevent resource exhaustion

See [docs/security.md](docs/security.md) for complete security documentation.

## 🌐 Multi-Cloud & High Availability

Mantl supports global, multi-region deployments with automatic failover:

- **Route53 Health Checks**: Automatic traffic routing based on endpoint health
- **GeoDNS**: Route users to nearest region
- **Cross-Region Replication**: Database and storage replication patterns
- **Disaster Recovery**: Automated backup and restore procedures

See [docs/multicloud.rst](docs/multicloud.rst) for architecture patterns.

## 📚 Documentation

### Getting Started
- [Platform Quickstart](docs/quickstart-platform.md) - Get up and running quickly
- [Overlays Guide](docs/overlays.md) - Environment and provider-specific configurations
- [Platform Apps](docs/platform-apps.rst) - Inventory of included applications

### Operations
- [Security Baseline](docs/security.md) - Security hardening and best practices
- [Secrets Management](docs/secrets-and-issuers.md) - External Secrets and cert-manager
- [Multi-Cloud](docs/multicloud.rst) - Multi-region deployment patterns

### CI/CD
- [Jenkins X Integration](docs/jenkins-x.rst) - Cloud-native CI/CD
- [Spinnaker Setup](docs/spinnaker-registries.md) - Multi-cloud deployment pipelines
- [Required CI Checks](docs/ci-required-checks.md) - Quality gates

### Advanced
- [OIDC & TLS](docs/oidc-tls-dns.rst) - Authentication and encryption
- [External DNS Identity](docs/external-dns-identity.md) - Workload Identity configuration
- [Backstage Integration](docs/backstage.md) - Developer portal

### Architecture Decision Records
See [docs/ara/](docs/ara/) for detailed architectural decisions:
- [ARA-0001: Kubernetes First](docs/ara/ARA-0001-kubernetes-first.md)
- [ARA-0002: GitOps and Delivery](docs/ara/ARA-0002-gitops-and-delivery.md)
- [ARA-0003: Policy as Code](docs/ara/ARA-0003-policy-as-code.md)
- [ARA-0004: Secrets Strategy](docs/ara/ARA-0004-secrets-strategy.md)
- [ARA-0008: Supply Chain Security](docs/ara/ARA-0008-supply-chain-security.md)

## 🛠️ Development

### Prerequisites
```bash
# Install development dependencies
pip install -r requirements.txt
pip install -r requirements-test.txt

# Install pre-commit hooks
pre-commit install
```

### Local Testing
```bash
# Terraform validation
make terraform-validate

# Policy testing
make policy-test

# Run all checks
make test
```

### Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup
- Code style guidelines
- Testing requirements
- PR process
- Release workflow

## 🔄 Migration from Legacy Mantl

If you're upgrading from the original Mesos/Marathon-based Mantl:

1. **Infrastructure**: Migrate from EC2/GCE instances to managed Kubernetes (EKS/GKE/AKS)
2. **Orchestration**: Move Marathon apps to Kubernetes Deployments
3. **Service Discovery**: Replace Consul with Kubernetes DNS and Gateway API
4. **Secrets**: Migrate Vault secrets to External Secrets Operator
5. **Monitoring**: Replace collectd with Prometheus/Grafana

Legacy documentation is preserved in `docs/` for reference.

## 📊 Project Status

- ✅ **Active Development**: Kubernetes platform features
- ✅ **Production Ready**: Multi-cloud infrastructure blueprints
- ✅ **CI/CD**: GitHub Actions workflows
- 🚧 **In Progress**: Enhanced observability stack
- 📋 **Planned**: Service mesh integration, cost optimization tools

## 🤝 Community & Support

- **Issues**: [GitHub Issues](https://github.com/thomasvincent/mantl/issues)
- **Discussions**: [GitHub Discussions](https://github.com/thomasvincent/mantl/discussions)
- **Security**: See [SECURITY.md](SECURITY.md) for reporting vulnerabilities

## 📜 License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

Mantl was originally created by Cisco Cloud and open-sourced in 2015. This modern iteration builds upon that foundation while embracing cloud-native technologies and GitOps practices.

Special thanks to:
- The original Mantl contributors
- The Kubernetes, ArgoCD, and CNCF communities
- All contributors to this modernization effort

---

**Built with ❤️ for Platform Engineers**
