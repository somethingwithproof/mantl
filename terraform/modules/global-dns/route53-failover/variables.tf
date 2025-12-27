variable "region" { type = string }
variable "zone_id" { type = string }
variable "record_name" { type = string }
variable "primary_target" { type = string }
variable "secondary_target" { type = string }
variable "primary_fqdn" { type = string }
variable "secondary_fqdn" { type = string }
variable "port" { type = number, default = 80 }
variable "health_path" { type = string, default = "/healthz" }
