<!-- SPDX-License-Identifier: Apache-2.0 -->
# Mantl Modernization Roadmap (2026)

Status: approved design, phased delivery
Date: 2026-06-02
Owner: maintainer

## Purpose

Mantl is positioned as an evolutionary successor to the original Cisco Mantl
microservices infrastructure platform. The original combined Terraform, Ansible,
and Mesos/Marathon/Consul/Vault. This successor is a Kubernetes-native platform
built on GitOps (ArgoCD), Terraform/OpenTofu, and a Go compliance operator, with
"Compliance-as-Code" as its stated differentiator.

This document records the agreed goal, an honest assessment of the current state,
and a phased plan to make the project a credible open-source product.

## Goal and success criteria

Identity: a real OSS product that an unaffiliated team could adopt.

A release is credible when all of the following hold:

- `go build ./...` and `go test ./...` are green on a public CI runner, with no
  vendored third-party tests polluting the module test scope.
- Every claim in the README maps to a feature that exists in code, or is labeled
  with an accurate maturity level (GA / beta / experimental).
- The compliance-as-code path runs end to end on at least one framework: a
  `ComplianceProfile` selects controls, controls map to policies that exist in
  the repo, policy violations surface as `Finding` CRDs, evidence is captured on
  a schedule, and an audit produces stored, hashed evidence.
- AWS, GCP, and Azure blueprints reach equivalent hardening (private cluster,
  audit logging, encryption at rest, workload identity, least-privilege IAM).
- Supply-chain, runtime-security, and developer-portal capabilities are present
  and documented, not aspirational.

## Current state (verified 2026-06-02, post-merge of #155)

Solid and working:

- `go build ./...` passes. CRD `DeepCopy` methods are generated.
- `FindingReconciler` reads Kyverno `PolicyReport` results and creates `Finding`
  CRDs. Spec-parser, bootstrap-DAG, and finding-controller tests pass.
- AWS EKS Terraform blueprint is production-grade (KMS-encrypted etcd, control
  plane audit logging, managed node groups, IRSA roles).
- Evidence capture + S3 upload (`pkg/evidence`) has allowlist validation and
  SHA256 hashing.
- ArgoCD/GitOps bootstrap path (`pkg/bootstrap`) is real.
- About 13 working Kyverno policies.

Broken, missing, or overstated:

1. Two failing tests: `pkg/evidence` `TestCaptureResource_ResourceID` (own code),
   and a vendored Falco Helm chart Go test under `platform/security/falco/`. The
   `platform/` tree holds roughly 2000 vendored chart files whose Go tests are
   pulled into `go test ./...`.
2. Compliance frameworks are declarations, not enforcement. SOC2/HIPAA/PCI/CIS
   YAML reference policy templates (for example `require-network-policy`,
   `require-pod-security-context`) that do not exist in the repo. The audit
   controller captures only hardcoded resource kinds; the `evidenceCollector`
   specs and cron schedules in the framework YAML are never read by any code, so
   "continuous compliance" has no scheduler.
3. Multi-cloud is overstated. AWS is near complete; GCP and Azure are bare module
   wrappers without hardening; Linode/Oracle/IBM/DigitalOcean are unverified.
4. README oversells. The `mantl compliance status` dashboard is fabricated output
   for a command that does not exist. "5-minute production setup" actually creates
   a local kind cluster.
5. Dead legacy code: vestigial Nomad/Consul under `infrastructure/nomad/`,
   disconnected from the Kubernetes path.
6. CI honesty gap: the main `validate` job runs on a `self-hosted` runner (so it
   does not run for outside contributors) and pins Go 1.23 while `go.mod` is 1.26.

Overall: the foundation is real, the compliance differentiator is largely unwired,
and the documentation describes features the code does not provide.

## Scope decisions

- Cloud parity targets AWS, GCP, and Azure only. Linode, Oracle, IBM, and
  DigitalOcean blueprints are demoted to a clearly labeled experimental tier.
- Added 2026 capabilities, in priority order: supply chain (SLSA/Sigstore),
  eBPF runtime security (Falco wired to findings, Cilium/Tetragon), and a
  Backstage internal developer portal.
