package compliance

import (
	"context"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/fleet"
	core "k8s.io/api/core/v1"
	"log/slog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type FleetPublisher struct {
	Client    client.Client
	Reader    client.Reader
	Sent      map[string]bool
	Remote    *fleet.Client
	Store     evidence.Store
	ClusterID string
}

func (p *FleetPublisher) NeedLeaderElection() bool { return true }
func (p *FleetPublisher) Start(ctx context.Context) error {
	var namespace core.Namespace
	reader := p.Reader
	if reader == nil {
		reader = p.Client
	}
	if err := reader.Get(ctx, client.ObjectKey{Name: "kube-system"}, &namespace); err != nil {
		return fmt.Errorf("verify enrolled cluster identity: %w", err)
	}
	if string(namespace.UID) != p.ClusterID {
		return fmt.Errorf("fleet enrollment does not match this cluster identity")
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		publish, cancel := context.WithTimeout(ctx, 20*time.Second)
		err := p.publish(publish)
		cancel()
		if err != nil && ctx.Err() == nil {
			slog.Warn("fleet metadata publication failed")
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
func (p *FleetPublisher) publish(ctx context.Context) error {
	var runs api.AuditRunList
	var evaluations api.ControlEvaluationList
	var findings api.FindingList
	if err := p.Client.List(ctx, &runs); err != nil {
		return err
	}
	if err := p.Client.List(ctx, &evaluations); err != nil {
		return err
	}
	if err := p.Client.List(ctx, &findings); err != nil {
		return err
	}
	var events []fleet.Event
	event := func(object client.Object, kind string, at time.Time) fleet.Event {
		return fleet.Event{ID: evidence.Hash([]byte(kind + "/" + string(object.GetUID()) + "/" + object.GetResourceVersion())), Kind: kind, ResourceUID: string(object.GetUID()), Namespace: object.GetNamespace(), Name: object.GetName(), ObservedAt: at}
	}
	for _, run := range runs.Items {
		if run.Status.EndTime == nil {
			continue
		}
		e := event(&run, "AuditRun", run.Status.EndTime.Time)
		e.Profile = run.Spec.Profile
		e.Result = run.Status.Phase
		e.EvidenceURI = run.Status.EvidenceURI
		e.ManifestHash = run.Status.ManifestHash
		events = append(events, e)
	}
	for _, evaluation := range evaluations.Items {
		if evaluation.Status.ObservedAt == nil {
			continue
		}
		e := event(&evaluation, "ControlEvaluation", evaluation.Status.ObservedAt.Time)
		e.Profile = evaluation.Spec.Profile
		e.Control = evaluation.Spec.Control
		e.Result = evaluation.Status.Result
		e.Coverage = evaluation.Status.Coverage
		e.Freshness = evaluation.Status.Freshness
		e.EvidenceURI = evaluation.Status.EvidenceURI
		events = append(events, e)
	}
	for _, finding := range findings.Items {
		at, err := time.Parse(time.RFC3339Nano, finding.Annotations["mantl.io/observed-at"])
		if err != nil {
			continue
		}
		e := event(&finding, "Finding", at)
		e.Control = finding.Spec.ControlID
		e.Result = finding.Spec.Status
		events = append(events, e)
	}
	if p.Sent == nil {
		p.Sent = map[string]bool{}
	}
	unsent := events[:0]
	for _, event := range events {
		if !p.Sent[event.ID] {
			unsent = append(unsent, event)
		}
	}
	events = unsent
	for offset := 0; offset < len(events); offset += 100 {
		end := offset + 100
		if end > len(events) {
			end = len(events)
		}
		if err := p.Remote.Journal(ctx, p.Store, p.ClusterID, events[offset:end], time.Now().UTC()); err != nil {
			return err
		}
		for _, event := range events[offset:end] {
			p.Sent[event.ID] = true
		}
	}
	if len(p.Sent) > 10000 {
		p.Sent = map[string]bool{}
	}
	return nil
}
