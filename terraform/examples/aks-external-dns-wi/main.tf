terraform {
  required_version = ">= 1.5.0"
  required_providers {
    azurerm = { source = "hashicorp/azurerm", version = ">= 4.0" }
  }
}

provider "azurerm" { features {} }

variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "issuer_url" { type = string }
variable "uami_name" { type = string }
variable "ksa_namespace" { type = string, default = "external-dns" }
variable "ksa_name" { type = string, default = "external-dns" }
variable "dns_zone_id" { type = string }

module "wi" {
  source            = "../../modules/iam/aks-workload-identity"
  resource_group_name = var.resource_group_name
  location          = var.location
  issuer_url        = var.issuer_url
  uami_name         = var.uami_name
  ksa_namespace     = var.ksa_namespace
  ksa_name          = var.ksa_name
  dns_zone_id       = var.dns_zone_id
}

output "client_id" { value = module.wi.client_id }
