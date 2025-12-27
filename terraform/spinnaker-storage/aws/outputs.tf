output "front50" {
  value = {
    bucket               = aws_s3_bucket.front50.bucket
    region               = var.region
    access_key_id        = aws_iam_access_key.front50.id
    secret_access_key    = aws_iam_access_key.front50.secret
  }
  sensitive = true
}
