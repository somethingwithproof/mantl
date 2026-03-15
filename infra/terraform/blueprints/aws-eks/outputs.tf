################################################################################
# Cluster
################################################################################

output "cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "EKS cluster API endpoint"
  value       = module.eks.cluster_endpoint
  sensitive   = true
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded cluster CA certificate"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "cluster_oidc_issuer_url" {
  description = "OIDC issuer URL for the cluster"
  value       = module.eks.cluster_oidc_issuer_url
  sensitive   = true
}

output "cluster_oidc_provider_arn" {
  description = "OIDC provider ARN for IRSA"
  value       = module.eks.oidc_provider_arn
}

output "cluster_version" {
  description = "Kubernetes version of the cluster"
  value       = module.eks.cluster_version
}

################################################################################
# VPC
################################################################################

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnets
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = module.vpc.public_subnets
}

################################################################################
# IRSA Role ARNs
################################################################################

output "external_dns_role_arn" {
  description = "IAM role ARN for External DNS"
  value       = module.external_dns_irsa.iam_role_arn
}

output "external_secrets_role_arn" {
  description = "IAM role ARN for External Secrets"
  value       = module.external_secrets_irsa.iam_role_arn
}

output "cert_manager_role_arn" {
  description = "IAM role ARN for cert-manager"
  value       = module.cert_manager_irsa.iam_role_arn
}

output "cluster_autoscaler_role_arn" {
  description = "IAM role ARN for Cluster Autoscaler"
  value       = module.cluster_autoscaler_irsa.iam_role_arn
}

output "ebs_csi_role_arn" {
  description = "IAM role ARN for EBS CSI Driver"
  value       = module.ebs_csi_irsa.iam_role_arn
}

output "load_balancer_controller_role_arn" {
  description = "IAM role ARN for AWS Load Balancer Controller"
  value       = module.load_balancer_controller_irsa.iam_role_arn
}

################################################################################
# kubectl Configuration
################################################################################

output "configure_kubectl" {
  description = "Command to configure kubectl"
  value       = "aws eks update-kubeconfig --name ${module.eks.cluster_name} --region ${var.region}"
}

output "platform_irsa_annotations" {
  description = "ServiceAccount annotations for platform components"
  value = {
    external_dns = {
      "eks.amazonaws.com/role-arn" = module.external_dns_irsa.iam_role_arn
    }
    external_secrets = {
      "eks.amazonaws.com/role-arn" = module.external_secrets_irsa.iam_role_arn
    }
    cert_manager = {
      "eks.amazonaws.com/role-arn" = module.cert_manager_irsa.iam_role_arn
    }
    cluster_autoscaler = {
      "eks.amazonaws.com/role-arn" = module.cluster_autoscaler_irsa.iam_role_arn
    }
  }
}
