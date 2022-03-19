/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resumes

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/nukleros/operator-builder-tools/pkg/controller/phases"
	"github.com/nukleros/operator-builder-tools/pkg/controller/predicates"
	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
	"github.com/jefedavis/resume-operator/apis/resumes/v1alpha1/resume"
	"github.com/jefedavis/resume-operator/internal/dependencies"
	"github.com/jefedavis/resume-operator/internal/mutate"
)

// ProfileReconciler reconciles a Profile object.
type ProfileReconciler struct {
	client.Client
	Name         string
	Log          logr.Logger
	Controller   controller.Controller
	Events       record.EventRecorder
	FieldManager string
	Watches      []client.Object
	Phases       *phases.Registry
}

func NewProfileReconciler(mgr ctrl.Manager) *ProfileReconciler {
	return &ProfileReconciler{
		Name:         "Profile",
		Client:       mgr.GetClient(),
		Events:       mgr.GetEventRecorderFor("Profile-Controller"),
		FieldManager: "Profile-reconciler",
		Log:          ctrl.Log.WithName("controllers").WithName("resumes").WithName("Profile"),
		Watches:      []client.Object{},
		Phases:       &phases.Registry{},
	}
}

// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=resumes.jefedavis.dev,resources=profiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=resumes.jefedavis.dev,resources=profiles/status,verbs=get;update;patch

// Until Webhooks are implemented we need to list and watch namespaces to ensure
// they are available before deploying resources,
// See:
//   - https://github.com/vmware-tanzu-labs/operator-builder/issues/141
//   - https://github.com/vmware-tanzu-labs/operator-builder/issues/162

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ProfileReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := r.NewRequest(ctx, request)
	if err != nil {

		if !apierrs.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if err := phases.RegisterDeleteHooks(r, req); err != nil {
		return ctrl.Result{}, err
	}

	// execute the phases
	return r.Phases.HandleExecution(r, req)
}

func (r *ProfileReconciler) NewRequest(ctx context.Context, request ctrl.Request) (*workload.Request, error) {
	component := &resumesv1alpha1.Profile{}

	log := r.Log.WithValues(
		"kind", component.GetWorkloadGVK().Kind,
		"name", request.Name,
		"namespace", request.Namespace,
	)

	// get the component from the cluster
	if err := r.Get(ctx, request.NamespacedName, component); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to fetch workload")

			return nil, fmt.Errorf("unable to fetch workload, %w", err)
		}

		return nil, err
	}

	// create the workload request
	workloadRequest := &workload.Request{
		Context:  ctx,
		Workload: component,
		Log:      log,
	}

	return workloadRequest, nil
}

// GetResources resources runs the methods to properly construct the resources in memory.
func (r *ProfileReconciler) GetResources(req *workload.Request) ([]client.Object, error) {
	resourceObjects := []client.Object{}

	component, err := resume.ConvertWorkload(req.Workload)
	if err != nil {
		return nil, err
	}

	// create resources in memory
	resources, err := resume.Generate(*component)
	if err != nil {
		return nil, err
	}

	// run through the mutation functions to mutate the resources
	for _, resource := range resources {
		mutatedResources, skip, err := r.Mutate(req, resource)
		if err != nil {
			return []client.Object{}, err
		}

		if skip {
			continue
		}

		resourceObjects = append(resourceObjects, mutatedResources...)
	}

	return resourceObjects, nil
}

// GetEventRecorder returns the event recorder for writing kubernetes events.
func (r *ProfileReconciler) GetEventRecorder() record.EventRecorder {
	return r.Events
}

// GetFieldManager returns the name of the field manager for the controller.
func (r *ProfileReconciler) GetFieldManager() string {
	return r.FieldManager
}

// GetLogger returns the logger from the reconciler.
func (r *ProfileReconciler) GetLogger() logr.Logger {
	return r.Log
}

// GetName returns the name of the reconciler.
func (r *ProfileReconciler) GetName() string {
	return r.Name
}

// GetController returns the controller object associated with the reconciler.
func (r *ProfileReconciler) GetController() controller.Controller {
	return r.Controller
}

// GetWatches returns the objects which are current being watched by the reconciler.
func (r *ProfileReconciler) GetWatches() []client.Object {
	return r.Watches
}

// SetWatch appends a watch to the list of currently watched objects.
func (r *ProfileReconciler) SetWatch(watch client.Object) {
	r.Watches = append(r.Watches, watch)
}

// CheckReady will return whether a component is ready.
func (r *ProfileReconciler) CheckReady(req *workload.Request) (bool, error) {
	return dependencies.ProfileCheckReady(r, req)
}

// Mutate will run the mutate function for the workload.
func (r *ProfileReconciler) Mutate(
	req *workload.Request,
	object client.Object,
) ([]client.Object, bool, error) {
	return mutate.ProfileMutate(r, req, object)
}

func (r *ProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.InitializePhases()

	baseController, err := ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicates.WorkloadPredicates()).
		For(&resumesv1alpha1.Profile{}).
		Build(r)
	if err != nil {
		return fmt.Errorf("unable to setup controller, %w", err)
	}

	r.Controller = baseController

	return nil
}
