package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestComputeComplianceScore(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"all passing", "10,0,0\n5,0,0", "100% compliant (15/15 checks passing, 0 not passing)"},
		{"mixed", "8,2,0\n0,0,0", "80% compliant (8/10 checks passing, 2 not passing)"},
		{"errors count as not passing", "8,0,2", "80% compliant (8/10 checks passing, 2 not passing)"},
		{"all failing", "0,4,0", "0% compliant (0/4 checks passing, 4 not passing)"},
		{"no reports", "", "No compliance data (no PolicyReports found)"},
		{"reports but no results", "0,0,0\n0,0,0", "No compliance data (no policy results yet)"},
		{"malformed fails closed", "10,0,0\nbroken", "UNKNOWN (unexpected PolicyReport data)"},
		{"wrong field count fails closed", "10,0", "UNKNOWN (unexpected PolicyReport data)"},
		{"extra field fails closed", "8,2,0,extra", "UNKNOWN (unexpected PolicyReport data)"},
		{"negative fails closed", "-1,0,0", "UNKNOWN (unexpected PolicyReport data)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := computeComplianceScore(c.raw); got != c.want {
				t.Errorf("computeComplianceScore(%q) = %q, want %q", c.raw, got, c.want)
			}
		})
	}
}

func TestComplianceScore(t *testing.T) {
	orig := kubectlPolicyReportQuery
	t.Cleanup(func() { kubectlPolicyReportQuery = orig })

	t.Run("aggregates query output", func(t *testing.T) {
		kubectlPolicyReportQuery = func(context.Context) ([]byte, error) {
			return []byte("9,1,0"), nil
		}
		if got := complianceScore(); got != "90% compliant (9/10 checks passing, 1 not passing)" {
			t.Errorf("complianceScore() = %q", got)
		}
	})

	t.Run("query error fails closed", func(t *testing.T) {
		kubectlPolicyReportQuery = func(context.Context) ([]byte, error) {
			return nil, errors.New("connection refused")
		}
		if got := complianceScore(); !strings.HasPrefix(got, "UNKNOWN") {
			t.Errorf("complianceScore() = %q, want UNKNOWN prefix", got)
		}
	})
}
