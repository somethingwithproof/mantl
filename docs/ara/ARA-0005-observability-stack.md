# ARA-0005 Observability stack
Status: Accepted
Date: 2025-12-27

## Context
Legacy ELK is heavy to operate at scale; we want traces/metrics/logs correlation for full observability.

## Decision
Adopt OpenTelemetry Collector as the ingestion/forwarding plane with the following backend components:
- **Prometheus** (metrics) - CNCF Graduated, kube-prometheus-stack
- **Grafana** (visualization) - Bundled with Prometheus stack
- **Jaeger** (distributed tracing) - CNCF Graduated, Elasticsearch backend
- **Loki** (logs) - TBD, planned for Phase 2
- **Tempo** (tracing alternative) - TBD, optional alternative to Jaeger

Use OTLP protocol everywhere for vendor neutrality.

## Implementation Details

### Deployment Architecture
All observability components are deployed via ArgoCD in the platform-apps.yaml manifest:

1. **Prometheus** (sync-wave 4)
   - Namespace: `monitoring`
   - Includes: Prometheus Operator, Alertmanager, Grafana
   - Storage: 50Gi persistent volumes with 15-day retention
   - High availability: 2 replicas with pod disruption budgets

2. **OpenTelemetry Collector** (sync-wave 5)
   - Namespace: `observability`
   - Receives: OTLP, Jaeger, Zipkin protocols
   - Processors: K8s attributes, tail sampling, span metrics
   - Exporters: Prometheus (metrics), Jaeger (traces), Loki (logs when deployed)
   - High availability: 2-5 replicas with HPA

3. **Jaeger** (sync-wave 5)
   - Namespace: `observability`
   - Components: Collector, Query UI, Elasticsearch
   - Storage: Elasticsearch with 7-day retention
   - High availability: 2 replicas for collector and query
   - UI: Available at http://jaeger-query:16686

### Data Flow
```
Application (OTLP SDK)
  ↓ (OTLP/gRPC)
OpenTelemetry Collector
  ├─→ Prometheus (metrics)
  ├─→ Jaeger Collector (traces)
  └─→ Loki (logs, when deployed)
         ↓
    Grafana (unified visualization)
```

### Grafana Data Sources
Grafana is pre-configured with the following data sources:
- Prometheus (default) - metrics
- Jaeger - distributed traces
- Loki - logs (when deployed)
- Tempo - optional tracing backend

## Consequences

### Positive
- Unified telemetry model; vendor-neutral via OTLP
- Full correlation: traces → metrics → logs in single UI
- Easier canary/auto-remediation inputs for progressive delivery
- Production-ready: HA, autoscaling, persistent storage
- CNCF Graduated projects (Prometheus, Jaeger, OTel)

### Negative
- Additional storage requirements for Elasticsearch (Jaeger backend)
- Operational complexity with multiple storage backends
- Loki deployment still pending

## Alternatives Considered
- **Self-managed ELK/OpenSearch**: Higher ops cost; logs-only focus; no native OTLP
- **Tempo instead of Jaeger**: Lower storage cost (object storage), but less mature query UI
- **Cloud-native services** (Datadog, New Relic): Vendor lock-in; higher cost at scale
