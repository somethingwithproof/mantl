# Mantl Platform Architecture

## Complete Platform Stack

```mermaid
graph TB
    subgraph "External"
        User[👤 Users]
        Git[📦 Git Repository]
        Cloud[☁️ Cloud Secrets<br/>AWS/GCP/Azure]
        DNS_Provider[🌐 DNS Provider<br/>Route53/CloudDNS]
    end

    subgraph "Kubernetes Cluster"
        subgraph "Wave 1: Networking"
            Cilium[🔷 Cilium CNI<br/>eBPF + Gateway API<br/>CNCF Graduated]
        end

        subgraph "Wave 2: Certificate Management"
            CertManager[🔒 cert-manager<br/>TLS Automation<br/>CNCF Graduated]
        end

        subgraph "Wave 3: Security & Infrastructure"
            ESO[🔐 External Secrets<br/>Secret Sync<br/>CNCF Graduated]
            Kyverno[📋 Kyverno<br/>Policy Engine<br/>CNCF Graduated]
            Crossplane[⚙️ Crossplane<br/>IaC Day 2 Ops<br/>CNCF Incubating]
        end

        subgraph "Wave 4: Observability Backend"
            Prometheus[📊 Prometheus<br/>Metrics + Alertmanager<br/>CNCF Graduated]
            Loki[📝 Loki<br/>Log Aggregation<br/>Grafana Labs]
            Tempo[🔍 Tempo<br/>Distributed Tracing<br/>Grafana Labs]
            Grafana[📈 Grafana<br/>Visualization<br/>Grafana Labs]
        end

        subgraph "Wave 5: Telemetry & Security"
            OTel[📡 OpenTelemetry<br/>Collector<br/>CNCF Graduated]
            Falco[🛡️ Falco<br/>Runtime Security<br/>CNCF Graduated]
        end

        subgraph "Wave 6: Platform Services"
            Harbor[🐳 Harbor<br/>Registry On-Prem<br/>CNCF Graduated]
            ExternalDNS[🌍 ExternalDNS<br/>DNS Automation<br/>CNCF Incubating]
        end

        subgraph "GitOps Engine"
            ArgoCD[🔄 ArgoCD<br/>GitOps Platform + Apps<br/>CNCF Graduated]
        end

        subgraph "Application Layer"
            Apps[🚀 User Applications<br/>OTLP Instrumented]
        end
    end

    %% GitOps Flow
    Git -->|Pull Manifests| ArgoCD
    ArgoCD -->|Deploy Wave 1| Cilium
    ArgoCD -->|Deploy Wave 2| CertManager
    ArgoCD -->|Deploy Wave 3| ESO
    ArgoCD -->|Deploy Wave 3| Kyverno
    ArgoCD -->|Deploy Wave 3| Crossplane
    ArgoCD -->|Deploy Wave 4| Prometheus
    ArgoCD -->|Deploy Wave 4| Loki
    ArgoCD -->|Deploy Wave 4| Tempo
    ArgoCD -->|Deploy Wave 5| OTel
    ArgoCD -->|Deploy Wave 5| Falco
    ArgoCD -->|Deploy Wave 6| Harbor
    ArgoCD -->|Deploy Wave 6| ExternalDNS
    ArgoCD -->|Deploy Apps| Apps

    %% Secrets Flow
    Cloud -->|Sync Secrets| ESO
    ESO -->|Kubernetes Secrets| Apps

    %% Policy Enforcement
    Kyverno -->|Validate| Apps
    Kyverno -->|Verify Signatures| Apps

    %% Infrastructure Management
    Crossplane -->|Provision| Cloud

    %% DNS Automation
    ExternalDNS -->|Create Records| DNS_Provider
    Apps -->|Ingress| Cilium
    Cilium -->|Triggers| ExternalDNS

    %% TLS Automation
    CertManager -->|Request Certs| DNS_Provider
    CertManager -->|TLS Secrets| Apps

    %% Observability Data Flow
    Apps -->|OTLP gRPC :4317| OTel
    OTel -->|Metrics :8889| Prometheus
    OTel -->|Traces :4317| Tempo
    OTel -->|Logs :80| Loki

    Prometheus -->|Query| Grafana
    Loki -->|Query| Grafana
    Tempo -->|Query| Grafana

    %% Monitoring
    Cilium -->|ServiceMonitor| Prometheus
    ArgoCD -->|ServiceMonitor| Prometheus
    Falco -->|Alerts| Prometheus

    %% User Access
    User -->|HTTPS| Cilium
    User -->|UI| ArgoCD
    User -->|UI| Grafana

    classDef graduated fill:#2E7D32,stroke:#1B5E20,color:#fff
    classDef incubating fill:#F57C00,stroke:#E65100,color:#fff
    classDef nonCNCF fill:#1976D2,stroke:#0D47A1,color:#fff
    classDef gitops fill:#6A1B9A,stroke:#4A148C,color:#fff

    class Cilium,CertManager,Kyverno,ESO,Prometheus,OTel,Falco,Harbor,ArgoCD graduated
    class Crossplane,ExternalDNS incubating
    class Loki,Tempo,Grafana nonCNCF
    class Git,ArgoCD gitops
```

## Data Flow: Observability Stack

