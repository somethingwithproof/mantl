<!-- SPDX-License-Identifier: Apache-2.0 -->
# DISA Kubernetes STIG

Mantl includes experimental support for the Defense Information Systems Agency
(DISA) Kubernetes Security Technical Implementation Guide (STIG). The pinned
**V2R3** catalog contains all 91 requirement IDs, rule revisions, severities, and
Control Correlation Identifiers (CCIs) from the source benchmark. It is not a
claim of support for the latest DISA revision, STIG compliance, or authorization
to operate. This content is available from source; it is not in release v0.4.0.

## Scope and coverage

| Requirement | Implemented observation | Remaining responsibility |
| --- | --- | --- |
| V-242414 / CNTR-K8-000960 | Audit Pod hostPort values below 1024 across regular, init, and ephemeral containers in user namespaces | Review host networking, actual listeners, system workloads, and complete inventory |
| V-242415 / CNTR-K8-001160 | Audit secretKeyRef and envFrom.secretRef across those container types; collect scoped Pod snapshots | Review literal/application-supplied secrets, image configuration, runtime environment, and complete inventory |
| Other 89 requirements | Catalog entries with explicit manual-review mappings | Operator review of node/OS, control-plane, TLS, audit delivery, RBAC, patching, and other requirements |

Both mapped requirements also retain a manual-review mapping. Passing their
workload policies therefore leaves the full control **unknown with partial
coverage**; a policy violation remains **fail**. Every other requirement remains
unknown. The existing evaluator deliberately treats unsupported mappings as gaps.
Manual-review is a coverage marker, not an implemented attestation importer.
There is no automatic acceptance of a checklist, CKL/CKLB export, SCAP scanner,
or node/control-plane remediation in this integration.

`ComplianceProfile.status.state: Active` means policy references resolved in the
content directory; inspect `ControlEvaluation` for actual coverage, freshness,
and results. Resource snapshots alone do not establish a passing configuration.
Missing/stale observations remain unknown; approved exceptions annotate failures.
The catalog deliberately retains requirements about older Kubernetes mechanisms;
operators determine applicability against the pinned DISA guidance and their
distribution, rather than silently treating an obsolete mechanism as passing.

## Inspect offline

Obtain the source tree containing this guide and run its pinned toolchain:

```sh
mise install
mise exec -- go run ./cmd/mantl bundle build --version 0.5.0-stig.1 --output /tmp/mantl-stig-controls.tar.gz
mise exec -- go run ./cmd/mantl bundle unpack /tmp/mantl-stig-controls.tar.gz /tmp/mantl-stig-content
mise exec -- kustomize build policies/kyverno/stig-kubernetes
```

The example version identifies your local content bundle; it is not a published
release. Use a fresh unpack directory. The bundle contains the catalog under `frameworks/stig-kubernetes/` and
opt-in policies under `policies/stig-kubernetes/`, with a verified manifest.
Record its digest and mount verified content into the operator following
[content pinning](architecture-runtime.md#content-and-evaluations).
Building a bundle does not deploy it or enable STIG policies in its root
Kustomization; that root retains the existing SOC2/runtime topology.

## Enable through reviewed GitOps configuration

Install the operator, Kyverno, evidence identity/storage, and compliance CRDs
using [compliance operations](compliance-operations.md). A source-built operator
image includes this catalog automatically; the published v0.4.0 image does not.
Do not select the new framework until the mounted content includes it.

Add `policies/kyverno/stig-kubernetes` to a reviewed ArgoCD policy
overlay. The policies use **Audit** mode, do not change the built-in policy
topology. The host-port policy excludes `kube-system`, `kube-node-lease`, and
`kube-public`, matching the STIG check's system namespace boundary; it still
checks `default`. The Secret environment policy has no namespace exclusions.
Review the impact before selecting
Enforce mode; rejecting Secret environment references can interrupt workloads.
ArgoCD owns policy deployment; selecting a profile does not install policies.

Copy [profile-audit.yaml](../compliance/frameworks/stig-kubernetes/profile-audit.yaml)
into your environment repository. Replace `namespaces: [team-a]` with your
explicit workload scope, add a verified `spec.bundleDigest` if using a pinned
bundle, then have ArgoCD apply the profile and daily audit. The profile selects
all 91 controls so the remaining gaps stay visible. Policies observe Pods only
(controller autogeneration is disabled); child Pods are checked when created.

Daily audits schedule scoped Pod snapshots for the two mapped requirements.
They do not collect Secret object payloads or run `kubectl exec`. Raw snapshots
use the existing AWS S3 Object Lock backend and retention contract for every
cluster provider; this feature adds no native GCS/Azure backend. Review snapshot
sensitivity, bucket costs, and identity permissions before enabling collection.

```sh
mantl --context TARGET compliance status --namespace mantl-compliance-system
mantl --context TARGET compliance export stig-kubernetes-daily --output stig-evidence.tar.gz
```

These commands access the selected cluster; export requires stored completed
evidence and configured credentials. Follow the existing verification and
retention guidance in [compliance operations](compliance-operations.md).

## Source identity and maintenance

Source: [DISA Kubernetes V2R3 archive](https://dl.dod.cyber.mil/wp-content/uploads/stigs/zip/U_Kubernetes_V2R3_STIG.zip),
listed through the [DISA STIG library](https://www.cyber.mil/stigs/downloads).
The catalog's `spec.source` records archive and XCCDF SHA-256 hashes, and each
entry preserves its V-ID, CNTR-K8 identifier, rule revision, and CCIs. These
source metadata fields support content review; the runtime consumes IDs,
severities, and mappings and verifies the packaged bundle digest.

Updating DISA content requires comparing the new source and requirement changes,
reviewing mappings and applicability, selecting a distinct framework version,
and publishing a new immutable content bundle. Do not substitute a newer
benchmark under version `2.3`. Add collectors or mappings only with tests showing
that incomplete evidence cannot produce a passing full-control result.

The source content version advances to `1.0.2` for these new mappings; this change
does not publish a bundle. Offline policy tests run with Kyverno CLI `1.16.1`,
matching the application version of the platform's pinned Kyverno chart `3.6.1`.
