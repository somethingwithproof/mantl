Multi-Cloud and Failover
========================

Goals
-----
- Keep the “multi-cloud spirit” by supporting reproducible blueprints across cloud providers.
- Provide patterns for resilient, cross-cluster deployments and failover.

Recommended Patterns
--------------------
- DNS-based global routing (e.g., Route 53 / Cloud DNS) with health checks
- Per-cluster ingress gateways exposing health endpoints
- Replicated state via cloud-native services or data planes (choose per workload)
- Optional multi-cluster service mesh for east-west traffic

Notes
-----
- Start with active/active stateless services; add data replication as required.
- Automate runbooks for regional evacuation and failback.