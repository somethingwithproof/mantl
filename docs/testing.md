<!-- SPDX-License-Identifier: Apache-2.0 -->
# Testing and validation boundaries

Use the pinned mise toolchain and [contributor setup](../CONTRIBUTING.md).
Tests have different scopes; none of the commands below is cloud acceptance.

## First-party unit tests

```sh
mise exec -- make go-test
mise exec -- .venv/bin/python -m pytest tests/unit/
```

`make go-test` excludes vendored chart Go packages under `platform/` (ADR 004).
Do not run `go test ./...`. `make go-coverage` measures the same first-party scope;
coverage percentages do not establish complete framework/control coverage.
Python tests exercise auxiliary scripts and contracts; the Go operator is the
runtime source of truth. Example application jobs have separate locked dependencies.
PR example coverage is scoped to selected examples; main runs the complete Python
example matrix. See the workflow inventory for selection and shared-fixture rules.

## Offline compiler and contracts

```sh
mise exec -- go run ./cmd/mantl plan examples/mantl-spec.yaml --dry-run --format json
mise exec -- go run ./cmd/mantl validate-clouds --source-dir .
mise exec -- make manifests generate
```

The first command uses flags introduced in **v0.5.0**. Static cloud validation inspects the
AWS/GCP/Azure contract without deployment. API generation checks drift, not
admission in a real cloud. Rendered Applications reference Git content that must
be reviewed and published separately.

## Disposable integration and packaging

[compliance-runtime.yml](../.github/workflows/compliance-runtime.yml) explicitly
runs integration-tagged tests in `integration/controlapi` and `pkg/fleet` against
an envtest Kubernetes API and disposable PostgreSQL, builds operator entrypoints,
and checks package snapshots. These fixtures test admission, isolation, and package
behavior without developer cloud credentials. They are not a full kind platform
deployment or cloud acceptance suite.

The integration job supplies `KUBEBUILDER_ASSETS` and `MANTL_TEST_DATABASE_URL`.
Use the exact workflow fixture setup when reproducing it locally; do not point the
tests at production databases or clusters. Its local database disables TLS for the
fixture; the fleet deployment requires verified TLS and a role that cannot bypass RLS.

### Local compliance lifecycle

`TestLocalComplianceLifecycle` joins the actual reconcilers, Kubernetes resource
reader, S3 storage adapter, receipt verification and archive export against a
disposable API. It checks a policy violation, versioned finding history, fresh
collection, an approved exception that preserves failure, repaired resource bytes,
verified export, stale/missing evidence and rejection of corrupted stored bytes.

With the contributor toolchain installed, obtain the same disposable API binaries
as CI and run only this fixture:

```sh
mkdir -p /tmp/mantl-envtest-assets
export KUBEBUILDER_ASSETS="$(mise exec -- go run sigs.k8s.io/controller-runtime/tools/setup-envtest@f9589b9f2b9dddf8532b432bb8315f2820ab9971 use 1.35.0 --bin-dir /tmp/mantl-envtest-assets -p path)"
mise exec -- go test -tags=integration ./integration/controlapi -run TestLocalComplianceLifecycle -v
```

The setup command downloads API server/etcd binaries. The test starts and stops
its own API; it needs no kube-context or cloud credentials. Policy reports, approved
exception status and Job completion are seeded fixtures. The S3 API is simulated
with version/retention metadata; no real object store or kubelet is exercised.
This proves local component integration, not Kyverno enforcement, approval-role
authorization, container execution, AWS Object Lock/IAM or cloud acceptance.

`mise exec -- goreleaser release --snapshot --clean` builds local archives/packages.
`ci/verify-cli-packages.sh dist` needs Docker and verifies native package lifecycles
in isolated containers. It may pull pinned images, but does not publish or sign
artifacts. [Release verification](releases.md) is a separate trust check.

## Cloud acceptance gap

Deployment, workload identity, private access, audit delivery, encryption, upgrade,
recovery, and retained evidence need environment-specific acceptance records.
`mantl acceptance verify` authenticates retained receipts; it does not independently
perform or observe the deployment. See [runtime limits](architecture-runtime.md#provider-acceptance-and-oscal).

For actual CI selection and requirements use [the workflow inventory](../.github/workflows/README.md)
and [required checks](ci-required-checks.md), rather than historical E2E descriptions.
