package compliance

import (
	"context"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

func TestEvaluationReconcileSelectionAndInvalidContentFailClosed(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	requireNoError(t, os.MkdirAll(filepath.Join(dir, "fixture", "policies"), 0700))
	catalog := filepath.Join(dir, "fixture", "framework.yaml")
	requireNoError(t, os.WriteFile(catalog, []byte("spec:\n  version: '1'\n  controls:\n  - id: C1\n    mappings:\n    - policyRef:\n        template: policy\n  - id: C2\n"), 0600))
	requireNoError(t, os.WriteFile(filepath.Join(dir, "fixture", "policies", "policy.yaml"), []byte("metadata:\n  name: policy\nspec:\n  rules:\n  - name: rule\n"), 0600))
	scheme := runtime.NewScheme()
	requireNoError(t, api.AddToScheme(scheme))
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	profile := &api.ComplianceProfile{ObjectMeta: metav1.ObjectMeta{Name: "profile", Namespace: "team", UID: "profile-uid"}, Spec: api.ComplianceProfileSpec{Framework: "fixture", Version: "1"}}
	report := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "wgpolicyk8s.io/v1alpha2", "kind": "PolicyReport", "metadata": map[string]interface{}{"name": "report", "namespace": "team"}, "results": []interface{}{map[string]interface{}{"policy": "policy", "rule": "rule", "result": "pass", "timestamp": map[string]interface{}{"seconds": now.Unix()}, "resources": []interface{}{map[string]interface{}{"namespace": "team", "name": "object", "kind": "Pod"}}}}}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&api.ControlEvaluation{}).WithObjects(profile, report).Build()
	r := &EvaluationReconciler{Client: c, FrameworkDir: dir, Now: func() time.Time { return now }}
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(profile)}
	result, err := r.Reconcile(ctx, req)
	requireNoError(t, err)
	if result.RequeueAfter != 30*time.Second {
		t.Fatalf("missing refresh: %+v", result)
	}
	assertResult := func(control, want string) {
		t.Helper()
		var list api.ControlEvaluationList
		requireNoError(t, c.List(ctx, &list))
		for _, e := range list.Items {
			if e.Spec.Control == control {
				if e.Status.Result != want || e.Spec.BundleDigest == "" || len(e.OwnerReferences) != 1 {
					t.Fatalf("wrong evaluation: %+v", e)
				}
				return
			}
		}
		t.Fatalf("missing control %s", control)
	}
	assertResult("C1", "pass")
	assertResult("C2", "unknown")
	requireNoError(t, c.Get(ctx, req.NamespacedName, profile))
	profile.Spec.IncludeControls = []string{"C1"}
	requireNoError(t, c.Update(ctx, profile))
	_, err = r.Reconcile(ctx, req)
	requireNoError(t, err)
	assertResult("C2", "not-applicable")
	requireNoError(t, os.Remove(catalog))
	if _, err = r.Reconcile(ctx, req); err == nil {
		t.Fatal("missing catalog passed")
	}
	assertResult("C1", "error")
	assertResult("C2", "error")
	_, err = r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: "absent", Namespace: "team"}})
	requireNoError(t, err)
}

func TestEvaluationMappingReadFailureDoesNotPublishStatus(t *testing.T) {
	r := &EvaluationReconciler{}
	data := &controlEvaluationInput{profile: api.ComplianceProfile{}, framework: &framework.Framework{}, index: map[string]string{"missing": filepath.Join(t.TempDir(), "absent")}}
	control := framework.Control{ID: "C", Mappings: []framework.Mapping{{PolicyRef: &framework.PolicyRef{Template: "missing"}}}}
	if err := r.reconcileSelectedControl(context.Background(), control, data); err == nil {
		t.Fatal("missing mapped file inferred a result")
	}
	data.profile.Spec.IncludeControls = []string{"other"}
	requireNoError(t, r.reconcileSelectedControl(context.Background(), control, data))
}
