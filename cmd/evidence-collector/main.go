// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/collection"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func run() error {
	var task api.CollectorTask
	var uid, bucket, key, receipt, cluster string
	var allow bool
	flag.StringVar(&task.Resource, "resource", "", "Allowlisted Kubernetes resource")
	flag.StringVar(&task.Namespace, "namespace", "", "Explicit evidence namespace")
	flag.StringVar(&task.Control, "control", "", "Control ID")
	flag.StringVar(&task.ID, "task-id", "", "Durable task identity")
	flag.StringVar(&cluster, "cluster-id", "", "Trusted cluster identity")
	flag.StringVar(&uid, "run-uid", "", "AuditRun UID")
	flag.StringVar(&bucket, "bucket", "", "Evidence bucket")
	flag.StringVar(&key, "kms-key", "", "Optional KMS key identifier")
	flag.StringVar(&receipt, "receipt", "/dev/termination-log", "Metadata-only completion receipt")
	flag.BoolVar(&allow, "allow-cluster", false, "Explicit cluster evidence opt-in")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	kube, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(kube)
	if err != nil {
		return err
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	result, err := collection.Execute(ctx, &evidence.KubernetesReader{Client: client, AllowCluster: allow}, &evidence.S3Store{Client: s3.NewFromConfig(cfg), Bucket: bucket, KMSKey: key}, task, uid, cluster, time.Now().UTC())
	if err != nil {
		return err
	}
	return collection.WriteReceipt(receipt, result)
}
func main() {
	if err := run(); err != nil {
		slog.Error("collector failed", "error", err)
		os.Exit(1)
	}
}
