variable "resource_group_name" { type = string }
variable "cluster_name" { type = string }
variable "kubernetes_version" { type = string, default = "1.30.0" }
variable "subnet_id" { type = string }
