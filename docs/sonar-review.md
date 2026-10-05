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
