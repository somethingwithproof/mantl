# Phase 2b-1: Audit Scheduler Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:executing-plans. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Make recurring `ComplianceAudit`s actually recur. Today a terminal-phase early-return means an audit runs once and never again, so `Frequency: daily|weekly` is ignored.

**Architecture:** Add pure, cluster-free scheduling helpers (`requeueInterval`, `isTerminal`, `planAudit`) and wire them into `ComplianceAuditReconciler.Reconcile`: a terminal audit that is non-recurring is done; a recurring one requeues and reruns when its interval elapses. All decision logic is pure and unit-tested; the reconciler just acts on the decision.

**Tech Stack:** Go 1.26, controller-runtime (`ctrl.Result{RequeueAfter}`).

**Scope:** Scheduler only. Does NOT implement evidence collectors, author policies, generate CRDs, or compute a score (later Phase 2b slices). Run `make go-test`, never `go test ./...`.

---

### Task 1: Pure scheduling helpers

**Files:**
- Create: `controllers/compliance/schedule.go`
- Create: `controllers/compliance/schedule_test.go`

- [ ] **Step 1: Write the failing test**

Create `controllers/compliance/schedule_test.go`:

```go
package compliance

import (
	"testing"
	"time"
)

func TestRequeueInterval(t *testing.T) {
	cases := []struct {
		freq      string
		want      time.Duration
		recurring bool
	}{
		{"daily", 24 * time.Hour, true},
		{"weekly", 7 * 24 * time.Hour, true},
		{"Daily", 24 * time.Hour, true},
		{"manual", 0, false},
		{"", 0, false},
		{"hourly", 0, false},
	}
	for _, c := range cases {
		got, rec := requeueInterval(c.freq)
		if got != c.want || rec != c.recurring {
			t.Errorf("requeueInterval(%q) = (%v,%v), want (%v,%v)", c.freq, got, rec, c.want, c.recurring)
		}
	}
}

func TestPlanAudit(t *testing.T) {
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
	hourAgo := now.Add(-time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	cases := []struct {
		name      string
		phase     string
		end       *time.Time
		freq      string
		wantAct   auditAction
		wantAfter time.Duration
	}{
		{"non-terminal runs", "Running", nil, "daily", actionRun, 0},
		{"pending runs", "", nil, "daily", actionRun, 0},
		{"terminal non-recurring done", "Completed", &hourAgo, "manual", actionDone, 0},
		{"terminal recurring, interval not elapsed, waits", "Completed", &hourAgo, "daily", actionWaitRequeue, 23 * time.Hour},
		{"terminal recurring, interval elapsed, reruns", "Completed", &twoDaysAgo, "daily", actionRun, 0},
		{"terminal recurring, no end time, reruns", "Failed", nil, "weekly", actionRun, 0},
	}
	for _, c := range cases {
		act, after := planAudit(c.phase, c.end, c.freq, now)
		if act != c.wantAct || after != c.wantAfter {
			t.Errorf("%s: planAudit = (%v,%v), want (%v,%v)", c.name, act, after, c.wantAct, c.wantAfter)
		}
	}
}
```

- [ ] **Step 2: Run it; confirm it fails (undefined symbols)**

Run: `cd ~/Desktop/mantl && go test ./controllers/compliance/ -run 'TestRequeueInterval|TestPlanAudit'`
Expected: build failure, `undefined: requeueInterval` / `planAudit` / `auditAction`.

- [ ] **Step 3: Implement `schedule.go`**

Create `controllers/compliance/schedule.go`:

