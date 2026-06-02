# Phase 2a: Compliance Wiring (SOC2 Linchpin Slice) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Make the SOC2 compliance path stop being a silent no-op: parse the real framework schema, resolve control templates to the policies that exist, surface what is unresolved, and let policy violations flow into `Finding`s.

**Architecture:** The profile controller reads a framework YAML and is supposed to enforce the referenced Kyverno policies. Today a Go/YAML schema mismatch makes it parse zero controls. This slice fixes the schema, adds a pure template-to-policy resolver, wires the result into status, and connects the `FindingReconciler` watch so `PolicyReport` failures become `Finding`s. Heavier work (authoring the 10 missing policies, evidence collectors, the audit scheduler, CRD generation, score surfacing) is explicitly deferred to Phase 2b and noted at the end.

**Tech Stack:** Go 1.26, controller-runtime, sigs.k8s.io/yaml, Kyverno `wgpolicyk8s.io` PolicyReports.

**Scope boundary (read first):** This slice is unit-testable without a live cluster. It does NOT author missing policies, implement evidence collectors, add a scheduler, generate CRDs, or compute a score. Those are Phase 2b. Run `make go-test` (never `go test ./...`).

---

### Task 1: Fix the framework Go schema to match the YAML

The `Framework` struct expects `controls[].policies` and `controls[].evidenceResources`, but `compliance/frameworks/soc2/framework.yaml` uses `controls[].mappings[].policyRef.template` and `controls[].mappings[].evidenceCollector`. `yaml.Unmarshal` silently yields empty slices, so the controller enforces nothing. Replace the struct and add a regression test that parses the real file.

**Files:**
- Modify: `controllers/compliance/profile_controller.go`
- Create: `controllers/compliance/framework_test.go`

- [ ] **Step 1: Write the failing test against the real framework file**

Create `controllers/compliance/framework_test.go`:

```go
package compliance

import (
	"os"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestFrameworkParsesRealSOC2(t *testing.T) {
	data, err := os.ReadFile("../../compliance/frameworks/soc2/framework.yaml")
	if err != nil {
		t.Fatalf("read framework: %v", err)
	}
	var fw Framework
	if err := yaml.Unmarshal(data, &fw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(fw.Controls) == 0 {
		t.Fatal("no controls parsed")
	}
	// CC6.1 must expose its two policy templates via mappings.
	var cc61 *Control
	for i := range fw.Controls {
		if fw.Controls[i].ID == "CC6.1" {
			cc61 = &fw.Controls[i]
		}
	}
	if cc61 == nil {
		t.Fatal("CC6.1 not found")
	}
	got := templatesOf(cc61)
	if len(got) != 2 || got[0] != "require-network-policy" || got[1] != "require-pod-security-context" {
		t.Fatalf("CC6.1 templates = %v, want [require-network-policy require-pod-security-context]", got)
	}
}

func templatesOf(c *Control) []string {
	var out []string
	for _, m := range c.Mappings {
		if m.PolicyRef != nil && m.PolicyRef.Template != "" {
			out = append(out, m.PolicyRef.Template)
		}
	}
	return out
}
```

- [ ] **Step 2: Run it; confirm it fails to compile (Control/Mappings undefined)**

Run: `cd ~/Desktop/mantl && go test ./controllers/compliance/ -run TestFrameworkParsesRealSOC2`
Expected: build failure, `undefined: Control` / `c.Mappings`.

- [ ] **Step 3: Replace the struct in `profile_controller.go`**

In `controllers/compliance/profile_controller.go`, replace the `Framework` and `EvidenceResource` type block (the `type Framework struct {...}` through the end of `type EvidenceResource struct {...}`) with:

```go
// Framework mirrors the on-disk compliance framework YAML
// (compliance/frameworks/<name>/framework.yaml).
type Framework struct {
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Controls []Control `json:"controls"`
}

// Control is a single framework control with its policy and evidence mappings.
type Control struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Severity string    `json:"severity"`
	Mappings []Mapping `json:"mappings"`
	// EvidenceResources is retained for the audit controller, which still reads
	// it. The framework YAML does not populate it (so it stays nil and that loop
	// is a no-op), but keeping the field avoids breaking audit_controller.go.
	// The audit controller migrates to Mappings/EvidenceCollector in Phase 2b.
	EvidenceResources []EvidenceResource `json:"evidenceResources,omitempty"`
}

// EvidenceResource is a Kubernetes resource captured as evidence. Retained for
// audit_controller.go compatibility; superseded by EvidenceCollector in Phase 2b.
type EvidenceResource struct {
	Kind      string `json:"kind"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Group     string `json:"group,omitempty"`
}

