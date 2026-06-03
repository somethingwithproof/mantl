# ADR 006: GitOps Owns Policy Application; the Compliance Operator Resolves and Reports

## Status
Accepted

## Context
Phase 2a made the `ComplianceProfileReconciler` resolve a framework's control
templates to Kyverno policy files and report coverage: `Status.ActivePolicies`
counts resolved policies and `Status.UnresolvedTemplates` lists controls with no
matching policy.

A natural next step appeared to be having the operator apply the resolved
ClusterPolicies to the cluster with server-side apply, so a profile would deploy
its policies rather than only report them. A pre-merge review of that change
surfaced two problems that make operator-driven application the wrong design for
this repository:

1. **Dual ownership with ArgoCD.** `clusters/production/apps/kyverno-policies.yaml`
   is an ArgoCD `Application` that already syncs `policies/kyverno` with
   server-side apply, `selfHeal: true`, and `prune: true`. The app-of-apps sets
   `ServerSideApply=true` and `FailOnSharedResource=true`. An operator that
   force-applies the same ClusterPolicies under a different field manager fights
   ArgoCD: ArgoCD self-heals fields back to Git, the operator re-forces them, and
   the shared-resource guard trips. An externally hardened policy field can be
   silently reverted to the vendored default.

2. **No application lifecycle.** Server-side apply mutates cluster state, but the
   operator had no inverse: no delete, no finalizer, no `OwnerReference`. Removing
   a `policyRef` from a framework or deleting a `ComplianceProfile` would leave
   every previously applied ClusterPolicy enforcing cluster-wide with no resource
   left to explain it. Cluster state would depend on apply history, not current
   spec, which is non-convergent.

## Decision
ArgoCD (GitOps) is the single owner of ClusterPolicy application and lifecycle.
The compliance operator MUST NOT create, server-side apply, force-own, or delete
ClusterPolicies.

The operator's role on the policy path is resolution and reporting:
- Resolve each control's `policyRef.template` to a policy file in the repo.
- Report coverage through `ComplianceProfile` status: `ActivePolicies` is the
  count of resolved policies, `UnresolvedTemplates` lists controls with no policy.

Surfacing a coverage gap is the operator's job. Closing it is a Git change that
ArgoCD syncs, not a direct cluster mutation by the operator.

## Rationale
- A single owner avoids the reconcile fight and the silent reversion of hardened
  fields that dual ownership produces.
- The operator does not reimplement prune, finalizers, and ownership tracking
  that ArgoCD already provides.
- Convergence: cluster policy state is a function of Git, not of the operator's
  apply history.

## Consequences
- The `ComplianceProfileReconciler` clusterpolicies RBAC is read-only
  (`get;list;watch`); it does not need `create`, `update`, `patch`, or `delete`.
- `ActivePolicies` means resolved-and-present in the repo, not applied to the
  cluster. Status is an attestation of coverage, not of live enforcement; live
  enforcement is observed separately through Kyverno PolicyReports (the
  `FindingReconciler` and the `mantl status` compliance score).
- A future auto-remediation feature that wants to close coverage gaps must do so
  in a GitOps-compatible way (for example, open a pull request adding the missing
  policy), not by applying to the cluster directly.
