package evidence

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"strings"
	"testing"
	"time"
)

func TestEvidenceRedactsWorkloadCredentials(t *testing.T) {
	deployment := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment", "metadata": map[string]interface{}{"name": "app", "namespace": "tenant", "annotations": map[string]interface{}{"token": "SENSITIVE_SENTINEL"}}, "spec": map[string]interface{}{"replicas": int64(2), "template": map[string]interface{}{"spec": map[string]interface{}{"containers": []interface{}{map[string]interface{}{"env": []interface{}{map[string]interface{}{"name": "TOKEN", "value": "SENSITIVE_SENTINEL"}}}}}}}}}
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), deployment)
	reader := KubernetesReader{Client: client}
	snap, err := reader.Capture(context.Background(), "deployments", "tenant", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(snap.Data, "SENSITIVE_SENTINEL") || !strings.Contains(snap.Data, "replicas") {
		t.Fatal("unsafe or empty evidence")
	}
	for _, resource := range []string{"secrets", "configmaps", "unknown"} {
		if _, err := reader.Capture(context.Background(), resource, "tenant", time.Now()); err == nil {
			t.Fatal("disallowed evidence read")
		}
	}
	if _, err := reader.Capture(context.Background(), "deployments", "", time.Now()); err == nil {
		t.Fatal("unscoped read accepted")
	}
	if _, err := reader.Capture(context.Background(), "clusterroles", "", time.Now()); err == nil {
		t.Fatal("cluster read accepted without opt-in")
	}
}
func TestEvidenceClusterScopeExplicit(t *testing.T) {
	typ, _ := Resource("clusterroles")
	client := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{typ.GVR: "ClusterRoleList"})
	reader := KubernetesReader{Client: client, AllowCluster: true}
	snap, err := reader.Capture(context.Background(), "clusterroles", "", metav1.Now().Time)
	if err != nil || snap == nil {
		t.Fatalf("capture: %v", err)
	}
	if _, err := reader.Capture(context.Background(), "clusterroles", "tenant", time.Now()); err == nil {
		t.Fatal("cluster evidence silently ignored scope")
	}
}
