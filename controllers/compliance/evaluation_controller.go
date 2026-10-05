package compliance

import (
	"context"
	"fmt"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/auditplan"
	"github.com/thomasvincent/mantl/pkg/evaluation"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/yaml"
	"sort"
	"time"
)

type EvaluationReconciler struct {
	client.Client
	FrameworkDir string
	Now          func() time.Time
}

func (r *EvaluationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var profile api.ComplianceProfile
	if err := r.Get(ctx, req.NamespacedName, &profile); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	now := time.Now().UTC()
	if r.Now != nil {
		now = r.Now().UTC()
	}
	fw, err := framework.LoadSelected(r.FrameworkDir, profile.Spec.Framework, profile.Spec.Version, profile.Spec.BundleDigest, profile.Spec.IncludeControls)
	if err != nil {
		var previous api.ControlEvaluationList
		if listErr := r.List(ctx, &previous, client.InNamespace(profile.Namespace), client.MatchingLabels{"mantl.io/profile": string(profile.UID)}); listErr != nil {
			return ctrl.Result{}, listErr
		}
		for i := range previous.Items {
			e := &previous.Items[i]
			e.Status = api.ControlEvaluationStatus{Result: "error", Coverage: "unavailable", Freshness: "unknown", Reason: "profile content is invalid or unavailable"}
			if updateErr := r.Status().Update(ctx, e); updateErr != nil {
				return ctrl.Result{}, updateErr
			}
		}
		return ctrl.Result{}, err
	}
	index, err := buildPolicyIndex([]string{filepath.Join(r.FrameworkDir, "..", "policies", "runtime"), filepath.Join(r.FrameworkDir, profile.Spec.Framework, "policies")})
	if err != nil {
		return ctrl.Result{}, err
	}
	observations, err := r.observations(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	var runs api.AuditRunList
	var exceptions api.ComplianceExceptionList
	if err = r.List(ctx, &runs, client.InNamespace(profile.Namespace)); err != nil {
		return ctrl.Result{}, err
	}
	if err = r.List(ctx, &exceptions, client.InNamespace(profile.Namespace)); err != nil {
		return ctrl.Result{}, err
	}
	sort.Slice(runs.Items, func(i, j int) bool { return runs.Items[i].Spec.ScheduledAt.After(runs.Items[j].Spec.ScheduledAt.Time) })
	namespaces := profile.Spec.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{profile.Namespace}
	}
	plan := auditplan.Build(fw, &profile, &api.ComplianceAudit{}, now)
	for _, control := range fw.Controls {
		if !framework.Selected(control.ID, profile.Spec.IncludeControls) {
			continue
		}
		input := evaluation.Input{Profile: profile.Name, Control: control.ID, Now: now, MaxAge: time.Duration(profile.Spec.MaxEvidenceAgeSeconds) * time.Second, Exceptions: exceptions.Items, Observations: observations}
		if len(control.Mappings) == 0 {
			input.Unsupported = true
		}
		for _, mapping := range control.Mappings {
			if mapping.PolicyRef == nil && mapping.EvidenceCollector == nil {
				input.Unsupported = true
			}
			if mapping.PolicyRef != nil {
				name := mapping.PolicyRef.Template
				if alias, ok := templateAliases[name]; ok {
					name = alias
				}
				file, ok := index[name]
				if !ok {
					input.Unsupported = true
					continue
				}
				data, readErr := os.ReadFile(file)
				if readErr != nil {
					return ctrl.Result{}, readErr
				}
				var policy struct {
					Metadata struct {
						Name string `json:"name"`
					} `json:"metadata"`
					Spec struct {
						Rules []struct {
							Name string `json:"name"`
						} `json:"rules"`
					} `json:"spec"`
				}
				if yaml.Unmarshal(data, &policy) != nil || policy.Metadata.Name == "" || len(policy.Spec.Rules) == 0 {
					input.Unsupported = true
					continue
				}
				for _, ns := range namespaces {
					for _, rule := range policy.Spec.Rules {
						input.Expected = append(input.Expected, policy.Metadata.Name+"/"+ns+"/"+rule.Name)
					}
				}
			}
			if mapping.EvidenceCollector != nil {
				input.EvidenceRequired = true
			}
		}
		for _, gap := range plan.Gaps {
			if len(gap) > len(control.ID) && gap[:len(control.ID)+1] == control.ID+":" {
				input.Unsupported = true
			}
		}
		for _, task := range plan.Tasks {
			if task.Control != control.ID {
				continue
			}
			found := false
			for _, run := range runs.Items {
				if run.Spec.Profile != profile.Name || run.Spec.Framework != profile.Spec.Framework || run.Spec.BundleDigest != fw.ContentDigest || run.Status.EndTime == nil {
					continue
				}
				result, ok := run.Status.Results[task.ID]
				if !ok {
					continue
				}
				found = true
				if result.Phase != "Completed" || result.CapturedAt == nil || run.Status.EvidenceURI == "" {
					input.Unsupported = true
					break
				}
				at := result.CapturedAt.Time
				if input.EvidenceAt == nil || at.Before(*input.EvidenceAt) {
					input.EvidenceAt = &at
				}
				input.EvidenceURI = run.Status.EvidenceURI
				break
			}
			if !found {
				input.Unsupported = true
			}
		}
		name := "control-" + evidence.Hash([]byte(string(profile.UID) + "/" + control.ID))[:32]
		var object api.ControlEvaluation
		getErr := r.Get(ctx, client.ObjectKey{Namespace: profile.Namespace, Name: name}, &object)
		if getErr != nil && !apierrors.IsNotFound(getErr) {
			return ctrl.Result{}, getErr
		}
		if apierrors.IsNotFound(getErr) {
			object.ObjectMeta = metav1.ObjectMeta{Name: name, Namespace: profile.Namespace, Labels: map[string]string{"mantl.io/profile": string(profile.UID)}}
			if err = ctrl.SetControllerReference(&profile, &object, r.Scheme()); err != nil {
				return ctrl.Result{}, err
			}
		}
		object.Spec = api.ControlEvaluationSpec{Profile: profile.Name, Control: control.ID, Framework: profile.Spec.Framework, BundleDigest: fw.ContentDigest}
		if apierrors.IsNotFound(getErr) {
			err = r.Create(ctx, &object)
		} else {
			err = r.Update(ctx, &object)
		}
		if err != nil {
			return ctrl.Result{}, err
		}
		object.Status = evaluation.Assess(input)
		if err = r.Status().Update(ctx, &object); err != nil {
			return ctrl.Result{}, err
		}
	}

	var previous api.ControlEvaluationList
	if err = r.List(ctx, &previous, client.InNamespace(profile.Namespace), client.MatchingLabels{"mantl.io/profile": string(profile.UID)}); err != nil {
		return ctrl.Result{}, err
	}
	selected := map[string]bool{}
	for _, control := range fw.Controls {
		if framework.Selected(control.ID, profile.Spec.IncludeControls) {
			selected[control.ID] = true
		}
	}
	for i := range previous.Items {
		object := &previous.Items[i]
		if !selected[object.Spec.Control] {
			at := metav1.NewTime(now)
			object.Status = api.ControlEvaluationStatus{Result: "not-applicable", Coverage: "excluded", Freshness: "unknown", ObservedAt: &at, Reason: "control excluded from the current profile"}
			if err = r.Status().Update(ctx, object); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}
func (r *EvaluationReconciler) observations(ctx context.Context) ([]evaluation.Observation, error) {
	var observations []evaluation.Observation
	for _, kind := range []string{"PolicyReportList", "ClusterPolicyReportList"} {
		list := &unstructured.UnstructuredList{}
		gvk := policyReportGVK
		gvk.Kind = kind
		list.SetGroupVersionKind(gvk)
		if err := r.List(ctx, list); err != nil {
			return nil, err
		}
		for _, report := range list.Items {
			results, found, err := unstructured.NestedSlice(report.Object, "results")
			if err != nil || !found {
				return nil, fmt.Errorf("invalid evaluation report")
			}
			for _, item := range results {
				result, ok := item.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid evaluation result")
				}
				policy, _, _ := unstructured.NestedString(result, "policy")
				rule, _, _ := unstructured.NestedString(result, "rule")
				state, _, _ := unstructured.NestedString(result, "result")
				seconds, _, _ := unstructured.NestedInt64(result, "timestamp", "seconds")
				resources, _, _ := unstructured.NestedSlice(result, "resources")
				for _, item := range resources {
					resource, ok := item.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("invalid evaluation scope")
					}
					ns, _, _ := unstructured.NestedString(resource, "namespace")
					name, _, _ := unstructured.NestedString(resource, "name")
					kind, _, _ := unstructured.NestedString(resource, "kind")
					at := time.Time{}
					if seconds > 0 {
						at = time.Unix(seconds, 0).UTC()
					}
					observations = append(observations, evaluation.Observation{Key: policy + "/" + ns + "/" + rule, Result: state, At: at, Resource: kind + "/" + ns + "/" + name})
				}
			}
		}
	}
	return observations, nil
}
func (r *EvaluationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).Named("control-evaluations").For(&api.ComplianceProfile{}).WithEventFilter(predicate.GenerationChangedPredicate{}).Complete(r)
}
