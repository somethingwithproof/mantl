# ARA-0001 Kubernetes-first platform
Status: Accepted
Date: 2025-12-27

Context
Mantl historically supported Mesos/Marathon and Swarm. Those ecosystems are deprecated and fragmented.

Decision
Adopt Kubernetes as the sole orchestrator. All platform and app primitives target Kubernetes APIs.

Consequences
- Simplifies platform surface area and talent hiring
- Enables reuse of CNCF ecosystem (Helm, Kustomize, Gateway API)
- Deprecates Mesos/Marathon/Chronos/Swarm; archived under `legacy/`

Alternatives
- Keep multi-orchestrator support: higher maintenance, limited value in 2026.
