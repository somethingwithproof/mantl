// SPDX-License-Identifier: Apache-2.0

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
