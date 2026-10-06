# SPDX-License-Identifier: Apache-2.0
terraform {
  required_version = ">= 1.6"
  required_providers {
    google = { source = "hashicorp/google", version = ">= 7.0, < 8.0" }
  }
}
