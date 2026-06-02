# Batch Processing Job for Nomad
# Parameterized batch job with GPU support
# https://developer.hashicorp.com/nomad/docs/job-specification

job "batch-processor" {
  region      = "global"
  datacenters = ["dc1"]
  namespace   = "batch"
  type        = "batch"
  priority    = 30

  # Parameterized job configuration
  parameterized {
    payload       = "required"
    meta_required = ["job_id", "input_path"]
    meta_optional = ["output_path", "priority", "gpu_required"]
  }

  # Periodic configuration (commented - use parameterized OR periodic)
  # periodic {
  #   crons            = ["0 */4 * * * *"]
  #   prohibit_overlap = true
  #   time_zone        = "UTC"
  # }

  # Constraints
  constraint {
    attribute = "${attr.kernel.name}"
    value     = "linux"
  }

  # GPU constraint when required
  constraint {
    attribute = "${meta.gpu}"
    operator  = "="
    value     = "true"
  }

  # Task group
  group "processor" {
    count = 1

    # Reschedule policy for batch
    reschedule {
      attempts  = 3
      interval  = "15m"
      delay     = "30s"
      unlimited = false
    }

    # Restart policy
    restart {
      attempts = 2
      interval = "5m"
      delay    = "15s"
      mode     = "fail"
    }

    # Network
    network {
      mode = "bridge"

      port "metrics" {
        to = 9090
      }
    }

    # Ephemeral disk for processing
    ephemeral_disk {
      size    = 50000  # 50GB for ML datasets
      sticky  = false
      migrate = false
    }

    # CSI Volume for input/output
    volume "data" {
      type            = "csi"
      source          = "batch-data"
      read_only       = false
      attachment_mode = "file-system"
      access_mode     = "multi-node-multi-writer"
    }

    # Processing task
    task "process" {
      driver = "docker"

      config {
        image = "ghcr.io/mantl/batch-processor:2.0.0"

        # GPU configuration
        runtime = "nvidia"

        # Mount configuration
        volumes = [
          "local/config:/app/config:ro"
        ]

        # Memory limit with swap
        memory_hard_limit = 32768  # 32GB

        # Security
        security_opt = [
          "no-new-privileges:true"
        ]

        # Capabilities
        cap_drop = ["ALL"]
      }

      # GPU device request
      resources {
        cpu    = 4000   # 4 cores
        memory = 16384  # 16GB

        # GPU allocation
        device "nvidia/gpu" {
          count = 1

          constraint {
            attribute = "${device.attr.memory}"
            operator  = ">="
            value     = "16000"  # 16GB VRAM
          }

          affinity {
            attribute = "${device.model}"
            value     = "A100"
            weight    = 50
          }
        }
      }

      # Volume mount
      volume_mount {
        volume      = "data"
        destination = "/data"
        read_only   = false
      }

      # Environment
      env {
        JOB_ID      = "${NOMAD_META_job_id}"
        INPUT_PATH  = "${NOMAD_META_input_path}"
        OUTPUT_PATH = "${NOMAD_META_output_path}"
        CUDA_VISIBLE_DEVICES = "${NOMAD_DEVICE_NVIDIA_GPU_0}"

        # ML framework settings
        TF_CPP_MIN_LOG_LEVEL = "2"
        PYTORCH_CUDA_ALLOC_CONF = "max_split_size_mb:512"
      }

      # Configuration template
      template {
        data = <<EOH
# Batch processing configuration
job:
  id: {{ env "JOB_ID" }}
  input_path: {{ env "INPUT_PATH" }}
  output_path: {{ env "OUTPUT_PATH" | default "/data/output" }}

processing:
  batch_size: 32
  num_workers: 4
  use_gpu: true

{{ with secret "kv/data/batch-processor/credentials" }}
aws:
  access_key: {{ .Data.data.aws_access_key }}
  secret_key: {{ .Data.data.aws_secret_key }}
  region: us-west-2
{{ end }}

s3:
  input_bucket: mantl-batch-input
  output_bucket: mantl-batch-output
EOH

        destination = "local/config/config.yaml"
        perms       = "0644"
      }

      # Vault integration
      vault {
        policies    = ["batch-processor"]
        change_mode = "noop"
      }

      # Kill timeout for long-running batch
      kill_timeout = "5m"

      # Logs
      logs {
        max_files     = 3
        max_file_size = 50
      }
    }

    # Metrics sidecar
    task "metrics" {
      driver = "docker"

      lifecycle {
        hook    = "prestart"
        sidecar = true
      }

      config {
        image = "prom/pushgateway:v1.7.0"
        ports = ["metrics"]
      }

      resources {
        cpu    = 50
        memory = 64
      }

      service {
        name = "batch-processor-metrics"
        port = "metrics"
        tags = ["prometheus", "batch"]

        check {
          type     = "http"
          path     = "/-/healthy"
          interval = "30s"
          timeout  = "5s"
        }
      }
    }
  }
}
