<!-- SPDX-License-Identifier: Apache-2.0 -->
# Verification record: Mantl v0.3.1

Verified on 2026-10-06 from the assets attached to
[the v0.3.1 release](https://github.com/somethingwithproof/mantl/releases/tag/v0.3.1).
The resolved tag and every verified CLI provenance statement identify source
commit `c202d881d67e68c5d8720afa00e46a0a230a16bd`. This record establishes artifact
verification, not cloud deployment acceptance or an independent SLSA-level assessment.

Cosign 3.1.3 verified `checksums.txt` using `checksums.sigstore.json`, the exact
workflow certificate identity and issuer shown below, repository
`somethingwithproof/mantl`, and the source commit above. All 26 checksum targets
matched their SHA256 values. The provenance statement's 26 subject digests also
matched that signed manifest exactly.

SLSA verifier 2.7.1 independently verified each of the eight CLI assets below
using `multiple.intoto.jsonl`, the repository, branch and builder shown below.
Every result identified the source commit above. The generator invocation is on
`main`; resolving the release tag and checking the exact source commit establishes
the v0.3.1 association. A source-tag assertion against that main-branch invocation
is not used.

| CLI asset | SHA256 | Provenance verification |
| --- | --- | --- |
| `mantl_0.3.1_darwin_amd64.tar.gz` | `20399a4ab4ce892fd2f022919ca805c787e42ae53177062c986cbe6cc2a9e6a2` | Passed |
| `mantl_0.3.1_darwin_arm64.tar.gz` | `397068983caaf09d9e82577b3f90bf94af8d53021ba6e3c346b03c999e3dd885` | Passed |
| `mantl_0.3.1_linux_amd64.deb` | `e0412c94cc107335ee54f20ba87f603f80faff5a3ddc1abc000421d7a98bf7b7` | Passed |
| `mantl_0.3.1_linux_amd64.rpm` | `3b528a2d6502a4ff5dc665cee9f7bba330c0b4772861f09bfec2b9085a6d630c` | Passed |
| `mantl_0.3.1_linux_amd64.tar.gz` | `12505eaa3ad5594ac5b18a99c5fd067e7c0f448c06413d932a4b9677af1cdae3` | Passed |
| `mantl_0.3.1_linux_arm64.deb` | `78492faec64f763cc3941c0e0821d851c3d4a69488eb4109c186fe2697ab204c` | Passed |
| `mantl_0.3.1_linux_arm64.rpm` | `b823d14c9ead0924d61a9398dddbafce061c31dee4766bca7652e7c1e75a898c` | Passed |
| `mantl_0.3.1_linux_arm64.tar.gz` | `9c75dcf814c64f43cb1f3c765badd72af8da3237e05df4a587d6c0b3c468e3db` | Passed |

## Reproduce verification

Download all assets into an empty directory, then use the tool versions above:

```sh
gh release download v0.3.1 --repo somethingwithproof/mantl
cosign verify-blob checksums.txt --bundle checksums.sigstore.json \
  --certificate-identity https://github.com/somethingwithproof/mantl/.github/workflows/release.yml@refs/heads/main \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-github-workflow-repository somethingwithproof/mantl \
  --certificate-github-workflow-sha c202d881d67e68c5d8720afa00e46a0a230a16bd
sha256sum -c checksums.txt
slsa-verifier verify-artifact mantl_0.3.1_*.tar.gz mantl_0.3.1_*.deb mantl_0.3.1_*.rpm \
  --provenance-path multiple.intoto.jsonl \
  --source-uri github.com/somethingwithproof/mantl \
  --source-branch main \
  --builder-id https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.1.0
```

Resolve the GitHub tag to the source commit above and require every verifier
result to show that same commit. Keep these checks distinct from
[local package lifecycle checks](releases.md#local-package-lifecycle-checks):
signatures authenticate publication, provenance binds artifacts to their builder
and source, while package checks exercise installation and removal.
