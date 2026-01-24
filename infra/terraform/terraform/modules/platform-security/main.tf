################################################################################
# Platform Security Module
# Adds enterprise-grade security to providers without native features
################################################################################

terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.25"
    }
  }
}

################################################################################
# Cilium + Hubble (Network Observability - Flow Logs Alternative)
################################################################################

resource "kubernetes_namespace" "cilium" {
  count = var.enable_cilium ? 1 : 0

  metadata {
    name = "cilium"
    labels = {
      "pod-security.kubernetes.io/enforce" = "privileged"
      "pod-security.kubernetes.io/audit"   = "privileged"
      "pod-security.kubernetes.io/warn"    = "privileged"
    }
  }
}

resource "helm_release" "cilium" {
  count = var.enable_cilium ? 1 : 0

  name       = "cilium"
  repository = "https://helm.cilium.io/"
  chart      = "cilium"
  version    = "1.15.0"
  namespace  = kubernetes_namespace.cilium[0].metadata[0].name

  values = [yamlencode({
    hubble = {
      enabled = true
      relay = {
        enabled = true
      }
      ui = {
        enabled = true
        ingress = {
          enabled = var.hubble_ui_ingress_enabled
          hosts   = var.hubble_ui_hosts
        }
      }
      metrics = {
        enabled = [
          "dns",
          "drop",
          "tcp",
          "flow",
          "port-distribution",
          "icmp",
          "http"
        ]
        serviceMonitor = {
          enabled = true
        }
      }
    }

    operator = {
      replicas = 1
    }

    # Network policy enforcement
    policyEnforcementMode = "default"

    # Enable Prometheus metrics
    prometheus = {
      enabled        = true
      serviceMonitor = {
        enabled = true
      }
    }

    # Flow logs export
    flowLogs = {
      enabled = var.enable_flow_logs_export

      # Export to external logging if configured
      dynamic "export" {
        for_each = var.flow_logs_s3_bucket != "" ? [1] : []
        content {
          s3 = {
            bucket = var.flow_logs_s3_bucket
            prefix = "cilium-flows/"
          }
        }
      }
    }
  })]

  depends_on = [kubernetes_namespace.cilium]
}

################################################################################
# Sealed Secrets (KMS Alternative)
################################################################################

resource "kubernetes_namespace" "sealed_secrets" {
  count = var.enable_sealed_secrets ? 1 : 0

  metadata {
    name = "sealed-secrets"
    labels = {
      "pod-security.kubernetes.io/enforce" = "restricted"
      "pod-security.kubernetes.io/audit"   = "restricted"
      "pod-security.kubernetes.io/warn"    = "restricted"
    }
  }
}

resource "helm_release" "sealed_secrets" {
  count = var.enable_sealed_secrets ? 1 : 0

  name       = "sealed-secrets"
  repository = "https://bitnami-labs.github.io/sealed-secrets"
  chart      = "sealed-secrets"
  version    = "2.15.2"
  namespace  = kubernetes_namespace.sealed_secrets[0].metadata[0].name

  values = [yamlencode({
    controller = {
      # Automatic key rotation
      keyrenewperiod = "720h" # 30 days

      # Backup encryption keys
      createRoleBinding = true
    }

    metrics = {
      serviceMonitor = {
        enabled = true
      }
    }

    # Additional security
    securityContext = {
      runAsNonRoot = true
      runAsUser    = 1001
      fsGroup      = 65534
    }
  })]
}

################################################################################
# External Secrets Operator
################################################################################

resource "kubernetes_namespace" "external_secrets" {
  count = var.enable_external_secrets ? 1 : 0

  metadata {
    name = "external-secrets"
    labels = {
      "pod-security.kubernetes.io/enforce" = "restricted"
      "pod-security.kubernetes.io/audit"   = "restricted"
      "pod-security.kubernetes.io/warn"    = "restricted"
    }
  }
}

