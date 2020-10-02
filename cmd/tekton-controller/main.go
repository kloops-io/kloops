package main

import (
	"context"
	"flag"
	"os"

	"github.com/go-logr/logr"
	build_v1alpha1 "github.com/kloops-io/kloops/apis/build/v1alpha1"
	tekton_v1alpha1 "github.com/kloops-io/kloops/apis/tekton/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = build_v1alpha1.AddToScheme(scheme)
	_ = tekton_v1alpha1.AddToScheme(scheme)
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)).WithName("tekton-controller"))

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

	reconciler := NewReconciler(mgr.GetClient(), mgr.GetScheme(), ctrl.Log.WithName("reconciler"))
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

type Reconciler struct {
	client client.Client
	scheme *runtime.Scheme
	logger logr.Logger
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger) *Reconciler {
	return &Reconciler{
		client: client,
		scheme: scheme,
		logger: logger,
	}
}

// SetupWithManager sets up the reconcilier with it's manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tekton_v1alpha1.Job{}).
		WithEventFilter(predicate.ResourceVersionChangedPredicate{}).
		Complete(r)
}

// Reconcile represents an iteration of the reconciliation loop
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, logger := context.Background(), r.logger.WithValues("request", req)
	logger.Info("reconcile job")
	var job tekton_v1alpha1.Job
	if err := r.client.Get(ctx, req.NamespacedName, &job); err != nil {
		logger.Info("failed to get job")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	owner := metav1.GetControllerOf(&job)
	var klJob build_v1alpha1.Job
	klReq := types.NamespacedName{
		Name:      owner.Name,
		Namespace: job.Namespace,
	}
	if err := r.client.Get(ctx, klReq, &klJob); err != nil {
		logger.Info("failed to get kl job")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if klJob.Status.StartTime.IsZero() {
		klJob.Status.StartTime = metav1.Now()
	}
	klJob.Status.State = build_v1alpha1.SuccessState
	err := r.client.Status().Update(ctx, &klJob)
	if err != nil {
		logger.Error(err, "failed to report status")
	}
	return ctrl.Result{}, nil
}
