# Mantl

## Purpose
Cloud-native platform for deploying globally distributed services with built-in Compliance-as-Code (SOC2, HIPAA, PCI-DSS, CIS).

## Stack
- Python 3.13+ for tooling and tests
- Terraform / OpenTofu for infrastructure
- Kubernetes (ArgoCD, Helm, Kustomize)
- Docker, Makefile-driven workflows
- Observability: Prometheus, Loki, Tempo, OpenTelemetry, Tetragon (eBPF)

## Build / Test
```bash
make install-dev                # Local kind cluster + full platform
pytest tests/ -v                # Python tests
ruff check .                    # Lint
ruff format --check .           # Format check
```

## Standards
- Python: 3.13+, ruff for lint/format (line-length 100), strict pytest
- Terraform: OpenTofu preferred, `tofu validate` + `tflint`
- Kubernetes manifests validated via tests
- Conventional Commits

## Conventions
- `platform/` for Kubernetes platform components
- `infra/` and `infrastructure/` for IaC
- `compliance/` for compliance frameworks
- `apps/` for example applications
- `tests/` for Python test suite
- `pyproject.toml` configures ruff, black, pytest, coverage
