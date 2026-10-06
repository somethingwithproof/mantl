# SPDX-License-Identifier: Apache-2.0
# HashiCorp Nomad Server Configuration
# Production-ready configuration for Mantl platform
# https://developer.hashicorp.com/nomad/docs/configuration

# Datacenter and region configuration
datacenter = "dc1"
region     = "global"
name       = "nomad-server-${node_number}"

# Data directory
data_dir = "/opt/nomad/data"

# Bind address - all interfaces
bind_addr = "0.0.0.0"

# Advertise address configuration
advertise {
  http = "{{ GetPrivateIP }}:4646"
  rpc  = "{{ GetPrivateIP }}:4647"
  serf = "{{ GetPrivateIP }}:4648"
}

# Log configuration
log_level = "INFO"
log_json  = true
log_file  = "/var/log/nomad/nomad.log"

log_rotate_bytes    = 104857600  # 100MB
log_rotate_duration = "24h"
log_rotate_max_files = 7

# Enable server mode
server {
  enabled = true

  # Bootstrap configuration
  bootstrap_expect = 3

  # Server join configuration for auto-clustering
  server_join {
    retry_join = [
      "provider=aws tag_key=nomad-role tag_value=server",
      "nomad-server-1.mantl.local:4648",
      "nomad-server-2.mantl.local:4648",
      "nomad-server-3.mantl.local:4648"
    ]
    retry_max      = 5
    retry_interval = "15s"
  }

  # Raft configuration
  raft_protocol = 3

  # Encrypt gossip traffic
  encrypt = "{{ nomad_gossip_key }}"

  # Redundancy zone for multi-AZ deployments
  redundancy_zone = "{{ availability_zone }}"

  # Upgrade migrations
  upgrade_version = "1.7.0"

  # Licensing (for Enterprise features)
  license_path = "/etc/nomad.d/license.hclic"

  # Job garbage collection
  job_gc_interval   = "5m"
  job_gc_threshold  = "4h"

  # Evaluation garbage collection
  eval_gc_threshold = "1h"

  # Node garbage collection
  node_gc_threshold = "24h"

  # Deployment garbage collection
  deployment_gc_threshold = "1h"

  # CSI plugin garbage collection
  csi_plugin_gc_threshold = "1h"

  # Root key garbage collection (for variables)
  root_key_gc_interval  = "10m"
  root_key_gc_threshold = "1h"

  # Heartbeat grace period
  heartbeat_grace = "30s"
  min_heartbeat_ttl = "10s"

  # Maximum heartbeat TTL
  max_heartbeat_ttl = "120s"

  # Failover heartbeat TTL
  failover_heartbeat_ttl = "300s"

  # Search configuration
  search {
    fuzzy_enabled   = true
    limit_query     = 200
    limit_results   = 1000
    min_term_length = 2
  }

  # Default scheduler configuration
  default_scheduler_config {
    scheduler_algorithm = "binpack"

    preemption_config {
      batch_scheduler_enabled   = true
      system_scheduler_enabled  = true
      service_scheduler_enabled = true
      sysbatch_scheduler_enabled = true
    }

    memory_oversubscription_enabled = true
  }

  # Plan rejection tracker
  plan_rejection_tracker {
    enabled        = true
    node_threshold = 100
    node_window    = "5m"
  }
}

# ACL configuration
acl {
  enabled    = true
  token_ttl  = "30s"
  policy_ttl = "60s"

  # Role/token GC
  token_gc_threshold = "1h"
  role_gc_threshold  = "1h"
}

# TLS configuration
tls {
  http      = true
  rpc       = true
  ca_file   = "/etc/nomad.d/ca.pem"
  cert_file = "/etc/nomad.d/server.pem"
  key_file  = "/etc/nomad.d/server-key.pem"

  verify_server_hostname = true
  verify_https_client    = true

  # Minimum TLS version
  tls_min_version = "tls12"

  # Cipher suites
  tls_cipher_suites = "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
}

# Vault integration
vault {
  enabled          = true
  address          = "https://vault.mantl.local:8200"
  create_from_role = "nomad-cluster"
  token            = ""  # Token is injected via environment

  ca_file   = "/etc/nomad.d/vault-ca.pem"
  cert_file = "/etc/nomad.d/vault-client.pem"
  key_file  = "/etc/nomad.d/vault-client-key.pem"

  # Token renewal
  allow_unauthenticated = false

  # Namespace support (Enterprise)
  namespace = "nomad"

  # Default identity for Workload Identity
  default_identity {
    aud  = ["vault.io"]
    env  = true
    file = true
    ttl  = "1h"
  }
}

