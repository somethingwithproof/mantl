#!/bin/bash
set -e

echo "Starting control node services..."

# Create directories if they don't exist
mkdir -p /etc/kubernetes/manifests
mkdir -p /etc/kubernetes/pki
mkdir -p /etc/kubernetes/nomad-integration
mkdir -p /var/lib/kubernetes/nomad-integration
mkdir -p /var/log/kubernetes
mkdir -p /var/log/nomad
mkdir -p /etc/nomad.d/server
mkdir -p /etc/nomad.d/client
mkdir -p /etc/vault.d
mkdir -p /etc/consul.d

# Start Consul
echo "Starting Consul server..."
if [ -f /etc/mantl/config/consul-server.hcl ]; then
  cp /etc/mantl/config/consul-server.hcl /etc/consul.d/consul.hcl
else
  cat > /etc/consul.d/consul.hcl <<EOF
data_dir = "/var/lib/consul"
log_level = "INFO"
server = true
bootstrap_expect = 1
bind_addr = "0.0.0.0"
client_addr = "0.0.0.0"
advertise_addr = "$(hostname -I | awk '{print $1}')"
ui = true
EOF
fi

nohup consul agent -config-file=/etc/consul.d/consul.hcl > /var/log/consul.log 2>&1 &
sleep 5
consul members

# Start Vault
echo "Starting Vault server..."
if [ -f /etc/mantl/config/vault.hcl ]; then
  cp /etc/mantl/config/vault.hcl /etc/vault.d/vault.hcl
else
  cat > /etc/vault.d/vault.hcl <<EOF
storage "file" {
  path = "/var/lib/vault"
}

listener "tcp" {
  address = "0.0.0.0:8200"
  tls_disable = true
}

ui = true
EOF
fi

nohup vault server -config=/etc/vault.d/vault.hcl > /var/log/vault.log 2>&1 &
sleep 5

# Initialize and unseal Vault
export VAULT_ADDR=http://127.0.0.1:8200
vault operator init -key-shares=1 -key-threshold=1 > /var/lib/vault/init.txt
VAULT_UNSEAL_KEY=$(grep "Unseal Key 1" /var/lib/vault/init.txt | awk '{print $NF}')
VAULT_ROOT_TOKEN=$(grep "Initial Root Token" /var/lib/vault/init.txt | awk '{print $NF}')
echo "Unsealing Vault..."
vault operator unseal $VAULT_UNSEAL_KEY
echo "Vault Root Token: $VAULT_ROOT_TOKEN"
export VAULT_TOKEN=$VAULT_ROOT_TOKEN

# Configure Vault for the Kubernetes integration
echo "Configuring Vault for Kubernetes integration..."
vault auth enable kubernetes
vault secrets enable -path=kv kv-v2
vault policy write nomad-server - <<EOF
path "kv/data/mantl/*" {
  capabilities = ["create", "update", "read", "delete", "list"]
}
EOF

# Save token for Nomad
vault token create -policy=nomad-server -period=72h > /etc/nomad.d/vault-token
NOMAD_VAULT_TOKEN=$(grep "token " /etc/nomad.d/vault-token | awk '{print $2}')
echo $NOMAD_VAULT_TOKEN > /etc/nomad.d/vault-token-value

# Create example secrets for the tests
vault kv put kv/mantl/kubernetes-example api_key="test-api-key" db_password="test-db-password"

# Start kubeadm
echo "Initializing Kubernetes control plane..."
kubeadm init --pod-network-cidr=10.244.0.0/16 --apiserver-advertise-address=$(hostname -I | awk '{print $1}') --apiserver-cert-extra-sans=localhost,control,172.20.0.10

# Configure kubectl
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config

# Make kubeconfig available to other containers
cp /etc/kubernetes/admin.conf /etc/mantl/config/kubeconfig
chmod 644 /etc/mantl/config/kubeconfig

# Create token for worker nodes to join
kubeadm token create abcdef.0123456789abcdef --ttl 2h

# Install Calico CNI
kubectl apply -f https://docs.projectcalico.org/v3.25/manifests/calico.yaml

# Allow scheduling on control node for testing
kubectl taint nodes --all node-role.kubernetes.io/control-plane-

# Create serviceaccount for Nomad
echo "Creating Kubernetes service account for Nomad..."
kubectl create serviceaccount nomad-integration -n kube-system

