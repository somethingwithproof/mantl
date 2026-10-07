<!-- SPDX-License-Identifier: Apache-2.0 -->
# Release installation and verification

Release Please manages SemVer and changelogs. Its release-created output invokes
the reusable release workflow directly; publishing does not depend on another
workflow being triggered by a GITHUB_TOKEN-created release event.

## Release v0.5.0

[v0.5.0](https://github.com/somethingwithproof/mantl/releases/tag/v0.5.0) adds
[offline previews, saved-plan apply and artifact comparisons](platform-planning.md),
limited-effect input notices, and verified finding-history export. It retains
scoped JSON platform status and the optional `--require-healthy` application gate.
The release also includes the experimental [DISA Kubernetes STIG catalog and
opt-in audit policies](stig-kubernetes.md), with explicit manual coverage gaps.
The new plan/apply flags are absent from v0.4.0. Saved inventories use
`mantl.io/plan/v1alpha2`; regenerate earlier v1alpha1 inventories before use.

The release checks out the exact validated tag and publishes:

- Linux/macOS amd64/arm64 CLI tar.gz archives;
- Linux amd64/arm64 deb and rpm packages;
- the digest-pinned operator installation manifest and optional fleet installation manifests and platform source bundle;
- a versioned control bundle with a content manifest;
- SHA256 checksums, SPDX SBOMs and keyless Cosign verification bundles;
- a multiarchitecture operator image and OCI bundles.

For a complete installation walkthrough see [the release quickstart](quickstart-platform.md).
Download all attached assets into a new directory using
`gh release download TAG --repo somethingwithproof/mantl`. The full checksum check
below requires every file listed in the signed manifest; a partial download needs
verification of its exact selected entries instead.
Authenticate `checksums.txt` using its Cosign bundle before trusting its hashes.
Require the GitHub Actions OIDC issuer and this exact workflow identity:

```sh
cosign verify-blob checksums.txt --bundle checksums.sigstore.json \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity https://github.com/somethingwithproof/mantl/.github/workflows/release.yml@refs/heads/main
sha256sum -c checksums.txt
```

On macOS use `shasum -a 256 -c checksums.txt` for the checksum step. Check that the
certificate/source provenance matches the release commit as well as the workflow
identity. Do not accept arbitrary certificate identities. release-identity.json records the resolved operator digest; control-oci-identity.json
and platform-oci-identity.json record OCI bundle identities. Keep the source tag
and these identities with the installation.

Extract the CLI archive and put mantl on PATH, or install its deb/rpm package with
your distribution's package manager. `mantl --version` reports the release and
commit. Extract the platform bundle and pass its directory through --source-dir
for cloud bootstrap. AWS credentials must come from workload identity or the
standard runtime credential chain, never release assets or manifests.

Provide your own evidence-storage ConfigMap and workload identity before applying
the operator manifest. Install Kyverno reports first. Record existing CRDs before
upgrading; schema removal/downgrade is not automatically safe. Retain immutable
evidence and do not shorten its retention during upgrade or uninstall.

Run `mise exec -- goreleaser release --snapshot --clean` for a local packaging
check. Snapshot builds do not publish or sign a release. A passing build is not
evidence of a completed cloud deployment or an independently assessed SLSA level.

Content releases can also use Publish control content independently of CLI tags.
This publishes the version in compliance/bundle-version.txt and refuses changed
content at an existing version. Registry tag immutability remains an administrative
requirement. Read architecture-runtime.md before mounting a separately pinned bundle.

Hosted signature/provenance publication is exercised only by an actual release;
local tests intentionally use snapshot packages. The optional fleet manifest fails
closed until TLS, database, enrollment, identity and network overlays are supplied.

## Prepare an unsigned release candidate locally

Current source includes `ci.verify_release_candidate`; v0.4.0 does not. Install the
contributor Python dependencies (including PyYAML) and the mise-pinned toolchain.
Use a dedicated output directory and explicit prerelease version:

```sh
mkdir -p dist
mise exec -- .venv/bin/python - <<'PY'
from pathlib import Path
import yaml
config = yaml.safe_load(Path('.goreleaser.yaml').read_text())
config['snapshot'] = {'version_template': '0.5.0-rc.1'}
config['dist'] = 'dist/candidate-0.5.0-rc.1'
Path('dist/candidate-config.yaml').write_text(yaml.safe_dump(config, sort_keys=False))
PY
mise exec -- goreleaser release --snapshot --config dist/candidate-config.yaml --clean
mise exec -- .venv/bin/python -m ci.verify_release_candidate 0.5.0-rc.1 \
  --directory dist/candidate-0.5.0-rc.1 --native-packages
```

Docker is required for `--native-packages`; omit that flag for archive/CLI checks
only. The verifier requires all eight CLI archives/packages and their matching
checksums, checks archive documentation assets, runs the native CLI, and exercises
saved-plan generation, unchanged comparison and rejection of changed input. Native
DEB/RPM checks cover installation, reinstallation and removal in isolated containers.

`candidate-verification.json` records the checked version, CLI schema, local source
diff hash, dirty-worktree status and which checks ran. It is an unsigned local test
report, not provenance. This command neither creates a Git tag nor publishes a
release. Hosted signing, provenance, operator/content assets and cloud acceptance
remain separate checks; a successful local CLI report does not establish them.

## Local package lifecycle checks

After building a snapshot, run `ci/verify-cli-packages.sh dist`. The check uses
the Docker engine's native architecture and digest-pinned Ubuntu/Rocky images.
It checks matching release versions in filenames, CLI output and package metadata,
installs and reinstalls both packages, verifies the CLI, license and release
instructions against the matching archive, exercises CLI help, then uninstalls
and checks that package-owned files and registration are gone. Containers have
no network or host credentials; artifacts are mounted read-only. Capabilities are
dropped, with only `DAC_OVERRIDE` retained for RPM to write its image's system
directories. Image pulls happen before the isolated checks.

An optional second argument selects `amd64` or `arm64`, for example
`ci/verify-cli-packages.sh /path/to/downloads arm64`. Cross-architecture execution
requires an existing compatible Docker emulation setup; the check does not install
emulators. Supply exactly one archive, DEB and RPM per selected architecture.
CI checks native snapshot packages and the release workflow checks native release
packages before uploading CLI artifacts. These checks establish package behavior,
not a deployed platform, cloud acceptance or successful version-to-version upgrades.

Release assembly scripts read from their own checkout and constrain all archive
and identity paths to that checkout's `dist` directory. Source-directory
overrides must identify that same checkout. Symlinked or noncanonical artifact
paths are rejected before external publication commands.
