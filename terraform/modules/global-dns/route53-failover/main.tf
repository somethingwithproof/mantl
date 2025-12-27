terraform {
  required_version = ">= 1.10.0"
  required_providers {
    aws = { source = "hashicorp/aws", version = ">= 6.0" }
  }
}

provider "aws" {
  region = var.region
}

resource "aws_route53_health_check" "primary" {
  fqdn              = var.primary_fqdn
  port              = var.port
  type              = "HTTP"
  resource_path     = var.health_path
  failure_threshold = 3
  request_interval  = 30
}

resource "aws_route53_health_check" "secondary" {
  fqdn              = var.secondary_fqdn
  port              = var.port
  type              = "HTTP"
  resource_path     = var.health_path
  failure_threshold = 3
  request_interval  = 30
}

resource "aws_route53_record" "primary" {
  zone_id = var.zone_id
  name    = var.record_name
  type    = "CNAME"
  set_identifier = "primary"
  ttl     = 30
  records = [var.primary_target]
  failover_routing_policy { type = "PRIMARY" }
  health_check_id = aws_route53_health_check.primary.id
}

resource "aws_route53_record" "secondary" {
  zone_id = var.zone_id
  name    = var.record_name
  type    = "CNAME"
  set_identifier = "secondary"
  ttl     = 30
  records = [var.secondary_target]
  failover_routing_policy { type = "SECONDARY" }
  health_check_id = aws_route53_health_check.secondary.id
}
