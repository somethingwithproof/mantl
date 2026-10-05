package compliance

import (
	"context"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestFindingLifecycle(t *testing.T) {
	scheme := runtime.NewScheme()
	requireNoError(t, api.AddToScheme(scheme))
	report := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "wgpolicyk8s.io/v1alpha2", "kind": "PolicyReport", "metadata": map[string]interface{}{"name": "report", "namespace": "tenant"}, "results": []interface{}{map[string]interface{}{"policy": "require-labels", "rule": "labels", "result": "fail", "resources": []interface{}{map[string]interface{}{"kind": "Pod", "namespace": "tenant", "name": "app"}}}}}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(report).Build()
	r := FindingReconciler{Client: c, Scheme: scheme}
	req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "report", Namespace: "tenant"}}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	var findings api.FindingList
	requireNoError(t, c.List(context.Background(), &findings))
	if len(findings.Items) != 1 || findings.Items[0].Spec.Status != "fail" || findings.Items[0].Spec.ID == "" {
		t.Fatal("finding missing")
	}
	for _, value := range findings.Items[0].Labels {
		if len(validation.IsValidLabelValue(value)) > 0 {
			t.Fatal("finding label rejected by Kubernetes")
		}
	}
	results := report.Object["results"].([]interface{})
	results[0].(map[string]interface{})["result"] = "pass"
	requireNoError(t, c.Update(context.Background(), report))
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	requireNoError(t, c.List(context.Background(), &findings))
	if findings.Items[0].Spec.Status != "resolved" {
		t.Fatal("repair not reflected")
	}
	requireNoError(t, c.Delete(context.Background(), report))
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	requireNoError(t, c.List(context.Background(), &findings))
	if findings.Items[0].Spec.Status != "unknown" {
		t.Fatal("source deletion falsely compliant")
	}
}
func TestFindingNamesAvoidTruncationCollisions(t *testing.T) {
	prefix := ""
	for i := 0; i < 300; i++ {
		prefix += "a"
	}
	if buildFindingName(prefix, "rule", "ns", "one") == buildFindingName(prefix, "rule", "ns", "two") {
		t.Fatal("truncation collision")
	}
}
