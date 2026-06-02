# Phase 1: Foundation and Integrity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the Mantl repository trustworthy for outside contributors: green and honest CI on public runners, documentation that matches the code, and clean Go module boundaries.

**Architecture:** Seven independent tasks. Each fixes one integrity defect and commits on its own. No task depends on another, so they can run in parallel or any order. All work happens in `~/Desktop/mantl`.

**Tech Stack:** Go 1.26, GitHub Actions, pre-commit, Markdown.

**Pre-commit note:** The repo's pre-commit `terraform-docs` hook is broken (Task 4 fixes it). Until Task 4 lands, commit Go/docs-only changes with `git commit --no-verify` to avoid the unrelated hook failure. Do not use `--no-verify` to skip a hook that is actually relevant to your change.

---

### Task 1: Fix the environment-dependent evidence test

The `pkg/evidence` validation tests call `CaptureResource`, which expects `kubectl` to fail. On a machine with a live cluster, `kubectl get deployments -n default` succeeds, so the test fails. Extract validation into a pure function and test that directly, removing the cluster dependency.

**Files:**
- Modify: `pkg/evidence/snapshot.go`
- Modify: `pkg/evidence/snapshot_test.go`

- [ ] **Step 1: Add the `validateCaptureArgs` function**

In `pkg/evidence/snapshot.go`, add this function immediately above `CaptureResource`:

```go
// validateCaptureArgs checks kind, name, and namespace against the evidence
// collection allowlist and syntax rules. It returns nil when the arguments are
// safe to pass to kubectl. Keeping validation separate from process execution
// lets tests exercise the rules without a live cluster.
func validateCaptureArgs(kind, name, namespace string) error {
	if !validKubeKindRe.MatchString(kind) {
		return fmt.Errorf("invalid resource kind %q", kind)
	}
	if !allowedEvidenceKinds[kind] {
		return fmt.Errorf("resource kind %q is not permitted for evidence collection", kind)
	}
	if name != "" {
		if len(name) > maxKubeNameLen {
			return fmt.Errorf("resource name exceeds maximum length of %d", maxKubeNameLen)
		}
		if !validKubeNameRe.MatchString(name) {
			return fmt.Errorf("invalid resource name %q", name)
		}
	}
	if namespace == "" {
		return fmt.Errorf("namespace is required for evidence collection")
	}
	if len(namespace) > maxNamespaceLen {
		return fmt.Errorf("namespace exceeds maximum length of %d", maxNamespaceLen)
	}
	if !validNamespaceRe.MatchString(namespace) {
		return fmt.Errorf("invalid namespace %q", namespace)
	}
	return nil
}
```

- [ ] **Step 2: Call the new function from `CaptureResourceWithContext`**

In `pkg/evidence/snapshot.go`, replace the inline validation block (the eight `if` checks from `if !validKubeKindRe.MatchString(kind)` through the `validNamespaceRe` check, ending before the `// Place our flags before "--"` comment) with:

```go
	if err := validateCaptureArgs(kind, name, namespace); err != nil {
		return nil, err
	}
```

- [ ] **Step 3: Replace the flaky test**

In `pkg/evidence/snapshot_test.go`, replace the entire `TestCaptureResource_ResourceID` function with:

```go
func TestCaptureResource_ResourceID(t *testing.T) {
	// Validation and resource-identifier construction must not depend on a live
	// cluster. Exercise them directly rather than through kubectl.
	cases := []struct {
		name    string
		kind    string
		resName string
		wantID  string
	}{
		{"kind-only when name is empty", "Deployment", "", "Deployment"},
		{"kind/name when name is set", "Deployment", "my-app", "Deployment/my-app"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateCaptureArgs(tc.kind, tc.resName, "default"); err != nil {
				t.Fatalf("validateCaptureArgs returned error for valid input: %v", err)
			}
			if got := BuildResourceID(tc.kind, tc.resName); got != tc.wantID {
				t.Errorf("BuildResourceID(%q, %q) = %q, want %q", tc.kind, tc.resName, got, tc.wantID)
			}
		})
	}
}
```

- [ ] **Step 4: Run the package tests**

Run: `cd ~/Desktop/mantl && go test ./pkg/evidence/ -v`
Expected: PASS. If the build fails with `"strings" imported and not used`, remove the now-unused `strings` import from `snapshot_test.go` and re-run.

- [ ] **Step 5: Commit**

```bash
cd ~/Desktop/mantl
git add pkg/evidence/snapshot.go pkg/evidence/snapshot_test.go
git commit --no-verify -m "test(evidence): make capture validation testable without a cluster"
```

---

