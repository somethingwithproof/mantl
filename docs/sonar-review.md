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
