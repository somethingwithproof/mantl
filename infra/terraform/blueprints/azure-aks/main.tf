################################################################################
# Azure AKS Production Blueprint - Security Hardened
################################################################################


provider "azurerm" {
  # Register only the namespaces used by this blueprint (AzureRM 5 defaults to none).
  resource_provider_registrations = "none"
  resource_providers_to_register = [
    "Microsoft.Compute",
    "Microsoft.ContainerService",
    "Microsoft.Insights",
    "Microsoft.KeyVault",
    "Microsoft.ManagedIdentity",
    "Microsoft.Network",
    "Microsoft.OperationalInsights",
    "Microsoft.Storage",
  ]

  features {
    key_vault {
      purge_soft_delete_on_destroy    = false
      recover_soft_deleted_key_vaults = true
    }
  }
}

locals {
  name = var.cluster_name

  common_tags = merge(var.tags, {
    managed_by  = "terraform"
    platform    = "mantl"
    environment = var.environment
    blueprint   = "azure"
  })
}

################################################################################
# Resource Group
################################################################################

resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location
  tags     = local.common_tags
}

################################################################################
# Virtual Network with Security Features
################################################################################

resource "azurerm_virtual_network" "main" {
  name                = "${local.name}-vnet"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  address_space       = [var.vnet_cidr]

  tags = local.common_tags
}

resource "azurerm_subnet" "aks_nodes" {
  name                 = "${local.name}-aks-nodes"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.aks_subnet_cidr]

  # Enforce private endpoint policies
  private_endpoint_network_policies = "Enabled"

  # Service endpoints for Azure services
  service_endpoint {
    service = "Microsoft.Storage"
  }
  service_endpoint {
    service = "Microsoft.KeyVault"
  }
  service_endpoint {
    service = "Microsoft.ContainerRegistry"
  }
}

# Network Security Group for AKS nodes
resource "azurerm_network_security_group" "aks_nodes" {
  name                = "${local.name}-aks-nodes-nsg"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = local.common_tags
}

resource "azurerm_subnet_network_security_group_association" "aks_nodes" {
  subnet_id                 = azurerm_subnet.aks_nodes.id
  network_security_group_id = azurerm_network_security_group.aks_nodes.id
}

# Network Watcher for Flow Logs
resource "azurerm_network_watcher" "main" {
  name                = "${local.name}-network-watcher"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = local.common_tags
}

# Storage account for flow logs
resource "azurerm_storage_account" "flow_logs" {
  identity {
    type = "SystemAssigned"
  }

  lifecycle {
    prevent_destroy = true
  }
  name                            = replace("${local.name}flowlogs", "-", "")
  resource_group_name             = azurerm_resource_group.main.name
  location                        = azurerm_resource_group.main.location
  account_tier                    = "Standard"
  account_replication_type        = "LRS"
  min_tls_version                 = "TLS1_2"
  public_network_access           = "Enabled"
  allow_nested_items_to_be_public = false

  network_rules {
    default_action             = "Deny"
    bypass                     = ["AzureServices", "Logging", "Metrics"]
    virtual_network_subnet_ids = [azurerm_subnet.aks_nodes.id]
    ip_rules                   = []
  }

  # Storage accounts are encrypted at rest with Microsoft-managed keys by
  # default. A custom encryption scope is a separate
  # azurerm_storage_encryption_scope resource, not an inline block.


  tags = local.common_tags
}

# AzureRM 5 manages queue diagnostics as a separate resource.
resource "azurerm_storage_account_queue_properties" "flow_logs" {
  storage_account_id = azurerm_storage_account.flow_logs.id
  logging {
    delete                = true
    read                  = true
    write                 = true
    version               = "1.0"
    retention_policy_days = var.flow_logs_retention_days
  }
}

# VNet Flow Logs for security monitoring
resource "azurerm_network_watcher_flow_log" "aks_nodes" {
  name                 = "${local.name}-aks-nodes-flow-log"
  network_watcher_name = azurerm_network_watcher.main.name
  resource_group_name  = azurerm_resource_group.main.name

  target_resource_id = azurerm_network_security_group.aks_nodes.id
  storage_account_id = azurerm_storage_account.flow_logs.id
  enabled            = true

  retention_policy {
    enabled = true
    days    = var.flow_logs_retention_days
  }

  traffic_analytics {
    enabled               = true
    workspace_id          = azurerm_log_analytics_workspace.main.workspace_id
    workspace_region      = azurerm_log_analytics_workspace.main.location
    workspace_resource_id = azurerm_log_analytics_workspace.main.id
    interval_in_minutes   = 10
  }

  tags = local.common_tags
}

################################################################################
# Log Analytics for Monitoring
################################################################################

resource "azurerm_log_analytics_workspace" "main" {
  lifecycle {
    prevent_destroy = true
  }
  name                = "${local.name}-logs"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = var.log_retention_days

  tags = local.common_tags
}

################################################################################
# Azure Key Vault for Secrets Encryption
################################################################################

data "azurerm_client_config" "current" {}

