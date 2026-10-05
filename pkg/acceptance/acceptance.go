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
	if err := validateRecord(record, now); err != nil {
		return report, err
	}
	for _, check := range Required {
		ref, ok := record.Checks[check]
		if !ok {
			report.Missing = append(report.Missing, check)
			continue
		}
		passed, err := verifyCheck(ctx, store, record, check, ref)
		if err != nil {
			return report, err
		}
		if passed {
			report.Verified = append(report.Verified, check)
		} else {
			report.Missing = append(report.Missing, check)
		}
	}
	if len(report.Missing) > 0 {
		return report, fmt.Errorf("provider acceptance suite is incomplete")
	}
	report.State = "verified-receipts"
	return report, nil
}

func validateRecord(record Record, now time.Time) error {
	if record.Schema != 1 || (record.Provider != "aws" && record.Provider != "gcp" && record.Provider != "azure") || !identity.MatchString(record.ClusterID) || !identity.MatchString(record.RunID) || !digest.MatchString(record.OperatorDigest) || !digest.MatchString(record.BundleDigest) || record.FinishedAt.IsZero() || record.FinishedAt.After(now.Add(time.Minute)) || now.Sub(record.FinishedAt) > 30*24*time.Hour {
		return fmt.Errorf("invalid or stale provider acceptance identity")
	}
	return nil
}

func validateReference(record Record, check string, ref evidence.ObjectRef) error {
	u, err := url.Parse(ref.URI)
	if err != nil || u.Scheme != "s3" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "/"+Prefix(record, check)+"/"+ref.Hash+".json" || ref.Version == "" || !digest.MatchString("sha256:"+ref.Hash) {
		return fmt.Errorf("invalid acceptance receipt reference")
	}
	return nil
}

func verifyCheck(ctx context.Context, store evidence.Store, record Record, check string, ref evidence.ObjectRef) (bool, error) {
	if err := validateReference(record, check, ref); err != nil {
		return false, err
	}
	minimum := record.FinishedAt.AddDate(0, 0, evidence.MinimumRetentionDays)
	ref.RetainUntil = &minimum
	data, err := store.Get(ctx, ref)
	if err != nil {
		return false, fmt.Errorf("verify %s receipt: %w", check, err)
	}
	if len(data) > 1<<20 || evidence.Hash(data) != ref.Hash {
		return false, fmt.Errorf("invalid acceptance receipt content")
	}
	var receipt Receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&receipt); err != nil {
		return false, fmt.Errorf("decode %s receipt: %w", check, err)
	}
	if err = validateReceipt(record, check, receipt); err != nil {
		return false, err
	}
	if err = verifyArtifacts(ctx, store, receipt); err != nil {
		return false, err
	}
	return receipt.Passed, nil
}

func validateReceipt(record Record, check string, receipt Receipt) error {
	runner, err := url.Parse(receipt.RunnerURI)
	if err != nil || runner.Scheme != "https" || runner.Host != "github.com" || runner.User != nil || runner.RawQuery != "" || runner.Fragment != "" || !strings.HasPrefix(runner.Path, "/somethingwithproof/mantl/actions/runs/") {
		return fmt.Errorf("invalid acceptance runner provenance")
	}
	if receipt.OperatorDigest != record.OperatorDigest || receipt.BundleDigest != record.BundleDigest || receipt.Provider != record.Provider || receipt.ClusterID != record.ClusterID || receipt.RunID != record.RunID || receipt.Check != check || receipt.ObservedAt.IsZero() || receipt.ObservedAt.After(record.FinishedAt) || record.FinishedAt.Sub(receipt.ObservedAt) > 24*time.Hour || receipt.RunnerURI == "" || len(receipt.Artifacts) == 0 || len(receipt.Artifacts) > 20 {
		return fmt.Errorf("acceptance receipt scope or provenance mismatch")
	}
	return nil
}

func verifyArtifacts(ctx context.Context, store evidence.Store, receipt Receipt) error {
	// Every receipt includes verifiable, retained source artifacts. This validates
	// a recorded suite result, rather than independently inspecting a live cloud.
	for _, artifact := range receipt.Artifacts {
		if artifact.Version == "" || !digest.MatchString("sha256:"+artifact.Hash) {
			return fmt.Errorf("invalid artifact identity")
		}
		retained := receipt.ObservedAt.AddDate(0, 0, evidence.MinimumRetentionDays)
		artifact.RetainUntil = &retained
		data, err := store.Get(ctx, artifact)
		if err != nil {
			return fmt.Errorf("verify %s artifact %s: %w", receipt.Check, artifact.URI, err)
		}
		if evidence.Hash(data) != artifact.Hash {
			return fmt.Errorf("acceptance artifact hash mismatch")
		}
	}
	return nil
}
