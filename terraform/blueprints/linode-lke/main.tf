terraform {
  required_version = ">= 1.5.0"
  required_providers {
    linode = { source = "linode/linode", version = ">= 2.0.0" }
  }
}

provider "linode" {}

resource "linode_lke_cluster" "this" {
  label       = var.cluster_name
  region      = var.region
  k8s_version = var.kubernetes_version
  pool {
    type  = "g6-standard-4"
    count = 2
  }
}

output "cluster_id" { value = linode_lke_cluster.this.id }
