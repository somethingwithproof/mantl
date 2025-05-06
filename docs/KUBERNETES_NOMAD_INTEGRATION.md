# Kubernetes-Nomad Integration in Mantl

This document describes the comprehensive integration between Kubernetes and HashiCorp Nomad in Mantl.

## Overview

Mantl provides a unified platform for container orchestration with support for multiple container orchestration systems. The integration between Kubernetes and Nomad offers a powerful hybrid orchestration capability with a full set of enterprise features:

1. **Cross-platform Service Discovery**: Automatic service discovery between Kubernetes and Nomad
2. **Shared Secrets Management**: Unified access to secrets with HashiCorp Vault
3. **Federated Metrics and Logging**: Integrated observability across platforms
4. **Multi-region/Multi-cluster Orchestration**: Seamless workload distribution
5. **GPU Workload Scheduling**: Hardware acceleration support
6. **Autoscaling Integration**: Dynamic workload scaling
7. **Custom Resource Definition Support**: Extended Kubernetes API integration
8. **Disaster Recovery with Automatic Failover**: High availability guarantees
9. **Zero-downtime Migration Tools**: Smooth service transitions
10. **Unified Policy Enforcement with OPA**: Consistent governance

## Components

The integration consists of the following components:

1. **Kubernetes Role and Configuration**
   - Integration tasks in the `kubernetes` role
   - Templates for Nomad configuration
   - Example Nomad job that deploys to Kubernetes

2. **Nomad Role and Configuration**
   - Nomad server and client configuration
   - Kubernetes integration plugin configuration
   - Service discovery between platforms

3. **Security Integration**
   - RBAC configuration for Nomad in Kubernetes
   - Service accounts and tokens for cross-platform authentication
   - Common Criteria compliance features
   - Open Policy Agent integration

## Feature Configuration

### Cross-platform Service Discovery

Enable automatic service discovery between Kubernetes and Nomad services.

```yaml
# Enable or disable service discovery (enabled by default)
kubernetes_nomad_service_discovery_enabled: true

# Label selector for identifying Kubernetes services to import into Nomad
kubernetes_nomad_service_discovery_label_selector: "mantl-service=true"

# Sync period for service discovery (how often to refresh services)
kubernetes_nomad_service_discovery_sync_period: "30s"

# Default check interval for discovered services
kubernetes_nomad_service_discovery_check_interval: "10s"
```

### Shared Secrets Management with Vault

Integrate Vault for secure secrets management across Kubernetes and Nomad.

```yaml
# Enable Vault integration
kubernetes_nomad_vault_integration_enabled: true

# Vault server address
vault_addr: "https://vault.service.consul:8200"

# Default TTL settings for Vault tokens
kubernetes_nomad_vault_default_ttl: "1h"
kubernetes_nomad_vault_max_ttl: "4h"

# Enable Kubernetes auth method for Vault
kubernetes_nomad_vault_k8s_auth: true
kubernetes_nomad_vault_k8s_role: "nomad"
```

### Federated Metrics and Logging

Configure unified metrics and logging across Kubernetes and Nomad.

```yaml
# Enable federated metrics
kubernetes_nomad_metrics_enabled: true

# Metrics collection interval
kubernetes_nomad_metrics_interval: "10s"

# Prometheus retention time
kubernetes_nomad_prometheus_retention: "24h"

# Optional integrations
kubernetes_nomad_statsd_enabled: false
kubernetes_nomad_datadog_enabled: false

# Fluentd for logging integration
kubernetes_nomad_fluentd_image: "fluent/fluentd:v1.14"
kubernetes_nomad_elasticsearch_host: "elasticsearch.service.consul"
kubernetes_nomad_elasticsearch_port: "9200"
```

### Multi-region Orchestration

Configure multi-region support for geographical distribution and failover.

