---
name: mantl-core
description: Load when starting work anywhere in the mantl repo. Establishes the architecture (Kubernetes operator + Terraform platform), the boundary between CRDs/controllers, pkg logic, and frameworks-as-data, the module map, and which specialized skill to load next.
---
<!-- SPDX-License-Identifier: Apache-2.0 -->

# mantl core skill

## What this project is

Mantl is a Go and Kubernetes platform with two halves that share a vocabulary but run in different planes:

1. A **compliance operator**: CRDs under `apis/compliance/v1alpha1` reconciled by controllers under `controllers/compliance/`, fed by compliance frameworks expressed as YAML data, enforced through Kyverno policies, and producing evidence stored in object storage.
2. A **cluster platform**: Terraform blueprints under `infra/terraform/blueprints/` that stand up hardened clusters, Helm charts under `charts/`, and the bootstrap/compile pipeline under `pkg/`.

The module path is `github.com/thomasvincent/mantl`. Go version is pinned in `go.mod`.

## The boundary that matters

Mantl separates three concerns. Keep changes on the correct side.

| Layer | Where | What lives here |
|---|---|---|
| API + controllers | `apis/`, `controllers/` | CRD type definitions (kubebuilder markers, DeepCopy), reconcile loops. Kubernetes-aware. Imports controller-runtime. |
| Domain logic | `pkg/compiler`, `pkg/bootstrap`, `pkg/evidence` | Plain Go. Spec parsing, render, DAG ordering, evidence capture and upload. Minimal Kubernetes coupling; testable without a cluster. |
| Frameworks as data | `compliance/frameworks/`, `policies/kyverno/` | SOC2 / HIPAA / PCI-DSS / CIS expressed as YAML. Controls reference policy templates and evidence collectors. No Go. |

The rule: compliance behavior is data, not code. A new control is a YAML entry that points at a policy template and an evidence collector, not a new branch in a reconciler. Controllers translate between Kubernetes resources and that data; `pkg/` does the IO and computation.

## Module map

```
apis/
  compliance/v1alpha1/    ComplianceProfile, Finding, ComplianceAudit + DeepCopy
  platform/v1alpha1/      MantlCluster and related platform types
controllers/compliance/
  profile_controller.go   reconciles ComplianceProfile
  finding_controller.go   maps Kyverno PolicyReports into Finding CRs
  audit_controller.go     reconciles ComplianceAudit
pkg/
  compiler/               spec parser + GitOps/Terraform renderers
  bootstrap/              dependency DAG for ordered bring-up
  evidence/               point-in-time resource snapshot + S3 upload (ADR 003)
compliance/frameworks/    soc2/, hipaa/, pci-dss/, cis-kubernetes/ (framework.yaml + profiles)
policies/kyverno/         shipped Kyverno policy templates + vendored kyverno chart
infra/terraform/
  blueprints/             per-cloud cluster blueprints (aws-eks is GA; gcp/azure beta)
  modules/, examples/     reusable modules and worked examples
charts/mantl-platform/    Helm chart for the platform install
platform/                 vendored third-party Helm charts (Loki, Falco); not Mantl's code
docs/adr/                 architecture decision records (003, 004, 005)
```

## Architecture decision records

Three accepted ADRs govern current scope. Read the relevant one before changing the area it covers.

- **ADR 003** evidence storage contract: raw evidence goes to S3-compatible object storage; the `Finding` CRD holds a URI pointer only, never the payload.
- **ADR 004** test scope: `make go-test` excludes vendored `platform/` packages so upstream chart tests do not gate Mantl CI.
- **ADR 005** cloud parity: AWS EKS is GA; GCP GKE and Azure AKS target the same hardening baseline (beta); other clouds are experimental with no parity or CI guarantee.

## Build and test

```bash
make go-test          # canonical Go test target (excludes platform/)
```

`GO_PKGS` is `go list ./...` filtered to drop `github.com/thomasvincent/mantl/platform/`. Do not run bare `go test ./...`; it pulls in vendored chart tests that fail and are not ours to fix (ADR 004).

## When to load another skill

- Touching CRD types, controllers, framework YAML, Kyverno policies, or evidence flow: load `mantl-compliance`.
- Writing or hardening a Terraform blueprint: load `mantl-terraform`.
- Writing or running tests: load `mantl-testing`.

## When in doubt

Check whether the change belongs in code or in framework data. If a "new control" needs a code change beyond a new policy template plus a collector entry, that is a signal to reconsider the design or write an ADR first.
