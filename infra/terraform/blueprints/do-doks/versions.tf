terraform {
  required_version = ">= 1.6"
  required_providers {
    digitalocean = { source = "digitalocean/digitalocean", version = ">= 2.70, < 3.0" }
    helm         = { source = "hashicorp/helm", version = "~> 2.12" }
    kubernetes   = { source = "hashicorp/kubernetes", version = "~> 2.25" }
  }
}
