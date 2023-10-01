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

// CreateDeploymentPdfConverter creates the pdf-converter Deployment resource.
func CreateDeploymentPdfConverter(
	parent *resumesv1alpha1.Profile,
) ([]client.Object, error) {
	resourceObjs := []client.Object{}
	resourceObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "pdf-converter",
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":      "pdf",
					"app.kubernetes.io/component": "converter",
					// controlled by field: profile.firstName
					// controlled by field: profile.lastName
					"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					"app.kubernetes.io/managed-by": "resume-operator",
					"app.kubernetes.io/part-of":    "resume",
					"app.kubernetes.io/created-by": "resume-controller-manager",
					// controlled by field: web.image.tag
					"app.kubernetes.io/version": parent.Spec.Web.Image.Tag,
				},
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.kubernetes.io/name":      "pdf",
						"app.kubernetes.io/component": "converter",
						// controlled by field: profile.firstName
						// controlled by field: profile.lastName
						"app.kubernetes.io/instance": "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app.kubernetes.io/name":      "pdf",
							"app.kubernetes.io/component": "converter",
							// controlled by field: profile.firstName
							// controlled by field: profile.lastName
							"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
							"app.kubernetes.io/managed-by": "resume-operator",
							"app.kubernetes.io/part-of":    "resume",
							"app.kubernetes.io/created-by": "resume-controller-manager",
							// controlled by field: web.image.tag
							"app.kubernetes.io/version": parent.Spec.Web.Image.Tag,
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name": "pdf-converter",
								// controlled by field: pdf.image.registry
								// controlled by field: pdf.image.name
								// controlled by field: pdf.image.tag
								"image": "" + parent.Spec.Pdf.Image.Registry + "" + parent.Spec.Pdf.Image.Name + ":" + parent.Spec.Pdf.Image.Tag + "",
								"env": []interface{}{
									map[string]interface{}{
										"name":  "TARGET_URL",
										"value": "http://resume-svc:8080",
									},
								},
								"ports": []interface{}{
									map[string]interface{}{
										"containerPort": 3000,
									},
								},
								// controlled by field: pdf.image.pullPolicy
								"imagePullPolicy": parent.Spec.Pdf.Image.PullPolicy,
								"securityContext": map[string]interface{}{
									"capabilities": map[string]interface{}{
										"add": []interface{}{
											"SYS_ADMIN",
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

// CreateServicePdfConverterSvc creates the pdf-converter-svc Service resource.
func CreateServicePdfConverterSvc(
	parent *resumesv1alpha1.Profile,
) ([]client.Object, error) {
	resourceObjs := []client.Object{}
	resourceObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "pdf-converter-svc",
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":      "pdf",
					"app.kubernetes.io/component": "converter",
					// controlled by field: profile.firstName
					// controlled by field: profile.lastName
					"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					"app.kubernetes.io/managed-by": "resume-operator",
					"app.kubernetes.io/part-of":    "resume",
					"app.kubernetes.io/created-by": "resume-controller-manager",
					// controlled by field: web.image.tag
					"app.kubernetes.io/version": parent.Spec.Web.Image.Tag,
				},
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app.kubernetes.io/name":      "pdf",
					"app.kubernetes.io/component": "converter",
					// controlled by field: profile.firstName
					// controlled by field: profile.lastName
					"app.kubernetes.io/instance": "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
				},
				"ports": []interface{}{
					map[string]interface{}{
						"port":       3000,
						"targetPort": 3000,
					},
				},
			},
		},
	}

	resourceObj.SetNamespace(parent.Namespace)

	resourceObjs = append(resourceObjs, resourceObj)

	return resourceObjs, nil
}