### Task 2: Exclude vendored charts from the Go test scope

`go test ./...` descends into vendored Helm charts under `platform/` (Loki, Falco) whose Go tests fail. Define the canonical test target to exclude only the vendored `platform/` tree, keeping the project's own `apis/platform/` package.

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Confirm which packages are vendored**

Run: `cd ~/Desktop/mantl && go list ./... | grep 'mantl/platform/'`
Expected output (the packages to exclude):
```
github.com/thomasvincent/mantl/platform/observability/loki/charts/loki-6.23.0/loki/test
github.com/thomasvincent/mantl/platform/security/falco/charts/falco-4.15.0/falco/charts/k8s-metacollector/tests/unit
github.com/thomasvincent/mantl/platform/security/falco/charts/falco-4.15.0/falco/tests/unit
```
Confirm `mantl/apis/platform/v1alpha1` is NOT in this list (the filter must keep it).

- [ ] **Step 2: Add a `go-test` make target**

Add to `Makefile` (place near the existing `test` targets):

```makefile
# GO_PKGS excludes vendored Helm charts under platform/ whose upstream Go tests
# are not ours to maintain. The apis/platform package is kept.
GO_PKGS := $(shell go list ./... | grep -v 'github.com/thomasvincent/mantl/platform/')

go-test: ## Run Go unit tests excluding vendored charts
	go test $(GO_PKGS)
```

- [ ] **Step 3: Verify the target is green**

Run: `cd ~/Desktop/mantl && make go-test`
Expected: PASS for all project packages; no `platform/...loki` or `platform/...falco` lines.

- [ ] **Step 4: Commit**

```bash
cd ~/Desktop/mantl
git add Makefile
git commit --no-verify -m "build: exclude vendored platform charts from go test scope"
```

---

### Task 3: Make CI honest and runnable by outside contributors

The `validate` job runs on a `self-hosted` runner (never runs for forks/outside PRs) and pins Go 1.23 while `go.mod` is 1.26. Move it to a public runner, read the Go version from `go.mod`, and use the filtered test target.

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Edit the `validate` job**

In `.github/workflows/ci.yml`, change line 11 from:
```yaml
    runs-on: [self-hosted, linux, x64]
```
to:
```yaml
    runs-on: ubuntu-latest
```

Change the `Set up Go` step (lines 15-19) from:
```yaml
    - name: Set up Go
      uses: actions/setup-go@v6
      with:
        go-version: '1.23'
        cache: true
```
to:
```yaml
    - name: Set up Go
      uses: actions/setup-go@v6
      with:
        go-version-file: go.mod
        cache: true
```

Change the `Run unit tests` step (lines 30-31) from:
```yaml
    - name: Run unit tests
      run: go test -v ./...
```
to:
```yaml
    - name: Run unit tests
      run: make go-test
```

- [ ] **Step 2: Validate the workflow syntax**

Run: `cd ~/Desktop/mantl && python3 -c "import yaml,sys; yaml.safe_load(open('.github/workflows/ci.yml')); print('yaml ok')"`
Expected: `yaml ok`

- [ ] **Step 3: Dry-run the commands the job will run**

Run: `cd ~/Desktop/mantl && go build -o bin/mantl ./cmd/mantl && make go-test && ./bin/mantl plan examples/mantl-spec.yaml && go build -o bin/compliance-operator ./cmd/compliance-operator && echo "all ci steps ok"`
Expected: ends with `all ci steps ok`.

- [ ] **Step 4: Commit**

```bash
cd ~/Desktop/mantl
git add .github/workflows/ci.yml
git commit --no-verify -m "ci: run validate on public runner with go.mod version and filtered tests"
```

---

### Task 4: Fix the broken pre-commit terraform-docs hook

The `terraform-docs/terraform-docs` hook at `rev: v0.19.0` fails with `InvalidManifestError: .pre-commit-hooks.yaml is not a file`. The maintained `antonbabenko/pre-commit-terraform` repo (already referenced at line 60) provides a working `terraform_docs` hook. Remove the broken block and add the working hook to the existing repo entry.

**Files:**
- Modify: `.pre-commit-config.yaml`

- [ ] **Step 1: Inspect both repo blocks**

Run: `cd ~/Desktop/mantl && sed -n '60,92p' .pre-commit-config.yaml`
Note the `antonbabenko/pre-commit-terraform` block (line 60) and its `hooks:` list, and the broken `terraform-docs/terraform-docs` block (lines 88-92).

- [ ] **Step 2: Remove the broken block**

