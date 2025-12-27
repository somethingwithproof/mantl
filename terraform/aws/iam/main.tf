variable "short_name" { default = "mantl" }

# Build scoped ARNs for use in policies (for ECR repository resources)
# These data sources require only standard AWS credentials during plan/apply.
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

locals {
  ecr_repo_arn = "arn:${data.aws_partition.current.partition}:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*"
}

resource "aws_iam_instance_profile" "control_profile" {
  name = "${var.short_name}-control-profile"
  role = aws_iam_role.control_role.name
}

resource "aws_iam_role_policy" "control_policy" {
  name   = "${var.short_name}-control-policy"
  role   = aws_iam_role.control_role.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "EC2ReadOnly",
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeTags",
        "ec2:DescribeAvailabilityZones",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeVpcs"
      ],
      "Resource": "*"
    },
    {
      "Sid": "ELBDescribe",
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:DescribeLoadBalancers",
        "elasticloadbalancing:DescribeTargetGroups",
        "elasticloadbalancing:DescribeTargetHealth",
        "elasticloadbalancing:DescribeListeners"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "control_role" {
  name               = "${var.short_name}-control-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Service": "ec2.amazonaws.com"},
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

output "control_iam_instance_profile" {
  value = aws_iam_instance_profile.control_profile.name
}

resource "aws_iam_instance_profile" "worker_profile" {
  name = "${var.short_name}-worker-profile"
  role = aws_iam_role.worker_role.name
}

resource "aws_iam_role_policy" "worker_policy" {
  name   = "${var.short_name}-worker-policy"
  role   = aws_iam_role.worker_role.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "EC2ReadOnly",
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeTags",
        "ec2:DescribeAvailabilityZones",
        "ec2:DescribeVolumes"
      ],
      "Resource": "*"
    },
    {
      "Sid": "ECRPull",
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken"
      ],
      "Resource": "*"
    },
    {
      "Sid": "ECRPullScoped",
      "Effect": "Allow",
      "Action": [
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:GetRepositoryPolicy",
        "ecr:DescribeRepositories",
        "ecr:ListImages",
        "ecr:BatchGetImage"
      ],
      "Resource": "${local.ecr_repo_arn}"
    }
  ]
}
EOF
}

resource "aws_iam_role" "worker_role" {
  name               = "${var.short_name}-worker-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Service": "ec2.amazonaws.com"},
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

output "worker_iam_instance_profile" {
  value = aws_iam_instance_profile.worker_profile.name
}
