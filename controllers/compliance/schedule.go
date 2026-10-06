// SPDX-License-Identifier: Apache-2.0

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
