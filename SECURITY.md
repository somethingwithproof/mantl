# Security Policy

## Supported versions

Security fixes are provided for the latest minor release. Older minors receive
fixes only at the maintainer's discretion.

| Version | Supported          |
|---------|--------------------|
| Latest minor (0.x) | Yes     |
| Older minors       | No      |

Mantl is pre-1.0. The supported line moves forward with each minor release. Pin to
a released tag and track the changelog for security-relevant changes.

## Reporting a vulnerability

Report vulnerabilities privately through GitHub Security Advisories:

1. Go to the repository's **Security** tab.
2. Select **Report a vulnerability**.
3. Provide a description, affected versions, reproduction steps, and impact.

Do not open a public issue or pull request for a vulnerability. This is especially
important for pre-authentication issues, where a public report gives an attacker a
working exploit before a fix is available.

If you cannot use GitHub Security Advisories, contact the maintainer through the
address listed on the GitHub profile and mark the message as a security report.

## Response timeline

The maintainer aims to:

- Acknowledge a report within 5 business days.
- Provide an initial assessment, including severity and whether the report is in
  scope, within 10 business days.
- Coordinate a fix and disclosure timeline with the reporter for valid issues.

These are targets, not guarantees. A single maintainer runs this project, so
timelines depend on availability.

## Scope

In scope:

- The compliance operator and its reconcilers (`controllers/`, `cmd/`).
- Package logic that handles evidence, spec compilation, and bootstrap (`pkg/`).
- Compliance frameworks and the policies they reference (`compliance/`,
  `policies/kyverno/`), where a flaw weakens enforcement or audit integrity.
- Terraform blueprints and modules for the parity clouds (`infra/terraform/`),
  where a default widens a cloud security boundary.
- Release and supply-chain workflows (`.github/workflows/`).

Out of scope:

- Vendored third-party charts under `platform/`. Report those to their upstream
  projects (for example Loki, Falco).
- Experimental cloud blueprints (Linode, Oracle, IBM, DigitalOcean), which carry no
  hardening guarantee.
- Issues that require a misconfiguration the project documents against, or access
  the threat model already assumes (for example a compromised cluster-admin).

## Supply-chain posture

Release artifacts are signed and attested:

- Container images and bundles are signed with Cosign using keyless OIDC
  (`.github/workflows/release.yml`).
- The release path generates SLSA provenance and SBOMs
  (`.github/workflows/slsa-provenance.yml`).

Verify a signed image before trusting it:

```bash
cosign verify \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  ghcr.io/thomasvincent/mantl:<tag>
```

Admission-time enforcement of image signatures is available through the Kyverno
policies in `policies/kyverno/`.
