OIDC, TLS, and Global DNS Failover
==================================

OIDC for Gate
-------------
Use `addons/spinnaker/overlays/oidc` to enable OIDC for Gate. Provide env vars `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, and replace `OIDC_ISSUER_DOMAIN` in the patch.

TLS for Gate/Deck
-----------------
Use `addons/spinnaker/overlays/ingress-nginx-tls` with pre-created secrets `deck-tls` and `gate-tls` in the `spinnaker` namespace.

Global DNS Failover (Route53)
-----------------------------
Use `terraform/modules/global-dns/route53-failover` to create a DNS record that fails over between two endpoints with HTTP health checks (e.g., Gate or an app public endpoint). Provide `zone_id`, `record_name`, `primary_target`, `secondary_target`, and health check FQDNs.
