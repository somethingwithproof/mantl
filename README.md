# Mantl

[![CI](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml/badge.svg)](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Mantl is a modern, cloud-native platform for rapidly deploying globally distributed services with built-in compliance automation.

## What Makes Mantl Different

Mantl isn't just another Kubernetes platform. It's the first to provide **Compliance-as-Code** as a core feature:

- 🛡️ **Continuous Compliance** - SOC2, HIPAA, PCI-DSS, CIS benchmarks enforced automatically
- 📊 **Audit-Ready Evidence** - Continuous evidence collection, not point-in-time snapshots
- 🔄 **GitOps Native** - Everything is code, everything is versioned, everything is auditable
- ☁️ **Multi-Cloud** - AWS, GCP, Azure, DigitalOcean, Oracle, IBM, OpenStack, bare metal

## ⚡ Quick Start (3 Commands)

Spin up a local development platform on a kind cluster in a few minutes:

```bash
git clone https://github.com/thomasvincent/mantl.git
cd mantl
make install-dev
```

That's it! This command:
- ✅ Creates a local kind cluster
- ✅ Installs ArgoCD and all platform components
- ✅ Deploys example applications
- ✅ Sets up observability stack

### Alternative Installation Methods

**Interactive Wizard** (for configuration):
```bash
make wizard
```

**CLI Tool** (for advanced users):
```bash
./scripts/mantl init
```

**Production Deployment** (cloud providers):
```bash
# 1. Provision infrastructure
cd terraform/blueprints/aws-eks  # or gcp-gke, azure-aks, etc.
tofu init && tofu apply

# 2. Bootstrap platform
./scripts/bootstrap-platform.sh --environment production --cloud aws

# 3. Enable compliance
kubectl apply -f compliance/frameworks/soc2/profile-standard.yaml
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Applications                               │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                      Platform Services                            │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────────┐  │
│  │  ArgoCD   │ │  Traefik  │ │  Harbor   │ │  Compliance   │  │
│  │  GitOps   │ │  Ingress  │ │ Registry  │ │    Module     │  │
│  └───────────┘ └───────────┘ └───────────┘ └───────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                       Observability                               │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────────┐  │
│  │Prometheus │ │   Loki    │ │   Tempo   │ │ OpenTelemetry │  │
│  │ Metrics   │ │   Logs    │ │  Traces   │ │   Collector   │  │
│  └───────────┘ └───────────┘ └───────────┘ └───────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                         Security                                  │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────────┐  │
│  │  Kyverno  │ │   Falco   │ │cert-manager││ Ext. Secrets  │  │
│  │  Policy   │ │  Runtime  │ │    TLS    │ │   Secrets     │  │
│  └───────────┘ └───────────┘ └───────────┘ └───────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                       Infrastructure                              │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────────┐  │
│  │  Cilium   │ │ OpenTofu  │ │External  │ │  Kubernetes   │  │
│  │   CNI     │ │   IaC     │ │   DNS    │ │  EKS/GKE/AKS  │  │
│  └───────────┘ └───────────┘ └───────────┘ └───────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Compliance Module

Mantl's differentiator is its Compliance-as-Code module. A `ComplianceProfile`
selects controls from a framework; controls map to admission policies; policy
violations surface as `Finding` resources; and audits capture point-in-time
evidence to object storage.

```bash
# Apply a compliance profile to the cluster
kubectl apply -f compliance/frameworks/soc2/profile-standard.yaml

# Inspect findings produced by policy violations
kubectl get findings -A
```

A `mantl compliance status` CLI summary is planned (see the roadmap); today,
findings and audits are inspected through `kubectl`.

Supported frameworks:
- ✅ SOC2 Type II
- ✅ CIS Kubernetes Benchmark v1.8
- 🚧 HIPAA (beta)
- 🚧 PCI-DSS v4.0 (beta)
- 📋 NIST 800-53 (planned)
- 📋 FedRAMP (planned)

### Maturity

Cloud blueprints:

| Cloud | Status |
|-------|--------|
| AWS (EKS) | GA |
| GCP (GKE) | beta (hardening in progress) |
| Azure (AKS) | beta (hardening in progress) |
| Linode, Oracle, IBM, DigitalOcean | experimental |

Compliance frameworks:

| Framework | Status |
|-----------|--------|
| SOC2 Type II | policies and evidence wiring in progress |
| CIS Kubernetes | policies in progress |
| HIPAA, PCI-DSS | declarations only, not yet enforced |

Maturity reflects what the code enforces today, not aspirations.

## Supported Clouds

| Provider | Service | Status |
|----------|---------|--------|
| AWS | EKS | ✅ Production |
| GCP | GKE | ✅ Production |
| Azure | AKS | ✅ Production |
| DigitalOcean | DOKS | ✅ Production |
| Oracle | OKE | ✅ Production |
| IBM | IKS | ✅ Production |
| Linode | LKE | ✅ Production |
| OpenStack | Magnum | 🚧 Beta |
| Bare Metal | k3s | 🚧 Beta |

## Modernization and Migration (Terraform 1.x)

The Terraform modules under `terraform/` are being modernized with security-focused defaults and feature flags to avoid breaking legacy behavior. Highlights:

- AWS (terraform/aws)
  - EBS encryption enabled by default; optional KMS key via `kms_key_id`.
  - Public ingress controlled via `allowed_cidrs` (defaults to ["0.0.0.0/0"], override to restrict).
  - Optional Auto Scaling Groups for workers (and other roles) via flags.
  - Optional VPC Flow Logs (enabled by default).
- GCE (terraform/gce)
  - Optional modern instance schema via `use_modern_gce_schema` (shielded VM, block project SSH keys, no public IP by default). Supports `modern_subnetwork_self_link` to attach to a modern subnetwork.
  - Optional modern custom-mode VPC/subnet via `use_modern_gce_network`; external firewall uses `allowed_cidrs`.
  - When enabling modern instances for a role, set the legacy counts to 0 in `gce.tf` to avoid duplicates (e.g., `control_count = 0`, `worker_count = 0`, `kubeworker_count = 0`).
- vSphere (terraform/vsphere)
  - Optional modern schema via `use_modern_vsphere_schema`, cloning from a template with optional `customization_spec_name`.
  - When enabling modern instances for a role, set the legacy counts to 0 to avoid duplicates.

See module READMEs for details and examples.

## Documentation

- [Quick Start Guide](docs/quickstart-platform.md)
- [Architecture Overview](docs/architecture-diagram.md)
- [Compliance Module](compliance/README.md)
- [Development Guide](DEVELOPMENT.md)
- [Architecture Decision Records](docs/ara/README.md)

## Components

All components are CNCF projects (Graduated or Incubating):

| Category | Component | CNCF Status |
|----------|-----------|-------------|
| GitOps | ArgoCD | Graduated |
| CNI | Cilium | Graduated |
| Metrics | Prometheus | Graduated |
| Logs | Loki | - |
| Traces | Tempo | - |
| Telemetry | OpenTelemetry | Incubating |
| Policy | Kyverno | Graduated |
| Runtime | Falco | Graduated |
| Registry | Harbor | Graduated |
| Secrets | External Secrets | Graduated |
| TLS | cert-manager | Graduated |
| DNS | External DNS | Incubating |

## Developer Experience

Mantl provides multiple ways to get started, from zero to production in minutes:

| Tool | Use Case | Command |
|------|----------|---------|
| **Makefile** | Quick commands | `make install-dev`, `make dashboards` |
| **Setup Wizard** | Interactive configuration | `make wizard` |
| **CLI Tool** | Advanced operations | `./scripts/mantl <command>` |
| **Bootstrap Script** | Automated deployment | `./scripts/bootstrap-platform.sh` |

### Example Applications

Production-ready examples demonstrating best practices:

- **api-service** - Backend API with OpenTelemetry tracing and Prometheus metrics
- **frontend** - Nginx frontend with Gateway API ingress
- **database-app** - Stateful application with PostgreSQL StatefulSet
- **hello** - Simple hello-world example

Deploy examples:
```bash
make deploy-examples
# or
kubectl apply -k applications/examples/<app-name>/base
```

## Environments

Mantl supports multiple environments with progressive complexity:

| Environment | Purpose | Components |
|-------------|---------|------------|
| **dev** | Local development (kind) | Core services only, minimal resources |
| **staging** | Pre-production testing | Production-like, full stack |
| **production** | Production workloads | Full stack, HA, multi-region ready |

Switch environments:
```bash
kubectl apply -f clusters/dev/app-of-apps.yaml
kubectl apply -f clusters/staging/app-of-apps.yaml
kubectl apply -f clusters/production/platform-apps.yaml
```

## Repository Structure

```
mantl/
├── compliance/          # Compliance-as-Code module (unique!)
│   ├── api/             # CRD definitions
│   ├── frameworks/      # SOC2, HIPAA, CIS, etc.
│   ├── operator/        # Kubernetes operator
│   └── evidence/        # Evidence collection
├── platform/            # Platform components
│   ├── base/            # Cilium, cert-manager, etc.
│   ├── observability/   # Prometheus, Loki, Tempo
│   ├── security/        # Falco, Harbor
│   └── secrets/         # External Secrets
├── policies/            # Kyverno policies
├── terraform/           # OpenTofu blueprints
│   └── blueprints/      # AWS, GCP, Azure, etc.
├── clusters/            # Cluster configurations (dev, staging, production)
├── applications/        # Example applications
├── scripts/             # CLI, bootstrap, wizard, validation
└── docs/                # Documentation
```

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.

## History

Mantl was originally created by Cisco in 2016 as a Mesos/Marathon-based PaaS. This 2026 reboot reimagines it as a Kubernetes-native platform with a focus on compliance automation and multi-cloud deployment.
