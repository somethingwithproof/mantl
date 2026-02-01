# Example Web Service Job for Nomad
# Production-ready job specification with Consul Connect
# https://developer.hashicorp.com/nomad/docs/job-specification

job "web-service" {
  region      = "global"
  datacenters = ["dc1"]
  namespace   = "production"
  type        = "service"
  priority    = 50

  # Deployment configuration
  update {
    max_parallel      = 2
    min_healthy_time  = "30s"
    healthy_deadline  = "5m"
    progress_deadline = "10m"
    auto_revert       = true
    auto_promote      = true
    canary            = 1
  }

  # Migrate configuration
  migrate {
    max_parallel     = 1
    health_check     = "checks"
    min_healthy_time = "15s"
    healthy_deadline = "5m"
  }

  # Reschedule configuration
  reschedule {
    attempts       = 10
    interval       = "1h"
    delay          = "30s"
    delay_function = "exponential"
    max_delay      = "1h"
    unlimited      = false
  }

  # Multi-region configuration
  multiregion {
    strategy {
      max_parallel = 1
      on_failure   = "fail_all"
    }

    region "us-west" {
      count       = 3
      datacenters = ["dc1"]
    }

    region "us-east" {
      count       = 3
      datacenters = ["dc2"]
    }
  }

  # Constraints
  constraint {
    attribute = "${attr.kernel.name}"
    value     = "linux"
  }

  constraint {
    operator = "distinct_hosts"
    value    = "true"
  }

  # Affinity - prefer nodes with SSD
  affinity {
    attribute = "${meta.storage}"
    value     = "ssd"
    weight    = 50
  }

  # Spread across availability zones
  spread {
    attribute = "${meta.availability_zone}"
    weight    = 100

    target "us-west-2a" {
      percent = 34
    }
    target "us-west-2b" {
      percent = 33
    }
    target "us-west-2c" {
      percent = 33
    }
  }

  # Task group
  group "web" {
    count = 3

    # Network configuration with Consul Connect
    network {
      mode = "bridge"

      port "http" {
        to = 8080
      }

      port "metrics" {
        to = 9090
      }

      # DNS configuration
      dns {
        servers = ["172.17.0.1"]
      }
    }

    # Consul Connect service mesh
    service {
      name = "web-service"
      port = "http"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.web.rule=Host(`web.mantl.local`)",
        "traefik.http.routers.web.tls=true"
      ]

      meta {
        version = "1.0.0"
        owner   = "platform-team"
      }

      # Service mesh sidecar
      connect {
        sidecar_service {
          proxy {
            # Upstream services
            upstreams {
              destination_name = "postgres"
              local_bind_port  = 5432
            }

            upstreams {
              destination_name = "redis"
              local_bind_port  = 6379
            }

            # Proxy configuration
            config {
              protocol = "http"
              limits {
                max_connections         = 1000
                max_pending_requests    = 1000
                max_concurrent_requests = 1000
              }
            }
          }
        }

        # Sidecar task resources
        sidecar_task {
          resources {
            cpu    = 100
            memory = 128
          }
        }
      }

      # Health checks
      check {
        name     = "http-health"
        type     = "http"
        port     = "http"
        path     = "/health"
        interval = "10s"
        timeout  = "3s"

        check_restart {
          limit           = 3
          grace           = "60s"
          ignore_warnings = false
        }
      }

      check {
        name     = "ready"
        type     = "http"
        port     = "http"
        path     = "/ready"
        interval = "5s"
        timeout  = "2s"
      }
    }

    # Prometheus metrics service
    service {
      name = "web-service-metrics"
      port = "metrics"
      tags = ["prometheus"]

      check {
        name     = "metrics-health"
        type     = "http"
        port     = "metrics"
        path     = "/metrics"
        interval = "30s"
        timeout  = "5s"
      }
    }

    # Ephemeral disk
    ephemeral_disk {
      size    = 1000  # 1GB
      sticky  = true
      migrate = true
    }

    # Restart policy
    restart {
      attempts = 3
      interval = "5m"
      delay    = "15s"
      mode     = "fail"
    }

    # Volume configuration
    volume "data" {
      type            = "csi"
      source          = "web-data"
      read_only       = false
      attachment_mode = "file-system"
      access_mode     = "single-node-writer"

      mount_options {
        fs_type     = "ext4"
        mount_flags = ["noatime"]
      }
    }

    # Scaling configuration
    scaling {
      enabled = true
      min     = 2
      max     = 10

      policy {
        evaluation_interval = "30s"
        cooldown            = "5m"

        # Target CPU utilization
        check "cpu" {
          source = "nomad-apm"
          group  = "web"

          strategy "target-value" {
            target = 70
          }
        }

        # Target memory utilization
        check "memory" {
          source = "nomad-apm"
          group  = "web"

          strategy "target-value" {
            target = 80
          }
        }

        # Target requests per second
        check "requests" {
          source = "prometheus"
          query  = "sum(rate(http_requests_total{job='web-service'}[5m]))"

          strategy "target-value" {
            target = 1000
          }
        }
      }
    }

    # Main application task
    task "app" {
      driver = "docker"

      config {
        image = "ghcr.io/mantl/web-service:1.0.0"

        ports = ["http", "metrics"]

        # Docker labels
        labels {
          service = "web-service"
          env     = "production"
        }

        # Mount host volumes
        volumes = [
          "local/config:/app/config:ro"
        ]

        # Security options
        security_opt = [
          "no-new-privileges:true"
        ]

        # Read-only root filesystem
        readonly_rootfs = true

        # Capabilities
        cap_drop = ["ALL"]
        cap_add  = ["NET_BIND_SERVICE"]

        # Ulimits
        ulimit {
          nofile = "65536:65536"
          nproc  = "4096:4096"
        }

        # Healthcheck
        healthchecks {
          disable = true  # We use Nomad checks
        }

        # Logging
        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "3"
          }
        }
      }

      # Volume mounts
      volume_mount {
        volume      = "data"
        destination = "/data"
        read_only   = false
      }

      # Environment variables
      env {
        SERVICE_NAME = "web-service"
        LOG_LEVEL    = "info"
        LOG_FORMAT   = "json"
        PORT         = "8080"
        METRICS_PORT = "9090"

        # Database connection via Consul Connect
        DB_HOST = "${NOMAD_UPSTREAM_ADDR_postgres}"
        REDIS_HOST = "${NOMAD_UPSTREAM_ADDR_redis}"
      }

      # Templates for configuration
      template {
        data = <<EOH
# Application configuration
server:
  port: {{ env "PORT" }}
  metrics_port: {{ env "METRICS_PORT" }}

database:
  host: {{ env "DB_HOST" }}
  name: {{ with secret "database/creds/web-service" }}{{ .Data.username }}{{ end }}
  password: {{ with secret "database/creds/web-service" }}{{ .Data.password }}{{ end }}

redis:
  host: {{ env "REDIS_HOST" }}
  password: {{ with secret "kv/data/web-service/redis" }}{{ .Data.data.password }}{{ end }}

{{ range service "web-service" }}
peer: {{ .Address }}:{{ .Port }}
{{ end }}
EOH

        destination = "local/config/config.yaml"
        perms       = "0644"
        change_mode = "signal"
        change_signal = "SIGHUP"
      }

      # Vault secrets
      vault {
        policies      = ["web-service"]
        change_mode   = "signal"
        change_signal = "SIGHUP"
      }

      # Identity for workload identity
      identity {
        env  = true
        file = true
      }

      # Resources
      resources {
        cpu    = 500
        memory = 512

        # Memory oversubscription
        memory_max = 1024
      }

      # Kill configuration
      kill_timeout = "30s"
      kill_signal  = "SIGTERM"

      # Shutdown delay for graceful drain
      shutdown_delay = "10s"

      # Logs
      logs {
        max_files     = 5
        max_file_size = 10
      }

      # Artifacts
      artifact {
        source      = "https://releases.mantl.local/web-service/config.tar.gz"
        destination = "local/artifacts"

        options {
          checksum = "sha256:abc123..."
        }
      }
    }

    # Log shipping sidecar
    task "log-shipper" {
      driver = "docker"

      lifecycle {
        hook    = "poststart"
        sidecar = true
      }

      config {
        image = "fluent/fluent-bit:2.2"

        volumes = [
          "local/fluent-bit.conf:/fluent-bit/etc/fluent-bit.conf:ro",
          "alloc/logs:/logs:ro"
        ]
      }

      template {
        data = <<EOH
[SERVICE]
    Flush         5
    Daemon        Off
    Log_Level     info

[INPUT]
    Name          tail
    Path          /logs/*.log
    Tag           web-service

[OUTPUT]
    Name          opensearch
    Host          opensearch.mantl.local
    Port          9200
    Index         web-service-logs
    Type          _doc
    tls           On
    tls.verify    Off
EOH

        destination = "local/fluent-bit.conf"
      }

      resources {
        cpu    = 50
        memory = 64
      }
    }
  }
}
