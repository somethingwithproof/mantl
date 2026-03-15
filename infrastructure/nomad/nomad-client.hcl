# HashiCorp Nomad Client Configuration
# Worker node configuration for Mantl platform
# https://developer.hashicorp.com/nomad/docs/configuration

# Datacenter and region
datacenter = "dc1"
region     = "global"
name       = "nomad-client-{{ node_id }}"

# Data directory
data_dir = "/opt/nomad/data"

# Bind address
bind_addr = "0.0.0.0"

# Advertise address
advertise {
  http = "{{ GetPrivateIP }}:4646"
  rpc  = "{{ GetPrivateIP }}:4647"
  serf = "{{ GetPrivateIP }}:4648"
}

# Logging
log_level = "INFO"
log_json  = true
log_file  = "/var/log/nomad/nomad.log"

log_rotate_bytes     = 104857600
log_rotate_duration  = "24h"
log_rotate_max_files = 7

# Enable client mode
client {
  enabled = true

  # Server discovery
  servers = [
    "nomad-server-1.mantl.local:4647",
    "nomad-server-2.mantl.local:4647",
    "nomad-server-3.mantl.local:4647"
  ]

  # Auto-join via Consul
  server_join {
    retry_join = [
      "provider=aws tag_key=nomad-role tag_value=server"
    ]
    retry_max      = 10
    retry_interval = "15s"
  }

  # Node class for scheduling
  node_class = "{{ node_class | default('general') }}"

  # Node pool (Nomad 1.6+)
  node_pool = "{{ node_pool | default('default') }}"

  # Network interface
  network_interface = "eth0"

  # Network speed (for scheduling)
  network_speed = 10000  # 10Gbps

  # CPU compute class
  cpu_total_compute = 0  # Auto-detect

  # Memory total (0 = auto-detect)
  memory_total_mb = 0

  # Disk total (0 = auto-detect)
  disk_total_mb = 0

  # Reserved resources for system
  reserved {
    cpu            = 500   # 500 MHz
    memory         = 1024  # 1GB
    disk           = 10240 # 10GB
    reserved_ports = "22,8300-8302,8500-8502,8600"
  }

  # Node metadata for constraints
  meta {
    "node.class"           = "{{ node_class | default('general') }}"
    "node.pool"            = "{{ node_pool | default('default') }}"
    "instance_type"        = "{{ instance_type }}"
    "availability_zone"    = "{{ availability_zone }}"
    "platform"             = "mantl"
    "os"                   = "{{ node_os | default('linux') }}"
    "arch"                 = "{{ node_arch | default('amd64') }}"
    "gpu"                  = "{{ has_gpu | default('false') }}"
    "gpu_type"             = "{{ gpu_type | default('none') }}"
    "spot"                 = "{{ is_spot | default('false') }}"
  }

  # Host volumes
  host_volume "docker-sock" {
    path      = "/var/run/docker.sock"
    read_only = true
  }

  host_volume "certs" {
    path      = "/etc/ssl/certs"
    read_only = true
  }

  host_volume "logs" {
    path      = "/var/log/nomad-jobs"
    read_only = false
  }

  host_volume "data" {
    path      = "/opt/nomad/data/volumes"
    read_only = false
  }

  # Chroot environment for exec driver
  chroot_env {
    "/bin"            = "/bin"
    "/etc"            = "/etc"
    "/lib"            = "/lib"
    "/lib32"          = "/lib32"
    "/lib64"          = "/lib64"
    "/run/resolvconf" = "/run/resolvconf"
    "/sbin"           = "/sbin"
    "/usr"            = "/usr"
  }

  # Bridge network configuration
  bridge_network_name          = "nomad"
  bridge_network_subnet        = "172.26.64.0/20"
  bridge_network_hairpin_mode  = false

  # CNI configuration
  cni_path          = "/opt/cni/bin"
  cni_config_dir    = "/opt/cni/config"

  # Allocation directory
  alloc_dir = "/opt/nomad/data/alloc"

  # State directory
  state_dir = "/opt/nomad/data/client"

  # Garbage collection
  gc_interval            = "1m"
  gc_disk_usage_threshold = 80.0
  gc_inode_usage_threshold = 70.0
  gc_max_allocs          = 50
  gc_parallel_destroys   = 2

  # Minimum dynamic port
  min_dynamic_port = 20000
  max_dynamic_port = 32000

  # Template configuration
  template {
    # Disable file sandbox
    disable_file_sandbox = false

    # Function denylist
    function_denylist = ["plugin", "writeToFile"]

    # Vault retry
    vault_retry {
      attempts = 5
      backoff  = "250ms"
      max_backoff = "1m"
    }
  }

  # Host network configuration
  host_network "public" {
    cidr           = "0.0.0.0/0"
    reserved_ports = "22"
  }

  host_network "private" {
    interface      = "eth0"
    reserved_ports = "4646,4647,4648"
  }

  # Drain on shutdown
  drain_on_shutdown {
    deadline           = "5m"
    force              = true
    ignore_system_jobs = false
  }

  # Artifact configuration
  artifact {
    http_read_timeout = "30s"
    gcs_timeout       = "30s"
    git_timeout       = "30s"
    s3_timeout        = "30s"

    # Set decompression file count limit
    decompression_size_limit = "100GB"
    decompression_file_count_limit = 4096

    # Disable filesystem isolation
    disable_filesystem_isolation = false

    # Set source allowlist
    set_environment_variables = "FOO,BAR"
  }

  # Users configuration for rootless containers
  # users {
  #   dynamic_user {
  #     min = 80000
  #     max = 89999
  #   }
  # }
}

