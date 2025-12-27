# Spinnaker Docker Registry Accounts

This overlay configures dockerRegistry artifact accounts for multiple registries using
Kubernetes Secret -> env var mapping with prefixes.

Secrets (via ESO)
- spinnaker-docker-registry (Docker Hub or custom)
- spinnaker-ecr-registry
- spinnaker-gcr-registry
- spinnaker-ghcr-registry

Env mapping (clouddriver)
- Each Secret is loaded with an envFrom prefix (DOCKER_, ECR_, GCR_, GHCR_)
- Keys in the secrets: registry, username, password (or token/json_key)
- SpinnakerService references variables like ${DOCKER_registry}, ${GHCR_username}

Apply
```
kubectl apply -k addons/spinnaker/overlays/docker-registries
```