// Mapping is one entry under a control: either a policy reference or an
// evidence collector. Type is "policy" or "evidence".
type Mapping struct {
	Type              string             `json:"type"`
	PolicyRef         *PolicyRef         `json:"policyRef,omitempty"`
	EvidenceCollector *EvidenceCollector `json:"evidenceCollector,omitempty"`
}

// PolicyRef names a Kyverno policy template and its enforcement mode.
type PolicyRef struct {
	Template string `json:"template"`
	Mode     string `json:"mode"`
}

// EvidenceCollector describes periodic evidence capture for a control.
// Parsed now; consumed in Phase 2b.
type EvidenceCollector struct {
	Type          string   `json:"type"`
	Schedule      string   `json:"schedule"`
	Query         string   `json:"query,omitempty"`
	Resources     []string `json:"resources,omitempty"`
	RetentionDays int      `json:"retentionDays,omitempty"`
}
```

- [ ] **Step 4: Update the reconcile body to read mappings**

In `profile_controller.go` Reconcile, replace the control-to-policy loop (the `var policiesToEnforce []string` block through its `l.Info("Dynamic mapping complete", ...)`) with:

```go
	var referenced []string
	for _, control := range framework.Controls {
		for _, m := range control.Mappings {
			if m.PolicyRef != nil && m.PolicyRef.Template != "" {
				referenced = append(referenced, m.PolicyRef.Template)
			}
		}
	}
	l.Info("Parsed framework", "framework", framework.Name, "templateCount", len(referenced))
```

- [ ] **Step 5: Adjust the status write to use the referenced count for now**

In the status block, change `profile.Status.ActivePolicies = int32(len(policiesToEnforce))` to `profile.Status.ActivePolicies = int32(len(referenced))`. (Task 3 refines this to resolved count.)

- [ ] **Step 6: Run the test**

Run: `cd ~/Desktop/mantl && go test ./controllers/compliance/ -run TestFrameworkParsesRealSOC2 -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
cd ~/Desktop/mantl
git add controllers/compliance/profile_controller.go controllers/compliance/framework_test.go
git commit -s -m "fix(compliance): parse framework mappings schema instead of silently dropping them"
```

---

### Task 2: Pure template-to-policy resolver

Add a pure function that, given the set of referenced template names and an index of available Kyverno policy names, returns resolved and unresolved templates. Keep it pure so it tests without a cluster.

**Files:**
- Create: `controllers/compliance/resolve.go`
- Create: `controllers/compliance/resolve_test.go`

- [ ] **Step 1: Write the failing test**

Create `controllers/compliance/resolve_test.go`:

```go
package compliance

import (
	"reflect"
	"testing"
)

func TestResolveTemplates(t *testing.T) {
	index := map[string]string{
		"require-network-policy":       "platform/security/kyverno/policies/require-network-policy.yaml",
		"restrict-image-registries":    "platform/security/kyverno/policies/restrict-image-registries.yaml",
	}
	aliases := map[string]string{
		"restrict-registries": "restrict-image-registries",
	}
	resolved, unresolved := resolveTemplates(
		[]string{"require-network-policy", "restrict-registries", "require-rbac-least-privilege"},
		index, aliases,
	)
	wantResolved := []string{
		"platform/security/kyverno/policies/require-network-policy.yaml",
		"platform/security/kyverno/policies/restrict-image-registries.yaml",
	}
	if !reflect.DeepEqual(resolved, wantResolved) {
		t.Errorf("resolved = %v, want %v", resolved, wantResolved)
	}
	if !reflect.DeepEqual(unresolved, []string{"require-rbac-least-privilege"}) {
		t.Errorf("unresolved = %v, want [require-rbac-least-privilege]", unresolved)
	}
}
```

- [ ] **Step 2: Run it; confirm it fails (undefined: resolveTemplates)**

Run: `cd ~/Desktop/mantl && go test ./controllers/compliance/ -run TestResolveTemplates`
Expected: build failure, `undefined: resolveTemplates`.

- [ ] **Step 3: Implement `resolve.go`**

Create `controllers/compliance/resolve.go`:

```go
package compliance