# ACL configuration
acl {
  enabled = true
}

# TLS configuration
tls {
  http      = true
  rpc       = true
  ca_file   = "/etc/nomad.d/ca.pem"
  cert_file = "/etc/nomad.d/client.pem"
  key_file  = "/etc/nomad.d/client-key.pem"

  verify_server_hostname = true
  verify_https_client    = true

  tls_min_version   = "tls12"
  tls_cipher_suites = "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
}

# Vault integration
vault {
  enabled = true
  address = "https://vault.mantl.local:8200"

  ca_file   = "/etc/nomad.d/vault-ca.pem"
  cert_file = "/etc/nomad.d/vault-client.pem"
  key_file  = "/etc/nomad.d/vault-client-key.pem"

  # Workload identity
  default_identity {
    aud  = ["vault.io"]
    env  = true
    file = true
    ttl  = "1h"
  }
}

# Consul integration
consul {
  address = "127.0.0.1:8501"
  ssl     = true

  ca_file   = "/etc/nomad.d/consul-ca.pem"
  cert_file = "/etc/nomad.d/consul-client.pem"
  key_file  = "/etc/nomad.d/consul-client-key.pem"

  auto_advertise       = true
  client_auto_join     = true

  # Client service registration
  client_service_name = "nomad-clients"

  # ACL token
  token = "{{ consul_token }}"

  # Consul Connect (service mesh)
  service_identity {
    aud = ["consul.io"]
    ttl = "1h"
  }

  task_identity {
    aud = ["consul.io"]
    ttl = "1h"
  }
}

# Telemetry
telemetry {
  prometheus_metrics         = true
  publish_allocation_metrics = true
  publish_node_metrics       = true
  collection_interval        = "10s"

  statsd_address = "127.0.0.1:8125"
}

# Plugin configuration
plugin_dir = "/opt/nomad/plugins"

plugin "docker" {
  config {
    endpoint = "unix:///var/run/docker.sock"

    extra_labels = ["job_name", "job_id", "task_group_name", "task_name", "namespace", "node_name"]

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

    volumes {
      enabled      = true
      selinuxlabel = "z"
    }

    allow_privileged = false

    allow_caps = [
      "audit_write", "chown", "dac_override", "fowner", "fsetid",
      "kill", "mknod", "net_bind_service", "setfcap", "setgid",
      "setpcap", "setuid", "sys_chroot"
    ]

    # Auth configuration for private registries
    auth {
      config = "/etc/docker/config.json"
    }

    # NVIDIA GPU support
    nvidia_runtime = "nvidia"

    infra_image = "gcr.io/google_containers/pause:3.9"

    # Pull timeout
    pull_activity_timeout = "5m"

    # Pids limit
    pids_limit = 0  # Use Docker default

    # Logging
    logging {
      type = "json-file"
      config {
        max-file = "3"
        max-size = "10m"
      }
    }
  }
}

plugin "exec" {
  config {
    # Set NoPivotRoot for containers without root
    no_pivot_root = false
  }
}

plugin "raw_exec" {
  config {
    enabled = false
  }
}

plugin "java" {
  config {
    # Java version to use
    jvm_path = "/usr/lib/jvm/java-17-openjdk-amd64/bin/java"
  }
}

plugin "qemu" {
  config {
    # QEMU image paths
    image_paths = ["/opt/nomad/images"]
  }
}

# Device plugins
plugin "nvidia-gpu" {
  config {
    enabled            = true
    fingerprint_period = "1m"
  }
}

# Fingerprint configuration
fingerprint {
  # Network fingerprint
  network_disallow_link_local = true
}
