// Package acceptance verifies provider acceptance records without promoting
// static validation or an unauthenticated assertion to deployment evidence.
package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var Required = []string{"installation", "workload-identity", "private-access", "audit-delivery", "encryption", "upgrade", "recovery", "evidence-retention"}
var digest = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
var identity = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,127}$`)

type Record struct {
	Schema         int                           `json:"schemaVersion"`
	Provider       string                        `json:"provider"`
	ClusterID      string                        `json:"clusterId"`
	RunID          string                        `json:"runId"`
	OperatorDigest string                        `json:"operatorDigest"`
	BundleDigest   string                        `json:"bundleDigest"`
	FinishedAt     time.Time                     `json:"finishedAt"`
	Checks         map[string]evidence.ObjectRef `json:"checks"`
}
type Receipt struct {
	OperatorDigest string               `json:"operatorDigest"`
	BundleDigest   string               `json:"bundleDigest"`
	Provider       string               `json:"provider"`
	ClusterID      string               `json:"clusterId"`
	RunID          string               `json:"runId"`
	Check          string               `json:"check"`
	Passed         bool                 `json:"passed"`
	ObservedAt     time.Time            `json:"observedAt"`
	RunnerURI      string               `json:"runnerUri"`
	Artifacts      []evidence.ObjectRef `json:"artifacts"`
}
type Report struct {
	Provider  string   `json:"provider"`
	ClusterID string   `json:"clusterId"`
	Verified  []string `json:"verified"`
	Missing   []string `json:"missing"`
	State     string   `json:"state"`
}

func Prefix(record Record, check string) string {
	return "audits/acceptance/" + record.Provider + "/" + record.ClusterID + "/" + record.RunID + "/" + check
}
func Verify(ctx context.Context, store evidence.Store, record Record, now time.Time) (Report, error) {
	report := Report{Provider: record.Provider, ClusterID: record.ClusterID, State: "incomplete", Verified: []string{}, Missing: []string{}}
	if record.Schema != 1 || (record.Provider != "aws" && record.Provider != "gcp" && record.Provider != "azure") || !identity.MatchString(record.ClusterID) || !identity.MatchString(record.RunID) || !digest.MatchString(record.OperatorDigest) || !digest.MatchString(record.BundleDigest) || record.FinishedAt.IsZero() || record.FinishedAt.After(now.Add(time.Minute)) || now.Sub(record.FinishedAt) > 30*24*time.Hour {
		return report, fmt.Errorf("invalid or stale provider acceptance identity")
	}
	for _, check := range Required {
		ref, ok := record.Checks[check]
		if !ok {
			report.Missing = append(report.Missing, check)
			continue
		}
		u, err := url.Parse(ref.URI)
		if err != nil || u.Scheme != "s3" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "/"+Prefix(record, check)+"/"+ref.Hash+".json" || ref.Version == "" || !digest.MatchString("sha256:"+ref.Hash) {
			return report, fmt.Errorf("invalid acceptance receipt reference")
		}
		minimum := record.FinishedAt.AddDate(0, 0, evidence.MinimumRetentionDays)
		ref.RetainUntil = &minimum
		data, err := store.Get(ctx, ref)
		if err != nil {
			return report, fmt.Errorf("verify %s receipt: %w", check, err)
		}
		if len(data) > 1<<20 || evidence.Hash(data) != ref.Hash {
			return report, fmt.Errorf("invalid acceptance receipt content")
		}
		var receipt Receipt
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&receipt); err != nil {
			return report, err
		}
		runner, runnerErr := url.Parse(receipt.RunnerURI)
		if runnerErr != nil || runner.Scheme != "https" || runner.Host != "github.com" || runner.User != nil || runner.RawQuery != "" || runner.Fragment != "" || !strings.HasPrefix(runner.Path, "/somethingwithproof/mantl/actions/runs/") {
			return report, fmt.Errorf("invalid acceptance runner provenance")
		}
		if receipt.OperatorDigest != record.OperatorDigest || receipt.BundleDigest != record.BundleDigest || receipt.Provider != record.Provider || receipt.ClusterID != record.ClusterID || receipt.RunID != record.RunID || receipt.Check != check || receipt.ObservedAt.IsZero() || receipt.ObservedAt.After(record.FinishedAt) || record.FinishedAt.Sub(receipt.ObservedAt) > 24*time.Hour || receipt.RunnerURI == "" || len(receipt.Artifacts) == 0 || len(receipt.Artifacts) > 20 {
			return report, fmt.Errorf("acceptance receipt scope or provenance mismatch")
		}
		// Every receipt must include verifiable, retained source artifacts. This
		// validates a recorded suite result, not an independent live cloud inspection.
		for _, artifact := range receipt.Artifacts {
			if artifact.Version == "" || !digest.MatchString("sha256:"+artifact.Hash) {
				return report, fmt.Errorf("invalid artifact identity")
			}
			retained := receipt.ObservedAt.AddDate(0, 0, evidence.MinimumRetentionDays)
			artifact.RetainUntil = &retained
			b, err := store.Get(ctx, artifact)
			if err != nil {
				return report, err
			}
			if evidence.Hash(b) != artifact.Hash {
				return report, fmt.Errorf("acceptance artifact hash mismatch")
			}
		}
		if !receipt.Passed {
			report.Missing = append(report.Missing, check)
		} else {
			report.Verified = append(report.Verified, check)
		}
	}
	if len(report.Missing) > 0 {
		return report, fmt.Errorf("provider acceptance suite is incomplete")
	}
	report.State = "verified-receipts"
	return report, nil
}