resource "helm_release" "external_secrets" {
  count = var.enable_external_secrets ? 1 : 0

  name       = "external-secrets"
  repository = "https://charts.external-secrets.io"
  chart      = "external-secrets"
  version    = "0.9.11"
  namespace  = kubernetes_namespace.external_secrets[0].metadata[0].name

  values = [yamlencode({
    installCRDs = true

    webhook = {
      port = 9443
    }

    certController = {
      enabled = true
    }

    # Prometheus metrics
    serviceMonitor = {
      enabled = true
    }
  })]
}

################################################################################
# Falco (Runtime Security - Enhanced Monitoring)
################################################################################

resource "kubernetes_namespace" "falco" {
  count = var.enable_falco ? 1 : 0

  metadata {
    name = "falco"
    labels = {
      "pod-security.kubernetes.io/enforce" = "privileged"
      "pod-security.kubernetes.io/audit"   = "privileged"
      "pod-security.kubernetes.io/warn"    = "privileged"
    }
  }
}

resource "helm_release" "falco" {
  count = var.enable_falco ? 1 : 0

  name       = "falco"
  repository = "https://falcosecurity.github.io/charts"
  chart      = "falco"
  version    = "4.0.0"
  namespace  = kubernetes_namespace.falco[0].metadata[0].name

  values = [yamlencode({
    driver = {
      kind = "modern_ebpf"
    }

    tty = true

    # Export alerts
    falco = {
      json_output       = true
      json_include_tags = true

      rules_file = [
        "/etc/falco/falco_rules.yaml",
        "/etc/falco/falco_rules.local.yaml",
        "/etc/falco/k8s_audit_rules.yaml",
      ]

      # Priority threshold
      priority = "warning"
    }

    # Falcosidekick for alert routing
    falcosidekick = {
      enabled = true

      config = {
        # Webhook for alerts
        webhook = {
          address = var.falco_webhook_url != "" ? var.falco_webhook_url : null
        }

        # Prometheus metrics
        prometheus = {
          enabled = true
        }
      }

      webui = {
        enabled = var.falco_ui_enabled
      }
    }

    # ServiceMonitor for Prometheus
    serviceMonitor = {
      enabled = true
    }

    # Custom rules for additional security
    customRules = {
      "custom-rules.yaml" = <<-EOF
        - rule: Unexpected outbound connection destination
          desc: Detect outbound connections to unexpected destinations
          condition: >
            outbound and not (
              fd.sip in (allowed_outbound_destination_ipaddrs) or
              fd.snet in (allowed_outbound_destination_subnets)
            )
          output: Outbound connection to unexpected destination (command=%proc.cmdline connection=%fd.name)
          priority: WARNING

        - rule: Secrets accessed
          desc: Detect access to Kubernetes secrets
          condition: >
            open_read and container and
            fd.name glob "/var/run/secrets/kubernetes.io/*"
          output: Secret accessed (user=%user.name file=%fd.name container=%container.info)
          priority: WARNING
      EOF
    }
  })]
}

################################################################################
# Tailscale VPN (Private Access Alternative)
################################################################################

resource "kubernetes_namespace" "tailscale" {
  count = var.enable_tailscale ? 1 : 0

  metadata {
    name = "tailscale"
    labels = {
      "pod-security.kubernetes.io/enforce" = "privileged"
      "pod-security.kubernetes.io/audit"   = "privileged"
      "pod-security.kubernetes.io/warn"    = "privileged"
    }
  }
}

# Tailscale auth key secret (must be created separately)
resource "kubernetes_secret" "tailscale_auth" {
  count = var.enable_tailscale && var.tailscale_auth_key != "" ? 1 : 0

  metadata {
    name      = "tailscale-auth"
    namespace = kubernetes_namespace.tailscale[0].metadata[0].name
  }

  data = {
    AUTH_KEY = var.tailscale_auth_key
  }

  type = "Opaque"
}

