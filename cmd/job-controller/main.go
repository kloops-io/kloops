package main

import (
	"flag"
	"os"

	build_v1alpha1 "github.com/kloops-io/kloops/apis/build/v1alpha1"
	"github.com/kloops-io/kloops/pkg/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = build_v1alpha1.AddToScheme(scheme)
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)).WithName("job-controller"))

	logger := ctrl.Log.WithName("main")
	namespace := flag.String("namespace", "default", "The namespace to watch.")
	metricsAddr := flag.String("metrics-addr", ":8080", "The address the metric endpoint binds to.")
	enableLeaderElection := flag.Bool("enable-leader-election", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")

	flag.Parse()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Namespace:          *namespace,
		MetricsBindAddress: *metricsAddr,
		Port:               9443,
		LeaderElection:     *enableLeaderElection,
		LeaderElectionID:   "1e9e8f6c.kloops.io",
	})
	if err != nil {
		logger.Error(err, "unable to create manager")
		os.Exit(1)
	}

	reconciler := controllers.NewJobReconciler(mgr.GetClient(), mgr.GetScheme(), ctrl.Log.WithName("reconciler"))
	if err = reconciler.SetupWithManager(mgr); err != nil {
		logger.Error(err, "failed to create controller")
	}

	stopCh := ctrl.SetupSignalHandler()

	logger.Info("starting manager")
	if err := mgr.Start(stopCh); err != nil {
		logger.Error(err, "failed to start manager")
		os.Exit(1)
	}
}
