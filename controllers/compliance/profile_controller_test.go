package compliance

import (
	"context"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

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
