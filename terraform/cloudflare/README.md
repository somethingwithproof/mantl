# Cloudflare Terraform (mantl/terraform/cloudflare)

Modernized for Terraform 1.x using the Cloudflare provider v5.x.

- Records created with `cloudflare_dns_record`.
- Zone resolved via `data.cloudflare_zone` with filter.
- Inputs are typed lists for IPs; ordering is preserved using `for_each` with indexed maps.
- CI workflow validates only this module.

Inputs:
- `domain` (string) – Cloudflare zone name.
- `short_name` (string) – DNS label prefix.
- `subdomain` (string, default "") – suffix like `.prod.example.com`.
- `control_ips`, `worker_ips`, `kubeworker_ips`, `edge_ips` (list(string)).
- `control_subdomain` (string, default `control`).
- `cloudflare_api_token` (sensitive) – provider auth (set via env or TF var).
