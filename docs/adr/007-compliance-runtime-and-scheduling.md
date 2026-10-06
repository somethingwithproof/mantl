<!-- SPDX-License-Identifier: Apache-2.0 -->
# ADR 007: Packaged compliance definitions and durable evidence scheduling

Status: Accepted for implementation; review required before merge.

Framework definitions and policy mappings ship with the operator. Both profile
and audit controllers load `<framework>/framework.yaml`, validate the requested
version, and reject unsafe paths. ArgoCD remains the policy owner (ADR 006).

Evidence collection uses an injected Kubernetes reader, not kubectl in the
operator image. Only allowlisted resources are readable. Namespaced collection
requires explicit namespaces (default: profile namespace); cluster RBAC evidence
requires an explicit operator opt-in. Workload evidence removes pod templates,
annotations, managed fields, and status before storage to avoid collecting inline
credentials. Unsupported collectors remain visible coverage gaps.

`frequency: framework` uses five-field cron schedules in UTC. A collector runs
once at first observation and once when due; missed intervals coalesce into one
run. Manual/daily/weekly audits retain their existing meaning and collect the full
selected snapshot set on each run. A single leader and serial reconciliation
prevent overlapping runs for one audit. The persisted StartTime identifies an
in-progress run across retries. Collector timestamps advance only after their
snapshot and the run manifest have been stored successfully.

S3 evidence requires versioning, Object Lock, COMPLIANCE retention and encryption.
The minimum is 2557 days, preserving ADR 003's seven-year baseline; shorter
framework retention hints do not weaken it. Object keys include audit identity,
run identity and content hash. A manifest holds versioned object references and
hashes; CRDs hold only the manifest URI and hash. Export verifies every referenced
object hash. Capturing unsupported or denied evidence produces a partial/failed
audit, never a passing compliance verdict.

Cloud validation is static contract evidence only. No cloud is promoted to GA
without separate deployment and recovery evidence. Runtime findings use an
explicit offline Falco JSON ingestion command; enabling a public event receiver
or automatic remediation is outside this decision.
