terraform {
  required_version = ">= 1.5.0"
  required_providers {
    azurerm = { source = "hashicorp/azurerm", version = ">= 4.0" }
  }
}

provider "azurerm" {
  features {}
}

module "aks" {
  source  = "Azure/aks/azurerm"
  version = "~> 6.0"

  resource_group_name = var.resource_group_name
  cluster_name        = var.cluster_name
  kubernetes_version  = var.kubernetes_version
  vnet_subnet_id      = var.subnet_id

  agents_pool_name    = "nodepool1"
  agents_count        = 2
  agents_vm_size      = "Standard_D4s_v5"
}

output "kube_config" { value = module.aks.kube_config }
output "host" { value = module.aks.host }
