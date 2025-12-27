terraform {
  required_version = ">= 1.3.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

resource "aws_s3_bucket" "front50" {
  bucket        = var.bucket_name
  force_destroy = var.force_destroy
}

resource "aws_s3_bucket_versioning" "front50" {
  bucket = aws_s3_bucket.front50.id
  versioning_configuration { status = "Enabled" }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "front50" {
  bucket = aws_s3_bucket.front50.id
  rule {
    apply_server_side_encryption_by_default { sse_algorithm = var.sse_algorithm }
  }
}

resource "aws_iam_user" "front50" {
  name = var.iam_user_name
}

resource "aws_iam_access_key" "front50" {
  user = aws_iam_user.front50.name
}

data "aws_iam_policy_document" "front50" {
  statement {
    effect = "Allow"
    actions = [
      "s3:AbortMultipartUpload",
      "s3:DeleteObject",
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:ListBucket",
      "s3:ListBucketMultipartUploads",
      "s3:PutObject"
    ]
    resources = [
      aws_s3_bucket.front50.arn,
      "${aws_s3_bucket.front50.arn}/*"
    ]
  }
}

resource "aws_iam_user_policy" "front50" {
  name   = "front50-policy"
  user   = aws_iam_user.front50.name
  policy = data.aws_iam_policy_document.front50.json
}

output "bucket_name" {
  value = aws_s3_bucket.front50.bucket
}

output "front50_access_key_id" {
  value     = aws_iam_access_key.front50.id
  sensitive = true
}

output "front50_secret_access_key" {
  value     = aws_iam_access_key.front50.secret
  sensitive = true
}
