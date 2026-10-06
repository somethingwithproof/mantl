<!-- SPDX-License-Identifier: Apache-2.0 -->
# AWS Terraform (mantl/terraform/aws)

This module has been partially modernized for Terraform 1.x. Changes include:

- Provider and version constraints added (Terraform >= 1.5, aws >= 5).
- IMDSv2 required on EC2 instances via `metadata_options`.
- EBS encryption enabled by default on all root and data volumes; optionally supply a customer-managed KMS key via `kms_key_id`.
- Security groups now accept `allowed_cidrs` (default `["*******/0"]`). Override to restrict access.
- Tags use `tags = {}` maps (legacy `tags {}` blocks are removed).
- Optional worker Auto Scaling using Launch Template + ASG via `use_autoscaling` (default `false`). When enabled, worker instances are created by ASG; when disabled, legacy `aws_instance` resources are used.

Notes and migration:
- This is a stepwise modernization; some legacy constructs remain to preserve behavior.
- Consider migrating control/edge groups to Launch Templates in a follow-up.
- Replace hard-coded AMIs with `data "aws_ami"` filters where possible.

Inputs (additions):
- `allowed_cidrs` (list(string)) – CIDRs for ingress.
- `kms_key_id` (string, optional) – KMS key ID or ARN to use for EBS encryption. If unset, the AWS-managed key for EBS is used.
- `use_autoscaling` (bool) – enable LT/ASG for workers.
- `use_autoscaling_control` (bool) – enable LT/ASG for control nodes.
- `use_autoscaling_edge` (bool) – enable LT/ASG for edge nodes.

Example to use a customer-managed key:

```hcl path=null start=null
kms_key_id = "arn:aws:kms:us-east-1:123456789012:key/abcd-1234-efgh-5678"
```
