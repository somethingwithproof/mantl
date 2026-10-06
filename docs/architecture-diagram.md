<!-- SPDX-License-Identifier: Apache-2.0 -->
# Architecture and responsibility boundaries

These diagrams show implemented ownership boundaries in the beta runtime.
For capability maturity use [the project table](../README.md#capabilities-and-maturity).
For deployment requirements use [the quickstart](quickstart-platform.md#deployment-is-a-separate-step).

## Configuration flow

```mermaid
flowchart TB
    Inputs["Operator: spec, Git revision, cloud and storage settings"] --> CLI["Mantl CLI"]
    CLI --> Variables["terraform.tfvars.json"]
    CLI --> Generated["Applications and tenant manifests"]
    Variables --> Blueprint["Existing provider blueprint"]
    Blueprint --> Provision["OpenTofu or Terraform"]
    Provision --> K8s["Kubernetes"]
    Generated --> Git["Reviewed repository content"]
    Git --> Argo["ArgoCD"]
    Argo --> K8s
    Argo --> Policies["Kyverno policies"]
    Policies --> Reports["PolicyReports"]
    Reports --> Controller["Compliance controllers"]
```

The compiler generates inputs and references. Publishing tenant files to Git,
selecting valid infrastructure settings, supplying credentials, and configuring
state are operator work. Explicit `apply` invokes provisioning and bootstrap;
`plan` does not. ArgoCD owns desired policies; the compliance operator observes
reports and evidence, without directly applying policies.

`MantlCluster` is the CLI input type, not an implemented cluster provisioning
reconciler. Compliance APIs have generated CRDs and controllers. Framework YAML
is content, and vendored `platform/` charts are upstream assets.

## Evidence lifecycle

```mermaid
flowchart LR
    Profile["Profile and reviewed control mappings"] --> Evaluation["ControlEvaluation"]
    Report["Kyverno PolicyReports"] --> Finding["Finding and lifecycle history"]
    Report --> Evaluation
    Schedule["ComplianceAudit schedule"] --> Run["Immutable AuditRun"]
    Run --> Jobs["Scoped collector Jobs"]
    Jobs --> Store["AWS S3: versions, hashes, retained objects"]
    Store --> Receipt["Verified receipts in Kubernetes metadata"]
    Receipt --> Evaluation
    Finding --> Store
    Store --> Export["Exact-version retrieval and verified archive"]
    Exception["Separately approved exception"] --> Evaluation
```

Kubernetes holds status, scope, references, and history heads; raw evidence lives
in S3. Object version IDs and hashes identify what was captured. Jobs cannot
choose an arbitrary collector image or storage identity. Approved exceptions
annotate results; they do not erase failures. Missing reports, mappings, timestamps,
or collector output remain visible as unknown/partial or stale state.

Only allowlisted Kubernetes resource snapshots are executed by the collectors.
Framework references to unsupported log/metric/query collection are gaps. Falco
JSON can be converted to a Finding manifest through a trusted ingestion path;
there is no automatically integrated public event receiver.

## Optional fleet and boundaries

The optional fleet service accepts enrolled mutual-TLS identities, verifies S3
journal references, and stores metadata in PostgreSQL with tenant row-level
security. PostgreSQL is a rebuildable index; it is not the raw evidence store.
No fleet UI or cluster-admin proxy is implemented.

AWS/GCP/Azure blueprints have static contract tests. They do not establish working
cloud deployments, recovery, or storage parity. The evidence implementation uses
AWS S3 even when Kubernetes runs on another provider. Crossplane/cloud failover
and the larger observability stack in older ARA diagrams are design/reference
material, not guaranteed by this implementation.

See [runtime architecture](architecture-runtime.md), [compliance operations](compliance-operations.md),
[policy ownership ADR](adr/006-policy-application-ownership.md), and
[storage ADR](adr/003-evidence-storage-contract.md).
