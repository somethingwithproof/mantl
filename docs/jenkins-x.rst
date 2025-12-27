Jenkins X Integration (Optional)
===============================

Overview
--------
Mantl can integrate with Jenkins X for GitOps-style CI/CD on Kubernetes. This is optional; GitHub Actions remains the default CI path.

Prerequisites
-------------
- A Kubernetes cluster (kind, or managed)
- `jx` CLI installed
- Git repository with appropriate permissions

Bootstrap (high level)
----------------------
1. Install Jenkins X operator into your cluster.
2. Connect your Git repository and configure environments (staging/production).
3. Configure promotion via Pull Requests.

Notes
-----
- Use short-lived credentials and OIDC where possible.
- See Jenkins X documentation for provider-specific boot commands and requirements.