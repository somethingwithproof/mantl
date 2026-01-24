# End-to-End Test Suite

Comprehensive end-to-end tests for the mantl platform, covering Terraform validation, cluster security, platform components, disaster recovery, and performance.

## Prerequisites

- Python 3.11+
- `kubectl` configured with access to a Kubernetes cluster
- `terraform` CLI (for Terraform validation tests)
- Cluster with platform components deployed (for platform/DR/performance tests)

## Installation

```bash
# Install test dependencies
uv pip install -r requirements.txt

# Or with pip
pip install -r requirements.txt
```

## Test Structure

```
e2e/
├── terraform/           # Terraform blueprint validation
├── security/            # Cluster security verification
├── platform/            # Platform component integration
├── dr/                  # Disaster recovery capabilities
├── performance/         # Performance benchmarks
├── conftest.py          # Pytest configuration
├── requirements.txt     # Test dependencies
└── README.md            # This file
```

## Running Tests

### Run All Tests

```bash
pytest tests/e2e/
```

### Run Specific Test Categories

```bash
# Terraform validation (no cluster required)
pytest tests/e2e/terraform/ -v

# Security tests (requires cluster)
pytest tests/e2e/security/ -v

# Platform component tests (requires deployed components)
pytest tests/e2e/platform/ -v

# Disaster recovery tests (requires Velero, CloudNativePG)
pytest tests/e2e/dr/ -v

# Performance tests (requires cluster)
pytest tests/e2e/performance/ -v
```

### Run Tests by Marker

```bash
# Run only tests that don't require a cluster
pytest -m "not requires_cluster"

# Run only security-related tests
pytest -m security

# Skip slow tests
pytest -m "not slow"

# Run only Terraform validation
pytest -m terraform
```

### Parallel Execution

```bash
# Run tests in parallel (4 workers)
pytest tests/e2e/ -n 4

# Run specific category in parallel
pytest tests/e2e/terraform/ -n auto
```

### Generate HTML Report

```bash
pytest tests/e2e/ --html=report.html --self-contained-html
```

## Test Categories

### 1. Terraform Validation (`terraform/`)

Tests that validate Terraform blueprints for all 8 cloud providers:
- AWS EKS
- GCP GKE
- Azure AKS
- DigitalOcean DOKS
- Linode LKE
- Oracle OKE
- IBM IKS
- OpenStack

**Tests:**
- `terraform init` succeeds
- `terraform validate` succeeds
- `terraform fmt -check` passes
- Required files exist (main.tf, variables.tf)
- Kubernetes version is 1.31.x
- Security features configured
- Provider-specific features (VPC Flow Logs, KMS, etc.)

**Run without cluster:**
```bash
pytest tests/e2e/terraform/
```

### 2. Security Verification (`security/`)

Tests that verify security configurations in deployed clusters:

**Network Policies:**
- NetworkPolicy API available
- Default deny policies in protected namespaces

**RBAC:**
- RBAC API enabled
- Default ServiceAccount token mounting disabled
- cluster-admin not bound to anonymous users

**Pod Security Standards:**
- PSS labels configured on namespaces
- No privileged pods in default namespace

**Image Security:**
- Kyverno policies installed (if applicable)
- No containers using `:latest` tag

**Audit Logging:**
- Audit policy configured (provider-specific)

**Monitoring:**
- Prometheus deployed
- Falco deployed (optional)

**Run:**
```bash
pytest tests/e2e/security/ -v
```

### 3. Platform Components (`platform/`)

Tests that verify platform services work correctly:

**Monitoring Stack:**
- Prometheus server running
- Grafana running (if deployed)
- Metrics accessible

**Backup Solution:**
- Velero installed
- Backup storage location configured
- Backup schedules configured
- Test backup creation

**Security Scanning:**
- Trivy Operator installed
- Vulnerability reports generated
- Critical vulnerabilities tracked

**Cost Management:**
- OpenCost installed
- Metrics available

**Chaos Engineering:**
- Chaos Mesh installed
- Chaos experiments can be listed

**Database Operator:**
- CloudNativePG operator installed
- PostgreSQL clusters can be listed

**Progressive Delivery:**
- Flagger installed
- Canary deployments can be listed

**Run:**
```bash
pytest tests/e2e/platform/ -v
```

### 4. Disaster Recovery (`dr/`)

Tests that verify disaster recovery capabilities:

