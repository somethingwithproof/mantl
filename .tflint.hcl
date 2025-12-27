# Global TFLint configuration for all Terraform modules in this repo
# Keep it minimal and portable — CI will run `tflint --init` to fetch plugins.

plugin "aws" {
  enabled = true
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

plugin "google" {
  enabled = true
  source  = "github.com/terraform-linters/tflint-ruleset-google"
}

plugin "azurerm" {
  enabled = true
  source  = "github.com/terraform-linters/tflint-ruleset-azurerm"
}

plugin "oci" {
  enabled = true
  source  = "github.com/terraform-linters/tflint-ruleset-oci"
}

plugin "digitalocean" {
  enabled = true
  source  = "github.com/terraform-linters/tflint-ruleset-digitalocean"
}
