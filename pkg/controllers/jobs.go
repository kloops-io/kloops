package controllers

import (
	"context"

	"github.com/go-logr/logr"
	build_v1alpha1 "github.com/kloops-io/kloops/apis/build/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client client.Client
	scheme *runtime.Scheme
	logger logr.Logger
}

// NewJobReconciler creates a Job reconciler
func NewJobReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger) *JobReconciler {
	return &JobReconciler{
		client: client,
		scheme: scheme,
		logger: logger,
	}
}

// SetupWithManager sets up the reconcilier with it's manager
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&build_v1alpha1.Job{}).
		WithEventFilter(predicate.ResourceVersionChangedPredicate{}).
		Complete(r)
}

// Reconcile represents an iteration of the reconciliation loop
func (r *JobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	logger := r.logger.WithValues("request", req)

	logger.Info("reconcile job")

	// get lighthouse job
	var job build_v1alpha1.Job
	if err := r.client.Get(ctx, req.NamespacedName, &job); err != nil {
		logger.Info("failed to get job")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if job.Spec.Resource != nil {
		var obj unstructured.Unstructured
		err := json.Unmarshal(job.Spec.Resource.Raw, &obj)
		// codec := serializer.NewCodecFactory(r.scheme)
		// decoder := codec.UniversalDecoder(r.scheme.PrioritizedVersionsAllGroups()...)
		// obj, err := runtime.Decode(decoder, job.Spec.Resource.Raw)
		obj.SetNamespace(job.Namespace)
		if err != nil {
			logger.Error(err, "failed to decode resource")
		}
		if err := ctrl.SetControllerReference(&job, &obj, r.scheme); err != nil {
			logger.Error(err, "failed to set owner reference")
			return ctrl.Result{}, err
		}
		if _, err := ctrl.CreateOrUpdate(ctx, r.client, &obj, func() error { return nil }); err != nil {
			logger.Error(err, "failed to create or update resource")
		}
	}
	return ctrl.Result{}, nil
}
