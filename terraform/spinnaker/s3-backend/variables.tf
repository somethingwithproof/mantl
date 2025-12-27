variable "region" {
  description = "AWS region for the S3 bucket"
  type        = string
}

variable "bucket_name_prefix" {
  description = "Prefix for unique S3 bucket name"
  type        = string
  default     = "mantl-spinnaker-front50-"
}

variable "force_destroy" {
  description = "Allow destroy even if bucket not empty (use with caution)"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Tags to apply to the bucket"
  type        = map(string)
  default     = {}
}
