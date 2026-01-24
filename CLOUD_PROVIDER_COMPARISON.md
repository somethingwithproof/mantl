# Cloud Provider Feature Parity Comparison

## Overview

The mantl platform supports 8 cloud providers with production-grade security. However, due to differences in cloud provider capabilities, **not all providers have identical features**. This document explains what's available on each platform.

## Feature Comparison Matrix

| Feature | AWS EKS | GCP GKE | Azure AKS | Oracle OKE | IBM IKS | DigitalOcean DOKS | Linode LKE | OpenStack |
|---------|---------|---------|-----------|------------|---------|-------------------|------------|-----------|
| **Kubernetes Version** | 1.31 ✓ | 1.31 ✓ | 1.31 ✓ | 1.31 ✓ | 1.31 ✓ | 1.31 ✓ | 1.31 ✓ | 1.31 ✓ |
| **Private Networking** | VPC ✓ | VPC ✓ | VNet ✓ | VCN ✓ | VPC ✓ | VPC ✓ | Firewall ✓ | Neutron ✓ |
| **Network Flow Logs** | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ | Varies¹ |
| **Secrets Encryption (KMS)** | AWS KMS ✓ | GCP KMS ✓ | Key Vault ✓ | OCI Vault ✓ | Key Protect ✓ | ✗² | ✗² | Barbican¹ |
| **Private Cluster/Endpoint** | ✓ | ✓ | ✓ | ✓ | ✓ | ✗³ | ✗³ | Load Balancer |
| **Separate Node Pools** | System/Workload ✓ | System/Workload ✓ | System/Workload ✓ | System/Workload ✓ | System/Workload/Spot ✓ | System/Workload/Spot ✓ | System/Workload ✓ | Magnum Labels |
| **Auto-scaling** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **High Availability** | Multi-AZ ✓ | Multi-Zone ✓ | Multi-Zone ✓ | Multi-AD ✓ | Multi-Zone ✓ | Regional ✓ | Regional + HA CP ✓ | Multi-Master |
| **Backup Storage** | S3 ✓ | GCS ✓ | Blob ✓ | Object Storage ✓ | COS ✓ | Spaces ✓ | Object Storage ✓ | Swift ✓ |
| **Workload Identity** | IRSA ✓ | Workload Identity ✓ | Managed Identity ✓ | Workload Identity ✓ | ✓ | ✗⁴ | ✗⁴ | Service Accounts |
| **Managed Control Plane** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ (HA optional) | ✓ |
| **Spot/Preemptible Instances** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗⁵ | Varies |
| **Load Balancer Integration** | ALB/NLB ✓ | GCP LB ✓ | Azure LB ✓ | OCI LB ✓ | IBM LB ✓ | DO LB ✓ | NodeBalancer ✓ | Octavia ✓ |
| **Monitoring Integration** | CloudWatch ✓ | Cloud Monitoring ✓ | Monitor ✓ | Logging ✓ | Log Analysis ✓ | DO Monitoring ✓ | ✗⁶ | Varies¹ |
| **Container Registry** | ECR ✓ | GCR ✓ | ACR ✓ | OCIR ✓ | IBM CR ✓ | DO Registry ✓ | ✗⁷ | Varies¹ |

**Legend:**
- ✓ = Fully supported and implemented
- ✗ = Not available on this provider
- Varies¹ = Depends on OpenStack distribution/configuration
- ✗² = No native KMS service available
- ✗³ = Private cluster feature not available
- ✗⁴ = Workload identity concept not available (use service accounts)
- ✗⁵ = Spot instances not available on Linode
- ✗⁶ = Native monitoring not available (use Prometheus stack)
- ✗⁷ = No native container registry (use external registry)

## Provider Tier Classification

### Tier 1: Enterprise Cloud Providers (Full Feature Set)
**AWS EKS, GCP GKE, Azure AKS, Oracle OKE, IBM IKS**

These providers offer complete enterprise security features:
- ✓ Network flow logs for security monitoring
- ✓ Native KMS for secrets encryption at rest
- ✓ Private cluster endpoints
- ✓ Workload identity integration
- ✓ Comprehensive monitoring and logging

