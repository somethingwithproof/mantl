# ARA-0009: Compliance-as-Code Architecture

## Status

Accepted

## Context

Enterprises adopting Mantl need to demonstrate compliance with various regulatory frameworks (SOC2, HIPAA, PCI-DSS, etc.). Traditional compliance approaches involve:

1. Manual evidence collection (screenshots, exports)
2. Point-in-time audits (annual, quarterly)
3. Spreadsheet-based control tracking
4. Disconnected tooling for policy enforcement vs. evidence collection

This creates significant operational burden, audit anxiety, and gaps between "compliant at audit time" vs. "continuously compliant."

## Decision

Implement a Compliance-as-Code module that:

1. **Defines compliance frameworks as Kubernetes Custom Resources**
   - ComplianceFramework: Defines controls (SOC2, HIPAA, etc.)
   - ComplianceProfile: Selects applicable controls for an organization
   - ControlMapping: Links abstract controls to concrete policies

2. **Automatically generates and enforces policies**
   - Operator generates Kyverno policies from control mappings
   - Policies enforce in real-time (not just audit)
   - Same policies produce evidence when triggered

3. **Continuously collects evidence**
   - Configuration snapshots (RBAC, NetworkPolicy, etc.)
   - Log-based evidence (Loki queries)
   - Metric-based evidence (Prometheus queries)
   - Scan results (vulnerabilities, signatures)
   - Human attestations for procedural controls

4. **Generates audit-ready reports**
   - Control-by-control status
   - Evidence attachments with integrity hashes
   - Finding remediation tracking
   - Executive summaries

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                 ComplianceFramework (CRD)                        │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ SOC2 Type II                                             │    │
│  │  ├── CC6: Logical Access                                 │    │
│  │  │    ├── CC6.1: Access Security → NetworkPolicy         │    │
│  │  │    └── CC6.8: Malware Prevention → Image Signing      │    │
│  │  └── CC7: System Operations                              │    │
│  │       └── CC7.1: Event Detection → Falco Rules           │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                 ComplianceProfile (CRD)                          │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ soc2-standard                                            │    │
│  │  ├── frameworkRef: soc2-type2                           │    │
│  │  ├── mode: enforce                                       │    │
│  │  ├── excludeNamespaces: [kube-system, monitoring]       │    │
│  │  └── evidenceStore: s3://compliance-evidence            │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                 Compliance Operator                              │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────────┐    │
│  │   Framework   │  │    Audit      │  │   Remediation     │    │
│  │  Controller   │  │  Controller   │  │   Controller      │    │
│  │               │  │               │  │                   │    │
│  │ • Validates   │  │ • Schedules   │  │ • Auto-fixes      │    │
│  │   frameworks  │  │   audits      │  │   findings        │    │
│  │ • Generates   │  │ • Aggregates  │  │ • Tracks manual   │    │
│  │   Kyverno     │  │   evidence    │  │   remediation     │    │
│  │   policies    │  │ • Generates   │  │                   │    │
│  │               │  │   reports     │  │                   │    │
│  └───────────────┘  └───────────────┘  └───────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────┐      ┌─────────────┐      ┌─────────────────────┐
│   Kyverno   │      │  Evidence   │      │   Kubernetes API    │
│             │      │   Store     │      │   (Remediation)     │
│ • Enforce   │      │             │      │                     │
│ • Audit     │      │ • S3/GCS    │      │ • Create policies   │
│ • Generate  │      │ • Encrypted │      │ • Update configs    │
│             │      │ • Immutable │      │                     │
└─────────────┘      └─────────────┘      └─────────────────────┘
```

## Control Mapping Strategy

The key innovation is the control mapping layer:

```yaml
# Abstract control
Control:
  id: CC6.1
  name: "Logical Access Security"
  description: "Entity implements logical access security..."

# Maps to concrete implementations
Mappings:
  - type: policy
    policyRef:
      template: require-network-policy
      mode: enforce
  
  - type: evidence
    evidenceCollector:
      type: config-snapshot
      schedule: "0 0 * * *"
      resources: [networkpolicies, clusterroles]
```

This creates a verifiable chain:
1. **Compliance Requirement** → What the auditor cares about
2. **Technical Control** → What we enforce in Kubernetes
3. **Evidence** → How we prove compliance

## Alternatives Considered

### 1. Static Policy Libraries
Pre-built Kyverno policies without the framework abstraction.

**Rejected because:**
- No control-to-policy mapping for auditors
- No evidence collection integration
- No compliance scoring/reporting

### 2. External GRC Platform Integration
Integrate with Vanta, Drata, or similar.

**Rejected because:**
- Vendor lock-in
- Limited Kubernetes-native evidence
- Additional cost
- Doesn't leverage existing observability stack

### 3. Manual Evidence Collection
Continue with traditional audit approaches.

**Rejected because:**
- Labor-intensive
- Point-in-time vs. continuous
- Error-prone
- Doesn't scale

## Consequences

### Positive
- Continuous compliance monitoring (not just audit-time)
- Automated evidence collection reduces toil
- Single source of truth for compliance state
- Reusable framework definitions across organizations
- Clear auditor-ready reports

### Negative
- Additional CRDs to manage
- Operator complexity
- Storage costs for evidence
- Some controls require human attestation

### Risks
- False sense of compliance if mappings are incomplete
- Evidence storage requires careful security
- Auditor acceptance of automated evidence

## Implementation Plan

1. **Phase 1**: Core CRDs and operator framework
2. **Phase 2**: SOC2 framework with key controls
3. **Phase 3**: Evidence collection infrastructure
4. **Phase 4**: Report generation
5. **Phase 5**: CLI and dashboard
6. **Phase 6**: Additional frameworks (HIPAA, PCI-DSS)

## References

- [AICPA Trust Services Criteria](https://www.aicpa.org/interestareas/frc/assuranceadvisoryservices/trustservices.html)
- [Kyverno Policy Library](https://kyverno.io/policies/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
