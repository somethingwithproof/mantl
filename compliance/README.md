# Mantl Compliance-as-Code

Automated compliance management for Kubernetes platforms. Define compliance requirements as code, enforce them automatically, and generate audit-ready evidence.

## Supported Frameworks

| Framework | Status | Controls |
|-----------|--------|----------|
| SOC2 Type II | ✅ Production | 64 controls |
| HIPAA | 🚧 Beta | 45 controls |
| PCI-DSS v4.0 | 🚧 Beta | 78 controls |
| CIS Kubernetes | ✅ Production | 124 controls |
| NIST 800-53 | 📋 Planned | - |
| FedRAMP | 📋 Planned | - |
| ISO 27001 | 📋 Planned | - |

## Quick Start

```bash
# Install the compliance operator
kubectl apply -k compliance/operator/

# Enable SOC2 compliance
kubectl apply -f compliance/frameworks/soc2/profile-standard.yaml

# Check compliance status
mantl compliance status

# Run an audit
mantl compliance audit --framework soc2 --output report.pdf

# View violations
mantl compliance findings --severity critical,high
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Compliance Dashboard                         │
│                   (Grafana / Custom UI)                          │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                    Compliance Operator                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │  Framework  │  │   Audit     │  │     Remediation         │  │
│  │  Controller │  │  Controller │  │     Controller          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
         │                  │                    │
         ▼                  ▼                    ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────────────────┐
│   Kyverno   │    │   Evidence  │    │    Existing Resources   │
│  (Enforce)  │    │   Store     │    │  (Auto-Remediation)     │
└─────────────┘    └─────────────┘    └─────────────────────────┘
         │                  ▲
         │                  │
         ▼                  │
┌─────────────────────────────────────────────────────────────────┐
│                    Evidence Collectors                           │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌───────┐  │
│  │  Falco  │  │  Loki   │  │ Prometheus│ │  Harbor │  │ K8s   │  │
│  │ Runtime │  │  Logs   │  │  Metrics  │ │  Scans  │  │ Audit │  │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘  └───────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Key Concepts

### ComplianceFramework
Defines a compliance standard (SOC2, HIPAA, etc.) with all its controls.

### ComplianceProfile
A selection of controls relevant to your organization. Not every control applies to every company.

### ControlMapping
Links abstract compliance requirements to concrete Kubernetes policies and evidence collectors.

### Evidence
Proof of compliance: configuration snapshots, logs, metrics, scan results, and attestations.

### Finding
A compliance violation with severity, remediation guidance, and tracking.

## Directory Structure

```
compliance/
├── api/                    # CRD definitions
│   └── v1alpha1/
├── operator/               # Kubernetes operator
│   ├── controllers/
│   └── webhooks/
├── frameworks/             # Compliance framework definitions
│   ├── soc2/
│   ├── hipaa/
│   ├── pci-dss/
│   └── cis-kubernetes/
├── evidence/               # Evidence collection
│   ├── collectors/
│   └── store/
├── reports/                # Report templates
│   ├── templates/
│   └── generators/
└── cli/                    # mantl-compliance CLI
```

## How It Works

1. **Enable a Framework**: Apply a `ComplianceProfile` that selects which controls to enforce
2. **Policy Generation**: The operator generates Kyverno policies from control mappings
3. **Continuous Monitoring**: Policies enforce compliance; violations create Findings
4. **Evidence Collection**: Scheduled jobs collect proof of compliance
5. **Audit Reports**: Generate auditor-ready reports on demand
6. **Remediation**: Auto-fix or track manual remediation of findings

## License

Apache 2.0