**Best For:** Enterprise production workloads requiring maximum security, compliance, and auditability.

### Tier 2: Developer-Focused Providers (Core Feature Set)
**DigitalOcean DOKS, Linode LKE**

These providers offer essential security features with excellent developer experience:
- ✓ Network isolation (VPC/Firewall)
- ✓ Auto-scaling and high availability
- ✓ Object storage for backups
- ✗ No network flow logs
- ✗ No native KMS service
- ✗ No private cluster endpoints

**Best For:** Development, staging, and production workloads where cost-efficiency and simplicity are priorities. Suitable for startups and small-to-medium businesses.

**Security Mitigation:**
- Use external secrets management (HashiCorp Vault, Sealed Secrets)
- Implement application-level encryption
- Use network policy enforcement for segmentation
- Deploy Falco for runtime security monitoring

### Tier 3: Private Cloud Provider (Flexible Feature Set)
**OpenStack**

OpenStack features depend heavily on the specific deployment:
- ✓ Complete control over infrastructure
- ✓ Magnum for Kubernetes orchestration
- ✓ Swift for object storage
- Varies: Flow logs depend on Neutron configuration
- Varies: KMS depends on Barbican availability
- Varies: Monitoring depends on deployment

**Best For:** On-premises deployments, regulatory requirements requiring data sovereignty, or organizations already using OpenStack.

## Detailed Feature Analysis

### Network Flow Logs

**Available:** AWS, GCP, Azure, Oracle, IBM
**Not Available:** DigitalOcean, Linode, OpenStack (standard)

Flow logs capture network traffic metadata for security analysis, compliance, and troubleshooting.

**Alternatives for providers without flow logs:**
- Deploy Cilium with Hubble for network observability
- Use eBPF-based tools (Falco, Tetragon)
- Implement ingress/egress logging at application layer

### Secrets Encryption (KMS)

**Available:** AWS KMS, GCP KMS, Azure Key Vault, OCI Vault, IBM Key Protect
**Not Available:** DigitalOcean, Linode
**Varies:** OpenStack (Barbican)

Native KMS encrypts Kubernetes secrets at rest in etcd.

**Alternatives for providers without KMS:**
- Deploy HashiCorp Vault with auto-unseal
- Use Sealed Secrets controller
- Implement External Secrets Operator
- Use SOPS for GitOps secrets management

### Private Cluster Endpoints

**Available:** AWS, GCP, Azure, Oracle, IBM
**Not Available:** DigitalOcean, Linode

Private endpoints restrict API server access to internal networks only.

**Alternatives for providers without private endpoints:**
- Use IP allowlisting (configured in blueprints)
- Deploy VPN or bastion hosts
- Implement mutual TLS authentication
- Use API server audit logging for monitoring

### Workload Identity

**Available:** AWS IRSA, GCP Workload Identity, Azure Managed Identity, Oracle, IBM
**Not Available:** DigitalOcean, Linode

Workload identity allows pods to authenticate to cloud services without static credentials.

**Alternatives:**
- Use Kubernetes service accounts with RBAC
- Deploy external-secrets operator
- Use init containers for credential management
- Implement credential rotation policies

### Spot/Preemptible Instances

**Available:** AWS, GCP, Azure, Oracle, IBM, DigitalOcean
**Not Available:** Linode

Spot instances provide significant cost savings (60-90% off) for interruptible workloads.

**Linode Alternative:**
- Use standard instances with aggressive auto-scaling
- Implement cost monitoring with OpenCost
- Use scheduled scaling for predictable workloads

## Implementation Status

All blueprints implement the **maximum security available** on each platform:

### AWS EKS ✓
- VPC with private subnets (3 AZs)
- VPC Flow Logs → CloudWatch
- AWS KMS for envelope encryption
- IRSA for workload identity
- S3 for backups with versioning

### GCP GKE ✓
- VPC with private networking
- VPC Flow Logs → Cloud Logging
- GCP KMS for envelope encryption
- Workload Identity enabled
- GCS for backups with versioning

### Azure AKS ✓
- VNet with private networking
- Network Watcher Flow Logs → Storage Account
- Azure Key Vault for secrets
- Managed Identity enabled
- Blob Storage for backups

