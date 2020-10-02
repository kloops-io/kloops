package controllers

import (
	"context"

	"github.com/go-logr/logr"
	build_v1alpha1 "github.com/kloops-io/kloops/apis/build/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
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

	if job.Spec.Resource.Raw != nil {
		codec := serializer.NewCodecFactory(r.scheme)
		decoder := codec.UniversalDecoder(r.scheme.PrioritizedVersionsAllGroups()...)
		obj, _ := runtime.Decode(decoder, job.Spec.Resource.Raw)
		if result, err := ctrl.CreateOrUpdate(ctx, r.client, obj, func() error { return nil }); err != nil {
			logger.Error(err, "failed to create or update resource")
		} else {
			logger.Info(string(result))
		}
	}

	// // filter on job agent
	// if job.Spec.Agent != configjob.TektonPipelineAgent {
	// 	return ctrl.Result{}, nil
	// }

	// // get job's pipeline runs
	// var pipelineRunList pipelinev1beta1.PipelineRunList
	// if err := r.client.List(ctx, &pipelineRunList, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
	// 	r.logger.Errorf("Failed list pipeline runs: %s", err)
	// 	return ctrl.Result{}, err
	// }

	// // if pipeline run does not exist, create it
	// if len(pipelineRunList.Items) == 0 {
	// 	if job.Status.State == lighthousev1alpha1.TriggeredState {
	// 		// construct a pipeline run
	// 		pipelineRun, err := makePipelineRun(ctx, job, r.namespace, r.logger, r.idGenerator, r.apiReader)
	// 		if err != nil {
	// 			r.logger.Errorf("Failed to make pipeline run: %s", err)
	// 			return ctrl.Result{}, err
	// 		}
	// 		// link it to the current lighthouse job
	// 		if err := ctrl.SetControllerReference(&job, pipelineRun, r.scheme); err != nil {
	// 			r.logger.Errorf("Failed to set owner reference: %s", err)
	// 			return ctrl.Result{}, err
	// 		}
	// 		// TODO: changing the status should be a consequence of a pipeline run being created
	// 		// update status
	// 		job.Status = lighthousev1alpha1.LighthouseJobStatus{
	// 			State:     lighthousev1alpha1.PendingState,
	// 			StartTime: metav1.Now(),
	// 		}
	// 		if err := r.client.Status().Update(ctx, &job); err != nil {
	// 			r.logger.Errorf("Failed to update LighthouseJob status: %s", err)
	// 			return ctrl.Result{}, err
	// 		}
	// 		// create pipeline run
	// 		if err := r.client.Create(ctx, pipelineRun); err != nil {
	// 			r.logger.Errorf("Failed to create pipeline run: %s", err)
	// 			return ctrl.Result{}, err
	// 		}
	// 	}
	// } else if len(pipelineRunList.Items) == 1 {
	// 	// if pipeline run exists, create it and update status
	// 	pipelineRun := pipelineRunList.Items[0]
	// 	r.logger.Infof("Reconcile PipelineRun %+v", pipelineRun)
	// 	// update build id
	// 	job.Labels[util.BuildNumLabel] = pipelineRun.Labels[util.BuildNumLabel]
	// 	if err := r.client.Update(ctx, &job); err != nil {
	// 		r.logger.Errorf("failed to update Project status: %s", err)
	// 		return ctrl.Result{}, err
	// 	}
	// 	if r.dashboardURL != "" {
	// 		job.Status.ReportURL = fmt.Sprintf("%s/#/namespaces/%s/pipelineruns/%s", trimDashboardURL(r.dashboardURL), r.namespace, pipelineRun.Name)
	// 	}
	// 	job.Status.Activity = ConvertPipelineRun(&pipelineRun)
	// 	if err := r.client.Status().Update(ctx, &job); err != nil {
	// 		r.logger.Errorf("Failed to update LighthouseJob status: %s", err)
	// 		return ctrl.Result{}, err
	// 	}
	// } else {
	// 	r.logger.Errorf("A lighthouse job should never have more than 1 pipeline run")
	// }

	return ctrl.Result{}, nil
}
