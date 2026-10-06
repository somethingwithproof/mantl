// SPDX-License-Identifier: Apache-2.0

package auditplan

import (
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/framework"
	"time"
)

// InlineTask and InlinePlan preserve the legacy inline adapter's collector keys
// and duplicate resource scopes. The isolated Job planner remains Build.
type InlineTask struct{ Key, Control, Resource, Namespace string }
type InlinePlan struct {
	Tasks           []InlineTask
	Gaps            []string
	Next            time.Duration
	Due, Incomplete map[string]bool
}

func BuildInline(fw *framework.Framework, profile *api.ComplianceProfile, audit *api.ComplianceAudit, now time.Time) InlinePlan {
	p := InlinePlan{Due: map[string]bool{}, Incomplete: map[string]bool{}}
	namespaces := profile.Spec.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{profile.Namespace}
	}
	for _, control := range fw.Controls {
		if !framework.Selected(control.ID, profile.Spec.IncludeControls) {
			continue
		}
		for i, mapping := range control.Mappings {
			if mapping.EvidenceCollector == nil {
				continue
			}
			p.collectInline(control.ID, i, mapping.EvidenceCollector, namespaces, audit, now)
		}
	}
	return p
}

func (p *InlinePlan) collectInline(control string, index int, collector *framework.EvidenceCollector, namespaces []string, audit *api.ComplianceAudit, now time.Time) {
	key := fmt.Sprintf("%s/%d", control, index)
	if !p.inlineDue(key, control, collector.Schedule, audit, now) {
		return
	}

	p.Due[key] = true
	if collector.Type != "config-snapshot" || collector.Query != "" || len(collector.Resources) == 0 {
		p.Gaps = append(p.Gaps, control+": unsupported collector "+collector.Type)
		return
	}
	for _, resource := range collector.Resources {
		typ, typeErr := evidence.Resource(resource)
		if typeErr != nil {
			p.Incomplete[key] = true
			p.Gaps = append(p.Gaps, control+": unsupported resource "+resource)
			continue
		}
		if typ.Cluster {
			p.Tasks = append(p.Tasks, InlineTask{key, control, resource, ""})
		} else {
			for _, ns := range namespaces {
				p.Tasks = append(p.Tasks, InlineTask{key, control, resource, ns})
			}
		}
	}
}

func (p *InlinePlan) inlineDue(key, control, schedule string, audit *api.ComplianceAudit, now time.Time) bool {
	if audit.Spec.Frequency != "framework" {
		return true
	}

	last, exists := audit.Status.CollectorTimes[key]
	base := now
	if exists {
		base = last.Time
	}
	when, scheduleErr := framework.Next(schedule, base)
	if scheduleErr != nil {
		p.Gaps = append(p.Gaps, control+": invalid schedule")
		return false
	}
	if exists && when.After(now) {
		wait := when.Sub(now)
		p.waitInline(wait)
		return false
	}
	// First observation runs immediately; missed intervals coalesce into one run.
	nextWhen, _ := framework.Next(schedule, now)
	wait := nextWhen.Sub(now)
	p.waitInline(wait)
	return true
}

func (p *InlinePlan) waitInline(wait time.Duration) {
	if p.Next == 0 || wait < p.Next {
		p.Next = wait
	}
}
