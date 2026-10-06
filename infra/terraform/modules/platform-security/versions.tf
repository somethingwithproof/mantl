# SPDX-License-Identifier: Apache-2.0
terraform {
  required_version = ">= 1.6"

  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "3.3.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "3.3.0"
    }
  }
}