resource "azurerm_key_vault" "main" {
  lifecycle {
    prevent_destroy = true
  }
  name                = "${local.name}-kv"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "premium"

  # Security features
  rbac_authorization_enabled      = true
  enabled_for_disk_encryption     = true
  enabled_for_deployment          = false
  enabled_for_template_deployment = false
  purge_protection_enabled        = true
  public_network_access_enabled   = false
  soft_delete_retention_days      = 90

  # Network access restrictions
  network_acls {
    bypass         = "AzureServices"
    default_action = var.enable_private_endpoint ? "Deny" : "Allow"

    virtual_network_subnet_ids = [azurerm_subnet.aks_nodes.id]
  }

  tags = local.common_tags
}

# Key for AKS secrets encryption
resource "azurerm_key_vault_key" "aks" {
  lifecycle {
    prevent_destroy = true
  }
  name         = "${local.name}-aks-key"
  key_vault_id = azurerm_key_vault.main.id
  key_type     = "RSA"
  key_size     = 2048

  key_opts = [
    "decrypt",
    "encrypt",
    "sign",
    "unwrapKey",
    "verify",
    "wrapKey",
  ]

  expiration_date = var.key_vault_key_expiration_date

  rotation_policy {
    automatic {
      time_before_expiry = "P30D"
    }
    expire_after         = "P1Y"
    notify_before_expiry = "P29D"
  }

  tags = local.common_tags
}

# Managed identity for AKS to access Key Vault
resource "azurerm_user_assigned_identity" "aks" {
  name                = "${local.name}-aks-identity"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = local.common_tags
}

resource "azurerm_role_assignment" "aks_key" {
  scope                = azurerm_key_vault_key.aks.resource_versionless_id
  role_definition_name = "Key Vault Crypto Service Encryption User"
  principal_id         = azurerm_user_assigned_identity.aks.principal_id
}

################################################################################
# Managed Identities for Platform Components
################################################################################

# External DNS
resource "azurerm_user_assigned_identity" "external_dns" {
  name                = "${local.name}-external-dns"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = local.common_tags
}

# cert-manager
resource "azurerm_user_assigned_identity" "cert_manager" {
  name                = "${local.name}-cert-manager"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = local.common_tags
}

# External Secrets Operator
resource "azurerm_user_assigned_identity" "external_secrets" {
  name                = "${local.name}-external-secrets"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = local.common_tags
}

# Grant External Secrets access to Key Vault
resource "azurerm_role_assignment" "external_secrets" {
  scope                = azurerm_key_vault.main.id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = azurerm_user_assigned_identity.external_secrets.principal_id
}

################################################################################
# AKS Cluster with Security Hardening
################################################################################

resource "azurerm_kubernetes_cluster" "main" {
  node_provisioning_profile {
    mode = "Manual"
  }

  name                = local.name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  dns_prefix          = local.name
  kubernetes_version  = var.kubernetes_version

  # Private cluster configuration
  private_cluster_enabled = var.enable_private_endpoint

  # API server authorized IP ranges (when not using private endpoint)
  dynamic "api_server_access_profile" {
    for_each = var.enable_private_endpoint ? [] : [1]
    content {
      authorized_ip_ranges = var.authorized_ip_ranges
    }
  }

  # System node pool (separate from workload)
  default_node_pool {
    name                 = "system"
    node_count           = var.system_node_count
    vm_size              = var.system_node_vm_size
    vnet_subnet_id       = azurerm_subnet.aks_nodes.id
    type                 = "VirtualMachineScaleSets"
    auto_scaling_enabled = true
    min_count            = var.system_node_min_count
    max_count            = var.system_node_max_count

    # Only system workloads
    only_critical_addons_enabled = true

    # Security features
    os_disk_type    = "Ephemeral"
    os_disk_size_gb = 100

    node_labels = {
      "node-role" = "system"
    }

    upgrade_settings {
      max_surge = "33%"
    }
  }

  # Identity configuration
  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.aks.id]
  }

  # API-server etcd encryption is separate from CSI-mounted secrets.
  key_management_service {
    key_vault_key_id         = azurerm_key_vault_key.aks.id
    key_vault_network_access = "Private"
  }

  # Key Vault secrets provider
  key_vault_secrets_provider {
    secret_rotation_enabled  = true
    secret_rotation_interval = "2m" # pragma: allowlist secret (rotation interval, not a credential)
  }

  # Workload identity (OIDC issuer)
  oidc_issuer_enabled       = true
  workload_identity_enabled = true

  # Network configuration
  network_profile {
    network_plugin    = "azure"
    network_policy    = "calico"
    service_cidr      = var.service_cidr
    dns_service_ip    = var.dns_service_ip
    load_balancer_sku = "standard"
    outbound_type     = "loadBalancer"
  }

  # Azure Policy for Kubernetes
  azure_policy_enabled = true

  # Microsoft Defender for Kubernetes
  microsoft_defender {
    log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
  }

  # Monitoring
  oms_agent {
    log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
  }

  # Azure AD integration
  azure_active_directory_role_based_access_control {
    azure_rbac_enabled     = true
    admin_group_object_ids = var.admin_group_object_ids
  }

  tags = local.common_tags

  depends_on = [
    azurerm_role_assignment.aks_key,
    azurerm_private_endpoint.key_vault,
  ]
}