```yaml
# Enable multi-region orchestration
kubernetes_nomad_multi_region_enabled: true

# Regions to use (can be a list of strings or objects with name/address)
kubernetes_nomad_regions:
  - "global"
  - name: "us-west"
    address: "nomad-us-west.example.com"
  - name: "eu-central"
    address: "nomad-eu.example.com"

# Failover strategy (options: failover, active-active)
kubernetes_nomad_multi_region_strategy: "failover"
kubernetes_nomad_multi_region_failover_timeout: "60s"
```

### GPU Workload Scheduling

Support for GPU workloads in both Kubernetes and Nomad.

```yaml
# Enable GPU scheduling support
kubernetes_nomad_gpu_enabled: true

# GPU vendor (nvidia, amd, intel)
kubernetes_nomad_gpu_vendor: "nvidia"

# Example GPU resource limits for the example job
kubernetes_nomad_example_gpu_limit: 1
kubernetes_nomad_example_gpu_memory: 4096
```

### Autoscaling Integration

Configure dynamic workload scaling based on metrics.

```yaml
# Enable autoscaling
kubernetes_nomad_autoscaling_enabled: true

# Scaling boundaries
kubernetes_nomad_autoscaling_min_replicas: 1
kubernetes_nomad_autoscaling_max_replicas: 10

# Scaling metric and target (CPU percentage)
kubernetes_nomad_autoscaling_metric: "cpu"
kubernetes_nomad_autoscaling_target_value: 70

# Scaling behavior
kubernetes_nomad_autoscaling_evaluation_interval: "30s"
kubernetes_nomad_autoscaling_cooldown: "2m"
kubernetes_nomad_autoscaling_scale_down_period: "3m"
```

### Custom Resource Definition Support

Enable support for Kubernetes Custom Resource Definitions in Nomad.

```yaml
# Enable CRD support
kubernetes_nomad_crd_enabled: true

# Enable custom resources by default
kubernetes_nomad_custom_resources: true

# API groups to enable for custom resources
kubernetes_nomad_custom_resource_groups:
  - "monitoring.coreos.com"
  - "cert-manager.io"
  - "storage.k8s.io"
```

### Disaster Recovery with Automatic Failover

Configure automatic failover and disaster recovery.

```yaml
# Enable disaster recovery
kubernetes_nomad_disaster_recovery_enabled: true

# Recovery settings
kubernetes_nomad_dr_threshold: "30s"
kubernetes_nomad_dr_heartbeat: "60s"
kubernetes_nomad_dr_detection: "health_check"
kubernetes_nomad_dr_failback_delay: "5m"

# Snapshot settings for state recovery
kubernetes_nomad_dr_snapshot_path: "/var/lib/nomad/snapshots"
kubernetes_nomad_dr_snapshot_interval: "1h"

# Job-specific settings
kubernetes_nomad_dr_auto_revert: true
kubernetes_nomad_dr_max_disconnection: "10m"
```

### Service Mesh and Zero-downtime Migration

Configure service mesh integration for zero-downtime migrations.

```yaml
# Enable service mesh integration
kubernetes_nomad_service_mesh_enabled: true

# Service mesh settings
kubernetes_nomad_connect_sidecar_image: "consul:latest"
kubernetes_nomad_service_mesh_protocol: "http"
kubernetes_nomad_service_mesh_metrics: true
kubernetes_nomad_service_mesh_metrics_port: 20200
```

### Unified Policy Enforcement with OPA

Configure Open Policy Agent integration for unified policy enforcement.

```yaml
# Enable OPA policy enforcement
kubernetes_nomad_opa_enabled: true

# OPA server settings
kubernetes_nomad_opa_url: "http://opa.service.consul:8181/v1/data"
kubernetes_nomad_opa_token: ""

# Policy behavior
kubernetes_nomad_opa_allow_override: false
kubernetes_nomad_opa_default_action: "deny"

# Policy paths to evaluate
kubernetes_nomad_opa_paths:
  - "kubernetes/admission"
  - "nomad/task"

# OPA image for sidecar
kubernetes_nomad_opa_image: "openpolicyagent/opa:latest"
```

## Usage Examples

