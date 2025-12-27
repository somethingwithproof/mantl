terraform {
  required_version = ">= 1.5.0"
  required_providers {
    linode = {
      source  = "linode/linode"
      version = ">= 2.0.0"
    }
  }
}

provider "linode" {}

# NOTE: This is a stub module for Linode. Add LKE cluster resources and networking.