Platform Apps via ArgoCD
========================

Overview
--------
The production cluster is bootstrapped by ArgoCD Applications defined in ``clusters/production/platform-apps.yaml``. This app-of-apps manifest deploys all core platform components in sync-waves to ensure proper dependency ordering.

Platform Components
-------------------

**Sync Wave 1: Networking**

- **cilium**: Installs Cilium CNI with Hubble observability and Gateway API support
  - Path: ``platform/base/cilium``
  - Namespace: ``kube-system``

**Sync Wave 2: Certificate Management**

- **cert-manager**: Installs cert-manager for TLS certificate automation
  - Path: ``platform/base/cert-manager``
  - Namespace: ``cert-manager``

**Sync Wave 3: Security & Secrets**

- **external-secrets**: Installs External Secrets Operator for cloud secrets integration
  - Path: ``platform/secrets/external-secrets``
  - Namespace: ``external-secrets``

- **kyverno**: Installs Kyverno for policy enforcement and admission control
  - Path: ``policies/kyverno``
  - Namespace: ``kyverno``

**Sync Wave 4: Observability - Metrics**

- **prometheus**: Installs kube-prometheus-stack (Prometheus, Alertmanager, Grafana)
  - Path: ``platform/observability/prometheus``
  - Namespace: ``monitoring``
  - Components: Prometheus Operator, 2 Prometheus replicas, Alertmanager, Grafana
  - Storage: 50Gi persistent volumes with 15-day retention

**Sync Wave 5: Observability - Tracing & Security**

- **otel-collector**: Installs OpenTelemetry Collector for unified telemetry ingestion
  - Path: ``platform/observability/otel-collector``
  - Namespace: ``observability``
  - Receives: OTLP, Jaeger, Zipkin protocols
  - Exports to: Prometheus (metrics), Jaeger (traces)

- **jaeger**: Installs Jaeger distributed tracing platform
  - Path: ``platform/observability/jaeger``
  - Namespace: ``observability``
  - Components: Collector, Query UI, Elasticsearch backend
  - Storage: 7-day trace retention in Elasticsearch
  - UI: ``http://jaeger-query:16686``

- **falco**: Installs Falco runtime security monitoring
  - Path: ``platform/security/falco``
  - Namespace: ``falco-system``

**Sync Wave 6: Registry & DNS**

- **harbor**: Installs Harbor container registry (on-prem deployments)
  - Path: ``platform/registry/harbor``
  - Namespace: ``harbor``
  - Note: Disable if using cloud-native registries (ECR, GCR, ACR)

- **external-dns**: Installs ExternalDNS for automatic DNS record management
  - Path: ``platform/dns/external-dns``
  - Namespace: ``external-dns``

Additional Apps
---------------
Individual ArgoCD applications for specific configurations are available in ``clusters/production/apps/``:

- **cert-manager-issuer-dns-***: DNS-01 ACME issuers for various providers
  - cloudflare, route53, gcp, azure, digitalocean, linode
  - Enable one DNS-01 issuer per cluster

- **external-dns-***: Provider-specific ExternalDNS configurations
  - digitalocean, linode, azure, gcp, aws

- **gateway-sample**: Sample Gateway API resources for testing
- **envoy-gateway**: Alternative Gateway API implementation
- **spinnaker-***: Spinnaker continuous delivery platform components

Deployment
----------
To deploy the entire platform stack::

  kubectl apply -f clusters/production/platform-apps.yaml

To deploy individual components::

  kubectl apply -f clusters/production/apps/<app-name>.yaml

Observability Stack
-------------------
The observability stack provides full telemetry correlation:

1. **Application instrumentation**: Use OpenTelemetry SDKs to emit OTLP data
2. **Collection**: OTel Collector receives and processes telemetry
3. **Storage**:
   - Metrics → Prometheus
   - Traces → Jaeger → Elasticsearch
   - Logs → (Loki planned)
4. **Visualization**: Grafana with pre-configured data sources

Access Grafana::

  kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

Access Jaeger UI::

  kubectl port-forward -n observability svc/jaeger-query 16686:16686

See ``docs/ara/ARA-0005-observability-stack.md`` for architecture details.

Notes
-----
- Secrets are sourced via ESO from cloud provider secret managers
- All components use ArgoCD automated sync with prune and self-heal enabled
- Sync waves ensure proper dependency ordering during bootstrap