################################################################################
# Workload Node Pool
################################################################################

resource "azurerm_kubernetes_cluster_node_pool" "workload" {
  name                  = "workload"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.main.id
  vm_size               = var.workload_node_vm_size
  vnet_subnet_id        = azurerm_subnet.aks_nodes.id

  auto_scaling_enabled = true
  min_count            = var.workload_node_min_count
  max_count            = var.workload_node_max_count

  # Cost optimization: spot instances
  priority        = var.workload_node_spot ? "Spot" : "Regular"
  eviction_policy = var.workload_node_spot ? "Delete" : null
  spot_max_price  = var.workload_node_spot ? -1 : null

  # Security features
  os_disk_type    = "Ephemeral"
  os_disk_size_gb = 100

  node_labels = {
    "node-role" = "workload"
  }

  node_taints = []

  upgrade_settings {
    max_surge = "33%"
  }

  tags = local.common_tags
}

################################################################################
# Outputs
################################################################################

output "cluster_id" {
  description = "AKS cluster ID"
  value       = azurerm_kubernetes_cluster.main.id
}

output "cluster_name" {
  description = "AKS cluster name"
  value       = azurerm_kubernetes_cluster.main.name
}

output "kube_config" {
  description = "Kubeconfig for the cluster"
  value       = azurerm_kubernetes_cluster.main.kube_config_raw
  sensitive   = true
}

output "host" {
  description = "Kubernetes API server endpoint"
  value       = azurerm_kubernetes_cluster.main.kube_config[0].host
}

output "key_vault_id" {
  description = "Azure Key Vault ID for secrets"
  value       = azurerm_key_vault.main.id
  sensitive   = true
}

output "external_dns_identity_client_id" {
  description = "Client ID for External DNS managed identity"
  value       = azurerm_user_assigned_identity.external_dns.client_id
}

output "cert_manager_identity_client_id" {
  description = "Client ID for cert-manager managed identity"
  value       = azurerm_user_assigned_identity.cert_manager.client_id
}

output "external_secrets_identity_client_id" {
  description = "Client ID for External Secrets managed identity"
  value       = azurerm_user_assigned_identity.external_secrets.client_id
}

output "log_analytics_workspace_id" {
  description = "Log Analytics workspace ID"
  value       = azurerm_log_analytics_workspace.main.id
}

# Record control-plane audit events independently of application monitoring.
resource "azurerm_monitor_diagnostic_setting" "aks_audit" {
  name                       = "${local.name}-audit"
  target_resource_id         = azurerm_kubernetes_cluster.main.id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id

  enabled_log {
    category = "kube-audit"
  }
  enabled_log {
    category = "kube-audit-admin"
  }
}

# Private Key Vault access for the API-server KMS integration.
resource "azurerm_private_dns_zone" "key_vault" {
  name                = "privatelink.vaultcore.azure.net"
  resource_group_name = azurerm_resource_group.main.name
  tags                = local.common_tags
}
resource "azurerm_private_dns_zone_virtual_network_link" "key_vault" {
  name                = "${local.name}-key-vault"
  private_dns_zone_id = azurerm_private_dns_zone.key_vault.id
  virtual_network_id  = azurerm_virtual_network.main.id
  tags                = local.common_tags
}
resource "azurerm_private_endpoint" "key_vault" {
  name                = "${local.name}-key-vault"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  subnet_id           = azurerm_subnet.aks_nodes.id
  private_service_connection {
    name                           = "${local.name}-key-vault"
    private_connection_resource_id = azurerm_key_vault.main.id
    subresource_names              = ["vault"]
    is_manual_connection           = false
  }
  private_dns_zone_group {
    name                 = "key-vault"
    private_dns_zone_ids = [azurerm_private_dns_zone.key_vault.id]
  }
  tags = local.common_tags
}

# Bind platform Kubernetes service accounts to their dedicated identities.
resource "azurerm_federated_identity_credential" "external_secrets" {
  name                      = "${local.name}-external-secrets"
  user_assigned_identity_id = azurerm_user_assigned_identity.external_secrets.id
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = azurerm_kubernetes_cluster.main.oidc_issuer_url
  subject                   = "system:serviceaccount:external-secrets:external-secrets"
}
resource "azurerm_federated_identity_credential" "external_dns" {
  name                      = "${local.name}-external-dns"
  user_assigned_identity_id = azurerm_user_assigned_identity.external_dns.id
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = azurerm_kubernetes_cluster.main.oidc_issuer_url
  subject                   = "system:serviceaccount:external-dns:external-dns"
}
resource "azurerm_federated_identity_credential" "cert_manager" {
  name                      = "${local.name}-cert-manager"
  user_assigned_identity_id = azurerm_user_assigned_identity.cert_manager.id
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = azurerm_kubernetes_cluster.main.oidc_issuer_url
  subject                   = "system:serviceaccount:cert-manager:cert-manager"
}