**Velero Backup/Restore:**
- Create test namespace with resources
- Create backup with Velero
- Delete namespace
- Restore from backup
- Verify resources restored

**Database Failover:**
- PostgreSQL cluster HA configured (>= 2 instances)
- Backup configured with barmanObjectStore

**High Availability:**
- Control plane components have multiple replicas
- PodDisruptionBudgets exist
- Pod anti-affinity configured for critical workloads

**Run:**
```bash
pytest tests/e2e/dr/ -v
```

### 5. Performance (`performance/`)

Tests that measure cluster performance:

**API Server:**
- API server reachable
- Response time < 1s average, < 2s p95
- List all pods performance

**Pod Startup:**
- Single pod startup < 60s
- Multiple pods (10) startup < 120s

**Network:**
- Pod-to-pod connectivity
- DNS resolution latency < 1s

**Storage:**
- PVC provisioning < 60s

**Scheduler:**
- Scheduler latency < 5s

**Resource Utilization:**
- Node CPU/memory not overutilized (< 90%)
- System pods resource usage tracked

**Run:**
```bash
pytest tests/e2e/performance/ -v
```

## Test Markers

Tests are automatically marked based on their location and characteristics:

- `terraform` - Terraform validation tests
- `security` - Security verification tests
- `platform` - Platform component tests
- `dr` - Disaster recovery tests
- `performance` - Performance tests
- `requires_cluster` - Requires access to a Kubernetes cluster
- `slow` - Tests that take longer to run
- `destructive` - Tests that may modify cluster state

## Configuration

Tests use environment variables for configuration:

```bash
# Kubernetes context (uses current context by default)
export KUBECONFIG=~/.kube/config

# Skip tests that require specific components
export SKIP_VELERO_TESTS=true
export SKIP_TRIVY_TESTS=true
export SKIP_CHAOS_MESH_TESTS=true
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  terraform-validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - name: Run Terraform tests
        run: |
          pip install -r tests/e2e/requirements.txt
          pytest tests/e2e/terraform/ -v

  cluster-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: azure/setup-kubectl@v4
      - name: Create kind cluster
        run: |
          kind create cluster --config=tests/fixtures/kind-config.yaml
      - name: Run cluster tests
        run: |
          pip install -r tests/e2e/requirements.txt
          pytest tests/e2e/security/ tests/e2e/platform/ -v
```

## Troubleshooting

### Tests Skip or Fail

**"kubectl not configured or cluster not accessible"**
- Ensure `kubectl` is installed and configured
- Verify cluster access: `kubectl cluster-info`

**"Velero not installed"**
- Tests skip if Velero is not found
- Deploy Velero or set `SKIP_VELERO_TESTS=true`

**"Trivy Operator not installed"**
- Tests skip if Trivy is not found
- Deploy Trivy or set `SKIP_TRIVY_TESTS=true`

**"Metrics server not available"**
- Performance tests require metrics-server
- Deploy: `kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml`

### Slow Tests

```bash
# Skip slow tests during development
pytest -m "not slow"

# Run only fast tests
pytest tests/e2e/security/ tests/e2e/platform/ -m "not slow"
```

### Debug Mode

```bash
# Verbose output with stdout/stderr
pytest tests/e2e/ -v -s

# Stop on first failure
pytest tests/e2e/ -x

# Show local variables on failure
pytest tests/e2e/ -l
```

## Contributing

When adding new tests:

1. Place tests in the appropriate category directory
2. Add docstrings explaining what the test verifies
3. Use appropriate markers (`@pytest.mark.slow`, `@pytest.mark.requires_cluster`)
4. Clean up resources in test fixtures or teardown
5. Use `pytest.skip()` for optional features not present in all clusters
6. Keep tests independent (no inter-test dependencies)

## Best Practices

- **Idempotent**: Tests should be runnable multiple times
- **Isolated**: Each test should be independent
- **Cleanup**: Always clean up resources (use fixtures)
- **Skip Gracefully**: Use `pytest.skip()` when prerequisites aren't met
- **Fast Feedback**: Keep tests as fast as possible
- **Clear Assertions**: Use descriptive assertion messages

## Support

For issues or questions about the test suite:
1. Check this README
2. Review test output and logs
3. Open an issue on GitHub with:
   - Test command used
   - Full test output
   - Cluster configuration
   - Platform component versions