# Consul integration for service discovery
consul {
  address = "127.0.0.1:8501"
  ssl     = true

  ca_file   = "/etc/nomad.d/consul-ca.pem"
  cert_file = "/etc/nomad.d/consul-client.pem"
  key_file  = "/etc/nomad.d/consul-client-key.pem"

  # Service registration
  auto_advertise       = true
  server_auto_join     = true
  client_auto_join     = true

  # Service name
  server_service_name = "nomad-servers"
  client_service_name = "nomad-clients"

  # ACL token
  token = "{{ consul_token }}"

  # Checks
  checks_use_advertise = true

  # Service identity for Consul Connect
  service_identity {
    aud = ["consul.io"]
    ttl = "1h"
  }

  # Task identity
  task_identity {
    aud = ["consul.io"]
    ttl = "1h"
  }
}

# Telemetry configuration
telemetry {
  # Prometheus metrics
  prometheus_metrics = true

  # StatsD
  statsd_address = "127.0.0.1:8125"

  # DataDog
  datadog_address = "127.0.0.1:8125"
  datadog_tags    = ["env:production", "service:nomad-server"]

  # Metrics collection
  collection_interval = "10s"

  # Publish allocation metrics
  publish_allocation_metrics = true
  publish_node_metrics       = true

  # Disable hostname
  disable_hostname = false

  # Filter default metrics (performance)
  filter_default = true
}

# Audit logging (Enterprise)
audit {
  enabled = true

  sink "file" {
    type               = "file"
    delivery_guarantee = "enforced"
    format             = "json"
    path               = "/var/log/nomad/audit.log"
    rotate_bytes       = 104857600
    rotate_duration    = "24h"
    rotate_max_files   = 30
  }

  filter "default" {
    type       = "HTTPEvent"
    endpoints  = ["*"]
    operations = ["*"]
    stages     = ["*"]
  }
}

# Sentinel (Enterprise - Policy as Code)
sentinel {
  import "http" {
    path = "http"
  }
  import "strings" {
    path = "strings"
  }
}

# UI configuration
ui {
  enabled = true

  # Consul integration in UI
  consul {
    ui_url = "https://consul.mantl.local:8501/ui"
  }

  # Vault integration in UI
  vault {
    ui_url = "https://vault.mantl.local:8200/ui"
  }

  # Custom label
  label {
    text             = "Mantl Platform"
    background_color = "#1a73e8"
    text_color       = "#ffffff"
  }

  # Content security policy
  content_security_policy {
    connect_src     = ["*"]
    default_src     = ["'none'"]
    form_action     = ["'none'"]
    frame_ancestors = ["'none'"]
    img_src         = ["'self'", "data:"]
    script_src      = ["'self'"]
    style_src       = ["'self'", "'unsafe-inline'"]
  }
}

# Limits configuration
limits {
  https_handshake_timeout   = "30s"
  http_max_conns_per_client = 200
  rpc_handshake_timeout     = "30s"
  rpc_max_conns_per_client  = 100
}

# Plugin configuration
plugin_dir = "/opt/nomad/plugins"

plugin "raw_exec" {
  config {
    enabled = false  # Disabled for security
  }
}

plugin "docker" {
  config {
    # Docker endpoint
    endpoint = "unix:///var/run/docker.sock"

    # Extra labels for Docker containers
    extra_labels = ["job_name", "job_id", "task_group_name", "task_name", "namespace", "node_name"]

    # GC configuration
    gc {
      image       = true
      image_delay = "5m"
      container   = true

      dangling_containers {
        enabled        = true
        dry_run        = false
        period         = "5m"
        creation_grace = "5m"
      }
    }

    # Volume configuration
    volumes {
      enabled      = true
      selinuxlabel = "z"
    }

    # Allow privileged containers (controlled via job policy)
    allow_privileged = false

    # Allow container capabilities
    allow_caps = [
      "audit_write",
      "chown",
      "dac_override",
      "fowner",
      "fsetid",
      "kill",
      "mknod",
      "net_bind_service",
      "setfcap",
      "setgid",
      "setpcap",
      "setuid",
      "sys_chroot"
    ]

    # NVIDIA GPU support
    nvidia_runtime = "nvidia"

    # Infra image
    infra_image = "gcr.io/google_containers/pause:3.9"
  }
}

# Autopilot configuration (Enterprise)
autopilot {
  cleanup_dead_servers      = true
  last_contact_threshold    = "200ms"
  max_trailing_logs         = 250
  min_quorum                = 3
  server_stabilization_time = "10s"

  # Enterprise features
  enable_redundancy_zones   = true
  disable_upgrade_migration = false
  enable_custom_upgrades    = true
}
