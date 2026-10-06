// SPDX-License-Identifier: Apache-2.0

package compliance

import (
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func validAuditFrequency(frequency string) bool {
	switch frequency {
	case "", "manual", "daily", "weekly", "framework":
		return true
	default:
		return false
	}
}

func regularAuditWait(audit *api.ComplianceAudit, now time.Time) (ctrl.Result, bool) {
	if audit.Spec.Frequency == "framework" {
		return ctrl.Result{}, false
	}
	var end *time.Time
	if audit.Status.EndTime != nil {
		end = &audit.Status.EndTime.Time
	}
	action, after := planAudit(audit.Status.Phase, end, audit.Spec.Frequency, now)
	switch action {
	case actionDone:
		return ctrl.Result{}, true
	case actionWaitRequeue:
		return ctrl.Result{RequeueAfter: after}, true
	default:
		return ctrl.Result{}, false
	}
}

func auditCompletionPhase(objects, gaps int) string {
	switch {
	case objects == 0 && gaps == 0:
		return "NoEvidence"
	case objects == 0:
		return "Failed"
	case gaps > 0:
		return "PartiallyCompleted"
	default:
		return "Completed"
	}
}
