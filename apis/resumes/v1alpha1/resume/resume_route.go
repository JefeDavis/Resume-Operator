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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
)

// CreateRouteResume creates the resume Gateway API route (HTTPRoute by default)
// when spec.route is configured. When spec.route is nil, no route is created
// and the Ingress is used instead.
func CreateRouteResume(
	parent *resumesv1alpha1.Profile,
) ([]client.Object, error) {
	resourceObjs := []client.Object{}

	route := parent.Spec.Route
	if route == nil {
		return resourceObjs, nil
	}

	apiVersion := route.APIVersion
	if apiVersion == "" {
		apiVersion = "gateway.networking.k8s.io/v1"
	}

	kind := route.Kind
	if kind == "" {
		kind = "HTTPRoute"
	}

	hostname := route.Hostname
	if hostname == "" {
		hostname = parent.Spec.BaseURL
	}

	labels := map[string]any{
		"app.kubernetes.io/name":       "hugo",
		"app.kubernetes.io/component":  "webfront",
		"app.kubernetes.io/part-of":    "resume",
		"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
		"app.kubernetes.io/managed-by": "resume-operator",
		"app.kubernetes.io/created-by": "resume-controller-manager",
		"app.kubernetes.io/version":    parent.Spec.Web.Image.Tag,
	}
	for k, v := range route.Labels {
		labels[k] = v
	}

	annotations := map[string]any{}
	for k, v := range route.Annotations {
		annotations[k] = v
	}

	parentRefs := make([]any, 0, len(route.ParentRefs))
	for _, ref := range route.ParentRefs {
		parentRef := map[string]any{
			"name": ref.Name,
		}

		if ref.Namespace != "" {
			parentRef["namespace"] = ref.Namespace
		}

		if ref.SectionName != "" {
			parentRef["sectionName"] = ref.SectionName
		}

		if ref.Port != 0 {
			parentRef["port"] = int64(ref.Port)
		}

		if ref.Group != "" {
			parentRef["group"] = ref.Group
		}

		if ref.Kind != "" {
			parentRef["kind"] = ref.Kind
		}

		parentRefs = append(parentRefs, parentRef)
	}

	resourceObj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]any{
				"name":        "resume",
				"labels":      labels,
				"annotations": annotations,
			},
			"spec": map[string]any{
				"parentRefs": parentRefs,
				"hostnames": []any{
					hostname,
				},
				"rules": []any{
					// /convert routes to the pdf-converter; declared first so the
					// more specific prefix is evaluated ahead of the catch-all.
					map[string]any{
						"matches": []any{
							map[string]any{
								"path": map[string]any{
									"type":  "PathPrefix",
									"value": "/convert",
								},
							},
						},
						"backendRefs": []any{
							map[string]any{
								"name": "pdf-converter-svc",
								"port": int64(3000),
							},
						},
					},
					// / routes to the resume site.
					map[string]any{
						"matches": []any{
							map[string]any{
								"path": map[string]any{
									"type":  "PathPrefix",
									"value": "/",
								},
							},
						},
						"backendRefs": []any{
							map[string]any{
								"name": "resume-svc",
								"port": int64(8080),
							},
						},
					},
				},
			},
		},
	}

	resourceObj.SetNamespace(parent.Namespace)

	resourceObjs = append(resourceObjs, resourceObj)

	return resourceObjs, nil
}
