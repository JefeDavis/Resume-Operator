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

package resume

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
)

// applyPodCustomizations injects the user-configurable pod labels, pod
// annotations, pod-level security context, and (optionally) a container-level
// security context into a Deployment's pod template.
//
// containerSecurityContext is applied to the container named containerName when
// non-nil. Pass nil to leave container security contexts untouched (e.g. the
// pdf-converter keeps its own hardened defaults).
func applyPodCustomizations(
	parent *resumesv1alpha1.Profile,
	deployment *unstructured.Unstructured,
	containerName string,
	containerSecurityContext *corev1.SecurityContext,
) error {
	spec, ok := deployment.Object["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("deployment %q missing spec", containerName)
	}

	template, ok := spec["template"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("deployment %q missing spec.template", containerName)
	}

	metadata, ok := template["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("deployment %q missing spec.template.metadata", containerName)
	}

	podSpec, ok := template["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("deployment %q missing spec.template.spec", containerName)
	}

	// merge additional pod labels into the existing pod template labels
	if len(parent.Spec.PodLabels) > 0 {
		labels, ok := metadata["labels"].(map[string]interface{})
		if !ok {
			labels = map[string]interface{}{}
			metadata["labels"] = labels
		}

		for k, v := range parent.Spec.PodLabels {
			labels[k] = v
		}
	}

	// set pod annotations
	if len(parent.Spec.PodAnnotations) > 0 {
		annotations, ok := metadata["annotations"].(map[string]interface{})
		if !ok {
			annotations = map[string]interface{}{}
			metadata["annotations"] = annotations
		}

		for k, v := range parent.Spec.PodAnnotations {
			annotations[k] = v
		}
	}

	// set the pod-level security context
	if parent.Spec.PodSecurityContext != nil {
		psc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(parent.Spec.PodSecurityContext)
		if err != nil {
			return fmt.Errorf("unable to convert podSecurityContext, %w", err)
		}

		podSpec["securityContext"] = psc
	}

	// set the container-level security context on the requested container
	if containerSecurityContext != nil {
		sc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(containerSecurityContext)
		if err != nil {
			return fmt.Errorf("unable to convert securityContext, %w", err)
		}

		containers, ok := podSpec["containers"].([]interface{})
		if !ok {
			return fmt.Errorf("deployment %q missing spec.template.spec.containers", containerName)
		}

		for _, c := range containers {
			container, ok := c.(map[string]interface{})
			if !ok {
				continue
			}

			if container["name"] == containerName {
				container["securityContext"] = sc
			}
		}
	}

	return nil
}