- Out of scope for now: an AI/LLM operations layer.

## Approach

Foundation first, then a vertical slice, then breadth, then new capabilities.
Integrity precedes features because an OSS product cannot ask for trust while the
README is inaccurate and CI is red. After the foundation is honest and green, one
fully wired compliance path proves the architecture before money is spent on
multi-cloud breadth and new subsystems.

## Phased plan

### Phase 1: Foundation and integrity

Outcome: a repository a stranger can trust. Green, honest CI on public runners;
documentation that matches the code; clean module boundaries.

- Fix the failing `pkg/evidence` test.
- Remove vendored chart trees from the Go test scope (exclude `platform/` vendored
  charts from `go test ./...`, via module boundary or a filtered test target).
- Move the CI `validate` job to a public runner; align the Go version with
  `go.mod`; ensure build, test, and lint run on every PR.
- README truth pass: remove fabricated CLI output, correct the "production"
  claim, add a per-cloud and per-framework maturity matrix.
- Remove or quarantine the dead Nomad/Consul code with a documented rationale.
- Establish ADR discipline: record the decisions in this roadmap as ADRs.

Exit criteria: `go build ./...` and `go test ./...` green on a public runner;
README maturity matrix merged; no vendored tests in module scope.

### Phase 2: Compliance-as-code vertical slice (AWS + SOC2)

Outcome: one framework works end to end, proving the architecture.

- Author the Kyverno (or Gatekeeper) policies that the SOC2 framework references,
  so every mapped control resolves to a policy that exists.
- Wire the audit controller to read `evidenceCollector` specs from the framework
  (config-snapshot at minimum), not only hardcoded kinds.
- Implement the audit scheduler (CronJob or controller-managed schedule) so the
  cron expressions in framework YAML drive evidence collection.
- Connect `PolicyReport` to `Finding` to a compliance score surfaced through a
  real CLI command or status subresource.
- Controller tests for the profile, finding, and audit reconcilers covering the
  wired paths.

Exit criteria: applying the SOC2 standard profile to a kind cluster produces
findings from policy violations and scheduled, stored evidence.

### Phase 3: Multi-cloud parity (AWS, GCP, Azure)

Outcome: three clouds at equivalent hardening.

- Bring GCP GKE and Azure AKS to the AWS hardening baseline: private cluster,
  control-plane audit logging, encryption at rest, workload identity, and
  least-privilege IAM bindings for platform components.
- A shared variables and outputs contract across the three blueprints.
- `terraform validate` plus policy checks for all three in CI.
- Demote the remaining clouds to an experimental tier in docs and directory
  layout.

Exit criteria: AWS, GCP, and Azure pass validation and security checks in CI and
expose the same platform component contract.

### Phase 4: 2026 capabilities

Outcome: supply-chain, runtime-security, and developer-portal capabilities that
are real and documented.

- Supply chain: finish Sigstore/cosign keyless signing and SLSA provenance for
  build artifacts; generate and attest SBOMs; enforce signature verification at
  admission. Treat all of this as a hard-stop security domain requiring review.
- eBPF runtime security: wire the already-vendored Falco to the findings pipeline;
  add Cilium or Tetragon network policy and runtime threat detection feeding the
  compliance evidence path.
- Backstage IDP: stand up the developer portal with a software catalog and
  golden-path templates, replacing the current doc stub.

Exit criteria: each capability has working code, a test or smoke check, and
documentation; supply-chain changes carry a dedicated security review.

## Risks

- Vendored `platform/` tree is large and may hide more module-scope surprises;
  Phase 1 must establish a firm boundary so later phases are not derailed.
- Supply-chain and IAM work touches hard-stop security domains; these require
  delegated implementation and explicit review, not direct edits.
- Compliance wiring depends on cluster-side CRDs (Kyverno `PolicyReport`,
  `wgpolicyk8s.io`); Phase 2 must decide whether to require those CRDs or degrade
  gracefully.

## Process

Each phase ships as one or more small, atomic PRs with verification evidence.
Work in hard-stop domains (CI signing, IAM, Terraform providers, compliance
enforcement) is delegated and reviewed. Each phase produces or updates ADRs under
`docs/adr/`.
