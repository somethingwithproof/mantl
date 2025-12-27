variable "project" {
  description = "GCP project id"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "location" {
  description = "Bucket location (region or multi-region, e.g., US)"
  type        = string
  default     = "US"
}

variable "bucket_name" {
  description = "Globally unique bucket name"
  type        = string
}

variable "force_destroy" {
  description = "Allow destroy even if bucket not empty"
  type        = bool
  default     = false
}
