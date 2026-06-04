################################################################################
# General
################################################################################

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
  default     = "mantl"
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "kubernetes_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.31" # Updated to latest stable version
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}

################################################################################
# VPC
################################################################################

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "single_nat_gateway" {
  description = "Use a single NAT Gateway (cost savings for non-prod)"
  type        = bool
  default     = false
}

################################################################################
# EKS Cluster
################################################################################

variable "cluster_endpoint_public_access" {
  description = "Enable public access to cluster API endpoint (NOT recommended for production)"
  type        = bool
  default     = false # Changed to false for security - use bastion/VPN for access
}

################################################################################
# Node Groups - System
################################################################################

variable "system_node_instance_types" {
  description = "Instance types for system node group"
  type        = list(string)
  default     = ["m6i.large", "m5.large"]
}

variable "system_node_min_size" {
  description = "Minimum number of system nodes"
  type        = number
  default     = 2
}

variable "system_node_max_size" {
  description = "Maximum number of system nodes"
  type        = number
  default     = 4
}

variable "system_node_desired_size" {
  description = "Desired number of system nodes"
  type        = number
  default     = 2
}

################################################################################
# Node Groups - Workload
################################################################################

variable "workload_node_instance_types" {
  description = "Instance types for workload node group"
  type        = list(string)
  default     = ["m6i.xlarge", "m5.xlarge"]
}

variable "workload_node_capacity_type" {
  description = "Capacity type for workload nodes (ON_DEMAND or SPOT)"
  type        = string
  default     = "ON_DEMAND"
}

variable "workload_node_min_size" {
  description = "Minimum number of workload nodes"
  type        = number
  default     = 2
}

variable "workload_node_max_size" {
  description = "Maximum number of workload nodes"
  type        = number
  default     = 10
}

variable "workload_node_desired_size" {
  description = "Desired number of workload nodes"
  type        = number
  default     = 3
}

################################################################################
# IRSA - Route53 (for External DNS and cert-manager)
################################################################################

variable "route53_zone_arns" {
  description = "ARNs of Route53 hosted zones for External DNS and cert-manager"
  type        = list(string)
  default     = ["arn:aws:route53:::hostedzone/*"]
}

################################################################################
# IRSA - Secrets (for External Secrets)
################################################################################

variable "ssm_parameter_arns" {
  description = "ARNs of SSM parameters for External Secrets"
  type        = list(string)
  default     = ["arn:aws:ssm:*:*:parameter/*"]
}

variable "secrets_manager_arns" {
  description = "ARNs of Secrets Manager secrets for External Secrets"
  type        = list(string)
  default     = ["arn:aws:secretsmanager:*:*:secret:*"]
}
