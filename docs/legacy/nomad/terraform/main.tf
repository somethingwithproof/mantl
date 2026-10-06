# SPDX-License-Identifier: Apache-2.0
# HashiCorp Nomad Cluster Infrastructure
# OpenTofu configuration for AWS deployment
# https://developer.hashicorp.com/nomad/tutorials/enterprise/production-reference-architecture-vm-with-consul

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  backend "s3" {
    bucket         = "mantl-terraform-state"
    key            = "nomad/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    dynamodb_table = "mantl-terraform-locks"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "mantl"
      Component   = "nomad"
      ManagedBy   = "opentofu"
      Environment = var.environment
    }
  }
}

# Variables
variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "vpc_id" {
  description = "VPC ID for Nomad cluster"
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs for Nomad nodes"
  type        = list(string)
}

variable "nomad_server_count" {
  description = "Number of Nomad servers (should be 3 or 5)"
  type        = number
  default     = 3
}

variable "nomad_client_count" {
  description = "Number of Nomad clients"
  type        = number
  default     = 6
}

variable "nomad_version" {
  description = "Nomad version to install"
  type        = string
  default     = "1.7.3"
}

variable "consul_version" {
  description = "Consul version to install"
  type        = string
  default     = "1.17.1"
}

variable "kms_key_arns" {
  description = "ARNs of KMS keys that Nomad nodes can use for encryption"
  type        = list(string)
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Random resources
resource "random_id" "gossip_key" {
  byte_length = 32
}

# TLS certificates
resource "tls_private_key" "ca" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "ca" {
  private_key_pem = tls_private_key.ca.private_key_pem

  subject {
    common_name  = "Mantl Nomad CA"
    organization = "Mantl Platform"
  }

  validity_period_hours = 87600 # 10 years
  is_ca_certificate     = true

  allowed_uses = [
    "cert_signing",
    "crl_signing",
    "digital_signature",
    "key_encipherment",
  ]
}

resource "tls_private_key" "server" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_cert_request" "server" {
  private_key_pem = tls_private_key.server.private_key_pem

  subject {
    common_name  = "server.global.nomad"
    organization = "Mantl Platform"
  }

  dns_names = [
    "server.global.nomad",
    "localhost",
    "*.server.global.nomad",
  ]

  ip_addresses = ["127.0.0.1"]
}

resource "tls_locally_signed_cert" "server" {
  cert_request_pem   = tls_cert_request.server.cert_request_pem
  ca_private_key_pem = tls_private_key.ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca.cert_pem

  validity_period_hours = 8760 # 1 year

  allowed_uses = [
    "digital_signature",
    "key_encipherment",
    "server_auth",
    "client_auth",
  ]
}

resource "tls_private_key" "client" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_cert_request" "client" {
  private_key_pem = tls_private_key.client.private_key_pem

  subject {
    common_name  = "client.global.nomad"
    organization = "Mantl Platform"
  }

  dns_names = [
    "client.global.nomad",
    "localhost",
  ]

  ip_addresses = ["127.0.0.1"]
}

resource "tls_locally_signed_cert" "client" {
  cert_request_pem   = tls_cert_request.client.cert_request_pem
  ca_private_key_pem = tls_private_key.ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca.cert_pem

  validity_period_hours = 8760

  allowed_uses = [
    "digital_signature",
    "key_encipherment",
    "client_auth",
  ]
}

# Security Groups
resource "aws_security_group" "nomad_server" {
  name_prefix = "nomad-server-"
  vpc_id      = var.vpc_id
  description = "Security group for Nomad servers"

  # HTTP API
  ingress {
    from_port   = 4646
    to_port     = 4646
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Nomad HTTP API"
  }

  # RPC
  ingress {
    from_port   = 4647
    to_port     = 4647
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Nomad RPC"
  }

  # Serf WAN
  ingress {
    from_port   = 4648
    to_port     = 4648
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Nomad Serf TCP"
  }

  ingress {
    from_port   = 4648
    to_port     = 4648
    protocol    = "udp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Nomad Serf UDP"
  }

  # Consul ports
  ingress {
    from_port   = 8300
    to_port     = 8302
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Consul RPC/Serf"
  }

  ingress {
    from_port   = 8301
    to_port     = 8302
    protocol    = "udp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Consul Serf UDP"
  }

  ingress {
    from_port   = 8500
    to_port     = 8502
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Consul HTTP/gRPC"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = {
    Name = "nomad-server-sg"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_security_group" "nomad_client" {
  name_prefix = "nomad-client-"
  vpc_id      = var.vpc_id
  description = "Security group for Nomad clients"

  # Nomad client ports
  ingress {
    from_port   = 4646
    to_port     = 4648
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Nomad client ports"
  }

  # Dynamic ports for services
  ingress {
    from_port   = 20000
    to_port     = 32000
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Dynamic service ports"
  }

  # Consul client ports
  ingress {
    from_port   = 8300
    to_port     = 8302
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Consul ports"
  }

  ingress {
    from_port   = 8301
    to_port     = 8302
    protocol    = "udp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Consul Serf UDP"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = {
    Name = "nomad-client-sg"
  }

  lifecycle {
    create_before_destroy = true
  }
}

# IAM Role for Nomad nodes
resource "aws_iam_role" "nomad" {
  name_prefix = "nomad-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "nomad" {
  name_prefix = "nomad-"
  role        = aws_iam_role.nomad.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ec2:DescribeInstances",
          "ec2:DescribeTags",
          "autoscaling:DescribeAutoScalingGroups",
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:Encrypt",
          "kms:GenerateDataKey",
        ]
        Resource = var.kms_key_arns
      }
    ]
  })
}

resource "aws_iam_instance_profile" "nomad" {
  name_prefix = "nomad-"
  role        = aws_iam_role.nomad.name
}

