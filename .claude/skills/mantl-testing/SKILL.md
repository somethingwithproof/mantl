---
name: mantl-testing
description: Load when writing or running tests in the mantl repo. Covers the canonical make go-test target and why it excludes vendored platform/ (ADR 004), controller test structure, the rule that no test may depend on a live cluster or kubectl, and table-driven Go test style.
---
<!-- SPDX-License-Identifier: Apache-2.0 -->

# mantl testing skill

## The canonical target

```bash
make go-test    # go test $(GO_PKGS)
```

`GO_PKGS` is `go list ./...` filtered to drop `github.com/thomasvincent/mantl/platform/`. Always use this target. Do not run `go test ./...` directly.

## Why platform/ is excluded (ADR 004)

`platform/` vendors third-party Helm charts (Loki, Falco). Some carry their own Go tests that fail under our toolchain. Those failures are upstream concerns, not Mantl bugs, and they turned CI red for code Mantl does not own. The filter lives in the test target so chart updates cannot reintroduce the breakage. CI calls `make go-test`, not bare `go test`. The project's own `apis/platform/` package is in scope; only the vendored `platform/` tree is excluded.

## No live cluster, no kubectl

Tests must run with no Kubernetes cluster reachable and no `kubectl` on PATH. A test that shells out to `kubectl` or talks to a real API server is broken.

- Test reconcilers with a fake client (`sigs.k8s.io/controller-runtime/pkg/client/fake`) or envtest, not a live cluster.
- Construct CRD objects in-process and assert on the resulting client state, not on cluster reality.
- `finding_controller_test.go` is the pattern to follow for controller tests: build input objects, run `Reconcile`, assert the output.

Tests for `pkg/` are plain Go and need no Kubernetes at all. `pkg/compiler`, `pkg/bootstrap`, and `pkg/evidence` each keep their tests beside the code (`spec_parser_test.go`, `dag_test.go`, `snapshot_test.go`). Evidence tests must not perform real S3 calls; exercise the snapshot/build path and stub or skip the upload.

## Table-driven style

Match the Go-standard table-driven form. One test function, a slice of named cases, a subtest per case.

```go
func TestNormalizeSeverity(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "critical stays critical", in: "critical", want: "critical"},
		{name: "unknown falls back", in: "bogus", want: "medium"},
		{name: "empty falls back", in: "", want: "medium"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeSeverity(tc.in); got != tc.want {
				t.Errorf("normalizeSeverity(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

Conventions:

- Name each case so a failure points at the scenario.
- Use `t.Run` subtests; do not loop with bare asserts.
- Report with `t.Errorf`/`t.Fatalf` and the standard "got, want" phrasing.
- No assertion libraries unless the package already uses one; check the surrounding tests first.

## Controller tests

- Build the watched resources (a Kyverno `PolicyReport` for `finding_controller`, a `ComplianceProfile` for `profile_controller`) as in-memory objects.
- Seed them into a fake client, run `Reconcile`, then read back the resources the controller created or updated and assert on them.
- Reconcilers are idempotent; a useful case is reconciling twice and asserting the second pass is a no-op.

## Before claiming a change works

Run `make go-test` and read the output. Do not infer success from a compile. A change to a CRD type that skips DeepCopy regeneration compiles in some configurations but fails the controller build; the test run is what proves it.
