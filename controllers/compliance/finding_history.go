package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// persistFinding queues metadata before publishing a state transition. Failed
// storage leaves a durable queue; retries keep the same event identity.
func (r *FindingReconciler) persistFinding(ctx context.Context, f *api.Finding, creating bool) error {
	if creating {
		if err := r.Create(ctx, f); err != nil {
			return err
		}
	}
	if r.Store == nil {
		if creating {
			return nil
		}
		return r.Update(ctx, f)
	} // Legacy construction compatibility.
	if r.ClusterID == "" {
		return fmt.Errorf("cluster identity required for finding history")
	}
	desired := f.Spec
	if f.Status.LastQueuedState != desired.Status {
		if len(f.Status.Pending) >= 64 {
			f.Status.HistoryGap = true
			if err := r.Status().Update(ctx, f); err != nil {
				return err
			}
			f.Spec = desired
			f.Spec.Status = "unknown"
			f.Spec.Message = "Finding history backlog prevents a current lifecycle record"
			if err := r.Update(ctx, f); err != nil {
				return err
			}
			return fmt.Errorf("finding history backlog full")
		}
		at := time.Now().UTC()
		transition := api.FindingTransition{ID: evidence.Hash([]byte(string(f.UID) + "/" + desired.Status + "/" + at.Format(time.RFC3339Nano))), State: desired.Status, At: metav1.NewTime(at), Control: desired.ControlID, Framework: desired.Framework, Resource: desired.Resource}
		f.Status.Pending = append(f.Status.Pending, transition)
		f.Status.LastQueuedState = desired.Status
		if err := r.Status().Update(ctx, f); err != nil {
			return err
		}
		f.Spec = desired
	}
	if !creating {
		if err := r.Update(ctx, f); err != nil {
			return err
		}
	}
	for len(f.Status.Pending) > 0 {
		transition := f.Status.Pending[0]
		payload, err := json.Marshal(struct {
			Schema   int                   `json:"schemaVersion"`
			Finding  string                `json:"finding"`
			UID      string                `json:"uid"`
			Event    api.FindingTransition `json:"event"`
			Previous *api.HistoryReference `json:"previous,omitempty"`
		}{1, f.Namespace + "/" + f.Name, string(f.UID), transition, f.Status.HistoryHead})
		if err != nil {
			return err
		}
		upload, cancel := context.WithTimeout(ctx, 30*time.Second)
		ref, err := r.Store.Put(upload, "audits/findings/"+r.ClusterID+"/"+f.Namespace+"/"+string(f.UID)+"/"+transition.ID, payload, time.Now().UTC())
		cancel()
		if err != nil {
			return fmt.Errorf("persist finding history: %w", err)
		}
		f.Status.HistoryHead = &api.HistoryReference{URI: ref.URI, Version: ref.Version, SHA256: ref.Hash}
		f.Status.Pending = f.Status.Pending[1:]
		if err = r.Status().Update(ctx, f); err != nil {
			return err
		}
	}
	return nil
}
