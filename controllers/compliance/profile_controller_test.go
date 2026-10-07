// SPDX-License-Identifier: Apache-2.0

package compliance

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/controlbundle"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSTIGProfileResolvesCanonicalBundledPolicies(t *testing.T) {
	var content bytes.Buffer
	m, err := controlbundle.Build("../..", "1.0.2", &content)
	requireNoError(t, err)
	dir := filepath.Join(t.TempDir(), "content")
	_, err = controlbundle.Unpack(&content, dir, m.Digest())
	requireNoError(t, err)
	scheme := runtime.NewScheme()
	requireNoError(t, api.AddToScheme(scheme))
	profile := &api.ComplianceProfile{ObjectMeta: metav1.ObjectMeta{Name: "stig", Namespace: "tenant"}, Spec: api.ComplianceProfileSpec{Framework: "stig-kubernetes", Version: "2.3", BundleDigest: m.Digest()}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(profile).WithObjects(profile).Build()
	r := ComplianceProfileReconciler{Client: c, FrameworkDir: filepath.Join(dir, "frameworks")}
	req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "stig", Namespace: "tenant"}}
	_, err = r.Reconcile(context.Background(), req)
	requireNoError(t, err)
	requireNoError(t, c.Get(context.Background(), req.NamespacedName, profile))
	if profile.Status.ActivePolicies != 2 || len(profile.Status.UnresolvedTemplates) != 0 {
		t.Fatalf("STIG references failed to resolve: %+v", profile.Status)
	}
	index, err := frameworkPolicyIndex(r.FrameworkDir, "soc2")
	requireNoError(t, err)
	if _, ok := index["stig-disallow-secret-env"]; ok {
		t.Fatal("STIG-only policies leaked into another framework's index")
	}
}

func TestProfileInvalidConfigurationClearsPriorCoverage(t *testing.T) {
	scheme := runtime.NewScheme()
	requireNoError(t, api.AddToScheme(scheme))
	profile := &api.ComplianceProfile{ObjectMeta: metav1.ObjectMeta{Name: "profile", Namespace: "tenant"}, Spec: api.ComplianceProfileSpec{Framework: "soc2", Version: "bad"}, Status: api.ComplianceProfileStatus{State: "Active", ActivePolicies: 5}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(profile).WithObjects(profile).Build()
	r := ComplianceProfileReconciler{Client: c, FrameworkDir: "../../compliance/frameworks"}
	req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "profile", Namespace: "tenant"}}
	if _, err := r.Reconcile(context.Background(), req); err == nil {
		t.Fatal("invalid profile succeeded")
	}
	requireNoError(t, c.Get(context.Background(), req.NamespacedName, profile))
	if profile.Status.State != "Invalid" || profile.Status.ActivePolicies != 0 || len(profile.Status.UnresolvedTemplates) == 0 {
		t.Fatal("invalid profile retained prior successful status")
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
