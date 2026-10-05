# Archived Nomad Terraform security limits

This module is archived and is not part of the supported Kubernetes platform.
Do not apply it as evidence of production readiness. The archived configuration
also lacks the referenced server/client user-data templates; standalone Terraform
validation fails on those pre-existing missing files. Launch templates explicitly
disable public IP association and retain their role-specific security groups on
the primary network interface; Auto Scaling groups use private subnets.

The internal Network Load Balancer passes Nomad TCP traffic on port 4646. AWS's
S3 NLB access logs cover TLS listeners and TLS requests only; enabling that setting
would not log this TCP listener. Changing the listener protocol alone would also
require a certificate and a reviewed Nomad client/server TLS contract. Before any
reuse, implement independently verified VPC Flow Logs and application audit logs,
retention and access controls, and review a complete TLS deployment. No load
balancer request-logging coverage is claimed by this archive.

[AWS NLB access logging limitations](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-access-logs.html).