import "sort"

// resolveTemplates maps framework template names to policy file paths using a
// name index and a small alias table (for templates whose canonical policy ships
// under a different name). It returns resolved paths (sorted, deduped) and the
// template names with no policy. Pure: no I/O, so it tests without a cluster.
func resolveTemplates(templates []string, index, aliases map[string]string) (resolved, unresolved []string) {
	seen := map[string]bool{}
	for _, t := range templates {
		name := t
		if alias, ok := aliases[t]; ok {
			name = alias
		}
		if path, ok := index[name]; ok {
			if !seen[path] {
				seen[path] = true
				resolved = append(resolved, path)
			}
			continue
		}
		unresolved = append(unresolved, t)
	}
	sort.Strings(resolved)
	return resolved, unresolved
}
```

- [ ] **Step 4: Run the test**

Run: `cd ~/Desktop/mantl && go test ./controllers/compliance/ -run TestResolveTemplates -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd ~/Desktop/mantl
git add controllers/compliance/resolve.go controllers/compliance/resolve_test.go
git commit -s -m "feat(compliance): add pure template-to-policy resolver"
```

---

### Task 3: Build the policy index and surface unresolved templates in status

Wire the resolver into the reconciler: build the index from the policy directories on disk, resolve the framework's templates, set `ActivePolicies` to the resolved count, and record unresolved templates so an operator can see coverage gaps. This does not yet apply policies to the cluster (Phase 2b); it reports truthfully.

**Files:**
- Modify: `controllers/compliance/profile_controller.go`
- Modify: `apis/compliance/v1alpha1/types.go` (add a status field)
- Create: `controllers/compliance/policyindex.go`
- Create: `controllers/compliance/policyindex_test.go`

- [ ] **Step 1: Add a status field for unresolved templates**

In `apis/compliance/v1alpha1/types.go`, find `ComplianceProfileStatus` and add:

```go
	// UnresolvedTemplates lists framework templates with no matching policy.
	// +optional
	UnresolvedTemplates []string `json:"unresolvedTemplates,omitempty"`
```

Regenerate DeepCopy after this change (Step 6).

- [ ] **Step 2: Write the policy-index test**

Create `controllers/compliance/policyindex_test.go`:

```go
package compliance

import "testing"

func TestBuildPolicyIndex(t *testing.T) {
	// The repo's platform policy dir must yield a known policy by base name.
	idx, err := buildPolicyIndex([]string{"../../platform/security/kyverno/policies"})
	if err != nil {
		t.Fatalf("buildPolicyIndex: %v", err)
	}
	if _, ok := idx["require-network-policy"]; !ok {
		t.Fatalf("index missing require-network-policy; got keys %v", keys(idx))
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
```

- [ ] **Step 3: Run it; confirm it fails (undefined: buildPolicyIndex)**

Run: `cd ~/Desktop/mantl && go test ./controllers/compliance/ -run TestBuildPolicyIndex`
Expected: build failure, `undefined: buildPolicyIndex`.

- [ ] **Step 4: Implement `policyindex.go`**

Create `controllers/compliance/policyindex.go`:

```go
package compliance

import (
	"os"
	"path/filepath"
	"strings"
)

// buildPolicyIndex maps a policy's base file name (without .yaml) to its path,
// scanning the given directories. Missing directories are skipped so the index
// works whether policies live under platform/ or compliance/ trees.
func buildPolicyIndex(dirs []string) (map[string]string, error) {
	index := map[string]string{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".yaml")
			if _, exists := index[name]; !exists {
				index[name] = filepath.Join(dir, e.Name())
			}
		}
	}
	return index, nil
}

