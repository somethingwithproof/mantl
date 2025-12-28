variable "tenancy_ocid" {
  description = "OCI tenancy OCID"
  type        = string
}

variable "user_ocid" {
  description = "OCI user OCID"
  type        = string
}

variable "fingerprint" {
  description = "OCI API key fingerprint"
  type        = string
}

variable "private_key_path" {
  description = "Path to OCI API private key"
  type        = string
  default     = "~/.oci/oci_api_key.pem"
}

variable "region" {
  description = "OCI region (e.g., us-ashburn-1, us-phoenix-1)"
  type        = string
  default     = "us-ashburn-1"
}

variable "compartment_id" {
  description = "OCI compartment OCID for resources"
  type        = string
}

variable "cluster_name" {
  description = "Name of the OKE cluster"
  type        = string
  default     = "mantl-oke"
}

variable "kubernetes_version" {
  description = "Kubernetes version for OKE cluster"
  type        = string
  default     = "v1.31.1"
}

variable "node_count" {
  description = "Number of worker nodes across all ADs"
  type        = number
  default     = 3
}

variable "node_shape" {
  description = "Shape for worker nodes (e.g., VM.Standard.E4.Flex)"
  type        = string
  default     = "VM.Standard.E4.Flex"
}

variable "node_ocpus" {
  description = "Number of OCPUs per node (for Flex shapes)"
  type        = number
  default     = 2
}

variable "node_memory_gb" {
  description = "Memory in GB per node (for Flex shapes)"
  type        = number
  default     = 16
}

variable "node_image_id" {
  description = "OCID of the node image (Oracle Linux 8 recommended)"
  type        = string
  default     = "" # Must be provided or use data source
}
