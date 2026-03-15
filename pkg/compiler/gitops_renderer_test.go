package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestRenderApp_SyncWavePriority(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		wantWave string
	}{
		{"cert-manager gets wave 1", "cert-manager", "1"},
		{"external-secrets gets wave 1", "external-secrets", "1"},
		{"kyverno gets wave 2", "kyverno", "2"},
		{"monitoring gets wave 2", "monitoring", "2"},
		{"unknown component defaults to wave 3", "ingress-nginx", "3"},
		{"empty name defaults to wave 3", "custom-app", "3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := renderApp(dir, tt.appName, "some/path"); err != nil {
				t.Fatalf("renderApp(%q) error: %v", tt.appName, err)
			}

			data, err := os.ReadFile(filepath.Join(dir, tt.appName+".yaml"))
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}

			var app map[string]interface{}
			if err := yaml.Unmarshal(data, &app); err != nil {
				t.Fatalf("failed to parse yaml: %v", err)
			}

			meta := app["metadata"].(map[string]interface{})
			annotations := meta["annotations"].(map[string]interface{})
			got := annotations["argocd.argoproj.io/sync-wave"].(string)
			if got != tt.wantWave {
				t.Errorf("sync-wave = %q, want %q", got, tt.wantWave)
			}
		})
	}
}

func TestRenderApp_SyncWaveDeterministic(t *testing.T) {
	// Run renderApp multiple times for the same name and verify the wave
	// is always identical, confirming deterministic iteration.
	dir := t.TempDir()
	const appName = "kyverno"

	for i := 0; i < 20; i++ {
		subDir := filepath.Join(dir, "run")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := renderApp(subDir, appName, "some/path"); err != nil {
			t.Fatalf("renderApp iteration %d error: %v", i, err)
		}

		data, err := os.ReadFile(filepath.Join(subDir, appName+".yaml"))
		if err != nil {
			t.Fatal(err)
		}

		var app map[string]interface{}
		if err := yaml.Unmarshal(data, &app); err != nil {
			t.Fatal(err)
		}

		meta := app["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		got := annotations["argocd.argoproj.io/sync-wave"].(string)
		if got != "2" {
			t.Fatalf("iteration %d: sync-wave = %q, want %q", i, got, "2")
		}

		os.RemoveAll(subDir)
	}
}

func TestRenderApp_HasPrefixNotContains(t *testing.T) {
	// "addon-monitoring" should NOT match "monitoring" via HasPrefix because
	// the prefix is "addon", not "monitoring". This confirms HasPrefix
	// semantics (no substring matching).
	dir := t.TempDir()
	if err := renderApp(dir, "addon-monitoring", "some/path"); err != nil {
		t.Fatalf("renderApp error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "addon-monitoring.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	var app map[string]interface{}
	if err := yaml.Unmarshal(data, &app); err != nil {
		t.Fatal(err)
	}

	meta := app["metadata"].(map[string]interface{})
	annotations := meta["annotations"].(map[string]interface{})
	got := annotations["argocd.argoproj.io/sync-wave"].(string)
	if got != "3" {
		t.Errorf("addon-monitoring sync-wave = %q, want %q (should not match 'monitoring' prefix)", got, "3")
	}
}
