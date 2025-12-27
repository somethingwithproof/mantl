variable "region" { type = string }
variable "cluster_name" { type = string }
variable "kubernetes_version" { type = string, default = "1.30" }
variable "vpc_id" { type = string }
variable "subnet_ids" { type = list(string) }
