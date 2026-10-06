# SPDX-License-Identifier: Apache-2.0
# Core Configuration
variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
  default     = "mantl-openstack"
}

variable "environment" {
  description = "Environment name (e.g., production, staging, development)"
  type        = string
  default     = "production"
}

variable "kubernetes_version" {
  description = "Kubernetes version tag (e.g., v1.31.1)"
  type        = string
  default     = "v1.31.1" # Latest stable
}

# Cluster Template Configuration
variable "use_existing_template" {
  description = "Use an existing cluster template instead of creating new one"
  type        = bool
  default     = false
}

variable "existing_template_name" {
  description = "Name of existing cluster template (if use_existing_template is true)"
  type        = string
  default     = ""
}

variable "template_labels" {
  description = "Additional labels for cluster template"
  type        = map(string)
  default     = {}
}

# Network Configuration
variable "external_network_name" {
  description = "Name of the external network for floating IPs"
  type        = string
}

variable "cluster_cidr" {
  description = "CIDR block for cluster network"
  type        = string
  default     = "10.0.0.0/24"
}

variable "dns_nameserver" {
  description = "DNS nameserver for cluster nodes"
  type        = string
  default     = "8.8.8.8"
}

variable "network_driver" {
  description = "Network driver for Kubernetes (flannel or calico)"
  type        = string
  default     = "flannel"
  validation {
    condition     = contains(["flannel", "calico"], var.network_driver)
    error_message = "network_driver must be either flannel or calico"
  }
}

# Proxy Configuration (optional)
variable "http_proxy" {
  description = "HTTP proxy URL (optional)"
  type        = string
  default     = ""
}

variable "https_proxy" {
  description = "HTTPS proxy URL (optional)"
  type        = string
  default     = ""
}

variable "no_proxy" {
  description = "No proxy list (comma-separated, optional)"
  type        = string
  default     = ""
}

# Security Configuration
variable "allowed_ssh_cidrs" {
  description = "CIDR blocks allowed to SSH to nodes"
  type        = list(string)
  default     = [] # No SSH by default
}

variable "api_server_access_cidr" {
  description = "CIDR block allowed to access Kubernetes API server"
  type        = string
  default     = "0.0.0.0/0" # Restrict in production
}

# Master Node Configuration
variable "master_count" {
  description = "Number of master nodes (1 or 3 recommended for HA)"
  type        = number
  default     = 3
  validation {
    condition     = var.master_count == 1 || var.master_count == 3 || var.master_count == 5
    error_message = "master_count should be 1, 3, or 5 for proper HA"
  }
}

variable "master_flavor" {
  description = "OpenStack flavor for master nodes"
  type        = string
  default     = "m1.medium" # 2 vCPU, 4 GB RAM
}

# Worker Node Configuration
variable "initial_node_count" {
  description = "Initial number of worker nodes"
  type        = number
  default     = 3
}

variable "worker_flavor" {
  description = "OpenStack flavor for worker nodes"
  type        = string
  default     = "m1.large" # 4 vCPU, 8 GB RAM
}

# Auto-scaling Configuration
variable "enable_auto_scaling" {
  description = "Enable cluster auto-scaling"
  type        = bool
  default     = true
}

variable "min_node_count" {
  description = "Minimum number of worker nodes (for autoscaling)"
  type        = number
  default     = 2
}

variable "max_node_count" {
  description = "Maximum number of worker nodes (for autoscaling)"
  type        = number
  default     = 10
}

# Node Image
variable "node_image_name" {
  description = "Name of the Glance image for Kubernetes nodes (Fedora CoreOS recommended)"
  type        = string
  default     = "fedora-coreos-latest"
}

# Storage Configuration
variable "docker_volume_size" {
  description = "Size of Docker storage volume in GB"
  type        = number
  default     = 50
}

# SSH Access
variable "ssh_keypair_name" {
  description = "Name of OpenStack SSH keypair for node access"
  type        = string
}

# Additional Features
variable "enable_registry" {
  description = "Enable integrated Docker registry"
  type        = bool
  default     = false
}

variable "enable_floating_ip" {
  description = "Assign floating IP to master nodes"
  type        = bool
  default     = true
}

variable "enable_master_lb" {
  description = "Enable load balancer for master nodes"
  type        = bool
  default     = true
}

variable "enable_auto_healing" {
  description = "Enable auto-healing for failed nodes"
  type        = bool
  default     = true
}

# Backup Storage
variable "enable_backup_container" {
  description = "Create Swift container for backups"
  type        = bool
  default     = true
}

# Cluster Timeouts
variable "cluster_create_timeout" {
  description = "Timeout for cluster creation in minutes"
  type        = number
  default     = 60
}
