package main

import (
	"context"
	"flag"
	"github.com/go-logr/logr"
	"log/slog"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/controllers/compliance"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/fleet"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	// Register Kubernetes client authentication providers for operator credentials.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const controllerSetupFailure = "unable to create controller"

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var frameworkDir string
	var evidenceBucket string
	var kmsKey string
	var allowCluster bool
	var requireReports bool
	var inlineAudits bool
	var collectorImage, collectorAccount, controlNamespace string
	flag.StringVar(&controlNamespace, "control-namespace", "mantl-compliance-system", "Trusted namespace for audit scheduling and run execution")
	var fleetURL, fleetCert, fleetKey, fleetCA, clusterID string
	flag.StringVar(&fleetURL, "fleet-url", "", "Optional enrolled fleet API origin")
	flag.StringVar(&fleetCert, "fleet-certificate", "", "Runtime fleet client certificate path")
	flag.StringVar(&fleetKey, "fleet-key", "", "Runtime fleet key path")
	flag.StringVar(&fleetCA, "fleet-ca", "", "Trusted fleet server CA path")
	flag.StringVar(&clusterID, "fleet-cluster-id", "", "Enrolled kube-system namespace UID")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&frameworkDir, "framework-dir", "/etc/compliance/frameworks", "The directory containing framework mapping files.")
	flag.StringVar(&evidenceBucket, "evidence-bucket", "", "The S3 bucket to store compliance evidence.")
	flag.StringVar(&kmsKey, "evidence-kms-key", "", "Optional evidence KMS key ARN")
	flag.BoolVar(&inlineAudits, "inline-audits", false, "Legacy inline collection compatibility mode")
	flag.StringVar(&collectorImage, "collector-image", "", "Digest-pinned isolated collector image")
	flag.StringVar(&collectorAccount, "collector-service-account", "mantl-evidence-collector", "Pre-provisioned scoped identity in each collection namespace")
	flag.BoolVar(&requireReports, "require-policy-reports", true, "Fail startup until required Kyverno report CRDs are installed")
	flag.BoolVar(&allowCluster, "allow-cluster-evidence", false, "Allow explicit cluster RBAC evidence reads")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()
	ctrl.SetLogger(logr.FromSlogHandler(slog.NewJSONHandler(os.Stderr, nil)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "compliance.mantl.io",
	})
	failStartup(err, "unable to start manager")

	var clusterNamespace core.Namespace
	failStartup(mgr.GetAPIReader().Get(context.Background(), client.ObjectKey{Name: "kube-system"}, &clusterNamespace), "read immutable cluster identity")
	actualClusterID := string(clusterNamespace.UID)
	// Register ComplianceProfile controller
	failStartup((&compliance.ComplianceProfileReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		FrameworkDir: frameworkDir,
	}).SetupWithManager(mgr), controllerSetupFailure, "controller", "ComplianceProfile")

	dynamicClient, err := dynamic.NewForConfig(mgr.GetConfig())
	failStartup(err, "unable to create evidence reader")
	awsConfig, err := config.LoadDefaultConfig(context.Background())
	failStartup(err, "unable to configure evidence storage")
	if evidenceBucket == "" {
		setupLog.Error(nil, "--evidence-bucket is required")
		os.Exit(1)
	}
	evidenceStore := &evidence.S3Store{Client: s3.NewFromConfig(awsConfig), Bucket: evidenceBucket, KMSKey: kmsKey}

	// Register Finding controller
	failStartup((&compliance.FindingReconciler{
		Store: evidenceStore, ClusterID: actualClusterID, RequireReports: requireReports,
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		FrameworkDir: frameworkDir,
	}).SetupWithManager(mgr), controllerSetupFailure, "controller", "Finding")

	// Register the selected audit execution adapter.
	if inlineAudits {
		failStartup((&compliance.ComplianceAuditReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			EvidenceBucket: evidenceBucket,
			Reader:         &evidence.KubernetesReader{Client: dynamicClient, AllowCluster: allowCluster},
			Store:          evidenceStore,
			FrameworkDir:   frameworkDir,
		}).SetupWithManager(mgr), controllerSetupFailure, "controller", "ComplianceAudit")

	} else {
		if collectorImage == "" {
			setupLog.Error(nil, "--collector-image is required for isolated collection")
			os.Exit(1)
		}
		failStartup((&compliance.AuditScheduler{Client: mgr.GetClient(), FrameworkDir: frameworkDir, ControlNamespace: controlNamespace, CollectorImage: collectorImage, CollectorServiceAccount: collectorAccount, ClusterID: actualClusterID, Bucket: evidenceBucket, KMSKey: kmsKey}).SetupWithManager(mgr), "unable to register audit scheduler")
		failStartup((&compliance.AuditRunReconciler{Client: mgr.GetClient(), Store: evidenceStore, Image: collectorImage, Bucket: evidenceBucket, KMSKey: kmsKey, ServiceAccount: collectorAccount, AllowCluster: allowCluster, ControlNamespace: controlNamespace, ClusterID: actualClusterID}).SetupWithManager(mgr), "unable to register audit run controller")
	}

	failStartup((&compliance.EvaluationReconciler{Client: mgr.GetClient(), FrameworkDir: frameworkDir}).SetupWithManager(mgr), "unable to register control evaluations")
	if fleetURL != "" {
		remote, clientErr := fleet.NewClient(fleetURL, fleetCert, fleetKey, fleetCA)
		failStartup(clientErr, "configure fleet transport")
		failStartup(mgr.Add(&compliance.FleetPublisher{Client: mgr.GetClient(), Reader: mgr.GetAPIReader(), Remote: remote, Store: evidenceStore, ClusterID: clusterID}), "register fleet publisher")
	}
	failStartup(mgr.Add(compliance.NewComplianceMonitor(mgr.GetClient())), "unable to register compliance metrics")
	failStartup(mgr.AddHealthzCheck("healthz", healthz.Ping), "unable to set up health check")
	failStartup(mgr.AddReadyzCheck("readyz", healthz.Ping), "unable to set up ready check")

	setupLog.Info("starting manager")
	failStartup(mgr.Start(ctrl.SetupSignalHandler()), "problem running manager")
}

// failStartup terminates before starting reconciliation when a required dependency
// or controller cannot be configured. Every call retains its setup diagnostic.
func failStartup(err error, message string, fields ...any) {
	if err != nil {
		setupLog.Error(err, message, fields...)
		os.Exit(1)
	}
}
