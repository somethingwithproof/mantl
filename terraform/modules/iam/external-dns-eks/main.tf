terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = { source = "hashicorp/aws", version = ">= 5.0" }
  }
}

variable "cluster_oidc_provider_arn" { type = string }
variable "cluster_oidc_issuer" { type = string }
variable "namespace" { type = string }
variable "service_account" { type = string }
variable "role_name" { type = string }
variable "zone_ids" { type = list(string) }

data "aws_iam_policy_document" "trust" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals { type = "Federated" identifiers = [var.cluster_oidc_provider_arn] }
    condition {
      test     = "StringEquals"
      variable = "${replace(var.cluster_oidc_issuer, "https://", "")}:sub"
      values   = ["system:serviceaccount:${var.namespace}:${var.service_account}"]
    }
  }
}

resource "aws_iam_role" "this" {
  name               = var.role_name
  assume_role_policy = data.aws_iam_policy_document.trust.json
}

data "aws_iam_policy_document" "route53" {
  statement {
    actions   = ["route53:ChangeResourceRecordSets"]
    resources = [for z in var.zone_ids : "arn:aws:route53:::hostedzone/${z}"]
  }
  statement {
    actions   = ["route53:ListHostedZones", "route53:ListResourceRecordSets"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "route53" {
  name   = "external-dns-route53"
  policy = data.aws_iam_policy_document.route53.json
}

resource "aws_iam_role_policy_attachment" "attach" {
  role       = aws_iam_role.this.name
  policy_arn = aws_iam_policy.route53.arn
}

output "role_arn" { value = aws_iam_role.this.arn }
