package compiler

import (
	api "github.com/thomasvincent/mantl/apis/platform/v1alpha1"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTenantIsolationAndCommittedSource(t *testing.T) {
	cluster := newTestCluster("small")
	cluster.Spec.GitOps = api.GitOpsSpec{Repository: "https://github.com/example/gitops.git", Revision: "v1.2.3", TenantPath: "tenants/dev"}
	cluster.Spec.Tenants = []api.TenantSpec{{Name: "team", Admins: []string{"alice"}}}
	dir := t.TempDir()
	if err := RenderGitOps(cluster, dir); err != nil {
		t.Fatal(err)
	}
	app, _ := os.ReadFile(filepath.Join(dir, "gitops/platform-tenants.yaml"))
	if !strings.Contains(string(app), "tenants/dev") || !strings.Contains(string(app), "v1.2.3") || !strings.Contains(string(app), "example/gitops") {
		t.Fatal("generated source not bound to committed path/revision")
	}
	tenant, _ := os.ReadFile(filepath.Join(dir, "tenants/team.yaml"))
	for _, kind := range []string{"NetworkPolicy", "ResourceQuota", "LimitRange", "RoleBinding"} {
		if !strings.Contains(string(tenant), kind) {
			t.Fatal("missing isolation", kind)
		}
	}
	if err := RenderGitOps(cluster, dir); err != nil {
		t.Fatal(err)
	}
	k, _ := os.ReadFile(filepath.Join(dir, "tenants/kustomization.yaml"))
	if strings.Contains(string(k), "- kustomization.yaml") {
		t.Fatal("recursive Kustomization generated")
	}
}
func TestSpecRejectsUnsafeTenantAndProvider(t *testing.T) {
	cluster := newTestCluster("small")
	cluster.Spec.Tenants = []api.TenantSpec{{Name: "../escape"}}
	if err := Validate(cluster); err == nil {
		t.Fatal("unsafe tenant accepted")
	}
	cluster.Spec.Tenants = nil
	cluster.Spec.Kubernetes.Distribution = "gke"
	if err := Validate(cluster); err == nil {
		t.Fatal("mismatched provider accepted")
	}
}
