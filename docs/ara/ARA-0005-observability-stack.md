<!-- SPDX-License-Identifier: Apache-2.0 -->
# ARA-0005 Observability stack
Status: Accepted
Date: 2025-12-27

## Context
Legacy ELK is heavy to operate at scale; we want traces/metrics/logs correlation for full observability.

## Decision
Adopt OpenTelemetry Collector as the ingestion/forwarding plane with the following backend components:
- **Prometheus** (metrics) - CNCF Graduated, kube-prometheus-stack
- **Grafana** (visualization) - Bundled with Prometheus stack
- **Tempo** (distributed tracing) - Grafana Labs, object storage backend
- **Loki** (logs) - Grafana Labs, object storage backend

Use OTLP protocol everywhere for vendor neutrality and leverage Grafana's integrated stack (Prometheus, Loki, Tempo) for seamless correlation.

## Implementation Details

### Deployment Architecture
All observability components are deployed via ArgoCD in the platform-apps.yaml manifest:

1. **Prometheus** (sync-wave 4)
   - Namespace: `monitoring`
   - Includes: Prometheus Operator, Alertmanager, Grafana
   - Storage: 50Gi persistent volumes with 15-day retention
   - High availability: 2 replicas with pod disruption budgets

2. **Loki** (sync-wave 4)
   - Namespace: `observability`
   - Components: Distributor, Ingester, Querier, Compactor
   - Storage: Object storage (S3, GCS, Azure Blob, MinIO)
   - Retention: Configurable, typically 30 days
   - Query: Via Grafana Explore with LogQL

3. **Tempo** (sync-wave 4)
   - Namespace: `observability`
   - Components: Distributor, Ingester, Querier, Compactor
   - Storage: Object storage (S3, GCS, Azure Blob, MinIO)
   - Retention: Configurable, typically 7-14 days
   - Query: Via Grafana Explore with TraceQL

4. **OpenTelemetry Collector** (sync-wave 5)
   - Namespace: `observability`
   - Receives: OTLP, Jaeger, Zipkin protocols
   - Processors: K8s attributes, tail sampling, span metrics
   - Exporters: Prometheus (metrics), Tempo (traces), Loki (logs)
   - High availability: 2-5 replicas with HPA

### Data Flow
```
Application (OTLP SDK)
  ↓ (OTLP/gRPC port 4317)
OpenTelemetry Collector
  ├─→ Prometheus (metrics)
  ├─→ Tempo (traces)
  └─→ Loki (logs)
         ↓
    Grafana (unified visualization with native correlation)
```

### Grafana Data Sources
Grafana is pre-configured with the following data sources:
- Prometheus (default) - metrics with PromQL
- Tempo - distributed traces with TraceQL
- Loki - logs with LogQL
- All three support trace-to-metrics-to-logs correlation

## Consequences

### Positive
- Unified telemetry model; vendor-neutral via OTLP
- Full correlation: traces → metrics → logs in single UI with native Grafana integration
- Easier canary/auto-remediation inputs for progressive delivery
- Production-ready: HA, autoscaling, object storage backend
- CNCF Graduated projects (Prometheus, OTel)
- Consistent storage layer: object storage for both traces and logs
- TraceQL and LogQL provide powerful query languages
- Lower operational overhead: no Elasticsearch cluster to manage

### Negative
- Tempo and Loki are not CNCF projects (Grafana Labs)
- Requires object storage backend (S3, GCS, MinIO, etc.)
- Less mature query UI compared to Jaeger for traces

## Alternatives Considered
- **Self-managed ELK/OpenSearch**: Higher ops cost; logs-only focus; no native OTLP
- **Jaeger for tracing**: More mature UI, but requires Elasticsearch cluster; higher storage costs
- **Cloud-native services** (Datadog, New Relic): Vendor lock-in; higher cost at scale
- **Separate storage backends**: Using different backends for logs/traces increases operational complexity
