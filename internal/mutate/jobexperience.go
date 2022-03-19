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

package mutate

import (
	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// JobExperienceMutate performs the logic to mutate resources that belong to the parent.
func JobExperienceMutate(
	r workload.Reconciler,
	req *workload.Request,
	object client.Object,
) (replacedObjects []client.Object, skip bool, err error) {
	return []client.Object{object}, false, nil
}
