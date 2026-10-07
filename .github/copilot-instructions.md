<!-- SPDX-License-Identifier: Apache-2.0 -->
# Mantl implementation and review instructions

Read [AGENTS.md](../AGENTS.md) for the complete architecture, conventions and
commands. The two root instruction files must stay synchronized.

- Keep CRD/reconciler concerns in `apis/` and `controllers/`, pure behavior in
  `pkg/`, and framework content in `compliance/frameworks/`.
- Preserve collection scope, trusted Job ownership, immutable evidence receipts,
  tenant identity, mTLS/database TLS and fleet row-level security.
- Flag missing/stale evidence classified as passing, exceptions counted as passes,
  unsupported coverage claims, and raw evidence placed in CRDs/etcd.
- Distinguish compiler output from apply, static provider validation from cloud
  acceptance, STIG assets from full implementation, and evidence from certification.
- Test with `mise exec -- make go-test`; regenerate API output with
  `mise exec -- make manifests generate`. Do not test/refactor vendored `platform/`.
- Preserve provider pins, GitOps policy ownership and exact-tag signed artifacts.
  Required checks must pass on the latest head; never suppress meaningful failures.
- Preserve legacy `master`; no deployment, settings change or release publication
  is implied by a code review. Keep local assistant state and secrets untracked.
