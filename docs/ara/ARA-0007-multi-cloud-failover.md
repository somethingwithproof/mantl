# ARA-0007 Multi-cloud and failover
Status: Accepted
Date: 2025-12-27

Context
Mantl’s spirit is multi-cloud and globally distributed services.

Decision
- Provide cluster blueprints for AWS/GCP/Azure/DO/Linode/on-prem.
- Use DNS-based global routing and health checks for active/active or active/passive.
- Replicate only the data stores that justify cross-region/cloud complexity.

Consequences
- Resilience to regional/provider outages
- Higher cost/complexity where data replication is required

Alternatives
- Single cloud with regional redundancy only
