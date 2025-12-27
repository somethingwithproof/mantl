# Mantl Compliance CLI Design

## Overview

The `mantl compliance` CLI provides unified compliance management.

## Installation

```bash
brew install mantl-io/tap/mantl
```

## Commands

### Framework Management

```bash
# List frameworks
mantl compliance frameworks list

# Enable framework
mantl compliance frameworks enable soc2-type2

# Show framework details
mantl compliance frameworks describe soc2-type2
```

### Control Management

```bash
# List controls
mantl compliance controls list --framework soc2-type2

# Filter by status
mantl compliance controls list --status failing

# Show control details
mantl compliance controls describe SOC2-CC6.1
```

### Audit Management

```bash
# Start audit
mantl compliance audit start \
  --framework soc2-type2 \
  --period-start 2024-01-01 \
  --period-end 2024-06-30

# Check status
mantl compliance audit status audit-soc2-h1-2024

# Generate report
mantl compliance audit report audit-soc2-h1-2024 --format pdf

# Export evidence package
mantl compliance audit export audit-soc2-h1-2024 --output ./evidence/
```

### Finding Management

```bash
# List findings
mantl compliance findings list --severity critical,high

# Acknowledge finding
mantl compliance findings acknowledge FND-001 --assignee @security-team

# Resolve finding
mantl compliance findings resolve FND-001 --resolution "Fixed in PR #456"
```

### Evidence Management

```bash
# List evidence
mantl compliance evidence list --control SOC2-CC6.1

# Download evidence
mantl compliance evidence download sha256:a1b2c3...

# Verify integrity
mantl compliance evidence verify sha256:a1b2c3...
```

### Dashboard

```bash
# Show status
mantl compliance status

# Watch real-time
mantl compliance status --watch

# Export JSON for CI/CD
mantl compliance status --output json
```

## CI/CD Integration

```yaml
name: Compliance Gate
on: [pull_request]
jobs:
  compliance:
    runs-on: ubuntu-latest
    steps:
      - uses: mantl-io/compliance-action@v1
        with:
          framework: soc2-type2
          fail-on: critical,high
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 10 | Critical findings |
| 11 | High findings |
| 20 | Audit failed |
