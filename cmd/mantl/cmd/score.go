package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// policyReportSummaryJSONPath emits one "pass,fail,error" line per report from
// each report's aggregated summary counts. warn and skip are intentionally
// excluded from the score (see computeComplianceScore).
const policyReportSummaryJSONPath = `jsonpath={range .items[*]}{.summary.pass},{.summary.fail},{.summary.error}{"\n"}{end}`

// kubectlPolicyReportQuery returns the raw summary lines (stdout only) for both
// namespaced PolicyReports and cluster-scoped ClusterPolicyReports. It is a
// package var so tests can compute a score without a cluster.
var kubectlPolicyReportQuery = func(ctx context.Context) ([]byte, error) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return nil, fmt.Errorf("kubectl not found in PATH: %w", err)
	}
	out, err := exec.CommandContext(ctx, "kubectl", "get",
		"policyreports,clusterpolicyreports", "-A", "-o", policyReportSummaryJSONPath).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, err
	}
	return out, nil
}

// complianceScore aggregates Kyverno report summaries into a one-line score. It
// does not wait and fails closed: a query error or unexpected data yields
// UNKNOWN rather than a misleading number.
func complianceScore() string {
	ctx, cancel := context.WithTimeout(context.Background(), statusCheckTimeout)
	defer cancel()

	out, err := kubectlPolicyReportQuery(ctx)
	if err != nil {
		return fmt.Sprintf("UNKNOWN (cluster unreachable: %v)", err)
	}
	return computeComplianceScore(string(out))
}

// computeComplianceScore sums the per-report "pass,fail,error" lines and formats
// a score. error is counted as not-passing (an unevaluated control is not a
// passing one), so the score cannot inflate by dropping failures. warn and skip
// are excluded: skip means not-applicable and warn is advisory. Pure;
// unparseable or negative data fails closed to UNKNOWN, and no results is
// reported as no data, never as 100%.
func computeComplianceScore(raw string) string {
	var pass, notPassing int64
	rows := 0
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) != 3 {
			return "UNKNOWN (unexpected PolicyReport data)"
		}
		p, okP := parseCount(fields[0])
		f, okF := parseCount(fields[1])
		e, okE := parseCount(fields[2])
		if !okP || !okF || !okE {
			return "UNKNOWN (unexpected PolicyReport data)"
		}
		pass += p
		notPassing += f + e
		rows++
	}

	if rows == 0 {
		return "No compliance data (no PolicyReports found)"
	}
	total := pass + notPassing
	if total == 0 {
		return "No compliance data (no policy results yet)"
	}
	pct := int64(float64(pass) / float64(total) * 100)
	return fmt.Sprintf("%d%% compliant (%d/%d checks passing, %d not passing)", pct, pass, total, notPassing)
}

// parseCount parses a non-negative integer count. Empty, non-numeric, or
// negative input is rejected so callers fail closed.
func parseCount(s string) (int64, bool) {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil || v < 0 {
		return 0, false
	}
	return v, true
}
