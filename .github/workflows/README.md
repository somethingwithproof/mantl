# GitHub Actions CI/CD Workflows

## Overview

The mantl platform uses GitHub Actions for continuous integration and deployment. All workflows are secure and follow GitHub security best practices.

## Workflows

### 1. Terraform Validation (`terraform-validate.yml`)

**Triggers:**
- Pull requests affecting `infra/terraform/**`
- Pushes to main branch

**Jobs:**
- `validate-providers`: Validates all 8 cloud provider blueprints in parallel
  - Terraform fmt check
  - Terraform init (backend=false)
  - Terraform validate
  - Kubernetes version verification (1.31.x)
- `validate-modules`: Validates Terraform modules
- `tflint`: Runs TFLint for best practices
- `checkov`: Security scanning with Checkov

**Matrix Strategy:**
```yaml
matrix:
  provider:
    - aws-eks
    - gcp-gke
    - azure-aks
    - do-doks
    - linode-lke
    - oci-oke
    - ibm-iks
    - openstack
```

### 2. E2E Tests (`e2e-tests.yml`)

**Triggers:**
- Pull requests affecting `tests/**` or `infra/**`
- Pushes to main
- Daily at 2 AM UTC
- Manual dispatch

**Jobs:**
- `terraform-validation`: Runs pytest terraform validation tests
- `kind-cluster-tests`: Matrix of security/platform/dr/performance tests on kind
- `integration-tests`: Full integration suite (scheduled only)

**Test Categories:**
- Security: RBAC, network policies, PSS, image security
- Platform: Prometheus, Velero, Trivy, OpenCost, etc.
- DR: Backup/restore, database failover, HA
- Performance: API server, pod startup, network, storage

### 3. Security Scanning (`security-scan.yml`)

**Triggers:**
- Pull requests
- Pushes to main
- Daily at 3 AM UTC

**Scans:**
- **Trivy**: Container image and filesystem vulnerabilities
- **Checkov**: Infrastructure as code security
- **Semgrep**: Static analysis for code vulnerabilities
- **YAML Lint**: YAML file validation

**SARIF Upload:**
All results uploaded to GitHub Security tab for tracking.

### 4. Pre-commit Checks (`pre-commit.yml`)

**Triggers:**
- All pull requests

**Checks:**
- Python formatting (ruff)
- Terraform formatting
- YAML linting
- Trailing whitespace
- Large files detection
- Commit message format

### 5. Release Automation (`release.yml`)

**Triggers:**
- Tags matching `v*.*.*`

**Jobs:**
- Create GitHub release
- Generate changelog
- Build and publish Helm charts
- Update documentation
- Notify Slack/Discord

### 6. Documentation (`docs.yml`)

**Triggers:**
- Changes to `*.md` files
- Changes to Terraform files

**Jobs:**
- Generate terraform-docs
- Update API documentation
- Deploy to GitHub Pages
- Check for broken links

## Security Practices

### No Untrusted Input in Commands

❌ **UNSAFE:**
```yaml
- run: echo "${{ github.event.issue.title }}"
```

✅ **SAFE:**
```yaml
- env:
    TITLE: ${{ github.event.issue.title }}
  run: echo "$TITLE"
```

### Secrets Management

- Use GitHub Secrets for sensitive data
- Never commit credentials
- Rotate secrets regularly
- Use OIDC for cloud provider auth

### Dependency Pinning

- Pin action versions: `uses: actions/checkout@v4`
- Pin tool versions: `terraform_version: 1.10.0`
- Use dependabot for updates

## Local Testing

### Run Terraform Validation Locally

```bash
# All providers
for provider in aws-eks gcp-gke azure-aks do-doks linode-lke oci-oke ibm-iks openstack; do
  cd infra/terraform/blueprints/$provider
  terraform fmt -check
  terraform init -backend=false
  terraform validate
  cd -
done

# Or use act
act -j validate-providers
```

### Run E2E Tests Locally

```bash
# Create kind cluster
kind create cluster --name mantl-test

# Run tests
pytest tests/e2e/security/ -v
pytest tests/e2e/platform/ -v
```

