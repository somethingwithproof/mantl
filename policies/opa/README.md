# Conftest Policies

This directory contains example OPA/Conftest policies for Dockerfiles, Terraform, and Kubernetes manifests.

Run locally:

```
docker run --rm -v "$PWD":/project openpolicyagent/conftest test -p policies/opa Dockerfile
```
