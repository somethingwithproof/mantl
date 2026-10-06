// SPDX-License-Identifier: Apache-2.0

package compiler

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
)

const PlanSchemaVersion = "mantl.io/plan/v1alpha2"
const MaxPlanBytes = 8 << 20
const MaxArtifactBytes = 8 << 20
const maxPlanArtifacts = 4096
const maxPlanContentBytes = 32 << 20

// ValidatePlan validates inventory structure before comparison or file access.
func ValidatePlan(plan Plan) error {
	if plan.SchemaVersion != PlanSchemaVersion {
		return fmt.Errorf("unsupported plan schema; regenerate the plan with this CLI")
	}
	if plan.Platform == "" || plan.Provider == "" || plan.Distribution == "" || !validSHA256(plan.SpecSHA256) {
		return fmt.Errorf("plan requires platform identity and a spec SHA256")
	}
	if len(plan.Artifacts) == 0 || len(plan.Artifacts) > maxPlanArtifacts {
		return fmt.Errorf("invalid plan artifact count")
	}
	seen := map[string]bool{}
	var total int
	for _, artifact := range plan.Artifacts {
		if artifact.Path == "." || path.Clean(artifact.Path) != artifact.Path || strings.Contains(artifact.Path, "\\") || strings.HasPrefix(artifact.Path, "/") || strings.HasPrefix(artifact.Path, "../") || artifact.Path == ".." || seen[artifact.Path] {
			return fmt.Errorf("artifact paths must be unique, canonical relative paths")
		}
		if artifact.Size <= 0 || artifact.Size > MaxArtifactBytes || !validSHA256(artifact.SHA256) {
			return fmt.Errorf("invalid artifact size or hash for %s", artifact.Path)
		}
		seen[artifact.Path] = true
		total += artifact.Size
	}
	if total > maxPlanContentBytes {
		return fmt.Errorf("plan artifact content exceeds size limit")
	}
	return nil
}

func validSHA256(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && value == strings.ToLower(value)
}

// DecodePlan bounds metadata and rejects unknown fields and trailing JSON.
func DecodePlan(reader io.Reader) (Plan, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxPlanBytes+1))
	if err != nil {
		return Plan{}, fmt.Errorf("read saved plan: %w", err)
	}
	if len(data) > MaxPlanBytes {
		return Plan{}, fmt.Errorf("saved plan exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var plan Plan
	if err := decoder.Decode(&plan); err != nil {
		return Plan{}, fmt.Errorf("decode saved plan: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return Plan{}, fmt.Errorf("saved plan must contain exactly one JSON object")
	}
	if err := ValidatePlan(plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

// MatchPlan requires the reviewed metadata to match current compilation exactly.
func MatchPlan(reviewed, current Plan) error {
	if err := ValidatePlan(reviewed); err != nil {
		return err
	}
	if err := ValidatePlan(current); err != nil {
		return err
	}
	before, err := json.Marshal(reviewed)
	if err != nil {
		return fmt.Errorf("encode reviewed plan: %w", err)
	}
	after, err := json.Marshal(current)
	if err != nil {
		return fmt.Errorf("encode current plan: %w", err)
	}
	if !bytes.Equal(before, after) {
		return fmt.Errorf("saved plan differs from the current spec or compiler; regenerate and review it")
	}
	return nil
}

type ArtifactChange struct {
	Path   string `json:"path"`
	Action string `json:"action"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

type PlanComparison struct {
	Changed      bool             `json:"changed"`
	InputChanged bool             `json:"inputChanged"`
	Artifacts    []ArtifactChange `json:"artifacts"`
}

// ComparePlans reports artifact identity changes; it does not query live state.
func ComparePlans(before, after Plan) (PlanComparison, error) {
	for _, plan := range []Plan{before, after} {
		if err := ValidatePlan(plan); err != nil {
			return PlanComparison{}, err
		}
	}
	comparison := PlanComparison{InputChanged: before.SpecSHA256 != after.SpecSHA256, Artifacts: []ArtifactChange{}}
	old := map[string]Artifact{}
	for _, artifact := range before.Artifacts {
		old[artifact.Path] = artifact
	}
	for _, artifact := range after.Artifacts {
		previous, exists := old[artifact.Path]
		if !exists {
			comparison.Artifacts = append(comparison.Artifacts, ArtifactChange{Path: artifact.Path, Action: "added", After: artifact.SHA256})
		} else if previous.SHA256 != artifact.SHA256 || previous.Size != artifact.Size {
			comparison.Artifacts = append(comparison.Artifacts, ArtifactChange{Path: artifact.Path, Action: "changed", Before: previous.SHA256, After: artifact.SHA256})
		}
		delete(old, artifact.Path)
	}
	for _, artifact := range old {
		comparison.Artifacts = append(comparison.Artifacts, ArtifactChange{Path: artifact.Path, Action: "removed", Before: artifact.SHA256})
	}
	sort.Slice(comparison.Artifacts, func(i, j int) bool { return comparison.Artifacts[i].Path < comparison.Artifacts[j].Path })
	comparison.Changed = comparison.InputChanged || len(comparison.Artifacts) != 0
	return comparison, nil
}
