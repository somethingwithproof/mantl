variable "region" {
  type        = string
  description = "AWS region"
}

variable "bucket_name" {
  type        = string
  description = "S3 bucket name for Front50"
}

variable "iam_user_name" {
  type        = string
  description = "IAM user name for Front50 access"
  default     = "spinnaker-front50"
}

variable "force_destroy" {
  type        = bool
  description = "Allow terraform to destroy non-empty bucket"
  default     = false
}

variable "sse_algorithm" {
  type        = string
  description = "SSE algorithm (AES256 or aws:kms)"
  default     = "AES256"
}
