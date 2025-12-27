# Mantl 2026: CNCF Graduated Projects Analysis

**Analysis Date**: December 27, 2025
**Repository**: thomasvincent/mantl

---

## Final Stack Decision

After analysis, we're adopting a **focused, production-ready stack** rather than maximizing CNCF coverage.

### Core Components (Keep)

| Component | Version | CNCF Status | Purpose |
|-----------|---------|-------------|---------|
| Kubernetes | 1.31+ | Graduated | Orchestration |
| Cilium | 1.18.5 | Graduated | CNI + Service Mesh + L2 LB |
| ArgoCD | 2.13.x | Graduated | GitOps |
| cert-manager | 1.19.x | Graduated | TLS automation |
| Kyverno | 3.6.x | Graduated | Policy enforcement |
| External Secrets | 1.2.x | Incubating | Secret sync |
| External DNS | 1.19.x | Incubating | DNS automation |

### Observability & Security Additions

| Component | Version | CNCF Status | Purpose |
|-----------|---------|-------------|---------|
| Prometheus | 65.2.0 (stack) | Graduated | Metrics + Alerting |
| OpenTelemetry | 0.115.x | Graduated | Telemetry pipeline |
| Jaeger | 3.3.1 | Graduated | Distributed tracing |
| Loki | 6.x | — | Log aggregation |
| Tempo | 2.x | — | Tracing (alternative) |
| Grafana | 11.x | — | Visualization |
| Falco | 4.15.0 | Graduated | Runtime security |
| Harbor | 1.16.0 | Graduated | On-prem registry |

---

## What We're NOT Adding (and Why)

| Component | Reason |
|-----------|--------|
| OPA | Kyverno handles policy; redundant |
| Flux | ArgoCD handles GitOps; redundant |
| Istio/Linkerd | Cilium provides service mesh |
| Crossplane | Terraform/OpenTofu handles IaC |
| Knative | No serverless requirements |
| SPIRE | Cloud workload identity suffices |

---

## Directory Structure

```
platform/
├── base/
│   ├── argocd/           ✅ GitOps
│   ├── cert-manager/     ✅ Certificates
│   └── cilium/           ✅ CNI + L2 LB
├── dns/
│   └── external-dns/     ✅ DNS sync
├── observability/
│   ├── prometheus/       🆕 Metrics + Grafana
│   ├── otel-collector/   🆕 Telemetry pipeline
│   ├── jaeger/           🆕 Distributed tracing
│   ├── loki/             🆕 Log aggregation
│   └── tempo/            🆕 Tracing (alternative)
├── registry/
│   └── harbor/           🆕 On-prem registry
├── secrets/
│   └── external-secrets/ ✅ Secret sync
└── security/
    └── falco/            🆕 Runtime security

policies/
└── kyverno/              ✅ Policy enforcement
```

---

## CNCF Coverage

**Graduated Projects Used**: 10
- Kubernetes, Cilium, ArgoCD, cert-manager, Kyverno
- Prometheus, OpenTelemetry, Jaeger, Falco, Harbor

**Incubating Projects Used**: 2
- External Secrets, External DNS

**Non-CNCF Components**: 3
- Grafana (visualization), Loki (logs), Tempo (tracing alternative)

This is a focused stack optimized for production use, not badge collection.

---

## Deployment Order

```
1. Cilium           (sync-wave: 1) - CNI first
2. cert-manager     (sync-wave: 2) - TLS prereq
3. External Secrets (sync-wave: 3) - Secrets
4. Kyverno          (sync-wave: 3) - Policy
5. Prometheus       (sync-wave: 4) - Metrics + Grafana
6. Loki             (sync-wave: 4) - Log aggregation
7. Tempo            (sync-wave: 4) - Tracing backend
8. OpenTelemetry    (sync-wave: 5) - Telemetry pipeline
9. Jaeger           (sync-wave: 5) - Distributed tracing
10. Falco           (sync-wave: 5) - Security
11. Harbor          (sync-wave: 6) - Registry
12. External DNS    (sync-wave: 6) - DNS
```

---

## Configuration Required

### Harbor (On-Prem)
```yaml
# platform/registry/harbor/values.yaml
expose:
  ingress:
    hosts:
      core: harbor.YOUR-DOMAIN.com
externalURL: https://harbor.YOUR-DOMAIN.com
```

### Cilium L2 (On-Prem)
```yaml
# platform/base/cilium/l2-announcement.yaml
spec:
  blocks:
    - start: 192.168.X.100  # Your IP range
      stop: 192.168.X.200
```

### Falco Alerts
```yaml
# platform/security/falco/values.yaml
falcosidekick:
  config:
    slack:
      webhookurl: "https://hooks.slack.com/..."
```
