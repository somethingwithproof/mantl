package compliance

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"time"
)

// ComplianceMonitor exposes aggregate state, never resource payloads or identities.
type ComplianceMonitor struct {
	Client                              client.Client
	gaps, failed, age, findings, errors prometheus.Gauge
}

func NewComplianceMonitor(c client.Client) *ComplianceMonitor {
	m := &ComplianceMonitor{Client: c}
	gauge := func(name, help string) prometheus.Gauge {
		g := prometheus.NewGauge(prometheus.GaugeOpts{Name: name, Help: help})
		metrics.Registry.MustRegister(g)
		return g
	}
	m.gaps = gauge("mantl_compliance_coverage_gaps", "Unresolved policy templates plus audit collector gaps")
	m.failed = gauge("mantl_compliance_audits_failed", "Audits in failed or partially completed state")
	m.age = gauge("mantl_compliance_evidence_age_seconds", "Age of oldest audit manifest; -1 when no evidence exists")
	m.findings = gauge("mantl_compliance_findings_open", "Findings with failure status")
	m.errors = gauge("mantl_compliance_monitor_error", "1 when current compliance metadata cannot be read")
	m.age.Set(-1)
	return m
}
func (m *ComplianceMonitor) NeedLeaderElection() bool { return false }
func (m *ComplianceMonitor) Start(ctx context.Context) error {
	timer := time.NewTicker(30 * time.Second)
	defer timer.Stop()
	for {
		m.refresh(ctx)
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
		}
	}
}
func (m *ComplianceMonitor) refresh(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var profiles api.ComplianceProfileList
	var audits api.ComplianceAuditList
	var findings api.FindingList
	if m.Client.List(ctx, &profiles) != nil || m.Client.List(ctx, &audits) != nil || m.Client.List(ctx, &findings) != nil {
		m.errors.Set(1)
		return
	}
	m.errors.Set(0)
	gaps, failed, open := 0, 0, 0
	for _, p := range profiles.Items {
		gaps += len(p.Status.UnresolvedTemplates)
	}
	auditGaps, auditFailures, oldest := auditMetrics(audits.Items)
	gaps += auditGaps
	failed += auditFailures
	age := oldest
	for _, f := range findings.Items {
		if f.Spec.Status == "fail" {
			open++
		}
	}
	m.gaps.Set(float64(gaps))
	m.failed.Set(float64(failed))
	m.findings.Set(float64(open))
	m.age.Set(age)
}

func auditMetrics(audits []api.ComplianceAudit) (gaps, failed int, age float64) {
	age = -1
	for _, a := range audits {
		gaps += len(a.Status.CoverageGaps)
		if a.Status.Phase == "Failed" || a.Status.Phase == "PartiallyCompleted" {
			failed++
		}
		if a.Status.EvidenceURI != "" && a.Status.EndTime != nil {
			seconds := time.Since(a.Status.EndTime.Time).Seconds()
			if seconds > age {
				age = seconds
			}
		}
	}
	return gaps, failed, age
}
