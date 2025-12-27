variable "cloudflare_api_token" {
  description = "Cloudflare API token (least privilege)"
  type        = string
  sensitive   = true
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}
