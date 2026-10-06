// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestComplianceStatusSeparatesEvidenceAndCoverage(t *testing.T) {
	complianceStatusCmd.SetContext(context.Background())
	old := queryCompliance
	defer func() { queryCompliance = old }()
	queryCompliance = func(_ context.Context, kind, name, ns string) ([]byte, error) { return []byte(`{"items":[]}`), nil }
	var out bytes.Buffer
	complianceStatusCmd.SetOut(&out)
	defer complianceStatusCmd.SetOut(nil)
	if err := complianceStatusCmd.RunE(complianceStatusCmd, nil); err != nil {
		t.Fatal(err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil || len(payload) != 6 {
		t.Fatal("missing compliance dimensions")
	}
	for _, kind := range []string{"complianceprofiles", "findings", "complianceaudits", "auditruns", "controlevaluations", "complianceexceptions"} {
		if _, ok := payload[kind]; !ok {
			t.Fatalf("missing %s", kind)
		}
	}
	queryCompliance = func(context.Context, string, string, string) ([]byte, error) {
		return nil, fmt.Errorf("cluster denied")
	}
	if err := complianceStatusCmd.RunE(complianceStatusCmd, nil); err == nil {
		t.Fatal("cluster error hidden")
	}
}
