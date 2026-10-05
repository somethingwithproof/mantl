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

Release assembly scripts read from their own checkout and constrain all archive
and identity paths to that checkout's `dist` directory. Source-directory
overrides must identify that same checkout. Symlinked or noncanonical artifact
paths are rejected before external publication commands.
