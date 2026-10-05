// Package fleet indexes immutable metadata journals under authenticated enrollment.
package fleet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"net/url"
	"regexp"
	"time"
)

var ErrConflict = errors.New("event identity already has different content")
var scopePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`)
var digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Scope struct {
	Tenant  string `json:"tenant"`
	Cluster string `json:"cluster"`
	Role    string `json:"role"`
}
type Event struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	ResourceUID  string    `json:"resourceUid"`
	Namespace    string    `json:"namespace"`
	Name         string    `json:"name"`
	ObservedAt   time.Time `json:"observedAt"`
	Profile      string    `json:"profile,omitempty"`
	Control      string    `json:"control,omitempty"`
	Result       string    `json:"result,omitempty"`
	Coverage     string    `json:"coverage,omitempty"`
	Freshness    string    `json:"freshness,omitempty"`
	EvidenceURI  string    `json:"evidenceUri,omitempty"`
	ManifestHash string    `json:"manifestHash,omitempty"`
}
type Journal struct {
	Schema  int     `json:"schemaVersion"`
	Cluster string  `json:"cluster"`
	Events  []Event `json:"events"`
}
type Entry struct {
	Scope  Scope              `json:"scope"`
	Event  Event              `json:"event"`
	Source evidence.ObjectRef `json:"source"`
}
type Index interface {
	Put(context.Context, Scope, []Event, evidence.ObjectRef) error
	List(context.Context, Scope, int, string, bool) ([]Entry, error)
}

func Prefix(cluster string) string { return "audits/fleet/" + cluster }
func Canonical(cluster string, events []Event) ([]byte, error) {
	return json.Marshal(Journal{Schema: 1, Cluster: cluster, Events: events})
}
func Verify(ctx context.Context, store evidence.Store, scope Scope, ref evidence.ObjectRef, now time.Time) ([]Event, error) {
	u, err := url.Parse(ref.URI)
	if err != nil || u.Scheme != "s3" || u.User != nil || u.RawQuery != "" || u.Path != "/"+Prefix(scope.Cluster)+"/"+ref.Hash+".json" || !digestPattern.MatchString(ref.Hash) {
		return nil, fmt.Errorf("invalid fleet journal reference")
	}
	b, err := store.Get(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("verify fleet journal: %w", err)
	}
	if len(b) > 1<<20 || evidence.Hash(b) != ref.Hash {
		return nil, fmt.Errorf("invalid fleet journal content")
	}
	var journal Journal
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&journal); err != nil {
		return nil, err
	}
	if journal.Schema != 1 || journal.Cluster != scope.Cluster || len(journal.Events) == 0 || len(journal.Events) > 100 {
		return nil, fmt.Errorf("invalid enrolled journal scope")
	}
	seen := map[string]bool{}
	for _, e := range journal.Events {
		if err := validateEvent(e, seen, ref, now); err != nil {
			return nil, err
		}

	}
	return journal.Events, nil
}

func validateEvent(e Event, seen map[string]bool, ref evidence.ObjectRef, now time.Time) error {
	if ref.RetainUntil == nil || ref.RetainUntil.Before(e.ObservedAt.AddDate(0, 0, evidence.MinimumRetentionDays)) {
		return fmt.Errorf("fleet journal retention is shorter than the event minimum")
	}
	if !digestPattern.MatchString(e.ID) || seen[e.ID] || !scopePattern.MatchString(e.ResourceUID) || len(e.Name) == 0 || len(e.Name) > 253 || !scopePattern.MatchString(e.Namespace) || len(e.Profile) > 253 || len(e.Control) > 253 || len(e.Result) > 32 || len(e.Coverage) > 32 || len(e.Freshness) > 32 || e.ObservedAt.IsZero() || e.ObservedAt.After(now.Add(5*time.Minute)) {
		return fmt.Errorf("invalid fleet event identity or time")
	}
	seen[e.ID] = true
	switch e.Kind {
	case "Finding", "ControlEvaluation", "AuditRun":
	default:
		return fmt.Errorf("unsupported fleet metadata kind")
	}
	if e.EvidenceURI != "" {
		u, err := url.Parse(e.EvidenceURI)
		if err != nil || u.Scheme != "s3" || u.User != nil || u.Host == "" || u.Fragment != "" || len(e.EvidenceURI) > 4096 || len(u.Query()) > 1 || u.RawQuery != "" && u.Query().Get("versionId") == "" {
			return fmt.Errorf("invalid evidence metadata reference")
		}
	}
	return nil
}