### Oracle OKE ✓
- VCN with NAT Gateway
- VCN Flow Logs → OCI Logging (90-day retention)
- OCI Vault with 256-bit AES encryption
- Private endpoint by default
- Object Storage for backups

### IBM IKS ✓
- VPC with manual address prefixes (3 zones)
- VPC Flow Logs → COS
- IBM Key Protect for encryption
- Private endpoint by default
- COS for backups with versioning

### DigitalOcean DOKS ✓
- VPC with private networking
- Firewall rules (restrictive by default)
- Container Registry integration
- DigitalOcean Monitoring alerts
- Spaces (S3-compatible) for backups
- **Note:** No flow logs or KMS available

### Linode LKE ✓
- Cloud Firewall (stateful rules)
- High availability control plane
- NodeBalancer for ingress
- Object Storage (S3-compatible) for backups
- **Note:** No VPC, flow logs, or KMS available

### OpenStack ✓
- Neutron network with security groups
- Magnum for Kubernetes orchestration
- Octavia load balancer integration
- Swift container for backups with versioning
- Auto-scaling and auto-healing via labels
- **Note:** Flow logs and KMS depend on deployment

## Security Recommendations by Provider

### Enterprise Providers (AWS/GCP/Azure/Oracle/IBM)
✓ Use all native security features
✓ Enable private endpoints in production
✓ Configure flow logs with long retention
✓ Use workload identity for cloud resource access
✓ Implement KMS envelope encryption

### Developer Providers (DigitalOcean/Linode)
✓ Deploy HashiCorp Vault or Sealed Secrets for sensitive data
✓ Use network policies aggressively
✓ Deploy Falco for runtime security
✓ Implement Cilium for network observability
✓ Use external secrets management
✓ Configure aggressive IP allowlisting
✓ Deploy Prometheus stack for monitoring

### Private Cloud (OpenStack)
✓ Enable Barbican if available
✓ Configure Neutron flow logs if supported
✓ Use Octavia for load balancing
✓ Implement project-level isolation
✓ Deploy comprehensive monitoring stack
✓ Use Swift versioning for backups

## Cost Considerations

| Provider | Relative Cost | Best For |
|----------|---------------|----------|
| AWS EKS | High | Enterprise, complex workloads |
| GCP GKE | High | Enterprise, data analytics |
| Azure AKS | High | Enterprise, Microsoft ecosystem |
| Oracle OKE | Medium-High | Enterprise, Oracle ecosystem |
| IBM IKS | Medium-High | Enterprise, IBM ecosystem |
| **DigitalOcean DOKS** | **Low** | **Startups, small-medium production** |
| **Linode LKE** | **Low** | **Startups, simple workloads** |
| OpenStack | Varies | On-premises, data sovereignty |

## Migration Path

If you need to migrate from a Tier 2 provider to Tier 1:

1. **Deploy on new provider** using existing application manifests
2. **Migrate secrets** to cloud KMS
3. **Update workload identity** from service accounts to cloud-native
4. **Enable flow logs** for new cluster
5. **Test disaster recovery** with Velero cross-cloud restore
6. **Cutover traffic** with DNS/load balancer changes

All blueprints use consistent:
- Kubernetes 1.31.x (portable manifests)
- Separate system/workload node pools
- Velero for backups (cloud-agnostic)
- Prometheus stack (cloud-agnostic)
- Network policies (cloud-agnostic)

## Summary

**Are all clouds on feature parity?** No, but each is production-ready:

- **Tier 1 providers** (AWS/GCP/Azure/Oracle/IBM) have full security features
- **Tier 2 providers** (DigitalOcean/Linode) have core features with security gaps addressable by platform tools
- **Tier 3 providers** (OpenStack) are flexible and depend on deployment

All blueprints implement **maximum available security** for each platform and are suitable for production use with appropriate security controls.

Choose based on:
- **Budget:** DigitalOcean/Linode for cost-efficiency
- **Compliance:** AWS/GCP/Azure/Oracle/IBM for enterprise requirements
- **Data Sovereignty:** OpenStack for on-premises control
