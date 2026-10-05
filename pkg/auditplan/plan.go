// Package auditplan computes deterministic collection scopes and schedules.
package auditplan

import (
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/framework"
	"time"
)

type Plan struct {
	Tasks []api.CollectorTask
	Gaps  []string
	Next  time.Duration
}

func Build(fw *framework.Framework, profile *api.ComplianceProfile, audit *api.ComplianceAudit, now time.Time) Plan {
	seen := map[string]bool{}
	p := Plan{Tasks: []api.CollectorTask{}, Gaps: []string{}}
	namespaces := profile.Spec.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{profile.Namespace}
	}
	for _, control := range fw.Controls {
		if !framework.Selected(control.ID, profile.Spec.IncludeControls) {
			continue
		}
		for index, mapping := range control.Mappings {
			if mapping.EvidenceCollector == nil {
				continue
			}
			key := fmt.Sprintf("%s/%d", control.ID, index)
			p.collect(control.ID, key, mapping.EvidenceCollector, audit, namespaces, now, seen)
		}
	}
	return p
}

func (p *Plan) due(control, key string, collector *framework.EvidenceCollector, audit *api.ComplianceAudit, now time.Time) bool {
	if audit.Spec.Frequency != "framework" {
		return true
	}
	last, exists := audit.Status.CollectorTimes[key]
	base := now
	if exists {
		base = last.Time
	}
	when, err := framework.Next(collector.Schedule, base)
	if err != nil {
		p.Gaps = append(p.Gaps, control+": invalid schedule")
		return false
	}
	if exists && when.After(now) {
		p.wait(when.Sub(now))
		return false
	}
	next, _ := framework.Next(collector.Schedule, now)
	p.wait(next.Sub(now))
	return true
}

func (p *Plan) collect(control, key string, collector *framework.EvidenceCollector, audit *api.ComplianceAudit, namespaces []string, now time.Time, seen map[string]bool) {
	if !p.due(control, key, collector, audit, now) {
		return
	}
	if collector.Type != "config-snapshot" || collector.Query != "" || len(collector.Resources) == 0 {
		p.Gaps = append(p.Gaps, control+": unsupported collector "+collector.Type)
		return
	}
	for _, name := range collector.Resources {
		resource, err := evidence.Resource(name)
		if err != nil {
			p.Gaps = append(p.Gaps, control+": unsupported resource "+name)
			continue
		}
		scope := namespaces
		if resource.Cluster {
			scope = []string{""}
		}
		p.addTasks(control, key, name, scope, seen)
	}
}

func (p *Plan) addTasks(control, key, resource string, namespaces []string, seen map[string]bool) {
	for _, namespace := range namespaces {
		id := evidence.Hash([]byte(key + "\x00" + resource + "\x00" + namespace))[:32]
		if seen[id] {
			continue
		}
		seen[id] = true
		p.Tasks = append(p.Tasks, api.CollectorTask{ID: id, Collector: key, Control: control, Resource: resource, Namespace: namespace})
	}
}
func (p *Plan) wait(d time.Duration) {
	if d > 0 && (p.Next == 0 || d < p.Next) {
		p.Next = d
	}
}
