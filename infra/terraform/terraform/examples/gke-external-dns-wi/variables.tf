variable "project" {
  type = string
}

variable "region" {
  type = string
}

variable "gsa_name" {
  type = string
}

variable "ksa_namespace" {
  type    = string
  default = "external-dns"
}

variable "ksa_name" {
  type    = string
  default = "external-dns"
}
