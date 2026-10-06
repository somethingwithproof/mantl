# Release installation and verification

Release Please manages SemVer and changelogs. Its release-created output invokes
the reusable release workflow directly; publishing does not depend on another
workflow being triggered by a GITHUB_TOKEN-created release event.

The release checks out the exact validated tag and publishes:

- Linux/macOS amd64/arm64 CLI tar.gz archives;
- Linux amd64/arm64 deb and rpm packages;
- the digest-pinned operator installation manifest and optional fleet installation manifests and platform source bundle;
- a versioned control bundle with a content manifest;
- SHA256 checksums, SPDX SBOMs and keyless Cosign verification bundles;
- a multiarchitecture operator image and OCI bundles.

Download assets using `gh release download TAG --repo somethingwithproof/mantl`.
Verify checksums with `sha256sum -c checksums.txt`. Verify the checksum signature
with Cosign and an exact GitHub Actions workflow identity for this repository;
do not accept arbitrary certificate identities. release-identity.json records the resolved operator digest; control-oci-identity.json
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

## Local package lifecycle checks

After building a snapshot, run `ci/verify-cli-packages.sh dist`. The check uses
the Docker engine's native architecture and digest-pinned Ubuntu/Rocky images.
It installs and reinstalls both packages, verifies the CLI, license and release
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