resource "helm_release" "tailscale_operator" {
  count = var.enable_tailscale ? 1 : 0

  name       = "tailscale-operator"
  repository = "https://pkgs.tailscale.com/helmcharts"
  chart      = "tailscale-operator"
  version    = "1.58.2"
  namespace  = kubernetes_namespace.tailscale[0].metadata[0].name

  set {
    name  = "oauth.clientId"
    value = var.tailscale_client_id
  }

  set_sensitive {
    name  = "oauth.clientSecret"
    value = var.tailscale_client_secret
  }

  values = [yamlencode({
    apiServerProxyConfig = {
      mode = "true"
    }
  })]
}

################################################################################
# Cert-Manager (For TLS certificates)
################################################################################

resource "kubernetes_namespace" "cert_manager" {
  count = var.enable_cert_manager ? 1 : 0

  metadata {
    name = "cert-manager"
    labels = {
      "pod-security.kubernetes.io/enforce" = "restricted"
      "pod-security.kubernetes.io/audit"   = "restricted"
      "pod-security.kubernetes.io/warn"    = "restricted"
    }
  }
}

resource "helm_release" "cert_manager" {
  count = var.enable_cert_manager ? 1 : 0

  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  version    = "v1.14.2"
  namespace  = kubernetes_namespace.cert_manager[0].metadata[0].name

  set {
    name  = "installCRDs"
    value = "true"
  }

  values = [yamlencode({
    prometheus = {
      enabled        = true
      servicemonitor = {
        enabled = true
      }
    }

    webhook = {
      securePort = 10250
    }
  })]
}

################################################################################
# Security Monitoring Dashboard
################################################################################

resource "kubernetes_config_map" "security_dashboard" {
  count = var.enable_security_dashboard ? 1 : 0

  metadata {
    name      = "security-monitoring-dashboard"
    namespace = "monitoring"
    labels = {
      grafana_dashboard = "1"
    }
  }

  data = {
    "security-monitoring.json" = jsonencode({
      title = "Security Monitoring"
      panels = [
        {
          title = "Cilium Flow Logs"
          type  = "graph"
          targets = [{
            expr = "rate(cilium_forward_count_total[5m])"
          }]
        },
        {
          title = "Falco Alerts"
          type  = "graph"
          targets = [{
            expr = "rate(falco_events[5m])"
          }]
        },
        {
          title = "Sealed Secrets Operations"
          type  = "graph"
          targets = [{
            expr = "rate(sealed_secrets_controller_unseal_requests_total[5m])"
          }]
        }
      ]
    })
  }
}

################################################################################
# Outputs
################################################################################

output "cilium_enabled" {
  description = "Whether Cilium is enabled"
  value       = var.enable_cilium
}

output "hubble_ui_url" {
  description = "Hubble UI URL (if ingress enabled)"
  value       = var.enable_cilium && var.hubble_ui_ingress_enabled ? "https://${var.hubble_ui_hosts[0]}" : null
}

output "sealed_secrets_enabled" {
  description = "Whether Sealed Secrets is enabled"
  value       = var.enable_sealed_secrets
}

output "external_secrets_enabled" {
  description = "Whether External Secrets Operator is enabled"
  value       = var.enable_external_secrets
}

output "falco_enabled" {
  description = "Whether Falco is enabled"
  value       = var.enable_falco
}

output "tailscale_enabled" {
  description = "Whether Tailscale VPN is enabled"
  value       = var.enable_tailscale
}

output "security_features_summary" {
  description = "Summary of enabled security features"
  value = {
    network_observability = var.enable_cilium ? "Cilium + Hubble (Flow Logs)" : "None"
    secrets_management    = var.enable_sealed_secrets ? "Sealed Secrets" : (var.enable_external_secrets ? "External Secrets" : "None")
    runtime_security      = var.enable_falco ? "Falco" : "None"
    private_access        = var.enable_tailscale ? "Tailscale VPN" : "None"
    tls_management        = var.enable_cert_manager ? "Cert-Manager" : "None"
  }
}
