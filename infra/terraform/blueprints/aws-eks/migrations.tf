# IAM 5-to-6 policy consolidation. Role resource addresses remain unchanged.
# Inspect the plan for policy replacements before upgrading deployed clusters.

moved {
  from = module.external_dns_irsa.aws_iam_policy.external_dns[0]
  to   = module.external_dns_irsa.aws_iam_policy.this[0]
}

moved {
  from = module.external_dns_irsa.aws_iam_role_policy_attachment.external_dns[0]
  to   = module.external_dns_irsa.aws_iam_role_policy_attachment.this[0]
}

moved {
  from = module.external_secrets_irsa.aws_iam_policy.external_secrets[0]
  to   = module.external_secrets_irsa.aws_iam_policy.this[0]
}

moved {
  from = module.external_secrets_irsa.aws_iam_role_policy_attachment.external_secrets[0]
  to   = module.external_secrets_irsa.aws_iam_role_policy_attachment.this[0]
}

moved {
  from = module.cert_manager_irsa.aws_iam_policy.cert_manager[0]
  to   = module.cert_manager_irsa.aws_iam_policy.this[0]
}

moved {
  from = module.cert_manager_irsa.aws_iam_role_policy_attachment.cert_manager[0]
  to   = module.cert_manager_irsa.aws_iam_role_policy_attachment.this[0]
}

moved {
  from = module.cluster_autoscaler_irsa.aws_iam_policy.cluster_autoscaler[0]
  to   = module.cluster_autoscaler_irsa.aws_iam_policy.this[0]
}

moved {
  from = module.cluster_autoscaler_irsa.aws_iam_role_policy_attachment.cluster_autoscaler[0]
  to   = module.cluster_autoscaler_irsa.aws_iam_role_policy_attachment.this[0]
}

moved {
  from = module.ebs_csi_irsa.aws_iam_policy.ebs_csi[0]
  to   = module.ebs_csi_irsa.aws_iam_policy.this[0]
}

moved {
  from = module.ebs_csi_irsa.aws_iam_role_policy_attachment.ebs_csi[0]
  to   = module.ebs_csi_irsa.aws_iam_role_policy_attachment.this[0]
}

moved {
  from = module.load_balancer_controller_irsa.aws_iam_policy.load_balancer_controller[0]
  to   = module.load_balancer_controller_irsa.aws_iam_policy.this[0]
}

moved {
  from = module.load_balancer_controller_irsa.aws_iam_role_policy_attachment.load_balancer_controller[0]
  to   = module.load_balancer_controller_irsa.aws_iam_role_policy_attachment.this[0]
}
