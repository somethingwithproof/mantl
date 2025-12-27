# ARA-0005 Observability stack
Status: Accepted
Date: 2025-12-27

Context
Legacy ELK is heavy to operate at scale; we want traces/metrics/logs correlation.

Decision
Adopt OpenTelemetry Collector as the ingestion/forwarding plane with Prometheus, Grafana, Loki, and Tempo. Use OTLP everywhere.

Consequences
- Unified telemetry model; vendor-neutral
- Easier canary/auto-remediation inputs later

Alternatives
- Self-managed ELK/OpenSearch: higher ops cost; logs-only focus
