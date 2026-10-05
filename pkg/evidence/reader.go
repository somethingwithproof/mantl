package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"time"
)

type ResourceType struct {
	GVR     schema.GroupVersionResource
	Kind    string
	Cluster bool
}

var resources = map[string]ResourceType{
	"networkpolicies":     {schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"}, "NetworkPolicy", false},
	"roles":               {schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}, "Role", false},
	"rolebindings":        {schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}, "RoleBinding", false},
	"clusterroles":        {schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, "ClusterRole", true},
	"clusterrolebindings": {schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "ClusterRoleBinding", true},
	"deployments":         {schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, "Deployment", false},
	"pods":                {schema.GroupVersionResource{Version: "v1", Resource: "pods"}, "Pod", false},
	"serviceaccounts":     {schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, "ServiceAccount", false},
	"services":            {schema.GroupVersionResource{Version: "v1", Resource: "services"}, "Service", false},
}

func Resource(name string) (ResourceType, error) {
	r, ok := resources[name]
	if !ok {
		return ResourceType{}, fmt.Errorf("resource %q is not permitted for evidence collection", name)
	}
	return r, nil
}

type Reader interface {
	Capture(context.Context, string, string, time.Time) (*Snapshot, error)
}
type KubernetesReader struct {
	Client       dynamic.Interface
	AllowCluster bool
}

func (r *KubernetesReader) Capture(ctx context.Context, resource, namespace string, at time.Time) (*Snapshot, error) {
	typ, err := Resource(resource)
	if err != nil {
		return nil, err
	}
	if typ.Cluster {
		if !r.AllowCluster || namespace != "" {
			return nil, fmt.Errorf("cluster evidence requires explicit opt-in and empty namespace")
		}
	} else if !validNamespaceRe.MatchString(namespace) || len(namespace) > 63 {
		return nil, fmt.Errorf("explicit valid evidence namespace required")
	}
	var list *unstructured.UnstructuredList
	if typ.Cluster {
		list, err = r.Client.Resource(typ.GVR).List(ctx, metav1.ListOptions{})
	} else {
		list, err = r.Client.Resource(typ.GVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("read evidence %s: %w", resource, err)
	}
	// Allowlist fields rather than attempt to enumerate all possible credential names.
	items := make([]map[string]interface{}, 0, len(list.Items))
	for _, item := range list.Items {
		safe := map[string]interface{}{"apiVersion": item.GetAPIVersion(), "kind": item.GetKind(), "metadata": map[string]interface{}{"name": item.GetName(), "namespace": item.GetNamespace(), "uid": string(item.GetUID()), "resourceVersion": item.GetResourceVersion()}}
		switch resource {
		case "networkpolicies":
			safe["spec"] = item.Object["spec"]
		case "roles", "clusterroles":
			safe["rules"] = item.Object["rules"]
			safe["aggregationRule"] = item.Object["aggregationRule"]
		case "rolebindings", "clusterrolebindings":
			safe["subjects"] = item.Object["subjects"]
			safe["roleRef"] = item.Object["roleRef"]
		case "deployments":
			replicas, _, _ := unstructured.NestedFieldCopy(item.Object, "spec", "replicas")
			safe["replicas"] = replicas
		case "pods":
			safe["phase"], _, _ = unstructured.NestedString(item.Object, "status", "phase")
		case "serviceaccounts":
			safe["automountServiceAccountToken"] = item.Object["automountServiceAccountToken"]
		case "services":
			safe["type"], _, _ = unstructured.NestedString(item.Object, "spec", "type")
		}
		items = append(items, safe)
	}
	data, err := json.Marshal(map[string]interface{}{"resource": resource, "namespace": namespace, "items": items})
	if err != nil {
		return nil, fmt.Errorf("encode evidence: %w", err)
	}
	hash := sha256.Sum256(data)
	return &Snapshot{Resource: namespace + "/" + resource, CapturedAt: at.UTC(), ContentHash: hex.EncodeToString(hash[:]), Data: string(data)}, nil
}