# Create ClusterRole for Nomad
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: nomad-integration
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets", "daemonsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses", "networkpolicies"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses", "persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["monitoring.coreos.com"]
  resources: ["servicemonitors", "prometheusrules"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["cert-manager.io"]
  resources: ["certificates", "issuers"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
EOF

# Create ClusterRoleBinding for Nomad
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: nomad-integration
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: nomad-integration
subjects:
- kind: ServiceAccount
  name: nomad-integration
  namespace: kube-system
EOF

# Create Kubernetes integration files for Nomad
echo "Creating Kubernetes integration files for Nomad..."
mkdir -p /etc/kubernetes/nomad-integration
kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' > /etc/kubernetes/nomad-integration/k8s-api-server
kubectl config view --minify --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 --decode > /etc/kubernetes/nomad-integration/k8s-ca.crt
kubectl -n kube-system get secret $(kubectl -n kube-system get serviceaccount nomad-integration -o jsonpath='{.secrets[0].name}') -o jsonpath='{.data.token}' | base64 --decode > /etc/kubernetes/nomad-integration/nomad-token

# Start Nomad server
echo "Starting Nomad server..."
# Base Nomad configuration
cat > /etc/nomad.d/base.hcl <<EOF
data_dir  = "/var/lib/nomad"
bind_addr = "0.0.0.0"
log_level = "INFO"

# Enable debug log levels in development
log {
  level = "INFO"
  file  = "/var/log/nomad/nomad.log"
}

# Enable Common Criteria audit logging
audit {
  enabled = true
  sink "file" {
    type               = "file"
    format             = "json"
    path               = "/var/log/nomad/audit.json"
    delivery_guarantee = "enforced"
    rotate_bytes       = 10485760
    rotate_duration    = "24h"
    rotate_max_files   = 10
  }
}

# Telemetry for metrics
telemetry {
  prometheus_metrics = true
  publish_allocation_metrics = true
  publish_node_metrics = true
  disable_hostname = true
  collection_interval = "10s"
}

# Consul integration
consul {
  address = "127.0.0.1:8500"
}

# Vault integration
vault {
  enabled = true
  address = "http://127.0.0.1:8200"
  token = "$(cat /etc/nomad.d/vault-token-value)"
  create_from_role = "nomad-server"
}
EOF

# Server-specific config
cat > /etc/nomad.d/server/server.hcl <<EOF
server {
  enabled = true
  bootstrap_expect = 1
  
  # Enterprise features mock
  license_path = "/etc/nomad.d/server/license.hcl"
  
  # Common Criteria features
  encrypt = "AbC3Def56Hijk7mnO8pqrStUv9WxyZ"
  
  # Set default CPU/RAM for jobs
  default_scheduler_config {
    memory_oversubscription_enabled = true
    preemption_config {
      system_scheduler_enabled = true
      batch_scheduler_enabled = true
      service_scheduler_enabled = true
    }
  }
}
EOF

# Client-specific config (for running jobs locally on control node)
cat > /etc/nomad.d/client/client.hcl <<EOF
client {
  enabled = true
  servers = ["127.0.0.1:4646"]
  
  # Enable Common Criteria security features
  cni_path = "/opt/cni/bin"
  
  # Reserved resources for system
  reserved {
    cpu = 500
    memory = 512
    disk = 1024
  }

  # Mock GPU device for testing
  host_volume "gpu" {
    path = "/usr/local/nvidia"
    read_only = false
  }
  
  meta {
    "kubernetes_enabled" = "true"
  }
}
EOF

# Generate Kubernetes integration config
cat > /etc/nomad.d/kubernetes-integration.hcl <<EOF
# Cross-platform service discovery
service_discovery "kubernetes" {
  enabled = true
  server_address = "$(cat /etc/kubernetes/nomad-integration/k8s-api-server)"
  token = "$(cat /etc/kubernetes/nomad-integration/nomad-token)"
  ca_file = "/etc/kubernetes/nomad-integration/k8s-ca.crt"
  namespace = "default"
  label_selector = "mantl-service=true"
  sync_period = "30s"
  service_defaults {
    external = false
    check_interval = "10s"
  }
}

# Kubernetes plugin configuration
plugin "kubernetes" {
  config {
    host = "$(cat /etc/kubernetes/nomad-integration/k8s-api-server)"
    service_account_token = "$(cat /etc/kubernetes/nomad-integration/nomad-token)"
    ca_file = "/etc/kubernetes/nomad-integration/k8s-ca.crt"
    namespace = "default"
    
    # Security settings
    tls_verify = true
    
    # Resource management settings
    cleanup_failed = true
    cleanup_deadline = "5m"
    
    # Resource limits
    memory_limit = 256
    cpu_limit = 200
    
    # GPU scheduling support
    gpu_support = true
    gpu_vendor = "nvidia"
    
    # Service mesh integration
    service_mesh {
      enabled = true
      connect_sidecar_image = "consul:latest"
      default_protocol = "http"
      metrics_enabled = true
      metrics_port = 20200
    }
    
    # CRD support
    custom_resources = true
    custom_resource_groups = [
      "monitoring.coreos.com",
      "cert-manager.io"
    ]
    
    # Autoscaling integration
    autoscaling {
      enabled = true
      min_replicas = 1
      max_replicas = 10
      metric = "cpu"
      target_value = 70
      scale_down_stabilization = "3m"
    }
  }
}

# Multi-region orchestration
multi_region {
  enabled = true
  regions = [
    {
      name = "global"
    }
  ]
  strategy = "failover"
  failover_timeout = "60s"
}

# Unified policy enforcement with OPA
policy {
  enabled = true
  opa_url = "http://127.0.0.1:8181/v1/data"
  allow_policy_override = false
  default_action = "deny"
  evaluation_paths = [
    "kubernetes/admission"
  ]
}

# Disaster recovery with automatic failover
disaster_recovery {
  enabled = true
  recovery_threshold = "30s"
  max_heartbeat_interval = "60s"
  failure_detection = "health_check"
  failback_delay = "5m"
  snapshot_path = "/var/lib/nomad/snapshots"
  snapshot_interval = "1h"
}
EOF

# Create example Nomad job for Kubernetes
cat > /var/lib/kubernetes/nomad-integration/kubernetes-example.nomad <<EOF
job "mantl-kubernetes-example" {
  datacenters = ["dc1"]
  
  # Multi-region orchestration
  multiregion {
    strategy {
      max_parallel = 1
      on_failure   = "fail_all"
    }
    region "global" {
      count = 1
      datacenters = ["dc1"]
    }
  }

  # Disaster recovery configuration
  disaster_recovery {
    targeted_evaluation = true
    auto_revert = true
    max_disconnection = "10m"
  }
  
  type = "service"

  constraint {
    attribute = "\${meta.kubernetes_enabled}"
    operator  = "="
    value     = "true"
  }
  
  meta {
    managed_by          = "mantl"
    kubernetes_enabled  = "true"
    version             = "latest"
    service_type        = "web"
    tier                = "frontend"
  }

  group "kubernetes-nginx" {
    count = 1

    network {
      port "http" {
        to = 80
      }
      
      port "metrics" {
        to = 9090
      }
    }

    service {
      name = "mantl-example-service"
      port = "http"
      tags = ["mantl", "example", "nginx", "mantl-service=true"]
      
      check {
        type     = "http"
        path     = "/"
        interval = "10s"
        timeout  = "2s"
      }
      
      meta {
        version = "latest"
        environment = "test"
        service_type = "web"
      }
    }
    
    # Example of a federated metrics service
    service {
      name = "mantl-example-metrics"
      port = "metrics"
      tags = ["mantl", "metrics", "prometheus"]
      
      check {
        type     = "http"
        path     = "/metrics"
        interval = "30s"
        timeout  = "5s"
      }
    }

    task "nginx" {
      driver = "kubernetes"

      # Vault integration for secure secrets
      vault {
        policies = ["nomad-server"]
        
        # Kubernetes auth method
        kubernetes_auth = {
          role = "nomad"
        }
      }

      # Kubernetes driver configuration
      config {
        image = "nginx:stable-alpine"
        
        # Pod configuration
        pod_spec = <<EOT
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: mantl-example
    managed-by: nomad
    mantl-service: "true"
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9090"
spec:
  containers:
  - name: nginx
    resources:
      limits:
        cpu: "200m"
        memory: "256Mi"
        nvidia.com/gpu: "1"
      requests:
        cpu: "100m"
        memory: "128Mi"
    securityContext:
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      runAsNonRoot: true
      seccompProfile:
        type: RuntimeDefault
    ports:
    - containerPort: 80
      name: http
    - containerPort: 9090
      name: metrics
    livenessProbe:
      httpGet:
        path: /
        port: 80
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /
        port: 80
      initialDelaySeconds: 5
      periodSeconds: 5
    env:
    - name: VAULT_ADDR
      value: "http://vault.service.consul:8200"
    volumeMounts:
    - name: vault-token
      mountPath: "/var/run/secrets/vault"
      readOnly: true
  volumes:
  - name: vault-token
    projected:
      sources:
      - serviceAccountToken:
          path: token
          expirationSeconds: 7200
          audience: vault
  serviceAccountName: nomad-k8s-sa
EOT

        # Example of custom resource definition usage
        custom_resources = [
          {
            apiVersion = "monitoring.coreos.com/v1"
            kind = "ServiceMonitor"
            metadata = {
              name = "mantl-example"
              labels = {
                app = "mantl-example"
              }
            }
            spec = {
              selector = {
                matchLabels = {
                  app = "mantl-example"
                }
              }
              endpoints = [
                {
                  port = "metrics"
                  interval = "30s"
                }
              ]
            }
          }
        ]
      }

      # Resource constraints
      resources {
        cpu    = 200
        memory = 256
        device "nvidia/gpu" {
          count = 1
          constraints {
            attribute = "\${device.attr.memory}"
            operator  = ">="
            value     = "1024"
          }
        }
      }

      # Example templates
      template {
        data = <<EOT
<!DOCTYPE html>
<html>
<head>
  <title>Mantl - Nomad and Kubernetes Integration</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
    h1 { color: #333; }
    .info { background-color: #f4f4f4; padding: 20px; border-radius: 5px; }
    .features { margin-top: 20px; }
    .feature { margin-bottom: 10px; padding: 10px; background-color: #e9f7fe; border-radius: 3px; }
    .enabled { color: green; font-weight: bold; }
    .disabled { color: gray; }
  </style>
</head>
<body>
  <h1>Mantl Example - Nomad and Kubernetes Integration</h1>
  <div class="info">
    <p>This page demonstrates successful integration between Nomad and Kubernetes in Mantl.</p>
    <p>Integration Features:</p>
    <ul>
      <li>Cross-platform Service Discovery</li>
      <li>Shared Vault Secrets Management</li>
      <li>Federated Metrics and Logging</li>
      <li>Multi-region Orchestration</li>
      <li>GPU Workload Scheduling</li>
      <li>Autoscaling Integration</li>
      <li>Custom Resource Definition Support</li>
      <li>Disaster Recovery with Automatic Failover</li>
      <li>Zero-downtime Migration Tools</li>
      <li>Unified Policy Enforcement with OPA</li>
    </ul>
  </div>
</body>
</html>
EOT
        destination = "local/index.html"
      }
      
      # Vault secrets integration
      template {
        data = <<EOT
{{ with secret "kv/data/mantl/kubernetes-example" }}
# This file contains sensitive information retrieved from Vault
# Do not expose or share this information

EXAMPLE_API_KEY={{ .Data.data.api_key }}
EXAMPLE_DB_PASSWORD={{ .Data.data.db_password }}
{{ end }}
EOT
        destination = "secrets/credentials.env"
        env = true
        change_mode = "restart"
      }
    }
    
    task "logging-sidecar" {
      driver = "kubernetes"
      
      config {
        image = "fluent/fluentd:v1.14"
        
        pod_spec = <<EOT
spec:
  containers:
  - name: fluentd
    resources:
      limits:
        cpu: "100m"
        memory: "128Mi"
      requests:
        cpu: "50m"
        memory: "64Mi"
    volumeMounts:
    - name: log-volume
      mountPath: /logs
  volumes:
  - name: log-volume
    emptyDir: {}
EOT
      }
      
      resources {
        cpu    = 100
        memory = 128
      }
    }
    
    task "policy-agent" {
      driver = "kubernetes"
      
      config {
        image = "openpolicyagent/opa:latest"
        
        args = [
          "run",
          "--server",
          "--addr=:8181",
          "--log-level=info",
          "/policies"
        ]
        
        pod_spec = <<EOT
spec:
  containers:
  - name: opa
    ports:
    - containerPort: 8181
    volumeMounts:
    - name: policies
      mountPath: /policies
  volumes:
  - name: policies
    configMap:
      name: opa-policies
EOT
      }
      
      resources {
        cpu    = 100
        memory = 128
      }
      
      template {
        data = <<EOT
package kubernetes.admission

default allow = false

# Allow only if deployment has proper security context
allow {
  input.request.kind.kind == "Pod"
  securityContext := input.request.object.spec.securityContext
  securityContext.runAsNonRoot == true
  securityContext.allowPrivilegeEscalation == false
}

# Always allow kube-system namespaces
allow {
  input.request.namespace == "kube-system"
}

# Allow Nomad-managed resources
allow {
  input.request.object.metadata.labels["managed-by"] == "nomad"
}
EOT
        destination = "local/policies/kubernetes.rego"
      }
    }
  }
}
EOF

# Start Nomad with all configs
nohup nomad agent -config=/etc/nomad.d/base.hcl -config=/etc/nomad.d/server/server.hcl -config=/etc/nomad.d/client/client.hcl -config=/etc/nomad.d/kubernetes-integration.hcl > /var/log/nomad/nomad.log 2>&1 &
sleep 10
nomad server members

# Start OPA for policy enforcement
nohup opa run --server --addr=:8181 > /var/log/opa.log 2>&1 &

# Wait for Kubernetes to be ready
echo "Waiting for Kubernetes to be ready..."
until kubectl get nodes | grep -q Ready; do
  sleep 5
done

echo "Waiting for Nomad to be ready..."
until nomad server members | grep -q alive; do
  sleep 5
done

# Deploy the example job to test integration
nomad job run /var/lib/kubernetes/nomad-integration/kubernetes-example.nomad

echo "Control node setup complete"

# Keep the container running
exec "$@"