variable "project" { type = string }
variable "region" { type = string }
variable "cluster_name" { type = string }
variable "network" { type = string }
variable "subnetwork" { type = string }
variable "ip_range_pods" { type = string }
variable "ip_range_services" { type = string }

variable "enable_private_endpoint" {
  description = "Disable the public endpoint; set false only for dev environments where bastion access is unavailable"
  type        = bool
  default     = true
}
