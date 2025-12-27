output "front50_s3_bucket" {
  description = "S3 bucket name for Spinnaker Front50"
  value       = aws_s3_bucket.spinnaker_front50.bucket
}

output "front50_s3_arn" {
  description = "S3 bucket ARN"
  value       = aws_s3_bucket.spinnaker_front50.arn
}
