package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestDesiredTopologyMatchesRenderedApplications(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		cluster := &v1alpha1.MantlCluster{ObjectMeta: metav1.ObjectMeta{Name: "fixture"}}
		if enabled {
			cluster.Spec.Features = v1alpha1.FeatureSpec{Observability: true, Security: true, Secrets: true, Compliance: true, ProgressiveDelivery: true}
			cluster.Spec.GitOps.Repository = "https://example.com/platform.git"
			cluster.Spec.GitOps.Revision = "v2.0.0"
			cluster.Spec.GitOps.OperatorPath = "operator/production"
			// Configured empty tenants must still have an application for removals.
			cluster.Spec.GitOps.TenantPath = "tenants/production"
		}
		directory := t.TempDir()
		if err := RenderGitOps(cluster, directory); err != nil {
			t.Fatal(err)
		}
		expected := GitOpsApplications(cluster)
		want := 1
		if enabled {
			want = 7
		}
		if len(expected) != want {
			t.Fatalf("got %d apps, want %d", len(expected), want)
		}
		entries, err := os.ReadDir(filepath.Join(directory, "gitops"))
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != len(expected) {
			t.Fatal("topology differs from rendered inventory")
		}
		for _, desired := range expected {
			verifyRenderedTopologyApplication(t, directory, desired)
		}
	}
	if GitOpsApplications(nil) != nil {
		t.Fatal("nil cluster has a desired topology")
	}
}

// Read the actual manifest bytes independently from the topology descriptors.
func verifyRenderedTopologyApplication(t *testing.T, directory string, desired GitOpsApplication) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(directory, "gitops", desired.Name+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var app topologyApplication
	if err := yaml.Unmarshal(raw, &app); err != nil {
		t.Fatal(err)
	}
	if app.Metadata.Name != desired.Name || app.Spec.Source.Repository != desired.Repository || app.Spec.Source.Revision != desired.Revision || app.Spec.Source.Path != desired.Path {
		t.Fatalf("rendered identity differs for %s: %+v", desired.Name, app)
	}
}

type topologyMetadata struct {
	Name string `json:"name"`
}
type topologySource struct {
	Repository string `json:"repoURL"`
	Revision   string `json:"targetRevision"`
	Path       string `json:"path"`
}
type topologySpec struct {
	Source topologySource `json:"source"`
}
type topologyApplication struct {
	Metadata topologyMetadata `json:"metadata"`
	Spec     topologySpec     `json:"spec"`
}