Delete these lines from `.pre-commit-config.yaml`:
```yaml
  - repo: https://github.com/terraform-docs/terraform-docs
    rev: v0.19.0
    hooks:
      - id: terraform-docs-go
        args: ["markdown", "table", "--output-file", "README.md", "."]
        files: ^infra/terraform/.*
```

- [ ] **Step 3: Add the working hook under the antonbabenko block**

In the `antonbabenko/pre-commit-terraform` block's `hooks:` list, add:
```yaml
      - id: terraform_docs
        args:
          - --hook-config=--path-to-file=README.md
          - --hook-config=--add-to-existing-file=true
          - --hook-config=--create-file-if-not-exist=true
```

- [ ] **Step 4: Verify pre-commit initializes cleanly**

Run: `cd ~/Desktop/mantl && pre-commit run terraform_docs --all-files; echo "exit=$?"`
Expected: the hook installs and runs without `InvalidManifestError`. A non-zero exit from the hook modifying docs is acceptable; an initialization error is not.

- [ ] **Step 5: Commit**

```bash
cd ~/Desktop/mantl
git add .pre-commit-config.yaml
git commit --no-verify -m "build: replace broken terraform-docs hook with maintained terraform_docs"
```

---

### Task 5: README truth pass and maturity matrix

The README shows a fabricated `mantl compliance status` dashboard for a command that does not exist, claims "production-ready in under 5 minutes" for what is a local kind cluster, and marks frameworks/clouds without honest maturity levels.

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Correct the quick-start claim**

In `README.md` line 19, change:
```
Get a complete production-ready platform running in **under 5 minutes**:
```
to:
```
Spin up a local development platform on a kind cluster in a few minutes:
```

- [ ] **Step 2: Replace the fabricated compliance dashboard**

Replace the `## Compliance Module` example block (the fenced `bash` block containing `mantl compliance status`, the ASCII dashboard, and `mantl compliance audit --output report.pdf`) with an honest description:

````markdown
## Compliance Module

Mantl's differentiator is its Compliance-as-Code module. A `ComplianceProfile`
selects controls from a framework; controls map to admission policies; policy
violations surface as `Finding` resources; and audits capture point-in-time
evidence to object storage.

```bash
# Apply a compliance profile to the cluster
kubectl apply -f compliance/frameworks/soc2/profile-standard.yaml

# Inspect findings produced by policy violations
kubectl get findings -A
```

A `mantl compliance status` CLI summary is planned (see the roadmap); today,
findings and audits are inspected through `kubectl`.
````

- [ ] **Step 3: Add a maturity matrix**

Immediately after the frameworks list, add:

```markdown
### Maturity

Cloud blueprints:

| Cloud | Status |
|-------|--------|
| AWS (EKS) | GA |
| GCP (GKE) | beta (hardening in progress) |
| Azure (AKS) | beta (hardening in progress) |
| Linode, Oracle, IBM, DigitalOcean | experimental |

Compliance frameworks:

| Framework | Status |
|-----------|--------|
| SOC2 Type II | policies and evidence wiring in progress |
| CIS Kubernetes | policies in progress |
| HIPAA, PCI-DSS | declarations only, not yet enforced |

Maturity reflects what the code enforces today, not aspirations.
```

- [ ] **Step 4: Verify no fabricated claims remain**

Run: `cd ~/Desktop/mantl && grep -n "Score:\|Evidence Items\|Last Audit\|report.pdf\|under 5 minutes\|production-ready" README.md; echo "exit=$?"`
Expected: no matches (grep exits non-zero / prints nothing).

- [ ] **Step 5: Commit**

```bash
cd ~/Desktop/mantl
git add README.md
git commit --no-verify -m "docs: align README with implemented features and add maturity matrix"
```

---

### Task 6: Quarantine dead Nomad/Consul code

`infrastructure/nomad/` holds legacy Mesos-era Nomad/Consul files disconnected from the Kubernetes path. Confirm nothing references them, then move them under a documented legacy directory rather than deleting outright (preserves heritage; signals intent).

**Files:**
- Move: `infrastructure/nomad/` to `docs/legacy/nomad/`
- Create: `docs/legacy/README.md`

- [ ] **Step 1: Confirm no active references**

Run: `cd ~/Desktop/mantl && grep -rn "infrastructure/nomad" --include="*.yaml" --include="*.yml" --include="*.go" --include="*.sh" --include="Makefile" . | grep -v "docs/superpowers"`
Expected: no output. If any active kustomization or script references it, stop and report rather than moving.

- [ ] **Step 2: Move the directory**

```bash
cd ~/Desktop/mantl
mkdir -p docs/legacy
git mv infrastructure/nomad docs/legacy/nomad
```

