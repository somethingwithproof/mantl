package cmd

import (
	"context"
	"errors"
	"testing"
)

func TestEvaluateHealth(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"all synced and healthy", "Synced,Healthy\nSynced,Healthy", "HEALTHY"},
		{"single healthy app", "Synced,Healthy", "HEALTHY"},
		{"empty health field fails closed", "Synced,", "UNKNOWN (incomplete application status)"},
		{"count mismatch fails closed", "Synced,Healthy\nSynced,", "UNKNOWN (incomplete application status)"},
		{"no applications", "", "UNKNOWN (no ArgoCD applications found)"},
		{"whitespace only", "  \n  ", "UNKNOWN (no ArgoCD applications found)"},
		{"out of sync", "OutOfSync,Healthy", "DEGRADED (applications not synced)"},
		{"progressing health", "Synced,Progressing", "DEGRADED (applications not healthy)"},
		{"one unhealthy among healthy", "Synced,Healthy\nSynced,Degraded", "DEGRADED (applications not healthy)"},
		{"sync checked before health", "OutOfSync,Degraded", "DEGRADED (applications not synced)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := evaluateHealth(tc.raw); got != tc.want {
				t.Errorf("evaluateHealth(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestPlatformHealth(t *testing.T) {
	orig := kubectlHealthQuery
	t.Cleanup(func() { kubectlHealthQuery = orig })

	t.Run("healthy stdout only", func(t *testing.T) {
		kubectlHealthQuery = func(context.Context) ([]byte, error) {
			return []byte("Synced,Healthy\nSynced,Healthy"), nil
		}
		if got := platformHealth(); got != "HEALTHY" {
			t.Errorf("platformHealth() = %q, want HEALTHY", got)
		}
	})

	t.Run("query error maps to unknown", func(t *testing.T) {
		kubectlHealthQuery = func(context.Context) ([]byte, error) {
			return nil, errors.New("connection refused")
		}
		if got := platformHealth(); got == "HEALTHY" {
			t.Errorf("platformHealth() = %q, must not be HEALTHY on error", got)
		}
	})
}
