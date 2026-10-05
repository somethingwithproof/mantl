package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
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
	now := time.Now().UTC()
	if r.Now != nil {
		now = r.Now().UTC()
	}
	var end *time.Time
	if audit.Status.EndTime != nil {
		end = &audit.Status.EndTime.Time
	}
	if audit.Spec.Frequency != "framework" {
		action, after := planAudit(audit.Status.Phase, end, audit.Spec.Frequency, now)
		if action == actionDone {
			return ctrl.Result{}, nil
		}
		if action == actionWaitRequeue {
			return ctrl.Result{RequeueAfter: after}, nil
		}
	}
	var profile api.ComplianceProfile
	if err := r.Get(ctx, client.ObjectKey{Name: audit.Spec.Profile, Namespace: audit.Namespace}, &profile); err != nil {
		return ctrl.Result{}, fmt.Errorf("load audit profile: %w", err)
	}
	fw, err := framework.Load(r.FrameworkDir, profile.Spec.Framework, profile.Spec.Version)
	if err != nil {
		return ctrl.Result{}, err
	}
	if r.Reader == nil || r.Store == nil {
		return ctrl.Result{}, fmt.Errorf("evidence reader and immutable store must be configured")
	}
	if audit.Spec.Frequency != "" && audit.Spec.Frequency != "manual" && audit.Spec.Frequency != "daily" && audit.Spec.Frequency != "weekly" && audit.Spec.Frequency != "framework" {
		return ctrl.Result{}, fmt.Errorf("unsupported audit frequency %q", audit.Spec.Frequency)
	}
	next := time.Duration(0)
	type task struct{ key, control, resource, namespace string }
	var tasks []task
	var gaps []string
	due := map[string]bool{}
	incomplete := map[string]bool{}
	namespaces := profile.Spec.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{profile.Namespace}
	}
	for _, control := range fw.Controls {
		if !framework.Selected(control.ID, profile.Spec.IncludeControls) {
			continue
		}
		for i, mapping := range control.Mappings {
			collector := mapping.EvidenceCollector
			if collector == nil {
				continue
			}
			key := fmt.Sprintf("%s/%d", control.ID, i)
			if audit.Spec.Frequency == "framework" {
				last, exists := audit.Status.CollectorTimes[key]
				base := now
				if exists {
					base = last.Time
				}
				when, scheduleErr := framework.Next(collector.Schedule, base)
				if scheduleErr != nil {
					gaps = append(gaps, control.ID+": invalid schedule")
					continue
				}
				if exists && when.After(now) {
					wait := when.Sub(now)
					if next == 0 || wait < next {
						next = wait
					}
					continue
				}
				// First observation runs immediately; missed intervals coalesce into one run.
				nextWhen, _ := framework.Next(collector.Schedule, now)
				wait := nextWhen.Sub(now)
				if next == 0 || wait < next {
					next = wait
				}
			}
			due[key] = true
			if collector.Type != "config-snapshot" || len(collector.Resources) == 0 {
				gaps = append(gaps, control.ID+": unsupported collector "+collector.Type)
				continue
			}
			for _, resource := range collector.Resources {
				typ, typeErr := evidence.Resource(resource)
				if typeErr != nil {
					incomplete[key] = true
					gaps = append(gaps, control.ID+": unsupported resource "+resource)
					continue
				}
				if typ.Cluster {
					tasks = append(tasks, task{key, control.ID, resource, ""})
				} else {
					for _, ns := range namespaces {
						tasks = append(tasks, task{key, control.ID, resource, ns})
					}
				}
			}
		}
	}
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
	manifest := evidence.Manifest{SchemaVersion: 1, Audit: audit.Namespace + "/" + audit.Name, Run: run, Framework: profile.Spec.Framework, FrameworkVersion: fw.Version, CapturedAt: audit.Status.StartTime.Time, Objects: []evidence.ObjectRef{}}
	failed := incomplete
	for _, t := range tasks {
		captureCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		snap, captureErr := r.Reader.Capture(captureCtx, t.resource, t.namespace, audit.Status.StartTime.Time)
		cancel()
		if captureErr != nil {
			gaps = append(gaps, t.control+": capture failed for "+t.namespace+"/"+t.resource)
			failed[t.key] = true
			continue
		}
		uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		ref, uploadErr := r.Store.Put(uploadCtx, prefix+"/"+t.control+"/"+t.namespace+"/"+t.resource, []byte(snap.Data), now)
		cancel()
		if uploadErr != nil {
			gaps = append(gaps, t.control+": upload failed for "+t.resource)
			failed[t.key] = true
			continue
		}
		ref.Control = t.control
		ref.Resource = t.namespace + "/" + t.resource
		manifest.Objects = append(manifest.Objects, ref)
	}
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
	audit.Status.EvidenceURI = ref.URI + "?versionId=" + url.QueryEscape(ref.Version)
	audit.Status.ManifestHash = ref.Hash
	audit.Status.FindingCount = int32(len(manifest.Objects))
	audit.Status.FailedCount = int32(len(gaps))
	audit.Status.CoverageGaps = gaps
	switch {
	case len(manifest.Objects) == 0 && len(gaps) == 0:
		audit.Status.Phase = "NoEvidence"
	case len(manifest.Objects) == 0:
		audit.Status.Phase = "Failed"
	case len(gaps) > 0:
		audit.Status.Phase = "PartiallyCompleted"
	default:
		audit.Status.Phase = "Completed"
	}
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
		if !failed[t.key] {
			audit.Status.CollectorTimes[t.key] = finished
		}
	}
	if err := r.Status().Update(ctx, &audit); err != nil {
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