```mermaid
sequenceDiagram
    participant App as Application<br/>(OTLP SDK)
    participant OTel as OpenTelemetry<br/>Collector
    participant Prom as Prometheus
    participant Tempo as Tempo
    participant Loki as Loki
    participant Grafana as Grafana
    participant User as Platform User

    App->>OTel: Send OTLP telemetry<br/>(gRPC :4317)

    par Metrics Path
        OTel->>Prom: Export metrics<br/>(RemoteWrite :8889)
        Prom->>Prom: Store 15 days<br/>(50Gi PV)
    and Traces Path
        OTel->>Tempo: Export traces<br/>(OTLP :4317)
        Tempo->>Tempo: Store 7 days<br/>(Object Storage)
    and Logs Path
        OTel->>Loki: Export logs<br/>(HTTP :80)
        Loki->>Loki: Store 31 days<br/>(Object Storage)
    end

    User->>Grafana: Query metrics
    Grafana->>Prom: PromQL query
    Prom-->>Grafana: Metrics data

    User->>Grafana: Query traces
    Grafana->>Tempo: TraceQL query
    Tempo-->>Grafana: Trace data

    User->>Grafana: Query logs
    Grafana->>Loki: LogQL query
    Loki-->>Grafana: Log data

    Note over Grafana: Unified view:<br/>Logs ↔ Metrics ↔ Traces
```

## CI/CD Workflow

```mermaid
flowchart LR
    subgraph "CI (GitHub Actions)"
        Code[💻 Code Push]
        Lint[✅ Lint<br/>black, ruff, yamllint]
        Test[🧪 Test<br/>pytest]
        Build[🏗️ Build Image]
        Scan[🔍 Scan<br/>Trivy]
        SBOM[📋 SBOM<br/>Syft]
        Sign[✍️ Sign<br/>Cosign OIDC]
        Push[📤 Push Image<br/>GHCR/ECR/GCR]
    end

    subgraph "CD (ArgoCD)"
        PR[📝 Create PR<br/>Update image tag]
        Review[👀 Review + Approve]
        Merge[🔀 Merge to main]
        Sync[🔄 ArgoCD Sync]
        Verify[✅ Kyverno Verify<br/>Image Signature]
        Deploy[🚀 Deploy to K8s]
    end

    Code --> Lint
    Lint --> Test
    Test --> Build
    Build --> Scan
    Scan --> SBOM
    SBOM --> Sign
    Sign --> Push
    Push --> PR
    PR --> Review
    Review --> Merge
    Merge --> Sync
    Sync --> Verify
    Verify --> Deploy

    style Code fill:#4CAF50
    style Deploy fill:#2196F3
    style Sign fill:#FF9800
    style Verify fill:#FF9800
```

## Multi-Cloud Architecture

```mermaid
graph TB
    subgraph "Cloud Providers"
        AWS[☁️ AWS EKS<br/>IRSA Auth]
        GCP[☁️ GCP GKE<br/>Workload Identity]
        Azure[☁️ Azure AKS<br/>Managed Identity]
        DO[☁️ DigitalOcean DOKS<br/>Token Auth]
        Linode[☁️ Linode LKE<br/>Token Auth]
    end

    subgraph "Terraform Blueprints"
        TF[📦 terraform/blueprints/<br/>Day 0: Cluster Bootstrap]
    end

    subgraph "Crossplane"
        CP[⚙️ Crossplane<br/>Day 2: Cloud Resources]
        RDS[💾 RDS Instance]
        S3[🪣 S3 Bucket]
        VPC[🌐 VPC/Network]
    end

    subgraph "Platform (All Clouds)"
        Platform[🔷 Mantl Platform<br/>ArgoCD + Cilium + Observability]
    end

    subgraph "Global DNS"
        Route53[🌍 Route53 / CloudDNS<br/>ExternalDNS managed]
    end

    TF -->|Creates| AWS
    TF -->|Creates| GCP
    TF -->|Creates| Azure
    TF -->|Creates| DO
    TF -->|Creates| Linode

    AWS -->|Deploys| Platform
    GCP -->|Deploys| Platform
    Azure -->|Deploys| Platform
    DO -->|Deploys| Platform
    Linode -->|Deploys| Platform

    Platform -->|Provisions via| CP
    CP -->|Creates| RDS
    CP -->|Creates| S3
    CP -->|Creates| VPC

    Platform -->|Updates| Route53

    style TF fill:#7B1FA2
    style CP fill:#F57C00
    style Platform fill:#1976D2
```

## Legend

- 🟢 **CNCF Graduated** (9 projects): Kubernetes, Cilium, ArgoCD, cert-manager, Kyverno, External Secrets, Prometheus, OpenTelemetry, Falco, Harbor
- 🟠 **CNCF Incubating** (2 projects): Crossplane, ExternalDNS
- 🔵 **Non-CNCF** (3 projects): Grafana, Loki, Tempo

## Architecture Principles

1. **GitOps-First**: All configuration in Git, deployed via ArgoCD
2. **CNCF-Native**: Prefer graduated/incubating projects for stability
3. **Multi-Cloud**: Consistent platform across AWS, GCP, Azure, DO, Linode
4. **Security by Default**: Policy enforcement, image signing, runtime security
5. **Observable**: Full telemetry correlation (logs ↔ metrics ↔ traces)
6. **Declarative**: Kubernetes APIs for apps and infrastructure (Crossplane)
