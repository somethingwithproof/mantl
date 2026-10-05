# Inputs for the experimental standalone VM module. No cluster bootstrap is performed.
variable "long_name" {
  type = string
}

variable "short_name" {
  type = string
}

variable "zone" {
  type = string
}

variable "region" {
  type = string
}

variable "network_ipv4" {
  type    = string
  default = "10.20.0.0/24"
}

variable "datacenter" {
  type    = string
  default = "gce"
}

variable "ssh_user" {
  type    = string
  default = "mantl"
}

variable "ssh_key" {
  type = string
}

variable "control_type" {
  type    = string
  default = "e2-standard-2"
}

variable "worker_type" {
  type    = string
  default = "e2-standard-2"
}

variable "control_volume_size" {
  type    = number
  default = 50
}

variable "worker_volume_size" {
  type    = number
  default = 50
}

variable "control_count" {
  type    = number
  default = 0
  validation {
    condition     = var.control_count >= 0 && floor(var.control_count) == var.control_count
    error_message = "Instance counts must be nonnegative integers."
  }
}

variable "worker_count" {
  type    = number
  default = 0
  validation {
    condition     = var.worker_count >= 0 && floor(var.worker_count) == var.worker_count
    error_message = "Instance counts must be nonnegative integers."
  }
}

variable "kubeworker_count" {
  type    = number
  default = 0
  validation {
    condition     = var.kubeworker_count >= 0 && floor(var.kubeworker_count) == var.kubeworker_count
    error_message = "Instance counts must be nonnegative integers."
  }
}
