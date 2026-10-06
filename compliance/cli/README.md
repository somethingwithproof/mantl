<!-- SPDX-License-Identifier: Apache-2.0 -->
# Mantl Compliance CLI

Command-line interface for managing compliance in Mantl clusters.

## Installation

```bash
# Via Homebrew
brew install mantl-io/tap/mantl

# Via Go
go install github.com/mantl-io/mantl/compliance/cli/cmd/mantl-compliance@latest

# Via container
docker run -it ghcr.io/mantl/compliance-cli:latest
```

## Commands

### Status Overview

```bash
# View overall compliance status
mantl compliance status

# Status for specific framework
mantl compliance status --framework soc2

# JSON output for automation
mantl compliance status --output json
```

Example output:
```
┌─────────────────────────────────────────────────────────────────┐
│                    Mantl Compliance Status                       │
├─────────────────────────────────────────────────────────────────┤
│ Framework: SOC2 Type II                                          │
│ Profile:   soc2-standard                                         │
│ Mode:      Enforce                                               │
├─────────────────────────────────────────────────────────────────┤
│ Compliance Score: 94%                                            │
│                                                                  │
│ Controls:  64 total │ 60 passing │ 4 failing                    │
│ Findings:  ⚠ 2 critical │ 2 high │ 5 medium │ 12 low            │
│ Evidence:  1,247 items │ Last collected: 2 hours ago            │
├─────────────────────────────────────────────────────────────────┤
│ Critical Issues:                                                 │
│  • CC6.1: 3 namespaces missing NetworkPolicy                    │
│  • CC6.8: 2 unsigned images deployed                            │
└─────────────────────────────────────────────────────────────────┘
```

### Run Audit

```bash
# Run full audit
mantl compliance audit --framework soc2

# Run audit and generate PDF report
mantl compliance audit --framework soc2 --output report.pdf

# Audit specific controls
mantl compliance audit --controls CC6.1,CC6.8,CC7.1
```

### View Findings

```bash
# List all findings
mantl compliance findings

# Filter by severity
mantl compliance findings --severity critical,high

# Filter by control
mantl compliance findings --control CC6.1

# Show remediation guidance
mantl compliance findings --id finding-abc123 --remediation
```

### Remediation

```bash
# Auto-remediate a finding (if supported)
mantl compliance remediate --id finding-abc123

# Dry-run remediation
mantl compliance remediate --id finding-abc123 --dry-run

# Mark finding as accepted risk
mantl compliance accept --id finding-abc123 --reason "Accepted per security review 2024-01"

# Mark as false positive
mantl compliance dismiss --id finding-abc123 --reason "Not applicable to our environment"
```

### Evidence Management

```bash
# List collected evidence
mantl compliance evidence list

# Collect evidence now
mantl compliance evidence collect

# Download evidence for a control
mantl compliance evidence download --control CC6.1 --output ./evidence/

# Verify evidence integrity
mantl compliance evidence verify --id evidence-xyz789
```

### Attestations

```bash
# List controls requiring attestation
mantl compliance attestation list --pending

# Create attestation
mantl compliance attestation create \
  --control CC6.3 \
  --statement "Quarterly access review completed" \
  --evidence ./access-review-q1.pdf

# View attestation history
mantl compliance attestation history --control CC6.3
```

### Reports

```bash
# Generate executive summary
mantl compliance report executive --output summary.pdf

# Generate detailed audit report
mantl compliance report audit --framework soc2 --output audit-report.pdf

# Generate evidence package for auditors
mantl compliance report evidence-package --output evidence.zip

# Schedule recurring reports
mantl compliance report schedule --frequency weekly --email security@company.com
```

### Framework Management

```bash
# List available frameworks
mantl compliance frameworks list

# Enable a framework
mantl compliance frameworks enable soc2 --profile standard

# Disable a framework
mantl compliance frameworks disable hipaa

# Update framework version
mantl compliance frameworks update soc2 --version 2023
```

## Configuration

```yaml
# ~/.mantl/compliance.yaml
defaults:
  framework: soc2
  output: table

evidence:
  bucket: s3://my-compliance-evidence
  region: us-east-1

notifications:
  slack:
    webhook: $SLACK_WEBHOOK_URL
    channel: "#compliance"

reporting:
  schedule: "0 9 * * 1"  # Weekly Monday 9am
  recipients:
    - security@company.com
    - ciso@company.com
```

## Exit Codes

| Code | Meaning |
|------|--------|
| 0 | Success, fully compliant |
| 1 | General error |
| 2 | Compliance failures (critical/high) |
| 3 | Compliance warnings (medium/low) |

## CI/CD Integration

```yaml
# GitHub Actions
- name: Check Compliance
  run: |
    mantl compliance status --output json > compliance.json
    mantl compliance audit --fail-on critical,high
```

```groovy
// Jenkinsfile
stage('Compliance Check') {
    steps {
        sh 'mantl compliance audit --output junit --output-file compliance.xml'
    }
    post {
        always {
            junit 'compliance.xml'
        }
    }
}
```
