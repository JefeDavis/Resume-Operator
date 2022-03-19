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

// CreateIngressResume creates the resume Ingress resource.
func CreateIngressResume(
	parent *resumesv1alpha1.Profile,
) ([]client.Object, error) {

	resourceObjs := []client.Object{}
	var resourceObj = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.k8s.io/v1",
			"kind":       "Ingress",
			"metadata": map[string]interface{}{
				"name": "resume",
				"labels": map[string]interface{}{
					//controlled by field: profile.firstName
					//controlled by field: profile.lastName
					"resume.jefedavis.dev/candidate": "" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
				},
			},
			"spec": map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						//controlled by field: baseURL
						"host": parent.Spec.BaseURL,
						"https": map[string]interface{}{
							"paths": []interface{}{
								map[string]interface{}{
									"pathType": "Prefix",
									"path":     "/",
									"backend": map[string]interface{}{
										"service": map[string]interface{}{
											"name": "resume-svc",
											"port": map[string]interface{}{
												"number": 8080,
											},
										},
									},
								},
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
