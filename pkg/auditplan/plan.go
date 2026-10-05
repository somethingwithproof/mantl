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
		for i, m := range control.Mappings {
			c := m.EvidenceCollector
			if c == nil {
				continue
			}
			key := fmt.Sprintf("%s/%d", control.ID, i)
			if audit.Spec.Frequency == "framework" {
				last, exists := audit.Status.CollectorTimes[key]
				base := now
				if exists {
					base = last.Time
				}
				when, err := framework.Next(c.Schedule, base)
				if err != nil {
					p.Gaps = append(p.Gaps, control.ID+": invalid schedule")
					continue
				}
				if exists && when.After(now) {
					p.wait(when.Sub(now))
					continue
				}
				next, _ := framework.Next(c.Schedule, now)
				p.wait(next.Sub(now))
			}
			if c.Type != "config-snapshot" || c.Query != "" || len(c.Resources) == 0 {
				p.Gaps = append(p.Gaps, control.ID+": unsupported collector "+c.Type)
				continue
			}
			for _, name := range c.Resources {
				resource, err := evidence.Resource(name)
				if err != nil {
					p.Gaps = append(p.Gaps, control.ID+": unsupported resource "+name)
					continue
				}
				scope := namespaces
				if resource.Cluster {
					scope = []string{""}
				}
				for _, ns := range scope {
					id := evidence.Hash([]byte(key + "\x00" + name + "\x00" + ns))[:32]
					if seen[id] {
						continue
					}
					seen[id] = true
					p.Tasks = append(p.Tasks, api.CollectorTask{ID: id, Collector: key, Control: control.ID, Resource: name, Namespace: ns})
				}
			}
		}
	}
	return p
}
func (p *Plan) wait(d time.Duration) {
	if d > 0 && (p.Next == 0 || d < p.Next) {
		p.Next = d
	}
}
