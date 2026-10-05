package bootstrap

import "testing"

func TestConvergenceRequiresCompletePairs(t *testing.T) {
	for _, raw := range []string{"", "Synced,", "Synced,Healthy\nSynced,", "Synced,Healthy\nOutOfSync,Healthy"} {
		if ApplicationsHealthy(raw) {
			t.Fatal("incomplete convergence accepted", raw)
		}
	}
	if !ApplicationsHealthy("Synced,Healthy\nSynced,Healthy") {
		t.Fatal("healthy applications rejected")
	}
}
