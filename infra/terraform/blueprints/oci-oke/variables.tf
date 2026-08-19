# OCI Provider Configuration
variable "tenancy_ocid" {
  description = "OCI tenancy OCID"
  type        = string
  sensitive   = true
}

variable "user_ocid" {
  description = "OCI user OCID"
  type        = string
  sensitive   = true
}

variable "fingerprint" {
  description = "OCI API key fingerprint"
  type        = string
  sensitive   = true
}

variable "private_key_path" {
  description = "Path to OCI API private key"
  type        = string
  default     = "~/.oci/oci_api_key.pem"
  sensitive   = true
}

variable "region" {
  description = "OCI region (e.g., us-ashburn-1, us-phoenix-1, eu-frankfurt-1, ap-tokyo-1)"
  type        = string
  default     = "us-ashburn-1"
}

variable "compartment_id" {
  description = "OCI compartment OCID for resources"
  type        = string
}

# Core Configuration
variable "cluster_name" {
  description = "Name of the OKE cluster"
  type        = string
  default     = "mantl-oke"
}

variable "environment" {
  description = "Environment name (e.g., production, staging, development)"
  type        = string
  default     = "production"
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}

variable "kubernetes_version" {
  description = "Kubernetes version for OKE cluster"
  type        = string
  default     = "v1.31.1" # Latest stable
}

# Network Configuration
variable "vcn_cidr" {
  description = "CIDR block for the VCN"
  type        = string
  default     = "10.0.0.0/16"
}

variable "pod_cidr" {
  description = "CIDR block for Kubernetes pods"
  type        = string
  default     = "10.244.0.0/16"
}

variable "service_cidr" {
  description = "CIDR block for Kubernetes services"
  type        = string
  default     = "10.96.0.0/16"
}

variable "use_flannel_cni" {
  description = "Use Flannel CNI (overlay) instead of OCI VCN-native pod networking"
  type        = bool
  default     = false # VCN-native recommended for production
}

# Security Configuration
variable "enable_private_endpoint" {
  description = "Make API endpoint private (accessible only within VCN)"
  type        = bool
  default     = true
}

variable "allowed_ssh_cidrs" {
  description = "CIDR blocks allowed to SSH to nodes (empty list = no SSH access)"
  type        = list(string)
  default     = [] # No SSH by default
}

variable "enable_flow_logs" {
  description = "Enable VCN Flow Logs for security monitoring"
  type        = bool
  default     = true
}

variable "flow_logs_retention_days" {
  description = "Number of days to retain flow logs"
  type        = number
  default     = 90
}

variable "enable_audit_logs" {
  description = "Enable OKE cluster audit logs"
  type        = bool
  default     = true
}

variable "audit_log_retention_days" {
  description = "Number of days to retain audit logs"
  type        = number
  default     = 90
}

variable "enable_vault" {
  description = "Enable OCI Vault for secrets encryption"
  type        = bool
  default     = true
}

variable "enable_image_policy" {
  description = "Enable image signature verification"
  type        = bool
  default     = false # Requires image signing setup
}

# System Node Pool Configuration
variable "system_node_shape" {
  description = "OCI compute shape for system nodes (e.g., VM.Standard.E4.Flex)"
  type        = string
  default     = "VM.Standard.E4.Flex"
}

variable "system_node_ocpus" {
  description = "Number of OCPUs per system node (for Flex shapes)"
  type        = number
  default     = 4
}

variable "system_node_memory_gb" {
  description = "Memory in GB per system node (for Flex shapes)"
  type        = number
  default     = 16
}

variable "system_node_count" {
  description = "Number of system nodes (across all availability domains)"
  type        = number
  default     = 3
}

# Workload Node Pool Configuration
variable "workload_node_shape" {
  description = "OCI compute shape for workload nodes"
  type        = string
  default     = "VM.Standard.E4.Flex"
}

variable "workload_node_ocpus" {
  description = "Number of OCPUs per workload node (for Flex shapes)"
  type        = number
  default     = 4
}

variable "workload_node_memory_gb" {
  description = "Memory in GB per workload node (for Flex shapes)"
  type        = number
  default     = 16
}

variable "workload_initial_node_count" {
  description = "Initial number of workload nodes (across all availability domains)"
  type        = number
  default     = 3
}

# Node Image
variable "node_image_id" {
  description = "OCID of the node image (Oracle Linux 8 recommended). Leave empty to use latest."
  type        = string
  default     = ""
}

# Backup Storage
variable "enable_backup_bucket" {
  description = "Create Object Storage bucket for backups"
  type        = bool
  default     = true
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30
}

variable "backup_retention_lock_time" {
  description = "Optional stable RFC3339 timestamp after which the backup retention rule becomes immutable"
  type        = string
  default     = null
}

variable "enable_object_events" {
  description = "Enable object storage events for backup bucket"
  type        = bool
  default     = false
}
