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

// CreateDeploymentResume creates the resume Deployment resource.
func CreateDeploymentResume(
	parent *resumesv1alpha1.Profile,
) ([]client.Object, error) {

	resourceObjs := []client.Object{}
	var resourceObj = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "resume",
				"labels": map[string]interface{}{
					//controlled by field: profile.firstName
					//controlled by field: profile.lastName
					"resume.jefedavis.dev/candidate": "" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
				},
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						//controlled by field: profile.firstName
						//controlled by field: profile.lastName
						"resume.jefedavis.dev/candidate": "" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							//controlled by field: profile.firstName
							//controlled by field: profile.lastName
							"resume.jefedavis.dev/candidate": "" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name": "resume",
								//controlled by field: image.registry
								//controlled by field: image.name
								//controlled by field: image.tag
								"image": "" + parent.Spec.Image.Registry + "" + parent.Spec.Image.Name + ":" + parent.Spec.Image.Tag + "",
								//controlled by field: image.pullPolicy
								"imagePullPolicy": parent.Spec.Image.PullPolicy,
								"resources": map[string]interface{}{
									"limits": map[string]interface{}{
										"cpu":    "500m",
										"memory": "128Mi",
									},
								},
								"volumeMounts": []interface{}{
									map[string]interface{}{
										"mountPath": "/site/data",
										"name":      "profile-mount",
									},
									map[string]interface{}{
										"mountPath": "/site/data/experience/",
										"name":      "experience-mount",
									},
									map[string]interface{}{
										"mountPath": "/site/data/certs",
										"name":      "certs-mount",
									},
									map[string]interface{}{
										"mountPath": "/site/config.toml",
										"subPath":   "config.toml",
										"name":      "config",
									},
								},
							},
						},
						"volumes": []interface{}{
							map[string]interface{}{
								"name": "profile-mount",
								"configMap": map[string]interface{}{
									"name": "resume-profile",
								},
							},
							map[string]interface{}{
								"name": "experience-mount",
								"configMap": map[string]interface{}{
									"name": "resume-experience",
								},
							},
							map[string]interface{}{
								"name": "certs-mount",
								"configMap": map[string]interface{}{
									"name": "resume-cert",
								},
							},
							map[string]interface{}{
								"name": "config",
								"configMap": map[string]interface{}{
									"name": "resume-config",
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