# Launch template for Nomad servers
resource "aws_launch_template" "nomad_server" {
  name_prefix   = "nomad-server-"
  image_id      = data.aws_ami.ubuntu.id
  instance_type = "m6i.large"

  iam_instance_profile {
    name = aws_iam_instance_profile.nomad.name
  }

  network_interfaces {
    device_index                = 0
    associate_public_ip_address = false
    security_groups             = [aws_security_group.nomad_server.id]
  }

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size           = 100
      volume_type           = "gp3"
      iops                  = 3000
      throughput            = 125
      encrypted             = true
      delete_on_termination = true
    }
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }

  monitoring {
    enabled = true
  }

  user_data = base64encode(templatefile("${path.module}/templates/nomad-server-userdata.sh.tpl", {
    nomad_version  = var.nomad_version
    consul_version = var.consul_version
    gossip_key     = base64encode(random_id.gossip_key.hex)
    ca_cert        = tls_self_signed_cert.ca.cert_pem
    server_cert    = tls_locally_signed_cert.server.cert_pem
    server_key     = tls_private_key.server.private_key_pem
    server_count   = var.nomad_server_count
    aws_region     = var.aws_region
    environment    = var.environment
  }))

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name       = "nomad-server"
      nomad-role = "server"
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Auto Scaling Group for Nomad servers
resource "aws_autoscaling_group" "nomad_server" {
  name_prefix         = "nomad-server-"
  desired_capacity    = var.nomad_server_count
  min_size            = var.nomad_server_count
  max_size            = var.nomad_server_count
  vpc_zone_identifier = var.private_subnet_ids

  launch_template {
    id      = aws_launch_template.nomad_server.id
    version = "$Latest"
  }

  health_check_type         = "EC2"
  health_check_grace_period = 300

  instance_refresh {
    strategy = "Rolling"
    preferences {
      min_healthy_percentage = 66
    }
  }

  tag {
    key                 = "Name"
    value               = "nomad-server"
    propagate_at_launch = true
  }

  tag {
    key                 = "nomad-role"
    value               = "server"
    propagate_at_launch = true
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Launch template for Nomad clients
resource "aws_launch_template" "nomad_client" {
  name_prefix   = "nomad-client-"
  image_id      = data.aws_ami.ubuntu.id
  instance_type = "m6i.xlarge"

  iam_instance_profile {
    name = aws_iam_instance_profile.nomad.name
  }

  network_interfaces {
    device_index                = 0
    associate_public_ip_address = false
    security_groups             = [aws_security_group.nomad_client.id]
  }

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size           = 200
      volume_type           = "gp3"
      iops                  = 6000
      throughput            = 250
      encrypted             = true
      delete_on_termination = true
    }
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }

  monitoring {
    enabled = true
  }

  user_data = base64encode(templatefile("${path.module}/templates/nomad-client-userdata.sh.tpl", {
    nomad_version  = var.nomad_version
    consul_version = var.consul_version
    gossip_key     = base64encode(random_id.gossip_key.hex)
    ca_cert        = tls_self_signed_cert.ca.cert_pem
    client_cert    = tls_locally_signed_cert.client.cert_pem
    client_key     = tls_private_key.client.private_key_pem
    aws_region     = var.aws_region
    environment    = var.environment
  }))

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name       = "nomad-client"
      nomad-role = "client"
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Auto Scaling Group for Nomad clients
resource "aws_autoscaling_group" "nomad_client" {
  name_prefix         = "nomad-client-"
  desired_capacity    = var.nomad_client_count
  min_size            = 3
  max_size            = 20
  vpc_zone_identifier = var.private_subnet_ids

  launch_template {
    id      = aws_launch_template.nomad_client.id
    version = "$Latest"
  }

  health_check_type         = "EC2"
  health_check_grace_period = 300

  instance_refresh {
    strategy = "Rolling"
    preferences {
      min_healthy_percentage = 80
    }
  }

  tag {
    key                 = "Name"
    value               = "nomad-client"
    propagate_at_launch = true
  }

  tag {
    key                 = "nomad-role"
    value               = "client"
    propagate_at_launch = true
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Network Load Balancer for Nomad servers
resource "aws_lb" "nomad" {
  name_prefix        = "nomad-"
  internal           = true
  load_balancer_type = "network"
  subnets            = var.private_subnet_ids

  enable_cross_zone_load_balancing = true

  tags = {
    Name = "nomad-nlb"
  }
}

resource "aws_lb_target_group" "nomad_http" {
  name_prefix = "nomad-"
  port        = 4646
  protocol    = "TCP"
  vpc_id      = var.vpc_id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 30
    port                = 4646
    protocol            = "TCP"
  }
}

resource "aws_lb_listener" "nomad_http" {
  load_balancer_arn = aws_lb.nomad.arn
  port              = 4646
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.nomad_http.arn
  }
}

resource "aws_autoscaling_attachment" "nomad_server" {
  autoscaling_group_name = aws_autoscaling_group.nomad_server.name
  lb_target_group_arn    = aws_lb_target_group.nomad_http.arn
}

# Outputs
output "nomad_lb_dns" {
  description = "DNS name of the Nomad load balancer"
  value       = aws_lb.nomad.dns_name
}

output "nomad_server_asg" {
  description = "Name of the Nomad server Auto Scaling Group"
  value       = aws_autoscaling_group.nomad_server.name
}

output "nomad_client_asg" {
  description = "Name of the Nomad client Auto Scaling Group"
  value       = aws_autoscaling_group.nomad_client.name
}

output "ca_cert" {
  description = "CA certificate for Nomad TLS"
  value       = tls_self_signed_cert.ca.cert_pem
  sensitive   = true
}
