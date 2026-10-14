// SPDX-License-Identifier: Apache-2.0

package compiler

import (
	api "github.com/thomasvincent/mantl/apis/platform/v1alpha1"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
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

func TestTenantReplanExcludesRemovedAndUnmanagedFiles(t *testing.T) {
	cluster := newTestCluster("small")
	cluster.Spec.GitOps.TenantPath = "tenants/dev"
	cluster.Spec.Tenants = []api.TenantSpec{{Name: "former", Admins: []string{"alice"}}}
	dir := t.TempDir()
	if err := RenderGitOps(cluster, dir); err != nil {
		t.Fatal(err)
	}
	unmanaged := filepath.Join(dir, "tenants/unmanaged.yaml")
	if err := os.WriteFile(unmanaged, []byte("owned by someone else"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tenants := range [][]api.TenantSpec{{{Name: "current"}}, nil} {
		cluster.Spec.Tenants = tenants
		if err := RenderGitOps(cluster, dir); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(dir, "tenants/kustomization.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		var index struct {
			Resources []string `json:"resources"`
		}
		if err := yaml.Unmarshal(data, &index); err != nil {
			t.Fatal(err)
		}
		if len(index.Resources) != len(tenants) || len(tenants) > 0 && index.Resources[0] != "current.yaml" {
			t.Fatalf("stale or unmanaged tenant included: %v", index.Resources)
		}
		if _, err := os.Stat(filepath.Join(dir, "gitops/platform-tenants.yaml")); err != nil {
			t.Fatal("tenant application must remain to reconcile removals", err)
		}
	}
	if data, err := os.ReadFile(unmanaged); err != nil || string(data) != "owned by someone else" {
		t.Fatal("unmanaged file changed", err)
	}
}

func TestSpecRejectsTenantFilenameCollisions(t *testing.T) {
	for _, tenants := range [][]api.TenantSpec{
		{{Name: "kustomization"}},
		{{Name: "team", Namespace: "first"}, {Name: "team", Namespace: "second"}},
	} {
		cluster := newTestCluster("small")
		cluster.Spec.GitOps.TenantPath = "tenants/dev"
		cluster.Spec.Tenants = tenants
		if err := Validate(cluster); err == nil {
			t.Fatalf("tenant filename collision accepted: %v", tenants)
		}
	}
}
func TestSpecRejectsReservedTenantNamespaces(t *testing.T) {
	for _, ns := range []string{"kube-system", "kube-public", "kube-node-lease", "default", "argocd", "mantl-compliance-system"} {
		for _, tenant := range []api.TenantSpec{{Name: "payments", Namespace: ns}, {Name: ns}} {
			cluster := newTestCluster("small")
			cluster.Spec.GitOps.TenantPath = "tenants/dev"
			cluster.Spec.Tenants = []api.TenantSpec{tenant}
			err := Validate(cluster)
			if err == nil || !strings.Contains(err.Error(), "reserved") {
				t.Fatalf("reserved namespace accepted: %+v: %v", tenant, err)
			}
		}
	}
}

func TestSpecRejectsConfiguredControlNamespace(t *testing.T) {
	cluster := newTestCluster("small")
	cluster.Spec.GitOps.TenantPath = "tenants/dev"
	cluster.Spec.ControlNamespace = "Not_A_Label"
	if err := Validate(cluster); err == nil || !strings.Contains(err.Error(), "controlNamespace") {
		t.Fatalf("invalid control namespace accepted: %v", err)
	}
	cluster.Spec.ControlNamespace = "compliance-system"
	cluster.Spec.Tenants = []api.TenantSpec{{Name: "payments", Namespace: "compliance-system"}}
	if err := Validate(cluster); err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("configured control namespace accepted as a tenant: %v", err)
	}
	cluster.Spec.Tenants = []api.TenantSpec{{Name: "payments"}}
	if err := Validate(cluster); err != nil {
		t.Fatal(err)
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