### Run Security Scans Locally

```bash
# Trivy
trivy fs --scanners vuln,secret,config .

# Checkov
checkov -d infra/terraform/

# Semgrep
semgrep --config=auto .
```

## Workflow Status Badges

Add to README.md:

```markdown
![Terraform Validation](https://github.com/username/mantl/actions/workflows/terraform-validate.yml/badge.svg)
![E2E Tests](https://github.com/username/mantl/actions/workflows/e2e-tests.yml/badge.svg)
![Security Scan](https://github.com/username/mantl/actions/workflows/security-scan.yml/badge.svg)
```

## Troubleshooting

### Terraform Init Fails

**Issue:** Provider download fails
**Solution:** Check provider versions in `required_providers`

### E2E Tests Timeout

**Issue:** Kind cluster not ready
**Solution:** Increase timeouts or reduce test scope

### Security Scan False Positives

**Issue:** Checkov reports non-issues
**Solution:** Add skip comments:
```hcl
# checkov:skip=CKV_AWS_1: Public S3 bucket required for static hosting
resource "aws_s3_bucket" "public" {
  # ...
}
```

## CI/CD Best Practices

1. **Fast Feedback**: Fail fast with quick checks first
2. **Parallel Execution**: Use matrix strategies
3. **Caching**: Cache dependencies (pip, terraform providers)
4. **Artifacts**: Upload test reports and logs
5. **Notifications**: Alert on failures
6. **Branch Protection**: Require passing checks before merge

## Required GitHub Settings

### Branch Protection Rules (main)

- ✅ Require pull request before merging
- ✅ Require status checks to pass:
  - `validate-providers`
  - `validate-modules`
  - `tflint`
  - `kind-cluster-tests`
- ✅ Require conversation resolution
- ✅ Require signed commits

### Repository Secrets

Configure these secrets in repository settings:

#### Cloud Provider Credentials (Optional, for integration tests)
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `DIGITALOCEAN_TOKEN`
- `LINODE_TOKEN`
- `AZURE_CREDENTIALS`
- `GCP_SA_KEY`

#### Notifications
- `SLACK_WEBHOOK_URL`
- `DISCORD_WEBHOOK_URL`

#### Release Automation
- `GITHUB_TOKEN` (automatically provided)

## Workflow Customization

### Skip CI

Add to commit message:
```
fix: update documentation [skip ci]
```

### Run Specific Tests

```yaml
# In PR comment
/test security
/test performance
/test all
```

### Manual Workflow Dispatch

```bash
# Using GitHub CLI
gh workflow run e2e-tests.yml

# With inputs
gh workflow run release.yml -f version=v1.2.3
```

## Performance Optimization

### Current CI/CD Timings

- Terraform Validation: ~5 minutes (parallel)
- E2E Tests: ~15 minutes (parallel)
- Security Scans: ~3 minutes
- **Total: ~15-20 minutes**

### Optimization Strategies

1. **Caching**:
   ```yaml
   - uses: actions/cache@v4
     with:
       path: ~/.terraform.d/plugin-cache
       key: terraform-${{ hashFiles('**/*.tf') }}
   ```

2. **Parallel Jobs**: Already implemented with matrix

3. **Conditional Execution**: Skip unnecessary jobs
   ```yaml
   if: contains(github.event.head_commit.message, '[terraform]')
   ```

4. **Self-Hosted Runners**: For faster execution (optional)

## Monitoring

### GitHub Actions Dashboard

View workflow runs:
```
https://github.com/username/mantl/actions
```

### Metrics to Track

- Success rate
- Average duration
- Failure reasons
- Flaky tests

### Alerts

Configure workflow failure notifications:
```yaml
- name: Notify on failure
  if: failure()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    webhook_url: ${{ secrets.SLACK_WEBHOOK_URL }}
```

## Future Improvements

- [ ] Add performance benchmarking
- [ ] Implement canary deployments
- [ ] Add chaos engineering tests
- [ ] Integrate with Argo CD
- [ ] Add compliance scanning
- [ ] Implement automatic dependency updates
- [ ] Add cost estimation for Terraform changes
