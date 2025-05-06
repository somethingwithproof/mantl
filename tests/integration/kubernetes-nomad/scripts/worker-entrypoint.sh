#!/bin/bash
set -e

echo "Starting worker node services..."

# Create directories if they don't exist
mkdir -p /etc/kubernetes
mkdir -p /etc/kubernetes/nomad-integration
mkdir -p /var/lib/kubernetes/nomad-integration
mkdir -p /var/log/kubernetes
mkdir -p /var/log/nomad
mkdir -p /etc/nomad.d/client
mkdir -p /etc/consul.d
mkdir -p /var/lib/nomad/snapshots

# Get the control node IP
CONTROL_NODE_IP=$(getent hosts $CONTROL_NODE | awk '{ print $1 }')
if [ -z "$CONTROL_NODE_IP" ]; then
    echo "Error: Cannot resolve control node IP"
    exit 1
fi

# Start Consul client
echo "Starting Consul client..."
if [ -f /etc/mantl/config/consul-client.hcl ]; then
  cp /etc/mantl/config/consul-client.hcl /etc/consul.d/consul.hcl
else
  cat > /etc/consul.d/consul.hcl <<EOF
data_dir = "/var/lib/consul"
log_level = "INFO"
server = false
bind_addr = "0.0.0.0"
client_addr = "0.0.0.0"
advertise_addr = "$(hostname -I | awk '{print $1}')"
retry_join = ["$CONTROL_NODE_IP"]
EOF
fi

nohup consul agent -config-file=/etc/consul.d/consul.hcl > /var/log/consul.log 2>&1 &
sleep 5
consul members

# Join Kubernetes cluster
echo "Joining Kubernetes cluster..."
mkdir -p /var/lib/kubelet
# For testing, we'll use kubeadm join with pre-shared token
echo "Waiting for control node to be ready..."
sleep 30
kubeadm join ${CONTROL_NODE_IP}:6443 --token abcdef.0123456789abcdef --discovery-token-unsafe-skip-ca-verification

# Get credentials for kubectl
mkdir -p $HOME/.kube
# For testing, copy the kubeconfig file from a shared volume
cp /etc/mantl/config/kubeconfig $HOME/.kube/config 2>/dev/null || echo "Kubeconfig not available, continuing without it"

# Start Nomad client
echo "Starting Nomad client..."
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
EOF

# Client-specific config
cat > /etc/nomad.d/client/client.hcl <<EOF
client {
  enabled = true
  servers = ["$CONTROL_NODE_IP:4646"]
  
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

# Get Nomad-Kubernetes integration files
mkdir -p /etc/kubernetes/nomad-integration
scp -o StrictHostKeyChecking=no $CONTROL_NODE:/etc/kubernetes/nomad-integration/* /etc/kubernetes/nomad-integration/

# Copy Kubernetes integration config from control node
scp -o StrictHostKeyChecking=no $CONTROL_NODE:/etc/nomad.d/kubernetes-integration.hcl /etc/nomad.d/

# Start Nomad client
nohup nomad agent -config=/etc/nomad.d/base.hcl -config=/etc/nomad.d/client/client.hcl -config=/etc/nomad.d/kubernetes-integration.hcl > /var/log/nomad/nomad.log 2>&1 &

echo "Worker node setup complete"

# Keep the container running
exec "$@"