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
	"time"

	"github.com/nukleros/operator-builder-tools/pkg/controller/phases"
	ctrl "sigs.k8s.io/controller-runtime"
)

// InitializePhases defines what phases should be run for each event loop. phases are executed
// in the order they are listed.
func (r *CertificationReconciler) InitializePhases() {
	// Create Phases
	r.Phases.Register(
		"Dependency",
		phases.DependencyPhase,
		phases.CreateEvent,
		phases.WithCustomRequeueResult(ctrl.Result{RequeueAfter: 5 * time.Second}),
	)

	r.Phases.Register(
		"Create-Resources",
		phases.CreateResourcesPhase,
		phases.CreateEvent,
	)

	r.Phases.Register(
		"Check-Ready",
		phases.CheckReadyPhase,
		phases.CreateEvent,
		phases.WithCustomRequeueResult(ctrl.Result{RequeueAfter: 5 * time.Second}),
	)

	r.Phases.Register(
		"Complete",
		phases.CompletePhase,
		phases.CreateEvent,
	)

	// Update Phases
	r.Phases.Register(
		"Dependency",
		phases.DependencyPhase,
		phases.UpdateEvent,
		phases.WithCustomRequeueResult(ctrl.Result{RequeueAfter: 5 * time.Second}),
	)

	r.Phases.Register(
		"Create-Resources",
		phases.CreateResourcesPhase,
		phases.UpdateEvent,
	)

	r.Phases.Register(
		"Check-Ready",
		phases.CheckReadyPhase,
		phases.UpdateEvent,
		phases.WithCustomRequeueResult(ctrl.Result{RequeueAfter: 5 * time.Second}),
	)

	r.Phases.Register(
		"Complete",
		phases.CompletePhase,
		phases.UpdateEvent,
	)

	// Delete Phases
	r.Phases.Register(
		"DeletionComplete",
		phases.DeletionCompletePhase,
		phases.DeleteEvent,
	)
}
