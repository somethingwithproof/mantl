package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAddToSchemeRegistersComplianceKinds(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{
		"ComplianceProfile", "ComplianceProfileList", "Finding", "FindingList",
		"ComplianceAudit", "ComplianceAuditList", "AuditRun", "AuditRunList",
		"ControlEvaluation", "ControlEvaluationList", "ComplianceException", "ComplianceExceptionList",
	} {
		t.Run(kind, func(t *testing.T) {
			gvk := GroupVersion.WithKind(kind)
			object, err := scheme.New(gvk)
			if err != nil {
				t.Fatalf("instantiate %s: %v", gvk, err)
			}
			kinds, _, err := scheme.ObjectKinds(object)
			if err != nil || len(kinds) != 1 || kinds[0] != gvk {
				t.Fatalf("registered kinds = %v, err = %v; want %s", kinds, err, gvk)
			}
		})
	}
	if _, err := scheme.New(GroupVersion.WithKind("WatchEvent")); err != nil {
		t.Fatalf("register metadata: %v", err)
	}
	if err := scheme.Convert(&metav1.ListOptions{}, &metav1.ListOptions{}, nil); err != nil {
		t.Fatalf("convert metadata: %v", err)
	}
}
