package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/collection"
	"github.com/thomasvincent/mantl/pkg/evidence"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sort"
	"time"
)

type AuditRunReconciler struct {
	client.Client
	Store                                 evidence.Store
	Image, Bucket, KMSKey, ServiceAccount string
	ControlNamespace, ClusterID           string
	AllowCluster                          bool
	MaxConcurrent                         int
	Now                                   func() time.Time
}

func (r *AuditRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var run api.AuditRun
	if err := r.Get(ctx, req.NamespacedName, &run); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if r.ControlNamespace == "" || run.Namespace != r.ControlNamespace || r.ClusterID == "" || run.Spec.ClusterID != r.ClusterID {
		return ctrl.Result{}, fmt.Errorf("audit run is outside the trusted control namespace or cluster")
	}
	if run.Status.EndTime != nil {
		return ctrl.Result{}, nil
	}
	if r.Store == nil || r.Image == "" || r.Bucket == "" || r.ServiceAccount == "" {
		return ctrl.Result{}, fmt.Errorf("isolated collectors require image, scoped identity and immutable storage")
	}
	if len(run.Spec.Tasks) > 1000 {
		return ctrl.Result{}, fmt.Errorf("audit run exceeds task limit")
	}
	now := time.Now().UTC()
	if r.Now != nil {
		now = r.Now().UTC()
	}
	if run.Status.Results == nil {
		run.Status.Results = map[string]api.TaskResult{}
	}
	image, account, bucket, key := r.Image, r.ServiceAccount, r.Bucket, r.KMSKey
	if run.Spec.CollectorImage != "" {
		if run.Spec.CollectorImage != r.Image || run.Spec.CollectorServiceAccount != r.ServiceAccount || run.Spec.EvidenceBucket != r.Bucket || run.Spec.EvidenceKMSKey != r.KMSKey {
			return ctrl.Result{}, fmt.Errorf("audit run identities differ from approved operator configuration")
		}
		image, account, bucket, key = run.Spec.CollectorImage, run.Spec.CollectorServiceAccount, run.Spec.EvidenceBucket, run.Spec.EvidenceKMSKey
	}
	pending := false
	active := 0
	limit := r.MaxConcurrent
	if limit < 1 {
		limit = 4
	}
	var existingJobs batch.JobList
	if err := r.List(ctx, &existingJobs, client.MatchingLabels{collection.RunLabel: string(run.UID)}); err != nil {
		return ctrl.Result{}, err
	}
	for _, job := range existingJobs.Items {
		terminal := job.Status.Succeeded > 0
		for _, condition := range job.Status.Conditions {
			terminal = terminal || condition.Type == batch.JobFailed && condition.Status == core.ConditionTrue
		}
		if !terminal {
			active++
		}
	}
	for _, task := range run.Spec.Tasks {
		if existing := run.Status.Results[task.ID]; existing.Phase == "Completed" || existing.Phase == "Failed" {
			continue
		}
		if task.Namespace == "" && !r.AllowCluster {
			run.Status.Results[task.ID] = api.TaskResult{Phase: "Failed"}
			continue
		}
		expected := collection.Job(&run, task, image, bucket, key, account)
		var job batch.Job
		err := r.Get(ctx, client.ObjectKeyFromObject(expected), &job)
		if apierrors.IsNotFound(err) {
			if active >= limit {
				pending = true
				continue
			}
			if err = r.Create(ctx, expected); err != nil {
				if apierrors.IsForbidden(err) || apierrors.IsInvalid(err) {
					run.Status.Results[task.ID] = api.TaskResult{Phase: "Failed"}
					continue
				}
				return ctrl.Result{}, err
			}
			active++
			pending = true
			continue
		}
		if err != nil {
			return ctrl.Result{}, err
		}
		if job.Labels[collection.RunLabel] != string(run.UID) || job.Labels[collection.TaskLabel] != task.ID || !sameCollector(&job, expected) {
			return ctrl.Result{}, fmt.Errorf("collector job does not match immutable run scope")
		}
		if job.Status.Succeeded == 0 {
			failed := false
			for _, condition := range job.Status.Conditions {
				if condition.Type == batch.JobFailed && condition.Status == core.ConditionTrue {
					failed = true
				}
			}
			if failed {
				run.Status.Results[task.ID] = api.TaskResult{Phase: "Failed"}
			} else {
				pending = true
			}
			continue
		}
		var pods core.PodList
		if err = r.List(ctx, &pods, client.InNamespace(job.Namespace), client.MatchingLabels{collection.RunLabel: string(run.UID), collection.TaskLabel: task.ID}); err != nil {
			return ctrl.Result{}, err
		}
		var receipt *collection.Receipt
		for _, pod := range pods.Items {
			owned := false
			for _, owner := range pod.OwnerReferences {
				if owner.Kind == "Job" && owner.UID == job.UID {
					owned = true
				}
			}
			if !owned {
				continue
			}
			for _, status := range pod.Status.ContainerStatuses {
				if status.Name != "collector" || status.State.Terminated == nil || status.State.Terminated.ExitCode != 0 {
					continue
				}
				var value collection.Receipt
				if err = json.Unmarshal([]byte(status.State.Terminated.Message), &value); err != nil {
					return ctrl.Result{}, fmt.Errorf("invalid collector receipt")
				}
				receipt = &value
			}
		}
		if receipt == nil {
			return ctrl.Result{}, fmt.Errorf("completed collector has no verified receipt")
		}
		ref := receipt.Ref
		u, err := url.Parse(ref.URI)
		if err != nil || u.Scheme != "s3" || u.Host != bucket || u.RawQuery != "" || u.Path != "/"+collection.Prefix(run.Spec.ClusterID, task.Namespace, string(run.UID), task.ID)+"/"+ref.Hash+".json" || receipt.CapturedAt.Before(run.Spec.ScheduledAt.Time) || receipt.CapturedAt.After(now.Add(time.Minute)) {
			return ctrl.Result{}, fmt.Errorf("collector receipt scope or time mismatch")
		}
		if ref.RetainUntil == nil || ref.RetainUntil.Before(receipt.CapturedAt.AddDate(0, 0, evidence.MinimumRetentionDays)) {
			return ctrl.Result{}, fmt.Errorf("collector receipt retention too short")
		}
		readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		data, err := r.Store.Get(readCtx, ref)
		cancel()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("verify collector evidence: %w", err)
		}
		var scope struct {
			Resource  string `json:"resource"`
			Namespace string `json:"namespace"`
		}
		if json.Unmarshal(data, &scope) != nil || scope.Resource != task.Resource || scope.Namespace != task.Namespace || evidence.Hash(data) != ref.Hash {
			return ctrl.Result{}, fmt.Errorf("collector evidence does not match task")
		}
		captured := metav1.NewTime(receipt.CapturedAt)
		result := api.TaskResult{Phase: "Completed", URI: ref.URI, Version: ref.Version, SHA256: ref.Hash, CapturedAt: &captured}
		if ref.RetainUntil != nil {
			retained := metav1.NewTime(*ref.RetainUntil)
			result.RetainUntil = &retained
		}
		run.Status.Results[task.ID] = result
	}
	run.Status.Phase = "Running"
	if err := r.Status().Update(ctx, &run); err != nil {
		return ctrl.Result{}, err
	}
	if pending {
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}
	manifest := evidence.Manifest{ClusterID: run.Spec.ClusterID, SchemaVersion: 1, Audit: run.Namespace + "/" + run.Spec.Audit, Run: string(run.UID), Framework: run.Spec.Framework, FrameworkVersion: run.Spec.FrameworkVersion, BundleDigest: run.Spec.BundleDigest, CapturedAt: run.Spec.ScheduledAt.Time, Objects: []evidence.ObjectRef{}, Gaps: append([]string{}, run.Spec.CoverageGaps...)}
	for _, task := range run.Spec.Tasks {
		result := run.Status.Results[task.ID]
		if result.Phase != "Completed" {
			manifest.Gaps = append(manifest.Gaps, task.Control+": collector failed for "+task.Namespace+"/"+task.Resource)
			continue
		}
		ref := evidence.ObjectRef{URI: result.URI, Hash: result.SHA256, Version: result.Version, Control: task.Control, Resource: task.Namespace + "/" + task.Resource, CapturedAt: &result.CapturedAt.Time}
		if result.RetainUntil != nil {
			ref.RetainUntil = &result.RetainUntil.Time
		}
		manifest.Objects = append(manifest.Objects, ref)
	}
	sort.Strings(manifest.Gaps)
	payload, err := json.Marshal(manifest)
	if err != nil {
		return ctrl.Result{}, err
	}
	putCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	ref, err := r.Store.Put(putCtx, "audits/runs/"+run.Spec.ClusterID+"/"+string(run.UID)+"/manifest", payload, now)
	cancel()
	if err != nil {
		return ctrl.Result{}, err
	}
	run.Status.EvidenceURI = ref.URI + "?versionId=" + url.QueryEscape(ref.Version)
	run.Status.ManifestHash = ref.Hash
	run.Status.CoverageGaps = manifest.Gaps
	switch {
	case len(manifest.Objects) == 0 && len(manifest.Gaps) == 0:
		run.Status.Phase = "NoEvidence"
	case len(manifest.Objects) == 0:
		run.Status.Phase = "Failed"
	case len(manifest.Gaps) > 0:
		run.Status.Phase = "PartiallyCompleted"
	default:
		run.Status.Phase = "Completed"
	}
	end := metav1.NewTime(now)
	run.Status.EndTime = &end
	return ctrl.Result{}, r.Status().Update(ctx, &run)
}
func sameCollector(actual, expected *batch.Job) bool {
	a, e := actual.Spec.Template.Spec, expected.Spec.Template.Spec
	if len(a.Containers) != 1 || len(e.Containers) != 1 || len(a.InitContainers) != 0 || len(a.Volumes) != 0 {
		return false
	}
	ac, ec := a.Containers[0], e.Containers[0]
	return a.ServiceAccountName == e.ServiceAccountName && a.RestartPolicy == e.RestartPolicy && reflect.DeepEqual(a.SecurityContext, e.SecurityContext) && reflect.DeepEqual(a.AutomountServiceAccountToken, e.AutomountServiceAccountToken) && ac.Name == ec.Name && ac.Image == ec.Image && reflect.DeepEqual(ac.Command, ec.Command) && reflect.DeepEqual(ac.Args, ec.Args) && len(ac.Env) == 0 && len(ac.EnvFrom) == 0 && len(ac.VolumeMounts) == 0 && reflect.DeepEqual(ac.SecurityContext, ec.SecurityContext) && reflect.DeepEqual(actual.Spec.ActiveDeadlineSeconds, expected.Spec.ActiveDeadlineSeconds)
}
func (r *AuditRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&api.AuditRun{}).WithEventFilter(predicate.GenerationChangedPredicate{}).Complete(r)
}
