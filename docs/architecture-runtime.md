# 2026–2027 runtime architecture

The new execution path and fleet service are beta. Local API and PostgreSQL
integration tests exercise admission and isolation; no cloud deployment or
recovery acceptance is claimed. ADR 008 defines the boundaries.

## Durable collection

The operator schedules immutable `AuditRun` inputs and creates at most four
concurrent collector Jobs per run. Each Job captures one allowlisted resource in
one namespace, uploads retained evidence, and returns only a version/hash receipt.
The operator verifies that receipt before recording completion. An unavailable
store leaves the run retryable. Admission/RBAC failures become explicit gaps.
Successful Jobs expire after one day; AuditRuns have no audit owner reference, so
deleting the schedule does not garbage-collect history. Archive and retire run
metadata deliberately; immutable S3 objects keep their retention independently.

Provision the collector ServiceAccount in every target namespace using
`mantl compliance collector-rbac NAMESPACE`. Review the emitted Role, then add
workload identity scoped to `audits/collectors/CLUSTER_UID/NAMESPACE/` in the evidence
bucket, versioning/lock inspection, object upload/read/retention, and optional KMS
permissions. No credentials are embedded in Jobs. Add reviewed DNS, Kubernetes API,
identity-provider and S3/KMS egress for your environment; default-deny tenant policy
does not automatically permit collector traffic. Cluster-scoped collection remains
an explicit opt-in and needs a separately scoped ClusterRole.

The operator executes runs only from --control-namespace (mantl-compliance-system
by default), which must be reserved for trusted platform administrators. Tenant
users must not create/modify audits, runs or profiles in this namespace. Runs cannot
select a collector image, account, bucket or KMS key different from the approved
operator configuration. The operator uses the kube-system namespace UID as cluster identity. Changing
storage configuration or collector identities during active runs is unsupported:
finish/retire active runs before migrating storage. `--inline-audits` preserves the
old execution adapter for migration. Existing ComplianceAudit latest pointers stay
readable; historical export uses `mantl compliance export RUN --run`.

## Content and evaluations

`compliance/bundle-version.txt` versions control content independently of the CLI.
`mantl bundle build`, `unpack`, and `verify` provide deterministic content manifests
and bounded extraction. The optional Publish control content workflow signs a
single-layer OCI artifact. Increment the content version for changed bytes;
publication refuses replacement of an existing version. Registry administration
must also preserve tag immutability. Use the returned OCI digest for downloads;
`spec.bundleDigest` is the separate digest of the content manifest, not the OCI
manifest. A version number alone is not a trust anchor.

After verifying the OCI signature under the exact repository/workflow identity,
pull `controls.tar.gz`, unpack with the expected content digest, and mount the
verified directory read-only at `/etc/compliance` in an operator overlay. Files must
be readable by UID 65532. New images contain a verified default bundle. The bundle
root Kustomization lets ArgoCD own the same policy content; the operator never
applies policies. The built-in root application propagates configured repository
and revision to its policy child. Pin release revisions in production.

ControlEvaluation separates result, coverage, freshness, evidence and exceptions.
Missing policies, rules, reports, collectors or timestamps yield unknown/partial
state. A failure in one resource cannot be hidden by another resource's pass.
Coverage describes configured mappings and rule observations, not independently
verified inventory coverage for every live workload. Only wgpolicyk8s.io/v1alpha2
reports are currently supported. Excluded controls are
marked not-applicable rather than retaining an old result.

Exceptions require a reviewed owner, justification, scope, expiry and an approved
spec generation. Bind the separate `deploy/operator/exception-approver-role.yaml`
to actual approvers; the operator cannot approve exceptions. Use
`mantl compliance exception approve NAME` through the authenticated kube-context.
Approval annotates the failing evaluation; it never converts fail into pass.
Resource-scoped exceptions are stored but control-level annotation currently uses
whole-control exceptions. No automated remediation or enforcement bypass exists.

Finding transitions are durably queued in status before immutable, linked metadata
records are written. A store outage preserves pending transitions; a full queue
marks a history gap and prevents a current compliance claim. `mantl compliance
history FINDING` verifies the complete recorded chain before printing it. Keep
Finding metadata and its history head in metadata backups. Falco-generated Finding
manifests currently need a trusted ingestion path and do not automatically acquire
this PolicyReport-controller lifecycle history.

## Optional fleet metadata service

`deploy/fleet/base` runs `/fleet` from the operator image. It requires an overlay:
replace the image with its release digest, replace the empty enrollment ConfigMap,
provide existing `fleet-tls` and `fleet-database` Secrets and a `fleet-storage`
ConfigMap, and configure scoped workload identity. TLS Secret keys are tls.crt,
tls.key and client-ca.crt; database Secret key url supplies
MANTL_FLEET_DATABASE_URL. Use verified database TLS (`sslmode=verify-full`) and a
non-superuser role without BYPASSRLS. The role initializes/owns the metadata table;
schema migrations are currently startup-managed, not a separate migration service.

Enrollment JSON maps client certificate SHA256 fingerprints to tenant, cluster
and collector/viewer role. Collector cluster IDs must match kube-system UIDs;
a cluster cannot belong to multiple tenants. A viewer with no cluster scope can
read its own tenant's fleet. Rotate/revoke identities by updating reviewed
configuration and restarting; never derive tenant identity from a request body.

The base Service is private and the NetworkPolicy denies egress. Add reviewed DNS,
PostgreSQL and S3/KMS egress and permit intended client namespaces; same-cluster
clients require `mantl.io/fleet-client=true`. Cross-cluster networking needs an
explicit authenticated private ingress overlay. TCP readiness confirms the
listener, not database recovery or full service health.

Enable the operator publisher with --fleet-url, --fleet-certificate, --fleet-key,
--fleet-ca and --fleet-cluster-id. Metadata first reaches immutable S3 journals;
the API verifies those references and stores a rebuildable PostgreSQL index under
forced tenant row-level security. It has no cluster-admin credentials or raw
evidence endpoint. `mantl fleet status` paginates state; `mantl fleet replay
RECEIPTS.json` rebuilds from archived journal references. Back up the receipt
inventory or enumerate immutable journal versions through an authorized storage
inventory before recovery. Periodic snapshots can miss intermediate states;
Finding history is the separate lifecycle record. The fleet UI is not implemented.

## Provider acceptance and OSCAL

`mantl acceptance verify RECORD.json --bucket BUCKET` verifies versioned retained
receipts for installation, workload identity, private access, audit delivery,
encryption, upgrade, recovery and evidence retention. Receipts bind provider,
cluster, run, image/content digest and GitHub runner provenance to source artifacts.
Verification reports `verified-receipts`, not independently observed deployment
success. No cloud suite has been run for this change; only local tests were requested.
AWS S3 remains the immutable evidence backend for all provider adapters. Native
GCS/Azure WORM and provider acceptance runners need separate implementation and
cloud environments before parity can be claimed.

`mantl oscal export` and `import` support a bounded catalog subset: control IDs and
titles. Imports produce unmapped drafts for review. They do not import assessment
results, infer mappings, or validate every external OSCAL schema requirement.
