<!-- SPDX-License-Identifier: Apache-2.0 -->
# ADR 008: Compliance platform architecture and release boundary

Status: Accepted for implementation; deployment acceptance remains pending.

## Context

Mantl has a working Go compiler and compliance runtime, but installation depends
on a checkout, audit history shares one mutable status object, collectors run
inside reconciliation, and framework content is coupled to the operator image.
The owner authorized the 2026–2027 architecture recommendations on 2026-10-04.

## Decision

Retain one Go module. Domain packages own control bundles, collection plans,
evaluation state and evidence verification. Controllers own Kubernetes lifecycle;
CLI and optional fleet HTTP adapters call the same domain contracts.

1. Release Please selects SemVer tags. A reusable release workflow checks out and
   validates the exact tag, builds Linux/macOS amd64/arm64 archives and Linux
   deb/rpm packages, publishes operator images and installation/platform assets,
   and records checksums, SBOMs and signatures. Build identity is embedded in Go
   binaries. Publishing privileges belong only to publishing jobs.
2. Framework content has a separate content version and deterministic manifest
   digest. Profiles may pin a content digest. Every new run records that digest.
   Unsupported mappings are explicit coverage gaps; packaging is not validation
   of a regulatory framework.
3. AuditRun records immutable run input and result references. The scheduler
   creates durable task identities; isolated Kubernetes Jobs execute bounded
   captures under scoped ServiceAccounts. Legacy inline execution remains an
   explicit compatibility mode during migration, not the new installation default.
4. Evaluations distinguish pass/fail/unknown/error/not-applicable, coverage and
   freshness. Findings retain violation history. Exceptions have an explicit
   approver, owner, justification, scope and expiry; they do not turn violations
   into passing evaluations.
5. Immutable evidence remains outside etcd with the ADR 003 retention minimum.
   Metadata indexes are rebuildable; tenant/cluster identity comes from trusted
   enrollment rather than request-supplied names. The optional fleet API holds no
   cluster administrator credentials and never serves raw evidence by default.
6. Cloud support requires explicit deployment, identity, audit-delivery, encryption,
   upgrade and recovery acceptance records. Static checks cannot satisfy these.
   Existing Terraform/OpenTofu foundations remain; Crossplane is not introduced
   without a concrete self-service resource requirement.

ArgoCD remains the policy owner (ADR 006). Remediation produces reviewed Git
changes. Existing CRDs remain readable during migration; new types are additive.
Control selection and tenant isolation must not weaken when metadata is missing.

GitHub Actions owns required pull-request checks, hosted Sonar analysis and release
publication. The root Jenkinsfile is a compatibility entrypoint for local Go,
Python and formatting checks on an existing Jenkins worker. It uses the same mise
pins and canonical test commands; it does not duplicate GitHub release jobs,
provision clusters, or publish notifications. The Jenkins GitOps example produces
a manifest artifact for review rather than pushing an unreviewed deployment change.

## Consequences

Release snapshots and local deterministic tests validate implementation. Hosted
release publication, real cloud acceptance, native non-S3 storage, and fleet
deployment require distinct verification and may not be claimed from unit tests.
The optional API and new job execution mode remain beta until integration and
recovery checks pass. No certification or SLSA level is asserted merely because
a workflow file exists.
