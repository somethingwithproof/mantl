# SPDX-License-Identifier: Apache-2.0
terraform {
  required_version = ">= 1.6"
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "2.6.2"
    }
  }
}
