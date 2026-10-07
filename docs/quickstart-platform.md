<!-- SPDX-License-Identifier: Apache-2.0 -->
# Getting started with Mantl

Start by generating reviewable configuration offline. The commands below target
release **v0.5.0**. This workflow needs no cloud account,
Kubernetes cluster, Terraform/OpenTofu, Docker, or Go.

## Install a verified release

Prerequisites: a shell, `tar`, [GitHub CLI](https://cli.github.com/) (`gh`) for download,
[Cosign](https://github.com/sigstore/cosign) (the repository pins 3.1.3), and
`sha256sum` on Linux or `shasum` on macOS. A browser can download the same assets
from [the release page](https://github.com/somethingwithproof/mantl/releases/tag/v0.5.0)
if `gh` is unavailable. Keep the complete set together for the checksum command.

Create a new directory, download all attached release assets, and authenticate
`checksums.txt` **before** using it:

```sh
mkdir mantl-v0.5.0
cd mantl-v0.5.0
gh release download v0.5.0 --repo somethingwithproof/mantl
cosign verify-blob checksums.txt --bundle checksums.sigstore.json \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity https://github.com/somethingwithproof/mantl/.github/workflows/release.yml@refs/heads/main
```

Then verify the signed manifest's listed files:

```sh
# Linux
sha256sum -c checksums.txt
# macOS: use this instead
shasum -a 256 -c checksums.txt
```

Require successful signature and checksum verification. `release-identity.json`
records the exact source commit for the release and the pinned operator image.
Compare that commit with the source tag and provenance. See [release verification](releases.md) for provenance,
SBOMs (software bills of materials), container identities, and package details.

Select the archive matching your OS/architecture:
`mantl_0.5.0_linux_amd64.tar.gz`, `mantl_0.5.0_linux_arm64.tar.gz`,
`mantl_0.5.0_darwin_amd64.tar.gz`, or `mantl_0.5.0_darwin_arm64.tar.gz`.
`uname -s` and `uname -m` identify the host; `x86_64` means amd64 and
`aarch64`/Apple `arm64` means arm64. For example, on Apple Silicon:

```sh
mkdir cli
tar -xzf mantl_0.5.0_darwin_arm64.tar.gz -C cli
./cli/mantl --version
```

Replace that filename on other hosts. Expected version: `0.5.0`, with the commit
matching `release-identity.json`. You can keep using `./cli/mantl`, place it in a directory on your PATH, or
install a verified Linux DEB/RPM through your distribution's package manager.
There is no Windows CLI asset.

## Generate configuration offline

Save the small [README specification](../README.md#a-small-example) as
`platform.yaml` in this directory; it does not require a repository checkout.
Then run:

```sh
./cli/mantl plan platform.yaml
cat .mantl/build/terraform.tfvars.json
cat .mantl/build/gitops/root-platform.yaml
cat .mantl/build/tenants/payments.yaml
```

Expect the six files listed in the README: variables, three Applications, a tenant
index, and the tenant's namespace/RBAC/network/quota/limit manifests. Inspect the
repository/revision/path references and admin identities before using them.
`plan` only validates and writes files; it does not resolve Git paths, evaluate
compliance, contact Kubernetes, or run Terraform.

For more examples, obtain the **matching tagged source checkout** explicitly:

```sh
git clone --branch v0.5.0 --depth 1 https://github.com/somethingwithproof/mantl.git mantl-source
./cli/mantl plan mantl-source/examples/mantl-spec.yaml
```

That fuller example includes illustrative tenant identities and a local operator
overlay; customize it before deployment. Generation reuses `.mantl/build`; use
separate working directories to keep examples apart. See [planning](platform-planning.md)
for file replacement behavior and spec fields.

## Preview and review before deployment

The v0.5.0 CLI supports plan previews, JSON inventories, saved-plan comparisons,
and reviewed `apply --plan` with preflight and scoped convergence. Try the offline
preview against the file you saved above:

```sh
./cli/mantl plan platform.yaml --dry-run --format json
```

Follow [preview and inventory](platform-planning.md#preview-and-inventory)
to save and compare reviewed inputs. **v0.4.0 does not accept these plan/apply
flags**. A compiler plan is not an infrastructure resource-change plan.

## Deployment is a separate step

The v0.5.0 `apply` creates resources and modifies a cluster. It runs Terraform/OpenTofu apply
without a saved Terraform plan approval step, installs pinned ArgoCD v3.5.3, submits
the generated Applications, and polls reported application health. Review changes
with your infrastructure tool and Git process before invoking it. Cloud deployment
and recovery acceptance have not been established; use disposable environments
for your own evaluation and account for cluster, network, compute, storage, and
long-lived evidence-retention costs.

Before cloud bootstrap you must:

1. Verify and extract `mantl-platform_v0.5.0.tar.gz` into a source directory. Its
   `infra/terraform/blueprints` supplies configurations; generated variables alone
   cannot provision a cluster. Review the selected blueprint's required variables,
   provider identity, private connectivity, state/backend configuration, and costs.
2. Install OpenTofu or Terraform and `kubectl`. Mantl resolves kubectl from
   `/usr/local/bin`, `/usr/bin`, or `/opt/homebrew/bin`; for a tool-manager install,
   set `MANTL_KUBECTL_PATH` to its trusted absolute executable path. Supply cloud credentials through
   the provider's runtime identity/credential chain, never committed files. Ensure
   your explicit kube-context authenticates to the intended cluster once it exists;
   Mantl does not install Terraform outputs into your kubeconfig.
3. Use a Git repository you control containing the referenced platform paths and
   reviewed overlays. Commit generated tenant files at `gitops.tenantPath`; the
   local build directory is not an ArgoCD source. Configure ArgoCD repository access.
4. Review tenant admin bindings and add required DNS/service egress rules. Confirm
   your CNI enforces NetworkPolicy; choose a provider-supported Kubernetes version.
5. If enabling compliance, supply a digest-pinned operator overlay, Kyverno reports,
   evidence configuration, scoped collector identities, and a versioned AWS S3
   Object Lock bucket. Follow [compliance installation](compliance-operations.md#install).

After those preparations, the CLI shape is:

```sh
# Provisioning / cluster modification: replace context and paths for your environment.
mantl --context YOUR_CONTEXT --source-dir /path/to/extracted/platform apply platform.yaml
# Read-only observation after bootstrap:
mantl --context YOUR_CONTEXT status platform.yaml --format json --require-healthy
```

The local `provider.kind: local` / `distribution: kind` path also changes your
machine: it requires Docker, kind, and kubectl, creates `kind-<metadata.name>`,
and downloads/installs platform components. It is not part of the offline quickstart.

Use [scoped status](platform-status.md) for the expected Application identities;
its health gate does not assess compliance or nested workload/provider readiness.
Continue with [operations](runbooks.md) and [troubleshooting](troubleshooting.md).
