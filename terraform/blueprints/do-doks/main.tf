terraform {
  required_version = ">= 1.5.0"
  required_providers {
    digitalocean = { source = "digitalocean/digitalocean", version = ">= 2.0.0" }
  }
}

provider "digitalocean" {}

resource "digitalocean_kubernetes_cluster" "this" {
  name    = var.cluster_name
  region  = var.region
  version = var.kubernetes_version
  node_pool {
    name       = "default"
    size       = "s-4vcpu-8gb"
    node_count = 2
  }
}

output "cluster_id" { value = digitalocean_kubernetes_cluster.this.id }
output "endpoint" { value = digitalocean_kubernetes_cluster.this.endpoint }
