<!-- SPDX-License-Identifier: Apache-2.0 -->
# Compliance content and runtime

Mantl's beta compliance operator reconciles configured control mappings,
PolicyReport findings, evidence collection, and evaluations. Start with
[compliance installation and operations](../docs/compliance-operations.md),
[the runtime architecture](../docs/architecture-runtime.md), and
[the evidence lifecycle diagram](../docs/architecture-diagram.md#evidence-lifecycle).

## Content is reviewed data

[frameworks](frameworks) contains SOC2, HIPAA, PCI-DSS, CIS Kubernetes, and
[experimental DISA Kubernetes STIG](../docs/stig-kubernetes.md) catalog
and profile YAML. A catalog entry or profile is not proof that its control has a
working policy mapping, collector, complete coverage, or fresh evidence. Consult
ControlEvaluation coverage and gaps; no framework certification is claimed.
The built-in GitOps policy topology uses reviewed SOC2 content. Other framework
assets do not establish equivalent end-to-end integration.

[The bundle version](bundle-version.txt) identifies content independently of the
CLI. `mantl bundle build`, `verify`, and `unpack` create/verify deterministic content
manifests. Pin the verified content digest and a trusted OCI identity before
mounting separate bundles; see [content and evaluations](../docs/architecture-runtime.md#content-and-evaluations).

## Current responsibilities

ArgoCD applies reviewed policy content; the operator observes Kyverno
`wgpolicyk8s.io/v1alpha2` reports. Scheduled AuditRuns use scoped collector Jobs
for allowlisted Kubernetes resource snapshots. Unsupported log/metrics/query
collectors are gaps, not passing checks. Exceptions require separate approval and
annotate failures without resolving them.

Raw evidence goes to versioned AWS S3 with Object Lock COMPLIANCE retention;
Kubernetes stores references/status. [Export](../docs/compliance-operations.md#observe-and-export)
retrieves exact versions and verifies hashes. Native GCS/Azure immutable backends
and cloud recovery acceptance are not implemented/established.

Use [deploy/operator](../deploy/operator) and verified release install assets.
The older `compliance/operator`, Python `compliance/cli`, report/dashboard/API
assets, and framework descriptions are not the supported Go installation or proof
of an integrated dashboard, PDF audit command, or automatic remediation.
