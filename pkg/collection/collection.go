// SPDX-License-Identifier: Apache-2.0

// Package collection defines isolated, bounded evidence execution and receipts.
package collection

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/evidence"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

const RunLabel = "mantl.io/audit-run"
const TaskLabel = "mantl.io/collector-task"

type Receipt struct {
	Ref        evidence.ObjectRef `json:"object"`
	CapturedAt time.Time          `json:"capturedAt"`
}

func Prefix(cluster, namespace, uid, id string) string {
	if namespace == "" {
		namespace = "_cluster"
	}
	return "audits/collectors/" + cluster + "/" + namespace + "/" + uid + "/" + id
}
func Execute(ctx context.Context, reader evidence.Reader, store evidence.Store, task api.CollectorTask, uid, cluster string, now time.Time) (Receipt, error) {
	var result Receipt
	if cluster == "" || strings.ContainsAny(cluster, "/\\") || uid == "" || strings.ContainsAny(uid, "/\\") || len(task.ID) != 32 || len(validation.IsDNS1123Label(task.ID)) != 0 {
		return result, fmt.Errorf("invalid collection identity")
	}
	snap, err := reader.Capture(ctx, task.Resource, task.Namespace, now)
	if err != nil {
		return result, fmt.Errorf("capture task: %w", err)
	}
	ref, err := store.Put(ctx, Prefix(cluster, task.Namespace, uid, task.ID), []byte(snap.Data), now)
	if err != nil {
		return result, fmt.Errorf("store task: %w", err)
	}
	ref.Control = task.Control
	ref.Resource = task.Namespace + "/" + task.Resource
	return Receipt{ref, snap.CapturedAt}, nil
}
func WriteReceipt(path string, receipt Receipt) error {
	b, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	if len(b) > 4096 {
		return fmt.Errorf("collector receipt too large")
	}
	return os.WriteFile(path, b, 0600)
}
func Job(run *api.AuditRun, task api.CollectorTask, image, bucket, key, account string) *batch.Job {
	ns := task.Namespace
	if ns == "" {
		ns = run.Namespace
	}
	labels := map[string]string{RunLabel: string(run.UID), TaskLabel: task.ID, "app.kubernetes.io/name": "mantl-evidence-collector", "app.kubernetes.io/part-of": "mantl", "app.kubernetes.io/version": "v1", "app.kubernetes.io/managed-by": "mantl-operator"}
	ttl := int32(86400)
	backoff := int32(2)
	deadline := int64(120)
	user := int64(65532)
	yes := true
	no := false
	args := []string{"--resource", task.Resource, "--namespace", task.Namespace, "--control", task.Control, "--task-id", task.ID, "--run-uid", string(run.UID), "--cluster-id", run.Spec.ClusterID, "--bucket", bucket, "--kms-key", key}
	if task.Namespace == "" {
		args = append(args, "--allow-cluster")
	}
	return &batch.Job{ObjectMeta: metav1.ObjectMeta{Name: "mantl-" + task.ID + "-" + evidence.Hash([]byte(run.UID))[:12], Namespace: ns, Labels: labels}, Spec: batch.JobSpec{
		BackoffLimit: &backoff, ActiveDeadlineSeconds: &deadline, TTLSecondsAfterFinished: &ttl,
		Template: core.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: labels}, Spec: core.PodSpec{
			RestartPolicy: core.RestartPolicyNever, ServiceAccountName: account, AutomountServiceAccountToken: &yes,
			SecurityContext: &core.PodSecurityContext{RunAsNonRoot: &yes, RunAsUser: &user, SeccompProfile: &core.SeccompProfile{Type: core.SeccompProfileTypeRuntimeDefault}},
			Containers: []core.Container{{Name: "collector", Image: image, Command: []string{"/collector"}, Args: args, TerminationMessagePath: "/dev/termination-log", TerminationMessagePolicy: core.TerminationMessageReadFile,
				SecurityContext: &core.SecurityContext{AllowPrivilegeEscalation: &no, ReadOnlyRootFilesystem: &yes, Capabilities: &core.Capabilities{Drop: []core.Capability{"ALL"}}},
				Resources:       core.ResourceRequirements{Requests: core.ResourceList{core.ResourceCPU: resourceQuantity("50m"), core.ResourceMemory: resourceQuantity("64Mi")}, Limits: core.ResourceList{core.ResourceCPU: resourceQuantity("500m"), core.ResourceMemory: resourceQuantity("256Mi")}},
			}},
		}},
	}}
}
