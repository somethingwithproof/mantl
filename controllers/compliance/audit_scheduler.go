package compliance

import (
	"context"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/auditplan"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
	"time"
)

type AuditScheduler struct {
	CollectorImage, CollectorServiceAccount, ClusterID, Bucket, KMSKey string
	client.Client
	ControlNamespace string
	FrameworkDir     string
	Now              func() time.Time
}

func (r *AuditScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace != r.ControlNamespace {
		return ctrl.Result{}, nil
	}
	var audit api.ComplianceAudit
	if err := r.Get(ctx, req.NamespacedName, &audit); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	now := time.Now().UTC()
	if r.Now != nil {
		now = r.Now().UTC()
	}
	if audit.Spec.Frequency != "" && audit.Spec.Frequency != "manual" && audit.Spec.Frequency != "daily" && audit.Spec.Frequency != "weekly" && audit.Spec.Frequency != "framework" {
		return ctrl.Result{}, fmt.Errorf("unsupported audit frequency")
	}
	if audit.Status.ActiveRun != "" {
		var run api.AuditRun
		if err := r.Get(ctx, client.ObjectKey{Namespace: audit.Namespace, Name: audit.Status.ActiveRun}, &run); err != nil {
			return ctrl.Result{}, fmt.Errorf("load active audit run: %w", err)
		}
		if run.Spec.AuditUID != string(audit.UID) {
			return ctrl.Result{}, fmt.Errorf("audit run ownership mismatch")
		}
		if run.Status.EndTime == nil {
			return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
		}
		audit.Status.Phase = run.Status.Phase
		audit.Status.EndTime = run.Status.EndTime
		audit.Status.EvidenceURI = run.Status.EvidenceURI
		audit.Status.ManifestHash = run.Status.ManifestHash
		audit.Status.CoverageGaps = run.Status.CoverageGaps
		audit.Status.FailedCount = int32(len(run.Status.CoverageGaps))
		audit.Status.FindingCount = 0
		if audit.Status.CollectorTimes == nil {
			audit.Status.CollectorTimes = map[string]metav1.Time{}
		}
		failed := map[string]bool{}
		for _, task := range run.Spec.Tasks {
			if run.Status.Results[task.ID].Phase != "Completed" {
				failed[task.Collector] = true
			}
			for _, gap := range run.Status.CoverageGaps {
				if strings.HasPrefix(gap, task.Control+":") {
					failed[task.Collector] = true
				}
			}
		}
		for _, task := range run.Spec.Tasks {
			if run.Status.Results[task.ID].Phase == "Completed" {
				audit.Status.FindingCount++
			}
			if !failed[task.Collector] {
				audit.Status.CollectorTimes[task.Collector] = *run.Status.EndTime
			}
		}
		audit.Status.ActiveRun = ""
		if err := r.Status().Update(ctx, &audit); err != nil {
			return ctrl.Result{}, err
		}
	}
	if audit.Spec.Frequency != "framework" {
		var end *time.Time
		if audit.Status.EndTime != nil {
			end = &audit.Status.EndTime.Time
		}
		action, after := planAudit(audit.Status.Phase, end, audit.Spec.Frequency, now)
		if action == actionDone {
			return ctrl.Result{}, nil
		}
		if action == actionWaitRequeue {
			return ctrl.Result{RequeueAfter: after}, nil
		}
	}
	if audit.Spec.Frequency == "framework" && audit.Status.EndTime != nil && len(audit.Status.CoverageGaps) > 0 {
		if wait := audit.Status.EndTime.Time.Add(time.Hour).Sub(now); wait > 0 {
			return ctrl.Result{RequeueAfter: wait}, nil
		}
	}
	var profile api.ComplianceProfile
	if err := r.Get(ctx, client.ObjectKey{Name: audit.Spec.Profile, Namespace: audit.Namespace}, &profile); err != nil {
		return ctrl.Result{}, err
	}
	fw, err := framework.LoadSelected(r.FrameworkDir, profile.Spec.Framework, profile.Spec.Version, profile.Spec.BundleDigest, profile.Spec.IncludeControls)
	if err != nil {
		return ctrl.Result{}, err
	}
	plan := auditplan.Build(fw, &profile, &audit, now)
	if audit.Spec.Frequency == "framework" && len(plan.Tasks) == 0 && len(plan.Gaps) == 0 {
		return ctrl.Result{RequeueAfter: plan.Next}, nil
	}
	if audit.Status.Phase != "Running" || audit.Status.StartTime == nil {
		start := metav1.NewTime(now)
		audit.Status.StartTime = &start
	}
	audit.Status.Phase = "Running"
	audit.Status.EndTime = nil
	audit.Status.EvidenceURI = ""
	audit.Status.ManifestHash = ""
	if err = r.Status().Update(ctx, &audit); err != nil {
		return ctrl.Result{}, err
	}
	name := "audit-" + evidence.Hash([]byte(string(audit.UID) + "/" + audit.Status.StartTime.Format(time.RFC3339Nano)))[:32]
	run := api.AuditRun{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: audit.Namespace, Labels: map[string]string{"mantl.io/audit-uid": string(audit.UID)}}, Spec: api.AuditRunSpec{
		ClusterID: r.ClusterID, CollectorImage: r.CollectorImage, CollectorServiceAccount: r.CollectorServiceAccount, EvidenceBucket: r.Bucket, EvidenceKMSKey: r.KMSKey, Audit: audit.Name, AuditUID: string(audit.UID), Profile: profile.Name, Framework: profile.Spec.Framework, FrameworkVersion: fw.Version, BundleDigest: fw.ContentDigest, ScheduledAt: *audit.Status.StartTime, Tasks: plan.Tasks, CoverageGaps: plan.Gaps,
	}}
	if err = r.Create(ctx, &run); err != nil && !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}
	audit.Status.ActiveRun = name
	if err = r.Status().Update(ctx, &audit); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}
func (r *AuditScheduler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).Named("audit-scheduler").For(&api.ComplianceAudit{}).WithEventFilter(predicate.GenerationChangedPredicate{}).Complete(r)
}
