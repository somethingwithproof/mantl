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
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

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
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	var clusterNamespace core.Namespace
	if err = mgr.GetAPIReader().Get(context.Background(), client.ObjectKey{Name: "kube-system"}, &clusterNamespace); err != nil {
		setupLog.Error(err, "read immutable cluster identity")
		os.Exit(1)
	}
	actualClusterID := string(clusterNamespace.UID)
	// Register ComplianceProfile controller
	if err = (&compliance.ComplianceProfileReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		FrameworkDir: frameworkDir,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ComplianceProfile")
		os.Exit(1)
	}

	dynamicClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create evidence reader")
		os.Exit(1)
	}
	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		setupLog.Error(err, "unable to configure evidence storage")
		os.Exit(1)
	}
	if evidenceBucket == "" {
		setupLog.Error(nil, "--evidence-bucket is required")
		os.Exit(1)
	}
	evidenceStore := &evidence.S3Store{Client: s3.NewFromConfig(awsConfig), Bucket: evidenceBucket, KMSKey: kmsKey}

	// Register Finding controller
	if err = (&compliance.FindingReconciler{
		Store: evidenceStore, ClusterID: actualClusterID, RequireReports: requireReports,
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		FrameworkDir: frameworkDir,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Finding")
		os.Exit(1)
	}

	// Register the selected audit execution adapter.
	if inlineAudits {
		if err = (&compliance.ComplianceAuditReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			EvidenceBucket: evidenceBucket,
			Reader:         &evidence.KubernetesReader{Client: dynamicClient, AllowCluster: allowCluster},
			Store:          evidenceStore,
			FrameworkDir:   frameworkDir,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "ComplianceAudit")
			os.Exit(1)
		}

	} else {
		if collectorImage == "" {
			setupLog.Error(nil, "--collector-image is required for isolated collection")
			os.Exit(1)
		}
		if err = (&compliance.AuditScheduler{Client: mgr.GetClient(), FrameworkDir: frameworkDir, ControlNamespace: controlNamespace, CollectorImage: collectorImage, CollectorServiceAccount: collectorAccount, ClusterID: actualClusterID, Bucket: evidenceBucket, KMSKey: kmsKey}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to register audit scheduler")
			os.Exit(1)
		}
		if err = (&compliance.AuditRunReconciler{Client: mgr.GetClient(), Store: evidenceStore, Image: collectorImage, Bucket: evidenceBucket, KMSKey: kmsKey, ServiceAccount: collectorAccount, AllowCluster: allowCluster, ControlNamespace: controlNamespace, ClusterID: actualClusterID}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to register audit run controller")
			os.Exit(1)
		}
	}

	if err = (&compliance.EvaluationReconciler{Client: mgr.GetClient(), FrameworkDir: frameworkDir}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to register control evaluations")
		os.Exit(1)
	}
	if fleetURL != "" {
		remote, clientErr := fleet.NewClient(fleetURL, fleetCert, fleetKey, fleetCA)
		if clientErr != nil {
			setupLog.Error(clientErr, "configure fleet transport")
			os.Exit(1)
		}
		if err = mgr.Add(&compliance.FleetPublisher{Client: mgr.GetClient(), Reader: mgr.GetAPIReader(), Remote: remote, Store: evidenceStore, ClusterID: clusterID}); err != nil {
			setupLog.Error(err, "register fleet publisher")
			os.Exit(1)
		}
	}
	if err := mgr.Add(compliance.NewComplianceMonitor(mgr.GetClient())); err != nil {
		setupLog.Error(err, "unable to register compliance metrics")
		os.Exit(1)
	}
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
