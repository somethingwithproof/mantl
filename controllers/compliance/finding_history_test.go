// SPDX-License-Identifier: Apache-2.0

package compliance

import (
	"context"
	"encoding/json"
	"testing"

	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFindingHistorySurvivesStorageFailureAndRecordsRecovery(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	requireNoError(t, api.AddToScheme(scheme))
	f := &api.Finding{ObjectMeta: metav1.ObjectMeta{Name: "finding", Namespace: "team", UID: "finding-uid"}, Spec: api.FindingSpec{Status: "fail", ControlID: "C1", Resource: "Pod/team/app"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&api.Finding{}).WithObjects(f).Build()
	store := &runStore{objects: map[string][]byte{}, fail: true}
	r := &FindingReconciler{Client: c, Store: store, ClusterID: "cluster"}
	if err := r.persistFinding(ctx, f, false); err == nil {
		t.Fatal("storage failure hidden")
	}
	requireNoError(t, c.Get(ctx, client.ObjectKeyFromObject(f), f))
	if len(f.Status.Pending) != 1 || f.Status.HistoryHead != nil {
		t.Fatal("transition was not durably queued")
	}
	id := f.Status.Pending[0].ID
	store.fail = false
	requireNoError(t, r.persistFinding(ctx, f, false))
	if len(f.Status.Pending) != 0 || f.Status.HistoryHead == nil || len(store.objects) != 1 {
		t.Fatal("retry did not persist exactly one transition")
	}
	var first struct {
		Event api.FindingTransition `json:"event"`
	}
	requireNoError(t, json.Unmarshal(store.objects[f.Status.HistoryHead.URI], &first))
	if first.Event.ID != id {
		t.Fatal("retry changed the event identity")
	}
	head := *f.Status.HistoryHead
	f.Spec.Status = "resolved"
	requireNoError(t, r.persistFinding(ctx, f, false))
	var second struct {
		Event    api.FindingTransition `json:"event"`
		Previous *api.HistoryReference `json:"previous"`
	}
	requireNoError(t, json.Unmarshal(store.objects[f.Status.HistoryHead.URI], &second))
	if second.Event.State != "resolved" || second.Previous == nil || *second.Previous != head {
		t.Fatal("recovery lost the previous immutable record")
	}
}
