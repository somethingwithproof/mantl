# ARA-0007 Multi-cloud and failover
Status: Accepted
Date: 2025-12-27

Context
Mantl's spirit is multi-cloud and globally distributed services. Teams need the flexibility to run workloads on multiple cloud providers with consistent tooling.

Decision
**Cluster Provisioning** (`terraform/blueprints/`):
- **AWS EKS**: With IRSA for workload identity
- **GCP GKE**: With Workload Identity
- **Azure AKS**: With Managed Identity
- **DigitalOcean DOKS**: With API token authentication
- **Linode LKE**: With API token authentication
- **Oracle Cloud OKE**: With Instance Principal for workload identity
- **IBM Cloud IKS**: With IAM for authentication
- **Bare Metal/On-Prem**: k3s or kubeadm (planned)

**Infrastructure Management**:
- **Day 0 (Bootstrap)**: Terraform creates Kubernetes clusters
- **Day 2 (Operations)**: Crossplane manages cloud resources (databases, storage, networking) as Kubernetes CRs
- Crossplane provides single API across AWS, GCP, Azure, and 100+ providers

**Global Routing**:
- ExternalDNS for automatic DNS record management per cloud provider
- DNS-based global routing and health checks for active/active or active/passive
- Gateway API with Cilium for ingress across all clouds

**Data Replication**:
- Replicate only the data stores that justify cross-region/cloud complexity
- Object storage (S3, GCS, Azure Blob) for Loki and Tempo backends
- Multi-region databases via cloud-native services or Crossplane compositions

Consequences
- Resilience to regional/provider outages
- Consistent platform experience across clouds (same GitOps, policies, observability)
- Avoid vendor lock-in; teams can migrate workloads between clouds
- Higher cost/complexity where data replication is required
- Crossplane enables self-service infrastructure provisioning via GitOps

Alternatives
- **Single cloud with regional redundancy**: Simpler but vendor lock-in and regional-only resilience
- **Terraform for all infrastructure**: Requires separate workflow from application deployment; Crossplane unifies via Kubernetes API
- **Cloud-specific tools (CDK, Deployment Manager)**: Fragments tooling across clouds
