terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.67.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.38.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.3.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.4.0"
    }
  }

  # Uncomment to use S3 backend
  # backend "s3" {
  #   bucket         = "mantl-terraform-state"
  #   key            = "eks/terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "mantl-terraform-locks"
  # }
}
