// SPDX-License-Identifier: Apache-2.0

package compiler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
)

func TestCompileMatchesWrittenArtifacts(t *testing.T) {
	cluster := newTestCluster("full")
	cluster.Spec.Features = v1alpha1.FeatureSpec{Observability: true, Security: true, Compliance: true, Secrets: true, ProgressiveDelivery: true}
	cluster.Spec.GitOps = v1alpha1.GitOpsSpec{Repository: "https://example.com/platform.git", Revision: "v1.2.3", TenantPath: "tenants/production", OperatorPath: "operator/production"}
	cluster.Spec.Tenants = []v1alpha1.TenantSpec{{Name: "payments", Admins: []string{"alice@example.com"}}}
	original := cluster.DeepCopy()
	plan, err := Compile(cluster)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(original, cluster) {
		t.Fatal("compiler mutated its input")
	}
	second, err := Compile(cluster)
	if err != nil || !reflect.DeepEqual(plan, second) {
		t.Fatalf("compilation is not deterministic: %v", err)
	}
	if len(plan.Artifacts) != 10 || len(plan.Applications) != 7 {
		t.Fatalf("incomplete feature inventory: %+v", plan)
	}
	dir := t.TempDir()
	if err := RenderTerraform(cluster, dir); err != nil {
		t.Fatal(err)
	}
	if err := RenderGitOps(cluster, dir); err != nil {
		t.Fatal(err)
	}
	verifyPlanArtifacts(t, plan, dir)
	raw, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("alice@example.com")) || bytes.Contains(raw, []byte("Content")) {
		t.Fatal("JSON inventory includes artifact payloads")
	}
}

func TestCompileInvalidInputReturnsNoArtifacts(t *testing.T) {
	for _, invalid := range []*v1alpha1.MantlCluster{nil, {}, newTestCluster("invalid")} {
		plan, err := Compile(invalid)
		if err == nil || len(plan.Artifacts) != 0 {
			t.Fatalf("invalid input produced a plan: %+v, %v", plan, err)
		}
	}
}

func TestCompileEmptyTenantIndexAndDefaults(t *testing.T) {
	cluster := newTestCluster("small")
	cluster.Spec.GitOps.TenantPath = "tenants"
	plan, err := Compile(cluster)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Artifacts) != 4 || len(plan.Applications) != 2 {
		t.Fatalf("configured empty tenant topology lost: %+v", plan)
	}
	if !strings.Contains(strings.Join(plan.Notices, " "), "follows main") {
		t.Fatal("moving revision notice missing")
	}
	last := plan.Artifacts[len(plan.Artifacts)-1]
	if last.Path != "terraform.tfvars.json" {
		t.Fatal("unexpected artifact ordering")
	}
	for _, artifact := range plan.Artifacts {
		if artifact.Path == "tenants/kustomization.yaml" && !bytes.Contains(artifact.Content, []byte("resources: []")) {
			t.Fatal("empty tenant index is not explicit")
		}
	}
}

func verifyPlanArtifacts(t *testing.T, plan Plan, dir string) {
	t.Helper()
	for i, artifact := range plan.Artifacts {
		data, err := os.ReadFile(filepath.Join(dir, artifact.Path))
		if err != nil || !bytes.Equal(data, artifact.Content) {
			t.Fatalf("preview differs from written %s: %v", artifact.Path, err)
		}
		hash := sha256.Sum256(data)
		if artifact.SHA256 != hex.EncodeToString(hash[:]) || artifact.Size != len(data) {
			t.Fatalf("incorrect identity for %s", artifact.Path)
		}
		if i > 0 && plan.Artifacts[i-1].Path >= artifact.Path {
			t.Fatal("artifact paths are not unique and sorted")
		}
	}
}

func TestIgnoredConfigurationIsVisibleWithoutChangingArtifacts(t *testing.T) {
	cluster := newTestCluster("small")
	baseline, err := Compile(cluster)
	if err != nil {
		t.Fatal(err)
	}
	cluster.Spec.Networking.Exposure = "public"
	cluster.Spec.Networking.VpcID = "vpc-example"
	cluster.Spec.Profile.Compliance = "hipaa"
	changed, err := Compile(cluster)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(baseline.Artifacts, changed.Artifacts) {
		t.Fatal("unimplemented inputs unexpectedly altered generated configuration")
	}
	if baseline.SpecSHA256 == changed.SpecSHA256 {
		t.Fatal("review identity omitted ignored configuration")
	}
	for _, field := range []string{"profile.size", "networking.domain", "profile.compliance"} {
		if !strings.Contains(strings.Join(changed.Notices, "\n"), field) {
			t.Fatalf("accepted but limited field has no notice: %s", field)
		}
	}
}

func TestAccountAndExperimentalProviderNoticesAreScoped(t *testing.T) {
	cluster := newTestCluster("small")
	cluster.Spec.Provider.AccountID = "example-account"
	notices := strings.Join(compilationNotices(cluster), "\n")
	if !strings.Contains(notices, "AWS provider.accountId") {
		t.Fatal("ignored AWS account selection was silent")
	}
	cluster.Spec.Provider.Kind = "gcp"
	if strings.Contains(strings.Join(compilationNotices(cluster), "\n"), "provider.accountId") {
		t.Fatal("GCP project input incorrectly declared ignored")
	}
	cluster.Spec.Provider.Kind = "do"
	if !strings.Contains(strings.Join(compilationNotices(cluster), "\n"), "experimental") {
		t.Fatal("experimental provider maturity omitted")
	}
}