### Cross-platform Service Discovery Example

```hcl
# Nomad job with service discovery from Kubernetes
job "frontend" {
  group "app" {
    service {
      name = "frontend-app"
      port = "http"
      tags = ["mantl", "frontend", "mantl-service=true"]
    }
    
    task "server" {
      driver = "exec"
      
      template {
        data = <<EOT
UPSTREAM_SERVICE={{ range service "backend.kubernetes" }}
  {{ .Address }}:{{ .Port }}{{ end }}
EOT
        destination = "local/service.env"
        env = true
      }
    }
  }
}
```

### Vault Secrets Integration Example

```hcl
task "app" {
  driver = "kubernetes"
  
  vault {
    policies = ["app-policy"]
    kubernetes_auth = {
      role = "nomad"
    }
  }
  
  template {
    data = <<EOT
{{ with secret "kv/data/app/config" }}
API_KEY={{ .Data.data.api_key }}
DB_PASSWORD={{ .Data.data.db_password }}
{{ end }}
EOT
    destination = "secrets/app.env"
    env = true
  }
}
```

### GPU Workload Example

```hcl
task "ml-training" {
  driver = "kubernetes"
  
  config {
    image = "tensorflow/tensorflow:latest-gpu"
    
    pod_spec = <<EOT
spec:
  containers:
  - name: tensorflow
    resources:
      limits:
        nvidia.com/gpu: 1
EOT
  }
  
  resources {
    device "nvidia/gpu" {
      count = 1
      constraints {
        attribute = "${device.attr.memory}"
        operator  = ">="
        value     = "4096"
      }
    }
  }
}
```

### Multi-region Deployment Example

```hcl
job "global-service" {
  multiregion {
    strategy {
      max_parallel = 1
      on_failure   = "fail_all"
    }
    
    region "us-west" {
      count = 3
      datacenters = ["us-west-1", "us-west-2"]
    }
    
    region "eu-central" {
      count = 2
      datacenters = ["eu-central-1"]
    }
  }
  
  group "api" {
    count = 3
    
    task "server" {
      driver = "kubernetes"
      // ...
    }
  }
}
```

## Monitoring and Troubleshooting

### Logs

- Kubernetes audit logs: `/var/log/kubernetes/audit.log`
- Nomad audit logs: `/var/log/nomad/kubernetes-audit.json`
- Nomad server logs: `/var/log/nomad/nomad.log`
- Integration logs: `/var/log/nomad/nomad-k8s-integration.log`

### Common Issues

1. **Authentication Failures**
   - Check that the service account token is properly configured
   - Verify RBAC permissions
   - Ensure Vault authentication is configured correctly

2. **Service Discovery Issues**
   - Verify that the correct labels are applied to services
   - Check that sync periods are appropriate for your use case
   - Ensure network connectivity between Nomad and Kubernetes

3. **Resource Limits**
   - Ensure resource requests don't exceed cluster capacity
   - Check for namespace resource quotas
   - Verify GPU availability and driver installation

4. **Multi-region Issues**
   - Ensure all regions are reachable
   - Check that region addresses are correctly configured
   - Verify ACL tokens have appropriate permissions in all regions

5. **Policy Enforcement Failures**
   - Verify OPA is running and accessible
   - Check that policies are correctly defined
   - Review OPA logs for evaluation issues

## Testing

The integration is tested using Molecule tests:

```bash
cd roles/kubernetes
molecule test
```

Dedicated test suites exist for each feature:

- Service discovery tests
- Vault integration tests
- Multi-region tests
- GPU scheduling tests
- Autoscaling tests
- CRD support tests
- Disaster recovery tests
- OPA integration tests
- Metrics and logging tests

## Security Considerations

- The integration uses dedicated service accounts with limited privileges
- All communication is secured with TLS
- Token-based authentication is used between systems
- Audit logging is enabled for compliance tracking
- Common Criteria features ensure secure operation in regulated environments
- OPA policies provide unified enforcement of security standards
- Vault integration ensures secure secrets management