- [ ] **Step 3: Document the quarantine**

Create `docs/legacy/README.md`:

```markdown
# Legacy artifacts

This directory preserves code from earlier Mantl iterations that is not part of
the current Kubernetes-native platform.

## nomad/

Nomad and Consul configurations from the Mesos/Marathon-era design. Retained for
historical reference. Not wired into the GitOps or compliance paths and not
covered by CI. Do not deploy without revalidation.
```

- [ ] **Step 4: Verify the build is unaffected**

Run: `cd ~/Desktop/mantl && go build ./... && make go-test && echo "build+test ok"`
Expected: ends with `build+test ok`.

- [ ] **Step 5: Commit**

```bash
cd ~/Desktop/mantl
git add -A
git commit --no-verify -m "chore: quarantine legacy nomad/consul code under docs/legacy"
```

---

### Task 7: Record Phase 1 decisions as ADRs

The repo has one ADR. Record the decisions made in Phase 1 so later phases inherit the correct mental model. Follow the format of `docs/adr/003-evidence-storage-contract.md` (Status, Context, Decision, Rationale, Consequences).

**Files:**
- Create: `docs/adr/004-test-scope-vendored-charts.md`
- Create: `docs/adr/005-cloud-parity-scope.md`

- [ ] **Step 1: Write ADR 004**

Create `docs/adr/004-test-scope-vendored-charts.md`:

```markdown
# ADR 004: Go Test Scope Excludes Vendored Charts

## Status
Accepted

## Context
The `platform/` tree vendors third-party Helm charts (Loki, Falco). Some carry
their own Go tests. `go test ./...` ran those tests, and they failed, turning CI
red for code Mantl does not own.

## Decision
The canonical test target excludes the vendored `platform/` packages while
keeping the project's own `apis/platform/` package. It is defined in the Makefile
as `GO_PKGS := go list ./... | grep -v 'github.com/thomasvincent/mantl/platform/'`
and run via `make go-test`. CI calls `make go-test`, not `go test ./...`.

## Rationale
Vendored chart tests are upstream concerns. Mantl's CI should gate Mantl's code.
Filtering at the test target is less invasive than dropping `go.mod` files into
each vendored chart, which chart updates would overwrite.

## Consequences
Contributors must use `make go-test` rather than `go test ./...`. A future
restructure that relocates vendored charts out of the module would let this
filter be removed.
```

- [ ] **Step 2: Write ADR 005**

Create `docs/adr/005-cloud-parity-scope.md`:

```markdown
# ADR 005: Cloud Parity Scope

## Status
Accepted

## Context
The README advertised seven clouds. Only AWS EKS was production-grade; GCP and
Azure were bare module wrappers, and the rest were unverified. Maintaining seven
hardened blueprints is not realistic for the current team.

## Decision
Parity targets three clouds: AWS (EKS), GCP (GKE), and Azure (AKS), brought to a
common hardening baseline (private cluster, control-plane audit logging,
encryption at rest, workload identity, least-privilege IAM). Linode, Oracle, IBM,
and DigitalOcean are labeled experimental in docs and directory layout.

## Rationale
Three credible clouds serve more real users than seven half-built ones. Honest
maturity labels protect the project's credibility.

## Consequences
Phase 3 hardens GCP and Azure to the AWS baseline. Experimental blueprints carry
no parity or CI guarantees until a contributor adopts them.
```

- [ ] **Step 3: Verify ADRs render**

Run: `cd ~/Desktop/mantl && ls docs/adr/ && head -3 docs/adr/004-test-scope-vendored-charts.md docs/adr/005-cloud-parity-scope.md`
Expected: both files listed, each starting with its `# ADR` heading.

- [ ] **Step 4: Commit**

```bash
cd ~/Desktop/mantl
git add docs/adr/004-test-scope-vendored-charts.md docs/adr/005-cloud-parity-scope.md
git commit --no-verify -m "docs(adr): record test-scope and cloud-parity decisions"
```

---

## Definition of done for Phase 1

- [ ] `go build ./...` passes.
- [ ] `make go-test` passes with no vendored `platform/` packages in scope.
- [ ] `.github/workflows/ci.yml` `validate` job runs on `ubuntu-latest` with `go-version-file: go.mod`.
- [ ] `pre-commit run --all-files` initializes without `InvalidManifestError`.
- [ ] README contains no fabricated CLI output and includes the maturity matrix.
- [ ] `infrastructure/nomad/` no longer exists at the repo root; legacy code is under `docs/legacy/`.
- [ ] ADRs 004 and 005 exist.
- [ ] All changes pushed as a PR against `main` with the CI now running on a public runner.