```go
package compliance

import (
	"strings"
	"time"
)

// auditAction is the scheduler's decision for a ComplianceAudit reconcile.
type auditAction int

const (
	// actionRun means start (or restart) the audit now.
	actionRun auditAction = iota
	// actionWaitRequeue means the audit is terminal and recurring but its
	// interval has not elapsed; requeue after the returned duration.
	actionWaitRequeue
	// actionDone means the audit is terminal and not recurring; nothing to do.
	actionDone
)

// requeueInterval maps an audit Frequency to a recurrence interval. Unknown or
// manual frequencies are not recurring.
func requeueInterval(frequency string) (time.Duration, bool) {
	switch strings.ToLower(strings.TrimSpace(frequency)) {
	case "daily":
		return 24 * time.Hour, true
	case "weekly":
		return 7 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

// isTerminal reports whether a phase represents a finished audit run.
func isTerminal(phase string) bool {
	switch phase {
	case "Completed", "PartiallyCompleted", "Failed", "NoEvidence":
		return true
	default:
		return false
	}
}

// planAudit decides what the reconciler should do for an audit given its phase,
// last end time, frequency, and the current time. Pure: no I/O.
func planAudit(phase string, endTime *time.Time, frequency string, now time.Time) (auditAction, time.Duration) {
	if !isTerminal(phase) {
		return actionRun, 0
	}
	interval, recurring := requeueInterval(frequency)
	if !recurring {
		return actionDone, 0
	}
	if endTime == nil {
		return actionRun, 0
	}
	if elapsed := now.Sub(*endTime); elapsed < interval {
		return actionWaitRequeue, interval - elapsed
	}
	return actionRun, 0
}
```

- [ ] **Step 4: Run the tests**

Run: `cd ~/Desktop/mantl && go test ./controllers/compliance/ -run 'TestRequeueInterval|TestPlanAudit' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd ~/Desktop/mantl
git add controllers/compliance/schedule.go controllers/compliance/schedule_test.go
git commit -s -m "feat(compliance): add pure audit scheduling helpers"
```

---

### Task 2: Wire the scheduler into the audit reconciler

**Files:**
- Modify: `controllers/compliance/audit_controller.go`

- [ ] **Step 1: Replace the terminal early-return block**

In `controllers/compliance/audit_controller.go`, find the block (just after fetching the audit):

```go
	if audit.Status.Phase == "Completed" || audit.Status.Phase == "PartiallyCompleted" || audit.Status.Phase == "Failed" {
		return ctrl.Result{}, nil
	}
```

Replace it with:

```go
	var endTime *time.Time
	if audit.Status.EndTime != nil {
		endTime = &audit.Status.EndTime.Time
	}
	switch action, after := planAudit(audit.Status.Phase, endTime, audit.Spec.Frequency, time.Now()); action {
	case actionDone:
		return ctrl.Result{}, nil
	case actionWaitRequeue:
		return ctrl.Result{RequeueAfter: after}, nil
	case actionRun:
		// Clear the prior terminal state so a recurring run proceeds cleanly.
		audit.Status.Phase = "Pending"
		audit.Status.EndTime = nil
	}
```

- [ ] **Step 2: Requeue after a recurring run finalizes**

At the end of Reconcile, find the final `return ctrl.Result{}, nil` (after the finalize status update and the "Compliance Audit finalized" log). Replace it with:

```go
	if interval, recurring := requeueInterval(audit.Spec.Frequency); recurring {
		return ctrl.Result{RequeueAfter: interval}, nil
	}
	return ctrl.Result{}, nil
```

- [ ] **Step 3: Build and test**

Run: `cd ~/Desktop/mantl && go build ./... && go test ./controllers/compliance/`
Expected: build clean; package tests PASS. `time` is already imported in `audit_controller.go`.

- [ ] **Step 4: Commit**

```bash
cd ~/Desktop/mantl
git add controllers/compliance/audit_controller.go
git commit -s -m "feat(compliance): requeue and rerun recurring audits on their frequency"
```

---

### Definition of done

- [ ] `go build ./...` clean; `make go-test` green.
- [ ] `requeueInterval`/`planAudit` covered by table tests including the not-elapsed, elapsed, and non-recurring paths.
- [ ] A `daily`/`weekly` audit requeues after finalizing; a `manual` audit does not.
- [ ] Push as a PR; the local review gate and DCO pass.
