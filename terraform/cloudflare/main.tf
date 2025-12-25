# Inputs (modern types)
variable "domain" {
  description = "Cloudflare zone name (e.g., example.com)"
  type        = string
}

variable "short_name" {
  description = "Cluster short name (used as DNS label prefix)"
  type        = string
}

variable "subdomain" {
  description = "Optional subdomain suffix, prefixed with a dot when non-empty (e.g., .prod.example.com)"
  type        = string
  default     = ""
}

variable "control_subdomain" {
  description = "Subdomain label for control group A-records"
  type        = string
  default     = "control"
}

variable "control_ips" {
  description = "IPv4 addresses for control nodes"
  type        = list(string)
  default     = []
}

variable "worker_ips" {
  description = "IPv4 addresses for worker nodes"
  type        = list(string)
  default     = []
}

variable "kubeworker_ips" {
  description = "IPv4 addresses for kube worker nodes"
  type        = list(string)
  default     = []
}

variable "edge_ips" {
  description = "IPv4 addresses for edge nodes"
  type        = list(string)
  default     = []
}

# Resolve zone_id from domain
data "cloudflare_zone" "this" {
  filter = {
    name   = var.domain
    status = "active"
    paused = false
  }
}

locals {
  zone_id   = data.cloudflare_zone.this.id
  suffix    = var.subdomain
  # indexed maps to preserve order for formatted indices
  control   = { for idx, ip in var.control_ips    : idx => ip }
  workers   = { for idx, ip in var.worker_ips     : idx => ip }
  kubeworks = { for idx, ip in var.kubeworker_ips : idx => ip }
  edges     = { for idx, ip in var.edge_ips       : idx => ip }
}

# Individual records (A)
resource "cloudflare_dns_record" "control" {
  for_each = local.control
  zone_id  = local.zone_id
  name     = "${var.short_name}-control-${format("%02d", tonumber(each.key) + 1)}.node${local.suffix}"
  content  = each.value
  type     = "A"
  ttl      = 1
  proxied  = false
}

resource "cloudflare_dns_record" "worker" {
  for_each = local.workers
  zone_id  = local.zone_id
  name     = "${var.short_name}-worker-${format("%03d", tonumber(each.key) + 1)}.node${local.suffix}"
  content  = each.value
  type     = "A"
  ttl      = 1
  proxied  = false
}

resource "cloudflare_dns_record" "kubeworker" {
  for_each = local.kubeworks
  zone_id  = local.zone_id
  name     = "${var.short_name}-kubeworker-${format("%03d", tonumber(each.key) + 1)}.node${local.suffix}"
  content  = each.value
  type     = "A"
  ttl      = 1
  proxied  = false
}

resource "cloudflare_dns_record" "edge" {
  for_each = local.edges
  zone_id  = local.zone_id
  name     = "${var.short_name}-edge-${format("%02d", tonumber(each.key) + 1)}.node${local.suffix}"
  content  = each.value
  type     = "A"
  ttl      = 1
  proxied  = true
}

# Group records
resource "cloudflare_dns_record" "control_group" {
  for_each = local.control
  zone_id  = local.zone_id
  name     = "${var.control_subdomain}${local.suffix}"
  content  = each.value
  type     = "A"
  ttl      = 1
  proxied  = false
}

resource "cloudflare_dns_record" "wildcard_edge" {
  for_each = local.edges
  zone_id  = local.zone_id
  name     = "*${local.suffix}"
  content  = each.value
  type     = "A"
  ttl      = 1
  proxied  = true
}