// templateAliases maps framework template names to the policy base name that
// actually ships, for cases where the names differ.
var templateAliases = map[string]string{
	"restrict-registries":          "restrict-image-registries",
	"require-pod-security-context": "pss-restricted",
}
```

- [ ] **Step 5: Wire the resolver into Reconcile**

In `profile_controller.go` Reconcile, replace the `referenced` block from Task 1 Step 4 and the status write with:

```go
	var referenced []string
	for _, control := range framework.Controls {
		for _, m := range control.Mappings {
			if m.PolicyRef != nil && m.PolicyRef.Template != "" {
				referenced = append(referenced, m.PolicyRef.Template)
			}
		}
	}

	index, err := buildPolicyIndex([]string{
		"platform/security/kyverno/policies",
		"policies/kyverno",
		filepath.Join(r.FrameworkDir, profile.Spec.Framework, "policies"),
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	resolved, unresolved := resolveTemplates(referenced, index, templateAliases)
	l.Info("Resolved framework policies", "framework", framework.Name,
		"resolved", len(resolved), "unresolved", len(unresolved))

	profile.Status.State = "Active"
	profile.Status.ActivePolicies = int32(len(resolved))
	profile.Status.UnresolvedTemplates = unresolved
	if err := r.Status().Update(ctx, &profile); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
```

Remove the now-duplicated old status block.

- [ ] **Step 6: Regenerate DeepCopy and build**

Run:
```bash
cd ~/Desktop/mantl
go run sigs.k8s.io/controller-tools/cmd/controller-gen@latest object paths=./apis/compliance/v1alpha1/...
go build ./...
```
Expected: build clean, `zz_generated.deepcopy.go` updated for the new slice field.

- [ ] **Step 7: Run the package tests**

Run: `cd ~/Desktop/mantl && go test ./controllers/compliance/`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
cd ~/Desktop/mantl
git add controllers/compliance/profile_controller.go controllers/compliance/policyindex.go \
  controllers/compliance/policyindex_test.go apis/compliance/v1alpha1/types.go \
  apis/compliance/v1alpha1/zz_generated.deepcopy.go
git commit -s -m "feat(compliance): resolve framework templates to policies and report gaps in status"
```

---

### Task 4: Wire the FindingReconciler watch

`FindingReconciler.SetupWithManager` registers no event source, so its (correct) reconcile body never runs. Add a watch on the `PolicyReport` GVK. The type is not in the scheme, so use an unstructured watch, not `.For`.

**Files:**
- Modify: `controllers/compliance/finding_controller.go`

- [ ] **Step 1: Read the current SetupWithManager**

Run: `cd ~/Desktop/mantl && sed -n '155,170p' controllers/compliance/finding_controller.go`
Confirm it ends with `.Named("finding-aggregator").Complete(r)` and has no `.For`/`.Watches`.

- [ ] **Step 2: Add the unstructured watch**

In `finding_controller.go`, replace the `SetupWithManager` body with:

```go
func (r *FindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	report := &unstructured.Unstructured{}
	report.SetGroupVersionKind(policyReportGVK)
	return ctrl.NewControllerManagedBy(mgr).
		Named("finding-aggregator").
		Watches(report, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
```

- [ ] **Step 3: Ensure imports**

Confirm `finding_controller.go` imports `sigs.k8s.io/controller-runtime/pkg/handler` and `k8s.io/apimachinery/pkg/apis/meta/v1/unstructured`. Add any missing import.

- [ ] **Step 4: Build**

Run: `cd ~/Desktop/mantl && go build ./...`
Expected: clean. If `handler` is unused elsewhere it is now used here.

- [ ] **Step 5: Commit**

```bash
cd ~/Desktop/mantl
git add controllers/compliance/finding_controller.go
git commit -s -m "fix(compliance): give FindingReconciler a PolicyReport watch source"
```

---

### Definition of done for Phase 2a

- [ ] `go build ./...` clean; `make go-test` green.
- [ ] The SOC2 framework parses into populated `Mappings`/`PolicyRef`s (regression test).
- [ ] The profile controller reports a non-zero resolved policy count and lists unresolved templates in status.
- [ ] `FindingReconciler` has a real watch source.
- [ ] Push as a PR; the local review gate and DCO pass.

### Deferred to Phase 2b (do NOT attempt here)

- Author the 10 missing Kyverno policies and rename/relocate the 3 mismatched ones (audit listed them).
- Server-side apply resolved policies to the cluster from the profile controller.
- Implement evidence collectors by `type` (config-snapshot, log-query, metric-query) in the audit controller.
- Add the audit scheduler (`RequeueAfter` from `Frequency`/`schedule`, or a CronJob); remove the terminal-phase early return.
- Generate CRDs (`PROJECT` + `make manifests`); add the missing `ComplianceProfile` and `Finding` CRDs.
- Compute and surface a compliance score; replace the `status.go` stub with a real cluster query.
