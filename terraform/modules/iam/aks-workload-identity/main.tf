terraform {
  required_version = ">= 1.5.0"
  required_providers {
    azurerm = { source = "hashicorp/azurerm", version = ">= 4.0" }
  }
}

provider "azurerm" { features {} }

variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "uami_name" { type = string }
variable "issuer_url" { type = string } # AKS OIDC issuer URL
variable "ksa_namespace" { type = string }
variable "ksa_name" { type = string }
variable "dns_zone_id" { type = string } # /subscriptions/.../resourceGroups/.../providers/Microsoft.Network/dnszones/<zone>

resource "azurerm_user_assigned_identity" "uami" {
  name                = var.uami_name
  location            = var.location
  resource_group_name = var.resource_group_name
}

# Federated Identity Credential for the KSA subject
resource "azurerm_federated_identity_credential" "fic" {
  name                = "${var.uami_name}-fic"
  resource_group_name = var.resource_group_name
  parent_id           = azurerm_user_assigned_identity.uami.id
  audience            = ["api://AzureADTokenExchange"]
  issuer              = var.issuer_url
  subject             = "system:serviceaccount:${var.ksa_namespace}:${var.ksa_name}"
}

# Assign DNS Zone Contributor to the UAMI on the target zone
resource "azurerm_role_assignment" "dns_contrib" {
  scope                = var.dns_zone_id
  role_definition_name = "DNS Zone Contributor"
  principal_id         = azurerm_user_assigned_identity.uami.principal_id
}

output "client_id" { value = azurerm_user_assigned_identity.uami.client_id }
output "principal_id" { value = azurerm_user_assigned_identity.uami.principal_id }
output "uami_id" { value = azurerm_user_assigned_identity.uami.id }
