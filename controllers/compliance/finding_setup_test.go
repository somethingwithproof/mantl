package compliance

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestPolicyReportAvailable(t *testing.T) {
	t.Run("absent CRD reports unavailable without error", func(t *testing.T) {
		ok, err := policyReportAvailable(meta.NewDefaultRESTMapper(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected unavailable with an empty RESTMapper")
		}
	})

	t.Run("registered CRD reports available", func(t *testing.T) {
		m := meta.NewDefaultRESTMapper([]schema.GroupVersion{policyReportGVK.GroupVersion()})
		m.Add(policyReportGVK, meta.RESTScopeNamespace)
		ok, err := policyReportAvailable(m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected available after registering the mapping")
		}
	})

	t.Run("discovery failure propagates instead of reporting absent", func(t *testing.T) {
		ok, err := policyReportAvailable(failingMapper{})
		if err == nil {
			t.Fatal("expected the discovery error to propagate")
		}
		if ok {
			t.Fatal("expected unavailable on a discovery error")
		}
	})
}

// failingMapper is a RESTMapper whose discovery always fails with a non-no-match
// error, modeling a transient API/discovery outage at operator boot.
type failingMapper struct{}

var errDiscovery = errors.New("discovery unavailable")

func (failingMapper) KindFor(schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, errDiscovery
}
func (failingMapper) KindsFor(schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	return nil, errDiscovery
}
func (failingMapper) ResourceFor(schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	return schema.GroupVersionResource{}, errDiscovery
}
func (failingMapper) ResourcesFor(schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	return nil, errDiscovery
}
func (failingMapper) RESTMapping(schema.GroupKind, ...string) (*meta.RESTMapping, error) {
	return nil, errDiscovery
}
func (failingMapper) RESTMappings(schema.GroupKind, ...string) ([]*meta.RESTMapping, error) {
	return nil, errDiscovery
}
func (failingMapper) ResourceSingularizer(string) (string, error) {
	return "", errDiscovery
}
