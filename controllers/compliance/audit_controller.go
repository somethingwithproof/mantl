package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/auditplan"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/url"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sort"
	"time"
)

type ComplianceAuditReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	EvidenceBucket string // Legacy construction compatibility; Store owns configuration.
	FrameworkDir   string
	Reader         evidence.Reader
	Store          evidence.Store
	Now            func() time.Time
}

// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceaudits,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=compliance.mantl.io,resources=complianceaudits/status,verbs=get;update;patch
func (r *ComplianceAuditReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var audit api.ComplianceAudit
	if err := r.Get(ctx, req.NamespacedName, &audit); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !validAuditFrequency(audit.Spec.Frequency) {
		return ctrl.Result{}, fmt.Errorf("unsupported audit frequency %q", audit.Spec.Frequency)
	}
	now := time.Now().UTC()
	if r.Now != nil {
		now = r.Now().UTC()
	}
	if result, stop := regularAuditWait(&audit, now); stop {
		return result, nil
	}

	var profile api.ComplianceProfile
	if err := r.Get(ctx, client.ObjectKey{Name: audit.Spec.Profile, Namespace: audit.Namespace}, &profile); err != nil {
		return ctrl.Result{}, fmt.Errorf("load audit profile: %w", err)
	}
	fw, err := framework.LoadSelected(r.FrameworkDir, profile.Spec.Framework, profile.Spec.Version, profile.Spec.BundleDigest, profile.Spec.IncludeControls)
	if err != nil {
		return ctrl.Result{}, err
	}
	if r.Reader == nil || r.Store == nil {
		return ctrl.Result{}, fmt.Errorf("evidence reader and immutable store must be configured")
	}

	plan := auditplan.BuildInline(fw, &profile, &audit, now)
	next, tasks, gaps, due, incomplete := plan.Next, plan.Tasks, plan.Gaps, plan.Due, plan.Incomplete
	if len(due) == 0 && len(gaps) == 0 && audit.Spec.Frequency == "framework" {
		return ctrl.Result{RequeueAfter: next}, nil
	}
	if audit.Status.Phase != "Running" || audit.Status.StartTime == nil {
		start := metav1.NewTime(now)
		audit.Status.StartTime = &start
	}
	audit.Status.Phase = "Running"
	audit.Status.EndTime = nil
	audit.Status.EvidenceURI = ""
	audit.Status.ManifestHash = ""
	if err := r.Status().Update(ctx, &audit); err != nil {
		return ctrl.Result{}, fmt.Errorf("persist audit start: %w", err)
	}
	run := audit.Status.StartTime.Time.UTC().Format(time.RFC3339Nano)
	prefix := fmt.Sprintf("audits/%s/%s/%s/%s", audit.Namespace, audit.Name, audit.UID, run)
	manifest := evidence.Manifest{SchemaVersion: 1, Audit: audit.Namespace + "/" + audit.Name, Run: run, Framework: profile.Spec.Framework, FrameworkVersion: fw.Version, BundleDigest: fw.ContentDigest, CapturedAt: audit.Status.StartTime.Time, Objects: []evidence.ObjectRef{}}
	failed := incomplete
	gaps = append(gaps, r.captureInlineObjects(ctx, tasks, &manifest, prefix, now, failed)...)
	sort.Strings(gaps)
	manifest.Gaps = gaps
	payload, err := json.Marshal(manifest)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("encode audit manifest: %w", err)
	}
	uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	ref, err := r.Store.Put(uploadCtx, prefix+"/manifest", payload, now)
	cancel()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("persist audit manifest: %w", err)
	}
	return r.completeInlineAudit(ctx, &audit, manifest, ref, tasks, failed, gaps, next)
}

func (r *ComplianceAuditReconciler) completeInlineAudit(ctx context.Context, audit *api.ComplianceAudit, manifest evidence.Manifest, ref evidence.ObjectRef, tasks []auditplan.InlineTask, failed map[string]bool, gaps []string, next time.Duration) (ctrl.Result, error) {
	audit.Status.EvidenceURI = ref.URI + "?versionId=" + url.QueryEscape(ref.Version)
	audit.Status.ManifestHash = ref.Hash
	audit.Status.FindingCount = int32(len(manifest.Objects))
	audit.Status.FailedCount = int32(len(gaps))
	audit.Status.CoverageGaps = gaps
	audit.Status.Phase = auditCompletionPhase(len(manifest.Objects), len(manifest.Gaps))

	finishedAt := time.Now().UTC()
	if r.Now != nil {
		finishedAt = r.Now().UTC()
	}
	finished := metav1.NewTime(finishedAt)
	audit.Status.EndTime = &finished
	if audit.Status.CollectorTimes == nil {
		audit.Status.CollectorTimes = map[string]metav1.Time{}
	}
	// Unsupported collectors have no successful tasks and are deliberately not advanced.
	for _, t := range tasks {
		if !failed[t.Key] {
			audit.Status.CollectorTimes[t.Key] = finished
		}
	}
	if err := r.Status().Update(ctx, audit); err != nil {
		return ctrl.Result{}, fmt.Errorf("persist audit result: %w", err)
	}
	if audit.Spec.Frequency == "framework" {
		if next == 0 || len(gaps) > 0 && next > time.Hour {
			next = time.Hour
		}
		return ctrl.Result{RequeueAfter: next}, nil
	}
	if interval, recurring := requeueInterval(audit.Spec.Frequency); recurring {
		return ctrl.Result{RequeueAfter: interval}, nil
	}
	return ctrl.Result{}, nil
}

func (r *ComplianceAuditReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&api.ComplianceAudit{}).WithEventFilter(predicate.GenerationChangedPredicate{}).Complete(r)
}

func (r *ComplianceAuditReconciler) captureInlineObjects(ctx context.Context, tasks []auditplan.InlineTask, manifest *evidence.Manifest, prefix string, now time.Time, failed map[string]bool) []string {
	var gaps []string
	for _, t := range tasks {
		captureCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		snap, captureErr := r.Reader.Capture(captureCtx, t.Resource, t.Namespace, manifest.CapturedAt)
		cancel()
		if captureErr != nil {
			gaps = append(gaps, t.Control+": capture failed for "+t.Namespace+"/"+t.Resource)
			failed[t.Key] = true
			continue
		}
		uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		ref, uploadErr := r.Store.Put(uploadCtx, prefix+"/"+t.Control+"/"+t.Namespace+"/"+t.Resource, []byte(snap.Data), now)
		cancel()
		if uploadErr != nil {
			gaps = append(gaps, t.Control+": upload failed for "+t.Resource)
			failed[t.Key] = true
			continue
		}
		ref.Control = t.Control
		ref.Resource = t.Namespace + "/" + t.Resource
		manifest.Objects = append(manifest.Objects, ref)
	}
	return gaps
}
