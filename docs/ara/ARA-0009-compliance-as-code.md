<!-- SPDX-License-Identifier: Apache-2.0 -->
# ARA-0009: Compliance-as-Code Platform

## Status

Proposed

## Context

Organizations deploying Kubernetes face increasing regulatory requirements (SOC2, HIPAA, PCI-DSS, FedRAMP, GDPR). Current compliance approaches suffer from:

1. **Manual Evidence Collection**: Engineers spend 20-40% of audit prep time gathering screenshots and logs
2. **Point-in-Time Audits**: Compliance is verified annually, not continuously
3. **Policy-Control Disconnect**: Security policies exist in isolation from compliance frameworks
4. **Framework Silos**: Organizations implement duplicate controls for overlapping frameworks
5. **Remediation Lag**: Violations detected days/weeks after occurrence

Existing tools address fragments:
- **Kyverno/OPA**: Policy enforcement without compliance mapping
- **Falco**: Runtime detection without evidence correlation
- **Cloud Security Posture Management (CSPM)**: Cloud-specific, not Kubernetes-native
- **GRC Platforms**: Manual, expensive, not GitOps-compatible

## Decision

Implement a **Compliance-as-Code** platform as a core Mantl differentiator with:

1. **Unified Control Catalog**: Single source of truth mapping controls to policies to evidence
2. **Continuous Compliance**: Real-time posture assessment, not point-in-time
3. **Automated Evidence Collection**: GitOps-native evidence gathering and immutable storage
4. **Multi-Framework Mapping**: Implement once, satisfy multiple frameworks
5. **Remediation Automation**: Self-healing policies where safe

## Architecture

### Domain Model (DDD)

The compliance bounded context consists of:

| Entity | Description | Identifier |
|--------|-------------|------------|
| **Framework** | Compliance standard (SOC2, HIPAA) | `framework-id` |
| **Section** | Logical grouping within framework | `section-id` |
| **Control** | Specific requirement | `control-id` |
| **Policy** | Technical implementation (Kyverno) | `policy-name` |
| **FalcoRule** | Runtime detection rule | `rule-name` |
| **Evidence** | Proof of compliance (immutable) | `evidence-id` |
| **Finding** | Violation or gap | `finding-id` |
| **Audit** | Point-in-time assessment | `audit-id` |
| **Report** | Auditor-ready documentation | `report-id` |

### Control-to-Policy Mapping (Key Innovation)

The core value proposition is **bidirectional traceability**:

```yaml
# Single control can map to multiple policies
Control: SOC2-CC6.1 (Logical Access Security)
  - Policy: deny-privileged-containers
  - Policy: deny-privilege-escalation
  - Policy: require-network-policies
  - FalcoRule: detect-privilege-escalation
  - FalcoRule: detect-rbac-changes

# Single policy can satisfy multiple controls
Policy: deny-privileged-containers
  - Satisfies: SOC2-CC6.1
  - Satisfies: HIPAA-164.312(a)(1)
  - Satisfies: PCI-DSS-2.2.4
  - Satisfies: CIS-K8S-5.2.1
```

### Evidence Collection Strategy

| Evidence Type | Source | Collection Method | Frequency |
|---------------|--------|-------------------|-----------|
| Policy Compliance | Kyverno PolicyReport | Controller watch | Real-time |
| Runtime Events | Falco Alerts | Falcosidekick to Loki | Real-time |
| Audit Logs | K8s API Server | OTel Collector | Real-time |
| Configuration Snapshots | etcd | Scheduled export | Daily |
| Network Policies | Cilium | Hubble export | Hourly |
| RBAC State | K8s RBAC | Controller snapshot | Hourly |
| Container Images | Harbor | Trivy scan results | On push |
| Secrets Access | Vault | Audit log | Real-time |

## Implementation Phases

### Phase 1: Foundation (Weeks 1-4)
1. Custom Resource Definitions (CRDs)
2. Control catalog for SOC2 Type II
3. Kyverno policy pack with control mappings
4. Basic evidence collection

### Phase 2: Automation (Weeks 5-8)
1. Evidence Controller with immutable storage
2. Falco rules with control mappings
3. Finding aggregation and deduplication
4. CLI tooling (`mantl compliance`)

### Phase 3: Reporting (Weeks 9-12)
1. Report Controller with PDF generation
2. Auditor-ready templates
3. Gap analysis automation
4. Dashboard (Grafana)

### Phase 4: Multi-Framework (Weeks 13-16)
1. HIPAA framework and policies
2. PCI-DSS framework and policies
3. CIS Kubernetes Benchmark
4. Cross-framework control mapping

## Consequences

### Positive
- **Continuous Compliance**: Real-time posture vs annual audits
- **Reduced Audit Prep**: 80% reduction in manual evidence gathering
- **GitOps Native**: Policies and controls versioned alongside infrastructure
- **Multi-Framework Efficiency**: Implement once, satisfy many
- **Clear Differentiation**: No comparable OSS solution exists

### Negative
- **Initial Complexity**: Significant upfront investment
- **Framework Expertise**: Requires compliance domain knowledge
- **Maintenance Burden**: Framework updates require policy updates

### Risks
- **Framework Interpretation**: Controls are subjective; policies may not satisfy auditors
- **False Sense of Security**: Automated compliance does not equal actual security

### Mitigations
- Engage compliance consultants for control interpretation
- Clear documentation that tool assists but does not replace auditors
- Regular validation against real audit findings

## References
- SOC2 Trust Services Criteria (AICPA)
- HIPAA Security Rule (45 CFR Part 164)
- PCI-DSS v4.0
- CIS Kubernetes Benchmark v1.8
- NIST Cybersecurity Framework
