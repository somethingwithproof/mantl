// SPDX-License-Identifier: Apache-2.0

package compliance

import (
	"context"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/auditplan"
	"github.com/thomasvincent/mantl/pkg/evaluation"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
		if invalidateErr := r.invalidateEvaluations(ctx, &profile); invalidateErr != nil {
			return ctrl.Result{}, invalidateErr
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
	runs, exceptions, err := r.evaluationMetadata(ctx, profile.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	sort.Slice(runs.Items, func(i, j int) bool { return runs.Items[i].Spec.ScheduledAt.After(runs.Items[j].Spec.ScheduledAt.Time) })
	namespaces := profile.Spec.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{profile.Namespace}
	}
	plan := auditplan.Build(fw, &profile, &api.ComplianceAudit{}, now)
	data := controlEvaluationInput{profile: profile, framework: fw, namespaces: namespaces, index: index, observations: observations, exceptions: exceptions.Items, plan: plan, runs: runs.Items, now: now}
	for _, control := range fw.Controls {
		if err := r.reconcileSelectedControl(ctx, control, &data); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.excludeUnselectedEvaluations(ctx, &profile, fw, now); err != nil {
		return ctrl.Result{}, err
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
			parsed, err := evaluation.PolicyReportObservations(report)
			if err != nil {
				return nil, err
			}
			observations = append(observations, parsed...)
		}

	}
	return observations, nil
}
func (r *EvaluationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).Named("control-evaluations").For(&api.ComplianceProfile{}).WithEventFilter(predicate.GenerationChangedPredicate{}).Complete(r)
}

func (r *EvaluationReconciler) persistControlEvaluation(ctx context.Context, profile *api.ComplianceProfile, controlID, bundleDigest string, input evaluation.Input) error {
	var err error
	name := "control-" + evidence.Hash([]byte(string(profile.UID) + "/" + controlID))[:32]
	var object api.ControlEvaluation
	getErr := r.Get(ctx, client.ObjectKey{Namespace: profile.Namespace, Name: name}, &object)
	if getErr != nil && !apierrors.IsNotFound(getErr) {
		return getErr
	}
	if apierrors.IsNotFound(getErr) {
		object.ObjectMeta = metav1.ObjectMeta{Name: name, Namespace: profile.Namespace, Labels: map[string]string{profileLabel: string(profile.UID)}}
		if err := ctrl.SetControllerReference(profile, &object, r.Scheme()); err != nil {
			return err
		}
	}
	object.Spec = api.ControlEvaluationSpec{Profile: profile.Name, Control: controlID, Framework: profile.Spec.Framework, BundleDigest: bundleDigest}
	if apierrors.IsNotFound(getErr) {
		err = r.Create(ctx, &object)
	} else {
		err = r.Update(ctx, &object)
	}
	if err != nil {
		return err
	}
	object.Status = evaluation.Assess(input)
	if err := r.Status().Update(ctx, &object); err != nil {
		return err
	}
	return nil
}

func (r *EvaluationReconciler) invalidateEvaluations(ctx context.Context, profile *api.ComplianceProfile) error {
	var previous api.ControlEvaluationList
	if listErr := r.List(ctx, &previous, client.InNamespace(profile.Namespace), client.MatchingLabels{profileLabel: string(profile.UID)}); listErr != nil {
		return listErr
	}
	for i := range previous.Items {
		e := &previous.Items[i]
		e.Status = api.ControlEvaluationStatus{Result: "error", Coverage: "unavailable", Freshness: "unknown", Reason: "profile content is invalid or unavailable"}
		if updateErr := r.Status().Update(ctx, e); updateErr != nil {
			return updateErr
		}
	}
	return nil
}

func (r *EvaluationReconciler) excludeUnselectedEvaluations(ctx context.Context, profile *api.ComplianceProfile, fw *framework.Framework, now time.Time) error {
	var previous api.ControlEvaluationList
	if err := r.List(ctx, &previous, client.InNamespace(profile.Namespace), client.MatchingLabels{profileLabel: string(profile.UID)}); err != nil {
		return err
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
			if err := r.Status().Update(ctx, object); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *EvaluationReconciler) evaluationMetadata(ctx context.Context, namespace string) (api.AuditRunList, api.ComplianceExceptionList, error) {
	var runs api.AuditRunList
	var exceptions api.ComplianceExceptionList
	if err := r.List(ctx, &runs, client.InNamespace(namespace)); err != nil {
		return runs, exceptions, err
	}
	if err := r.List(ctx, &exceptions, client.InNamespace(namespace)); err != nil {
		return runs, exceptions, err
	}
	return runs, exceptions, nil
}

type controlEvaluationInput struct {
	profile      api.ComplianceProfile
	framework    *framework.Framework
	namespaces   []string
	index        map[string]string
	observations []evaluation.Observation
	exceptions   []api.ComplianceException
	plan         auditplan.Plan
	runs         []api.AuditRun
	now          time.Time
}

func (r *EvaluationReconciler) reconcileSelectedControl(ctx context.Context, control framework.Control, data *controlEvaluationInput) error {

	if !framework.Selected(control.ID, data.profile.Spec.IncludeControls) {
		return nil
	}
	input := evaluation.Input{Profile: data.profile.Name, Control: control.ID, Now: data.now, MaxAge: time.Duration(data.profile.Spec.MaxEvidenceAgeSeconds) * time.Second, Exceptions: data.exceptions, Observations: data.observations}
	expected, unsupported, evidenceRequired, mappingErr := evaluation.MappingExpectations(control, data.namespaces, data.index, templateAliases)
	if mappingErr != nil {
		return mappingErr
	}
	input.Expected, input.Unsupported, input.EvidenceRequired = expected, unsupported, evidenceRequired
	evaluation.AttachCollectionEvidence(&input, data.plan, data.runs, data.profile.Name, data.profile.Spec.Framework, data.framework.ContentDigest)
	if err := r.persistControlEvaluation(ctx, &data.profile, control.ID, data.framework.ContentDigest, input); err != nil {
		return err
	}

	return nil
}
