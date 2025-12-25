# GCE Terraform (mantl/terraform/gce)

Hardening for Terraform 1.x in progress.

- Introduced `allowed_cidrs` and applied to external firewall.
- Removed legacy `remote-exec` provisioners.
- CI workflow is path-scoped. Next phase will modernize network resources to current provider schema.

Inputs (addition):
- `allowed_cidrs` (list(string)) – CIDRs for external access (default `["0.0.0.0/0"]`).
