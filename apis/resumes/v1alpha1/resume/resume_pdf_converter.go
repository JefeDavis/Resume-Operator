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
		Object: map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name": "pdf-converter",
				"labels": map[string]any{
					"app.kubernetes.io/name":      "pdf",
					"app.kubernetes.io/component": "converter",
					// controlled by field: profile.firstName
					// controlled by field: profile.lastName
					"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					"app.kubernetes.io/managed-by": "resume-operator",
					"app.kubernetes.io/part-of":    "resume",
					"app.kubernetes.io/created-by": "resume-controller-manager",
					// controlled by field: web.image.tag
					"app.kubernetes.io/version": resumesv1alpha1.LabelValue(parent.Spec.Web.Image.Tag),
				},
			},
			"spec": map[string]any{
				"selector": map[string]any{
					"matchLabels": map[string]any{
						"app.kubernetes.io/name":      "pdf",
						"app.kubernetes.io/component": "converter",
						// controlled by field: profile.firstName
						// controlled by field: profile.lastName
						"app.kubernetes.io/instance": "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					},
				},
				"template": map[string]any{
					"metadata": map[string]any{
						"labels": map[string]any{
							"app.kubernetes.io/name":      "pdf",
							"app.kubernetes.io/component": "converter",
							// controlled by field: profile.firstName
							// controlled by field: profile.lastName
							"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
							"app.kubernetes.io/managed-by": "resume-operator",
							"app.kubernetes.io/part-of":    "resume",
							"app.kubernetes.io/created-by": "resume-controller-manager",
							// controlled by field: web.image.tag
							"app.kubernetes.io/version": resumesv1alpha1.LabelValue(parent.Spec.Web.Image.Tag),
						},
					},
					"spec": map[string]any{
						"containers": []any{
							map[string]any{
								"name": "pdf-converter",
								// controlled by field: pdf.image.registry
								// controlled by field: pdf.image.name
								// controlled by field: pdf.image.tag
								"image": imageRef(parent.Spec.Pdf.Image.Registry, parent.Spec.Pdf.Image.Name, parent.Spec.Pdf.Image.Tag),
								"env": []any{
									map[string]any{
										"name":  "TARGET_URL",
										"value": "http://resume-svc:8080",
									},
								},
								"ports": []any{
									map[string]any{
										"containerPort": 3000,
									},
								},
								// controlled by field: pdf.image.pullPolicy
								"imagePullPolicy": parent.Spec.Pdf.Image.PullPolicy,
								// The converter drives headless Chrome with --no-sandbox
								// (see the resume-pdf-converter image), so it no longer needs
								// the SYS_ADMIN capability. Harden the container instead.
								"securityContext": map[string]any{
									"runAsNonRoot":             true,
									"allowPrivilegeEscalation": false,
									"capabilities": map[string]any{
										"drop": []any{
											"ALL",
										},
									},
									"seccompProfile": map[string]any{
										"type": "RuntimeDefault",
									},
								},
								// writable scratch so the converter's temp PDFs and
								// headless Chrome work under a read-only root filesystem.
								"volumeMounts": []any{
									map[string]any{
										"mountPath": "/tmp",
										"name":      "tmp",
									},
								},
							},
						},
						"volumes": []any{
							map[string]any{
								"name":     "tmp",
								"emptyDir": map[string]any{},
							},
						},
					},
				},
			},
		},
	}

	resourceObj.SetNamespace(parent.Namespace)

	// inject pod labels/annotations and podSecurityContext; leave the container
	// securityContext alone so the pdf-converter keeps its hardened defaults.
	if err := applyPodCustomizations(parent, resourceObj, "pdf-converter", nil); err != nil {
		return nil, err
	}

	resourceObjs = append(resourceObjs, resourceObj)

	return resourceObjs, nil
}

// CreateServicePdfConverterSvc creates the pdf-converter-svc Service resource.
func CreateServicePdfConverterSvc(
	parent *resumesv1alpha1.Profile,
) ([]client.Object, error) {
	resourceObjs := []client.Object{}
	resourceObj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]any{
				"name": "pdf-converter-svc",
				"labels": map[string]any{
					"app.kubernetes.io/name":      "pdf",
					"app.kubernetes.io/component": "converter",
					// controlled by field: profile.firstName
					// controlled by field: profile.lastName
					"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					"app.kubernetes.io/managed-by": "resume-operator",
					"app.kubernetes.io/part-of":    "resume",
					"app.kubernetes.io/created-by": "resume-controller-manager",
					// controlled by field: web.image.tag
					"app.kubernetes.io/version": resumesv1alpha1.LabelValue(parent.Spec.Web.Image.Tag),
				},
			},
			"spec": map[string]any{
				"selector": map[string]any{
					"app.kubernetes.io/name":      "pdf",
					"app.kubernetes.io/component": "converter",
					// controlled by field: profile.firstName
					// controlled by field: profile.lastName
					"app.kubernetes.io/instance": "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
				},
				"ports": []any{
					map[string]any{
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
