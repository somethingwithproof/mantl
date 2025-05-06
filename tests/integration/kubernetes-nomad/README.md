# Kubernetes-Nomad Integration Tests

This directory contains Docker-based integration tests for the Kubernetes-Nomad integration in Mantl.

## Overview

The tests verify the comprehensive integration between Kubernetes and HashiCorp Nomad in Mantl, focusing on ten key enterprise features:

1. **Cross-platform Service Discovery**: Tests automatic service discovery between Kubernetes and Nomad
2. **Shared Secrets Management**: Tests Vault integration for secure secrets management
3. **Federated Metrics and Logging**: Tests integrated observability across platforms
4. **Multi-region/Multi-cluster Orchestration**: Tests workload distribution across regions
5. **GPU Workload Scheduling**: Tests hardware acceleration support
6. **Autoscaling Integration**: Tests dynamic workload scaling
7. **Custom Resource Definition Support**: Tests Kubernetes CRD integration
8. **Disaster Recovery with Automatic Failover**: Tests high availability features
9. **Zero-downtime Migration Tools**: Tests service mesh integration
10. **Unified Policy Enforcement with OPA**: Tests Open Policy Agent integration

## Architecture

The test environment consists of:
- 1 control node running both Kubernetes control plane and Nomad server
- 2 worker nodes running both Kubernetes and Nomad clients
- 1 test coordinator node that runs the integration tests

## Prerequisites

- Docker and Docker Compose
- At least 8GB of RAM
- At least 4 CPU cores

## Running the Tests

To run the full test suite:

```bash
cd /Users/thomasvincent/mantl/tests/integration/kubernetes-nomad
docker-compose up --build
```

To run individual test files:

```bash
cd /Users/thomasvincent/mantl/tests/integration/kubernetes-nomad
docker-compose up --build control worker1 worker2 -d
docker-compose run test pytest -v test/test_service_discovery.py
```

## Test Results

Test results are saved in the `results` directory with a full HTML report. You can open `results/report.html` in a browser to view detailed test results.

## Test Files

- `test_service_discovery.py`: Tests cross-platform service discovery
- `test_vault_integration.py`: Tests shared secrets management with Vault
- `test_metrics_logging.py`: Tests federated metrics and logging
- `test_multi_region.py`: Tests multi-region orchestration
- `test_gpu_support.py`: Tests GPU workload scheduling
- `test_autoscaling.py`: Tests autoscaling integration
- `test_crd.py`: Tests custom resource definition support
- `test_disaster_recovery.py`: Tests disaster recovery with automatic failover
- `test_service_mesh.py`: Tests service mesh integration
- `test_opa.py`: Tests unified policy enforcement with OPA
- `test_all_features.py`: Tests that all features work together correctly

## Docker Configuration

- `Dockerfile.control`: Control node with Kubernetes control plane, Nomad server, Consul, and Vault
- `Dockerfile.worker`: Worker node with Kubernetes node and Nomad client
- `Dockerfile.test`: Test environment with testing dependencies
- `docker-compose.yml`: Orchestrates the test environment

## Scripts

- `control-entrypoint.sh`: Initializes the control node services
- `worker-entrypoint.sh`: Initializes the worker node services
- `run-tests.sh`: Runs the integration tests

## Debugging

If tests fail, you can connect to the containers for debugging:

```bash
docker exec -it mantl-control /bin/bash
docker exec -it mantl-worker1 /bin/bash
docker exec -it mantl-worker2 /bin/bash
docker exec -it mantl-test /bin/bash
```

Log files are available in the containers at:
- Kubernetes: `/var/log/kubernetes/`
- Nomad: `/var/log/nomad/`
- Consul: `/var/log/consul.log`
- Vault: `/var/log/vault.log`
- OPA: `/var/log/opa.log`