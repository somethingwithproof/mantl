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

## Quick Start

```bash
# 1. Provision infrastructure (choose your cloud)
cd terraform/blueprints/aws-eks
tofu init && tofu apply

# 2. Bootstrap the platform
./scripts/bootstrap-cluster.sh

# 3. Enable compliance
kubectl apply -f compliance/frameworks/soc2/profile-standard.yaml

# 4. Check status
mantl compliance status
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
├── clusters/            # Cluster configurations
├── applications/        # Example applications
└── docs/                # Documentation
```

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.

## History

Mantl was originally created by Cisco in 2016 as a Mesos/Marathon-based PaaS. This 2026 reboot reimagines it as a Kubernetes-native platform with a focus on compliance automation and multi-cloud deployment.
