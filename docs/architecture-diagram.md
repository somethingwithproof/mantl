# Mantl Platform Architecture

**Last Updated**: 2025-12-27

## Overview

This document provides a comprehensive view of the Mantl platform architecture, including component relationships, data flows, deployment patterns, and multi-cloud support.

**Target Audience**: Platform engineers, DevOps teams, and solution architects evaluating or operating Mantl.

## Table of Contents

- [Platform Stack Overview](#platform-stack-overview)
- [Observability Data Flow](#observability-data-flow)
- [CI/CD Pipeline](#cicd-pipeline)
- [Multi-Cloud Architecture](#multi-cloud-architecture)
- [Architecture Principles](#architecture-principles)
- [Component Matrix](#component-matrix)
- [Related Documentation](#related-documentation)

## Platform Stack Overview

The Mantl platform consists of multiple layers deployed in dependency order using ArgoCD sync waves. This ensures components are available before dependent services attempt to use them.

### Deployment Waves

The platform deploys components in six waves:

1. **Wave 1: Networking** - CNI and Gateway API foundation
2. **Wave 2: Certificate Management** - TLS automation
3. **Wave 3: Security & Infrastructure** - Secrets, policies, and cloud resource management
4. **Wave 4: Observability Backend** - Metrics, logs, and traces storage
5. **Wave 5: Telemetry & Security** - Data collection and runtime security
6. **Wave 6: Platform Services** - Registry and DNS automation

### Architecture Diagram

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

### Key Integration Points

- **GitOps**: ArgoCD continuously reconciles cluster state with Git repository
- **Secrets Management**: External Secrets Operator syncs secrets from cloud providers
- **Policy Enforcement**: Kyverno validates resources and verifies image signatures
- **DNS Automation**: ExternalDNS creates DNS records from Kubernetes Ingress/Services
- **TLS Automation**: cert-manager provisions and renews certificates automatically
- **Observability**: OpenTelemetry Collector routes telemetry to appropriate backends

## Observability Data Flow

The observability stack provides unified visibility across metrics, logs, and traces using the OpenTelemetry standard.

### Telemetry Architecture

Applications instrument with OpenTelemetry SDKs and send all telemetry to the OpenTelemetry Collector via gRPC. The Collector then routes data to specialized backends:

- **Metrics** → Prometheus (15-day retention, 50Gi storage)
- **Traces** → Tempo (7-day retention, object storage)
- **Logs** → Loki (31-day retention, object storage)

Grafana provides a unified query interface with correlation between all three signal types.

### Sequence Diagram

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

### Benefits

- **Single Protocol**: Applications only need to implement OTLP
- **Correlation**: Navigate from logs → traces → metrics seamlessly
- **Vendor Neutral**: OpenTelemetry is a CNCF standard
- **Cost Optimized**: Different retention policies per signal type

## CI/CD Pipeline

Mantl implements a secure CI/CD pipeline with automated security scanning, signing, and GitOps-based deployment.

### Pipeline Stages

**Continuous Integration (GitHub Actions):**

1. **Lint** - Code quality checks (Black, Ruff, yamllint)
2. **Test** - Unit and integration tests (pytest)
3. **Build** - Container image creation
4. **Scan** - Security vulnerability scanning (Trivy)
5. **SBOM** - Software Bill of Materials generation (Syft)
6. **Sign** - Cryptographic signing with Cosign (OIDC keyless)
7. **Push** - Publish to container registry (GHCR, ECR, GCR)

**Continuous Deployment (ArgoCD):**

1. **PR Creation** - Automated or manual image tag update
2. **Review** - Human approval for production changes
3. **Merge** - Commit to main branch
4. **Sync** - ArgoCD detects changes and synchronizes
5. **Verify** - Kyverno validates image signatures
6. **Deploy** - Rollout to Kubernetes cluster

### Workflow Diagram

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

### Security Features

- **Signed Artifacts**: All images signed with Cosign (OIDC keyless signing)
- **Verification**: Kyverno policies enforce signature verification before deployment
- **SBOM**: Software Bill of Materials generated for supply chain transparency
- **Vulnerability Scanning**: Trivy scans for CVEs before images are published

## Multi-Cloud Architecture

Mantl provides consistent platform capabilities across seven cloud providers using a two-phase provisioning model.

### Supported Providers

| Provider | Kubernetes Service | Authentication Method | Status |
|----------|-------------------|----------------------|---------|
| AWS | Elastic Kubernetes Service (EKS) | IRSA (IAM Roles for Service Accounts) | ✅ Production |
| Google Cloud | Google Kubernetes Engine (GKE) | Workload Identity | ✅ Production |
| Azure | Azure Kubernetes Service (AKS) | Managed Identity | ✅ Production |
| DigitalOcean | DigitalOcean Kubernetes (DOKS) | Token Authentication | ✅ Production |
| Linode | Linode Kubernetes Engine (LKE) | Token Authentication | ✅ Production |
| Oracle Cloud | Oracle Kubernetes Engine (OKE) | Instance Principal | ✅ Production |
| IBM Cloud | IBM Kubernetes Service (IKS) | IAM Authentication | ✅ Production |

### Two-Phase Provisioning

**Phase 1: Cluster Bootstrap (Day 0)**
- Terraform/OpenTofu blueprints provision managed Kubernetes clusters
- Located in `terraform/blueprints/<provider>`
- Handles VPC, networking, IAM, and cluster creation

**Phase 2: Cloud Resources (Day 2)**
- Crossplane provisions cloud resources from Kubernetes
- Databases, storage buckets, VPCs created as Kubernetes custom resources
- Enables GitOps for infrastructure management

### Architecture Diagram

```mermaid
graph TB
    subgraph "Cloud Providers"
        AWS[☁️ AWS EKS<br/>IRSA Auth]
        GCP[☁️ GCP GKE<br/>Workload Identity]
        Azure[☁️ Azure AKS<br/>Managed Identity]
        DO[☁️ DigitalOcean DOKS<br/>Token Auth]
        Linode[☁️ Linode LKE<br/>Token Auth]
        OCI[☁️ Oracle Cloud OKE<br/>Instance Principal]
        IBM[☁️ IBM Cloud IKS<br/>IAM Auth]
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
    TF -->|Creates| OCI
    TF -->|Creates| IBM

    AWS -->|Deploys| Platform
    GCP -->|Deploys| Platform
    Azure -->|Deploys| Platform
    DO -->|Deploys| Platform
    Linode -->|Deploys| Platform
    OCI -->|Deploys| Platform
    IBM -->|Deploys| Platform

    Platform -->|Provisions via| CP
    CP -->|Creates| RDS
    CP -->|Creates| S3
    CP -->|Creates| VPC

    Platform -->|Updates| Route53

    style TF fill:#7B1FA2
    style CP fill:#F57C00
    style Platform fill:#1976D2
```

### Multi-Cloud Benefits

- **Provider Flexibility**: Switch providers without application changes
- **Disaster Recovery**: Deploy across multiple clouds for resilience
- **Cost Optimization**: Use different providers for different workloads
- **Vendor Independence**: Avoid lock-in with portable platform

## Architecture Principles

Mantl follows six core architectural principles:

### 1. GitOps-First

**All configuration lives in Git and is deployed via ArgoCD.**

- Infrastructure as Code (Terraform blueprints)
- Platform configuration (Kustomize overlays)
- Application manifests (Helm charts or plain YAML)
- Policy definitions (Kyverno policies)

### 2. CNCF-Native

**Prefer Cloud Native Computing Foundation graduated or incubating projects.**

This ensures:
- Production-grade stability
- Active community support
- Vendor neutrality
- Interoperability

### 3. Multi-Cloud Ready

**Provide consistent platform capabilities across all major cloud providers.**

- Unified abstractions (External Secrets, ExternalDNS)
- Provider-specific authentication (IRSA, Workload Identity, etc.)
- Portable workloads

### 4. Security by Default

**Security is built-in, not bolted on.**

- Policy enforcement (Kyverno)
- Image signature verification (Cosign)
- Runtime threat detection (Falco)
- Secret management (External Secrets)
- Network policies (Cilium)

### 5. Observable

**Full correlation between logs, metrics, and traces.**

- Single collection protocol (OTLP)
- Unified visualization (Grafana)
- Distributed tracing enabled
- Golden signals monitored

### 6. Declarative

**All resources managed through Kubernetes APIs.**

- Applications (Deployment, StatefulSet)
- Infrastructure (Crossplane)
- Policies (Kyverno)
- Monitoring (ServiceMonitor)

## Component Matrix

### CNCF Project Status

| Component | CNCF Status | Category | Purpose |
|-----------|-------------|----------|---------|
| **ArgoCD** | Graduated | GitOps | Continuous deployment |
| **Cilium** | Graduated | Networking | CNI, service mesh, Gateway API |
| **cert-manager** | Graduated | Security | TLS certificate automation |
| **External Secrets** | Graduated | Security | Cloud secrets synchronization |
| **Kyverno** | Graduated | Security | Policy enforcement |
| **Prometheus** | Graduated | Observability | Metrics collection |
| **OpenTelemetry** | Graduated | Observability | Telemetry collection |
| **Falco** | Graduated | Security | Runtime threat detection |
| **Harbor** | Graduated | Platform | Container registry |
| **Crossplane** | Incubating | Infrastructure | Cloud resource management |
| **ExternalDNS** | Incubating | Networking | DNS record automation |

### Non-CNCF Components

| Component | Vendor | Purpose |
|-----------|--------|---------|
| **Grafana** | Grafana Labs | Visualization and dashboards |
| **Loki** | Grafana Labs | Log aggregation |
| **Tempo** | Tempo Labs | Distributed tracing |

**Rationale**: These are industry-standard observability tools with excellent integration.

## Related Documentation

- **[Quick Start Guide](quickstart-platform.md)** - Deploy the platform in minutes
- **[Development Guide](../DEVELOPMENT.md)** - Contributing and local development
- **[Security Guide](security.md)** - Security features and configuration
- **[Architecture Decision Records](ara/README.md)** - Design decisions and rationale
- **[Compliance Module](../compliance/README.md)** - Compliance-as-Code documentation

## Next Steps

1. **Deploy**: Follow the [Quick Start Guide](quickstart-platform.md)
2. **Customize**: Review [Architecture Decision Records](ara/README.md) for design choices
3. **Secure**: Configure [security features](security.md)
4. **Contribute**: See [CONTRIBUTING.md](../CONTRIBUTING.md)

---

**Questions?** Open an issue on [GitHub](https://github.com/thomasvincent/mantl/issues).

## Network Topology

### AWS EKS Network Architecture

```mermaid
graph TB
    subgraph "Internet"
        Users[Users]
        GitHub[GitHub<br/>GitOps Source]
    end

    subgraph "AWS Region - us-east-1"
        subgraph "VPC 10.0.0.0/16"
            subgraph "Public Subnets"
                NAT1[NAT Gateway<br/>AZ-1]
                NAT2[NAT Gateway<br/>AZ-2]
                NAT3[NAT Gateway<br/>AZ-3]
                ALB[Application Load<br/>Balancer]
            end

            subgraph "Private Subnets - EKS Nodes"
                subgraph "AZ-1 - 10.0.0.0/20"
                    Node1[EKS Node<br/>System Pool]
                    Node2[EKS Node<br/>Workload Pool]
                end
                subgraph "AZ-2 - 10.0.16.0/20"
                    Node3[EKS Node<br/>System Pool]
                    Node4[EKS Node<br/>Workload Pool]
                end
                subgraph "AZ-3 - 10.0.32.0/20"
                    Node5[EKS Node<br/>System Pool]
                    Node6[EKS Node<br/>Workload Pool]
                end
            end

            subgraph "Intra Subnets - Control Plane"
                EKS_CP[EKS Control Plane<br/>Private Endpoint Only]
            end

            subgraph "Monitoring"
                VPC_Logs[VPC Flow Logs<br/>→ CloudWatch]
            end
        end

        subgraph "AWS Services"
            Route53[Route53<br/>DNS]
            SecretsManager[Secrets Manager]
            ECR[ECR<br/>Container Registry]
            KMS[KMS<br/>Encryption Keys]
        end
    end

    Users -->|HTTPS| ALB
    ALB --> Node2
    ALB --> Node4
    ALB --> Node6

    Node1 -->|Internet| NAT1
    Node2 -->|Internet| NAT1
    Node3 -->|Internet| NAT2
    Node4 -->|Internet| NAT2
    Node5 -->|Internet| NAT3
    Node6 -->|Internet| NAT3

    Node1 -.->|HTTPS| EKS_CP
    Node2 -.->|HTTPS| EKS_CP
    Node3 -.->|HTTPS| EKS_CP
    Node4 -.->|HTTPS| EKS_CP
    Node5 -.->|HTTPS| EKS_CP
    Node6 -.->|HTTPS| EKS_CP

    GitHub -->|Webhook| ALB

    Node1 -.->|IRSA| SecretsManager
    Node1 -.->|IRSA| Route53
    Node1 -.->|IRSA| ECR

    VPC_Logs -.->|Network Traffic| Node1
    VPC_Logs -.->|Network Traffic| Node2
    VPC_Logs -.->|Network Traffic| ALB

    style EKS_CP fill:#f9f,stroke:#333,stroke-width:4px
    style VPC_Logs fill:#ff9,stroke:#333,stroke-width:2px
```

**Security Features:**
- ✅ Private EKS endpoint (no public API access)
- ✅ Nodes in private subnets (no direct internet access)
- ✅ NAT Gateways for egress (redundant across AZs)
- ✅ VPC Flow Logs enabled (security monitoring)
- ✅ Network policies enforced by Cilium
- ✅ IRSA for pod-level AWS permissions (no node credentials)

## Security Boundaries

### Defense in Depth Architecture

```mermaid
graph TB
    subgraph "Layer 7: Application Security"
        AppAuth[Application<br/>Authentication]
        AppAuthz[Application<br/>Authorization]
        InputVal[Input Validation]
    end

    subgraph "Layer 6: API Gateway & Ingress"
        GatewayAPI[Gateway API<br/>Rate Limiting]
        WAF[Web Application<br/>Firewall - Optional]
        TLS[TLS Termination<br/>cert-manager]
    end

    subgraph "Layer 5: Service Mesh - Optional"
        mTLS[Mutual TLS<br/>Service-to-Service]
        AuthN[Service Identity<br/>SPIFFE/SPIRE]
    end

    subgraph "Layer 4: Policy Enforcement"
        Kyverno[Kyverno<br/>Admission Control]
        ImageVerify[Image Signature<br/>Verification]
        PolicyBlock[Block Unauthorized<br/>Registries]
    end

    subgraph "Layer 3: Runtime Security"
        Falco[Falco<br/>Syscall Monitoring]
        Tetragon[Tetragon<br/>eBPF Enforcement]
        PodSecurity[Pod Security<br/>Standards]
    end

    subgraph "Layer 2: Network Security"
        NetworkPol[Cilium Network<br/>Policies]
        SecGroup[Cloud Security<br/>Groups]
        PrivateNet[Private Networking]
    end

    subgraph "Layer 1: Infrastructure Security"
        IRSA[Workload Identity<br/>IRSA/WI]
        Encryption[Encryption at Rest<br/>KMS]
        VPCLogs[VPC Flow Logs<br/>Monitoring]
        PrivateEndpoint[Private API<br/>Endpoint]
    end

    AppAuth --> GatewayAPI
    GatewayAPI --> mTLS
    mTLS --> Kyverno
    Kyverno --> Falco
    Falco --> NetworkPol
    NetworkPol --> IRSA

    style Kyverno fill:#f96,stroke:#333,stroke-width:4px
    style Falco fill:#f96,stroke:#333,stroke-width:4px
    style NetworkPol fill:#f96,stroke:#333,stroke-width:4px
    style ImageVerify fill:#f96,stroke:#333,stroke-width:4px
```

**Security Layers Explained:**

1. **Infrastructure (Layer 1)**: Cloud-level security controls
   - Private API endpoints prevent public access
   - KMS encryption protects data at rest
   - Workload Identity eliminates long-lived credentials
   - VPC Flow Logs provide network forensics

2. **Network (Layer 2)**: Network segmentation and isolation
   - Cilium NetworkPolicies control pod-to-pod traffic
   - Security groups filter traffic at node level
   - Private subnets prevent direct internet access

3. **Runtime (Layer 3)**: Container runtime protection
   - Falco detects anomalous syscalls and file access
   - Tetragon enforces eBPF-based security policies
   - Pod Security Standards prevent privilege escalation

4. **Policy (Layer 4)**: Admission-time validation
   - Kyverno blocks pods before they start
   - Image signature verification ensures provenance
   - Registry restrictions prevent supply chain attacks

5. **Service Mesh (Layer 5)**: Service-to-service security (optional)
   - mTLS encrypts all service communication
   - SPIFFE provides service identity
   - Zero trust network model

6. **Gateway (Layer 6)**: Edge security
   - Rate limiting prevents DoS
   - WAF blocks common web attacks
   - TLS enforces encryption in transit

7. **Application (Layer 7)**: Application-level controls
   - Authentication verifies user identity
   - Authorization enforces access control
   - Input validation prevents injection attacks

## Disaster Recovery Architecture

### Multi-Region Failover Strategy

```mermaid
graph TB
    subgraph "Primary Region - us-east-1"
        Primary[EKS Cluster<br/>us-east-1<br/>🟢 Active]
        PrimaryDB[(Application DBs<br/>Primary)]
        PrimaryBackup[Automated Backups<br/>Velero]
    end

    subgraph "Secondary Region - us-west-2"
        Secondary[EKS Cluster<br/>us-west-2<br/>🟡 Standby]
        SecondaryDB[(Application DBs<br/>Replica)]
        SecondaryBackup[Backup Restore<br/>Velero]
    end

    subgraph "Global Services"
        Route53[Route53<br/>Global DNS<br/>Health Checks]
        S3[S3 Backup Storage<br/>Cross-Region Replication]
        ECR[ECR<br/>Multi-Region Registry]
    end

    subgraph "GitOps Source"
        Git[GitHub<br/>GitOps Source<br/>Geo-Redundant]
    end

    Route53 -->|Health Check OK| Primary
    Route53 -.->|Health Check FAIL| Secondary

    Primary --> PrimaryBackup
    PrimaryBackup -->|Daily Snapshots| S3
    S3 -->|Cross-Region Sync| SecondaryBackup
    SecondaryBackup --> Secondary

    PrimaryDB -->|Async Replication| SecondaryDB

    Git -->|ArgoCD| Primary
    Git -.->|ArgoCD Standby| Secondary

    Primary -.->|Image Pull| ECR
    Secondary -.->|Image Pull| ECR

    style Primary fill:#9f9,stroke:#333,stroke-width:4px
    style Secondary fill:#ff9,stroke:#333,stroke-width:2px
    style Route53 fill:#99f,stroke:#333,stroke-width:4px
```

**Recovery Objectives:**

- **RTO (Recovery Time Objective)**: < 1 hour
  - Infrastructure provisioning: 30 minutes (Terraform)
  - Platform deployment: 15 minutes (ArgoCD)
  - Application restore: 15 minutes (Velero)

- **RPO (Recovery Point Objective)**: < 24 hours
  - Daily automated backups via Velero
  - Application database replication varies by workload

**Failover Process:**

1. **Detection** (< 5 minutes)
   - Route53 health checks detect primary region failure
   - Monitoring alerts trigger incident response

2. **DNS Failover** (< 2 minutes)
   - Route53 automatically routes traffic to secondary region
   - No manual intervention required for DNS

3. **Cluster Activation** (< 30 minutes)
   - Terraform provisions infrastructure if needed (cold standby)
   - ArgoCD deploys platform components from Git
   - Velero restores application data from S3

4. **Application Recovery** (< 15 minutes)
   - Database failover to secondary region replicas
   - Application pods start and pass health checks
   - Services become available

5. **Verification** (< 5 minutes)
   - Smoke tests validate critical paths
   - Monitoring confirms services healthy

**Backup Strategy:**

- **Infrastructure**: Terraform state in S3 with versioning
- **Platform Config**: Git repository (GitHub redundant storage)
- **Application Data**: Velero daily backups to S3
- **Persistent Volumes**: EBS snapshots, S3 cross-region replication
- **Secrets**: External Secrets (provider-native backup)

**Testing:**

- **Monthly**: Backup restore verification
- **Quarterly**: Full DR drill to secondary region
- **Annually**: Complete primary region failover simulation

## Data Flow Diagram

### Observability Data Pipeline

```mermaid
graph LR
    subgraph "Applications"
        App1[App Pod 1<br/>OTLP SDK]
        App2[App Pod 2<br/>OTLP SDK]
        App3[App Pod 3<br/>OTLP SDK]
    end

    subgraph "Collection Layer"
        OTelCol[OpenTelemetry<br/>Collector<br/>🔄 Process & Route]
    end

    subgraph "Storage Layer"
        Prometheus[(Prometheus<br/>📊 Metrics<br/>15d retention)]
        Loki[(Loki<br/>📝 Logs<br/>30d retention)]
        Tempo[(Tempo<br/>🔍 Traces<br/>7d retention)]
    end

    subgraph "Analysis Layer"
        Grafana[Grafana<br/>📈 Dashboards<br/>Alerts<br/>Correlation]
        AlertManager[AlertManager<br/>🔔 Alert Routing<br/>PagerDuty/Slack]
    end

    subgraph "Long-Term Storage"
        S3[(S3<br/>☁️ Cold Storage<br/>1yr+ retention)]
    end

    App1 -->|OTLP/gRPC<br/>Metrics| OTelCol
    App2 -->|OTLP/gRPC<br/>Metrics| OTelCol
    App3 -->|OTLP/gRPC<br/>Metrics| OTelCol

    App1 -->|OTLP/gRPC<br/>Logs| OTelCol
    App2 -->|OTLP/gRPC<br/>Logs| OTelCol
    App3 -->|OTLP/gRPC<br/>Logs| OTelCol

    App1 -->|OTLP/gRPC<br/>Traces| OTelCol
    App2 -->|OTLP/gRPC<br/>Traces| OTelCol
    App3 -->|OTLP/gRPC<br/>Traces| OTelCol

    OTelCol -->|Prometheus<br/>Remote Write| Prometheus
    OTelCol -->|HTTP| Loki
    OTelCol -->|OTLP| Tempo

    Prometheus --> Grafana
    Loki --> Grafana
    Tempo --> Grafana

    Prometheus -->|Recording Rules| AlertManager
    AlertManager -->|Notifications| Grafana

    Prometheus -.->|Export| S3
    Loki -.->|Export| S3
    Tempo -.->|Export| S3

    style OTelCol fill:#9cf,stroke:#333,stroke-width:4px
    style Grafana fill:#f96,stroke:#333,stroke-width:4px
```

**Data Flow Characteristics:**

- **Single Protocol**: OTLP (OpenTelemetry Protocol) for all signals
- **Vendor Neutral**: Open standards, no vendor lock-in
- **Correlation**: Trace IDs link logs, metrics, and traces
- **Cardinality Control**: OTel Collector filters high-cardinality data
- **Cost Optimization**: Tiered storage (hot → warm → cold)

---

## Related Documentation

- **[Quickstart Guide](quickstart-platform.md)** - Deploy the platform
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions
- **[Runbooks](runbooks.md)** - Operational procedures
- **[Security Baseline](security.md)** - Security controls
- **[Testing Guide](../tests/README.md)** - Running tests

