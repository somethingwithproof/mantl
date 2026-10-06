# SPDX-License-Identifier: Apache-2.0
# Node Exporter System Job for Nomad
# Runs on every node for metrics collection
# https://developer.hashicorp.com/nomad/docs/job-specification

job "node-exporter" {
  region      = "global"
  datacenters = ["dc1", "dc2"]
  namespace   = "system"
  type        = "system"
  priority    = 80

  # Update configuration for system jobs
  update {
    max_parallel     = 1
    stagger          = "30s"
    min_healthy_time = "10s"
    healthy_deadline = "3m"
    auto_revert      = true
  }

  # Task group
  group "exporters" {
    # Network configuration
    network {
      mode = "host"

      port "node_exporter" {
        static = 9100
      }

      port "cadvisor" {
        static = 9101
      }

      port "promtail" {
        static = 9080
      }
    }

    # Node Exporter task
    task "node-exporter" {
      driver = "docker"

      config {
        image = "prom/node-exporter:v1.7.0"

        ports = ["node_exporter"]

        args = [
          "--path.procfs=/host/proc",
          "--path.sysfs=/host/sys",
          "--path.rootfs=/rootfs",
          "--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)",
          "--collector.textfile.directory=/var/lib/node_exporter/textfile_collector",
          "--web.listen-address=:9100",
          "--web.telemetry-path=/metrics"
        ]

        volumes = [
          "/proc:/host/proc:ro",
          "/sys:/host/sys:ro",
          "/:/rootfs:ro",
          "local/textfile:/var/lib/node_exporter/textfile_collector"
        ]

        # Run as root to access host filesystems
        privileged = false

        # Security
        cap_add = ["SYS_TIME"]
      }

      resources {
        cpu    = 100
        memory = 64
      }

      service {
        name = "node-exporter"
        port = "node_exporter"
        tags = ["prometheus", "metrics", "infrastructure"]

        meta {
          instance = "${node.unique.name}"
        }

        check {
          type     = "http"
          path     = "/metrics"
          interval = "30s"
          timeout  = "5s"
        }
      }

      # Generate node metadata for prometheus
      template {
        data = <<EOH
# HELP mantl_node_info Node information
# TYPE mantl_node_info gauge
mantl_node_info{node_id="{{ env "node.unique.id" }}",node_name="{{ env "node.unique.name" }}",datacenter="{{ env "node.datacenter" }}",node_class="{{ env "attr.nomad.node_class" }}"} 1
EOH

        destination = "local/textfile/node_info.prom"
        perms       = "0644"
      }
    }

    # cAdvisor task for container metrics
    task "cadvisor" {
      driver = "docker"

      config {
        image = "gcr.io/cadvisor/cadvisor:v0.49.1"

        ports = ["cadvisor"]

        args = [
          "--port=9101",
          "--housekeeping_interval=10s",
          "--docker_only=true",
          "--disable_metrics=disk,tcp,udp,percpu,sched,process,hugetlb,referenced_memory,resctrl",
          "--store_container_labels=false"
        ]

        volumes = [
          "/:/rootfs:ro",
          "/var/run:/var/run:ro",
          "/sys:/sys:ro",
          "/var/lib/docker/:/var/lib/docker:ro",
          "/dev/disk/:/dev/disk:ro"
        ]

        privileged = true

        devices = [
          {
            host_path      = "/dev/kmsg"
            container_path = "/dev/kmsg"
          }
        ]
      }

      resources {
        cpu    = 200
        memory = 256
      }

      service {
        name = "cadvisor"
        port = "cadvisor"
        tags = ["prometheus", "metrics", "containers"]

        check {
          type     = "http"
          path     = "/healthz"
          interval = "30s"
          timeout  = "5s"
        }
      }
    }

    # Promtail for log collection
    task "promtail" {
      driver = "docker"

      config {
        image = "grafana/promtail:2.9.0"

        ports = ["promtail"]

        args = [
          "-config.file=/etc/promtail/promtail.yaml",
          "-config.expand-env=true"
        ]

        volumes = [
          "local/promtail.yaml:/etc/promtail/promtail.yaml:ro",
          "/var/log:/var/log:ro",
          "/var/lib/docker/containers:/var/lib/docker/containers:ro",
          "/opt/nomad/data/alloc:/nomad/alloc:ro"
        ]
      }

      template {
        data = <<EOH
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki.mantl.local:3100/loki/api/v1/push
    tenant_id: mantl
    batchwait: 1s
    batchsize: 1048576
    timeout: 10s
    backoff_config:
      min_period: 500ms
      max_period: 5m
      max_retries: 10

scrape_configs:
  # System logs
  - job_name: system
    static_configs:
      - targets:
          - localhost
        labels:
          job: system
          node: {{ env "node.unique.name" }}
          __path__: /var/log/*.log

  # Docker container logs
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 30s
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        regex: '/(.*)'
        target_label: 'container'
      - source_labels: ['__meta_docker_container_log_stream']
        target_label: 'logstream'
      - source_labels: ['__meta_docker_container_label_com_hashicorp_nomad_job_name']
        target_label: 'nomad_job'
      - source_labels: ['__meta_docker_container_label_com_hashicorp_nomad_task_name']
        target_label: 'nomad_task'
      - source_labels: ['__meta_docker_container_label_com_hashicorp_nomad_task_group_name']
        target_label: 'nomad_group'
    pipeline_stages:
      - json:
          expressions:
            level: level
            msg: msg
            time: time
      - labels:
          level:
      - timestamp:
          source: time
          format: RFC3339

  # Nomad allocation logs
  - job_name: nomad_allocs
    static_configs:
      - targets:
          - localhost
        labels:
          job: nomad_allocs
          node: {{ env "node.unique.name" }}
          __path__: /nomad/alloc/*/alloc/logs/*.std*
    pipeline_stages:
      - regex:
          expression: '/nomad/alloc/(?P<alloc_id>[^/]+)/alloc/logs/(?P<task>[^.]+)\\.(?P<stream>std(out|err)).*'
      - labels:
          alloc_id:
          task:
          stream:
EOH

        destination = "local/promtail.yaml"
        perms       = "0644"
      }

      resources {
        cpu    = 100
        memory = 128
      }

      service {
        name = "promtail"
        port = "promtail"
        tags = ["logs", "loki"]

        check {
          type     = "http"
          path     = "/ready"
          interval = "30s"
          timeout  = "5s"
        }
      }
    }
  }
}
