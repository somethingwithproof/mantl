<!-- SPDX-License-Identifier: Apache-2.0 -->
# Sonar review decisions

Fix reported defects before classifying an exception. Keep the new-code quality
gate enabled, and record evidence for any finding that does not describe the
operation being performed.

## Python dependency lock generator

PR #324 finding `AaENLPYSU1QIPkv6xTIC` (`pythonsecurity:S2083`) is a false positive
in `ci/lock_python_dependencies.py`. Its trace identifies `Path.read_text()` on
the generated uv lock as an HTTP-controlled source, then treats the contents
argument of `Path.write_text()` as a destination path. This code has no HTTP
entrypoint. The destination is a separate `Path` object confined to the owning
checkout's `dist` directory; dependency inputs are restricted to the fixed
manifest list. The command-line interface accepts only `--check`.

`tests/unit/test_python_locks.py` verifies rejection of external outputs and
unknown manifests before subprocess execution, detection of stale locks without
changing existing locks, and preservation of all existing locks if resolution
fails. `tests/unit/test_release_assets.py` verifies confinement when output
directories contain symlinks. No file-content value is used to select a path.

## Frontend CI loopback HTTP

Findings `AaELVxneAzA3Gm6ElT5U` and `AaELVxneAzA3Gm6ElT5V`
(`githubactions:S5332`) concern the two curl
commands in `.github/workflows/example-frontend.yml`. These requests test the
disposable container's static index and health response. Docker binds its test
port explicitly to `127.0.0.1`; `docker port` supplies that loopback address to
curl. No credentials, customer data or remote transport are involved. Cleartext
HTTP is appropriate for this confined local smoke test. The application's
deployment ingress remains responsible for TLS outside the CI runner.

## SLSA generator tag requirement

Finding `AaELVxoFAzA3Gm6ElT5j` (`githubactions:S7637`) is an accepted compatibility
exception for `.github/workflows/slsa-provenance.yml`. The upstream generator
requires its reusable workflow to use a complete release tag so the SLSA verifier
can authenticate its builder identity. A commit SHA would break that supported
verification path. See the [upstream reference requirements](https://github.com/slsa-framework/slsa-github-generator/blob/main/README.md#referencing-slsa-builders-and-generators).

The exception covers only the upstream generator's `v2.1.0` release reference.
Other action references remain pinned to complete commit SHAs. Review this
exception when updating the generator or verifier, and remove it if upstream
supports SHA-pinned reusable workflows. Published provenance does not itself
establish a SLSA level for Mantl; an independent assessment is still required.

## SPIRE node attestation

Findings `AaELVxgOAzA3Gm6ElT4i` and `AaELVxgOAzA3Gm6ElT4j`
(`kubernetes:S6431`) concern the SPIRE agent's host PID and network namespaces.
These support process/cgroup attestation and the local kubelet route, as described
by the pinned SPIRE workload attestor. They are accepted privileges only for the
trusted node agent. The template now requires verified kubelet TLS and a mandatory
CA mount, drops capabilities and removes `nodes/proxy` authorization in favor of
`nodes/pods`. See [the prerequisites and trust-boundary review](../infrastructure/cilium/spire-security.md).
Local tests cover the trust mount and restricted RBAC; no live attestation is claimed.
Production acceptance is blocked until the environment-specific live evidence
listed in the prerequisite document is reviewed. These exceptions are conditional
and must be revisited on trust-boundary or supported-version changes.

## Archived Nomad TCP listener

Finding `AaELVxxQAzA3Gm6ElT8g` (`terraform:S6258`) concerns the archived Nomad
internal TCP Network Load Balancer. The AWS S3 access-log feature inspected by this
rule records TLS listeners and requests, not this TCP listener. This is an accepted
logging gap only while the archived, unsupported module remains undeployed.
It is not approved for reuse or production, and is not a claim of complete logs.
[Archive security limits](legacy/nomad/terraform/SECURITY.md) require independently
verified network and application audit logging before reuse. Re-review if the
listener protocol, provider logging capabilities or archive support status changes.

## GCE public-IP capabilities removed

Findings `AaELVxu5AzA3Gm6ElT72`, `AaELVxu5AzA3Gm6ElT73` and
`AaELVxu5AzA3Gm6ElT74` (`terraform:S6329`) previously concerned conditional
`access_config` blocks. The accepted opt-in design has been superseded: all three
blocks and the external service-ingress rule are removed. All VM roles are now
private-only; deprecated public-access inputs reject unsafe legacy settings.

Optional administrative ingress allows only logged TCP/22 from Google's IAP
service range to this module's tagged VMs, with separate reviewed tunnel IAM and
SSH authentication prerequisites. Mocked plan-only tests cover every VM role,
managed and supplied subnets, rejected legacy inputs and the scoped IAP rule.
This is a code fix, not a continuing public-IP risk acceptance. Confirm removal
in analysis of the merged revision; retain the original Sonar history and add
that evidence to the issue record. The module remains experimental, with no live
GCE deployment or access validation claimed. See
[the migration and acceptance limits](../infra/terraform/gce/README.md) and
[ADR 010](adr/010-security-exception-boundaries.md).
