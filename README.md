# Mantl

[![CI](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml/badge.svg)](https://github.com/thomasvincent/mantl/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/thomasvincent/mantl/branch/main/graph/badge.svg)](https://codecov.io/gh/thomasvincent/mantl)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Security Rating](https://img.shields.io/badge/security-A+-brightgreen)](docs/security.md)
[![Code Quality](https://img.shields.io/badge/quality-A+-brightgreen)](#code-quality)

Mantl is a modern, cloud-native platform for rapidly deploying globally distributed services with built-in compliance automation.

## What Makes Mantl Different

Mantl isn't just another Kubernetes platform. It's the first to provide **Compliance-as-Code** as a core feature:

- 🛡️ **Continuous Compliance** - SOC2, HIPAA, PCI-DSS, CIS benchmarks enforced automatically
- 📊 **Audit-Ready Evidence** - Continuous evidence collection, not point-in-time snapshots
- 🔄 **GitOps Native** - Everything is code, everything is versioned, everything is auditable
- ☁️ **Multi-Cloud** - AWS, GCP, Azure, DigitalOcean, Oracle, IBM, OpenStack, bare metal

## ⚡ Quick Start (3 Commands)

Get a complete production-ready platform running in **under 5 minutes**:

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
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    Tetragon (eBPF)                         │  │
│  │            Runtime Security Observability                  │  │
│  └───────────────────────────────────────────────────────────┘  │
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
│  │  Cilium   │ │Crossplane │ │ Karpenter │ │  Kubernetes   │  │
│  │   CNI     │ │   IaC     │ │ Autoscale │ │  EKS/GKE/AKS  │  │
│  └───────────┘ └───────────┘ └───────────┘ └───────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Compliance Module

Mantl's killer feature is the Compliance-as-Code module:

```bash
# View compliance status
$ mantl compliance status

┌────────────────────────────────────────────┐
│           Compliance Status             │
├────────────────────────────────────────────┤
│ SOC2 Type II           Score: 94%      │
│ CIS Kubernetes v1.8    Score: 97%      │
├────────────────────────────────────────────┤
│ Critical Findings: 2                   │
│ Evidence Items: 1,247                  │
│ Last Audit: 2 hours ago                │
└────────────────────────────────────────────┘

# Generate audit report
$ mantl compliance audit --framework soc2 --output report.pdf
```

Supported frameworks:
- ✅ SOC2 Type II
- ✅ CIS Kubernetes Benchmark v1.8
- 🚧 HIPAA (beta)
- 🚧 PCI-DSS v4.0 (beta)
- 📋 NIST 800-53 (planned)
- 📋 FedRAMP (planned)

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

## Modern Kubernetes Infrastructure

Mantl includes production-ready configurations for modern Kubernetes infrastructure management:

### Crossplane - Multi-Cloud Infrastructure as Code

Crossplane enables declarative infrastructure provisioning across AWS, GCP, and Azure using Kubernetes-native APIs.

**Features:**
- Multi-cloud provider support (AWS, GCP, Azure)
- Composite Resource Definitions (XRDs) for abstracted infrastructure
- Pre-built compositions for networks, databases, and Kubernetes clusters
- GitOps-native infrastructure management

**Example - Provision a VPC:**
```yaml
apiVersion: infrastructure.mantl.io/v1alpha1
kind: Network
metadata:
  name: production-vpc
spec:
  provider: aws
  region: us-west-2
  cidrBlock: "10.0.0.0/16"
  enableNatGateway: true
  availabilityZones: 3
```

**Example - Provision a Database:**
```yaml
apiVersion: infrastructure.mantl.io/v1alpha1
kind: Database
metadata:
  name: production-db
spec:
  provider: aws
  region: us-west-2
  engine: postgres
  engineVersion: "15"
  instanceClass: db.r6g.large
  storageGB: 100
  multiAZ: true
  enableEncryption: true
```

### Karpenter - Just-in-Time Node Provisioning

Karpenter provides fast, efficient node autoscaling for Kubernetes clusters with intelligent instance selection.

**Features:**
- Sub-minute node provisioning
- Cost-optimized instance selection
- Support for Spot, On-Demand, and mixed capacity
- GPU and ARM64 (Graviton) workload support
- Automatic node consolidation

**Pre-configured NodePools:**
| NodePool | Use Case | Instance Types |
|----------|----------|----------------|
| default | General workloads | c6i, m6i, r6i (medium-2xlarge) |
| compute-optimized | CPU-intensive | c7i (xlarge-8xlarge) |
| memory-optimized | Memory-intensive | r7i (xlarge-8xlarge) |
| gpu | ML/AI workloads | g5, p4d |
| spot | Cost optimization | Mixed (50%+ savings) |

**Deploy Karpenter resources:**
```bash
kubectl apply -k infrastructure/karpenter/
```

### Tetragon - eBPF Runtime Security

Cilium Tetragon provides deep runtime security observability using eBPF, enabling detection and prevention of security threats at the kernel level.

**Security Policies Included:**
- **Privilege Escalation Detection** - Monitors setuid/setgid calls to root
- **Container Escape Detection** - Detects namespace manipulation and ptrace injection
- **Sensitive File Access** - Monitors access to /etc/shadow, /etc/kubernetes, secrets
- **Cryptominer Detection** - Blocks known mining software execution
- **Network Security** - Monitors suspicious outbound connections
- **File Integrity Monitoring** - Detects modifications to system binaries
- **Shell Spawn Detection** - Alerts on unexpected shell execution in containers

**Example - View security events:**
```bash
kubectl logs -n tetragon -l app.kubernetes.io/name=tetragon -f | \
  tetra getevents -o compact
```

**Deploy Tetragon:**
```bash
kubectl apply -k infrastructure/tetragon/
```

## Documentation

### Getting Started
- [Quick Start Guide](docs/quickstart-platform.md)
- [Architecture Overview](docs/architecture-diagram.md)
- [Development Guide](DEVELOPMENT.md)

### Operational Runbooks
- [Disaster Recovery Procedures](docs/runbooks/disaster-recovery.md) - Complete DR procedures for all failure scenarios
- [Scaling Operations](docs/runbooks/scaling-operations.md) - HPA, VPA, cluster autoscaling procedures
- [Upgrade Procedures](docs/runbooks/upgrade-procedures.md) - Kubernetes, platform, and application upgrades
- [Troubleshooting Guide](docs/runbooks/troubleshooting-guide.md) - Common issues and solutions
- [Incident Response](docs/runbooks/incident-response.md) - Incident classification and response playbooks

### Additional Documentation
- [Compliance Module](compliance/README.md)
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
| Infrastructure | Crossplane | Incubating |
| Autoscaling | Karpenter | Graduated |
| eBPF Security | Tetragon | - |

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

#### E-Commerce Microservices (`examples/ecommerce-microservices/`)
- Full-stack application with Products API (FastAPI/Python) and React frontend
- PostgreSQL database integration with CloudNativePG
- Prometheus metrics and OpenTelemetry tracing
- Complete CI/CD with GitHub Actions
- Horizontal Pod Autoscaling

#### ML Inference Service (`examples/ml-inference-service/`)
- Machine learning model serving with FastAPI
- Single and batch prediction endpoints
- Model versioning and A/B testing support
- Optional GPU acceleration
- Autoscaling based on prediction latency

#### Batch Processing Job (`examples/batch-processing-job/`)
- CronJob for scheduled batch data processing
- Parallel processing with async workers
- S3 checkpointing for resume capability
- Idempotent processing patterns
- Prometheus Pushgateway metrics

#### Event-Driven Architecture (`examples/event-driven-arch/`)
- NATS JetStream message broker (clustered)
- Publisher service with REST API
- Subscriber service with consumer groups
- CloudEvents standard format
- At-least-once delivery guarantees

#### Static Website with CDN (`examples/static-website/`)
- Production nginx configuration
- CloudFront/Cloudflare CDN integration
- Aggressive caching and compression
- Security headers (CSP, X-Frame-Options)
- Responsive design

Deploy examples:
```bash
# Individual examples
kubectl apply -f examples/ml-inference-service/k8s/
kubectl apply -f examples/batch-processing-job/k8s/
kubectl apply -f examples/event-driven-arch/k8s/
kubectl apply -f examples/static-website/k8s/

# E-commerce microservices (requires Flux)
flux reconcile kustomization ecommerce-microservices
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
├── infrastructure/      # Modern Kubernetes infrastructure
│   ├── crossplane/      # Multi-cloud IaC with XRDs
│   ├── karpenter/       # Node autoscaling
│   └── tetragon/        # eBPF security policies
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
