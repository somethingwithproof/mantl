terraform {
  required_version = ">= 1.6"
  required_providers {
    linode     = { source = "linode/linode", version = ">= 3.0, < 4.0" }
    helm       = { source = "hashicorp/helm", version = "3.3.0" }
    kubernetes = { source = "hashicorp/kubernetes", version = "3.3.0" }
  }
}
