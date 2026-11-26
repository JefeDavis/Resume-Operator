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
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
)

// CreateDeploymentResume creates the resume Deployment resource.
func CreateDeploymentResume(
	parent *resumesv1alpha1.Profile,
) ([]client.Object, error) {
	resourceObjs := []client.Object{}

	// Hugo resolves the ConfigMap data at startup, so content changes must roll
	// the pod. Hash the Profile spec (which drives profile.yaml and config.toml)
	// into a pod-template annotation so an update triggers a new rollout.
	specJSON, err := json.Marshal(parent.Spec)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal Profile spec for config checksum, %w", err)
	}
	configChecksum := fmt.Sprintf("%x", sha256.Sum256(specJSON))
	resourceObj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name": "resume",
				"labels": map[string]any{
					"app.kubernetes.io/name":      "hugo",
					"app.kubernetes.io/component": "webfront",
					"app.kubernetes.io/part-of":   "resume",
					// controlled by field: profile.firstName
					// controlled by field: profile.lastName
					"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					"app.kubernetes.io/managed-by": "resume-operator",
					"app.kubernetes.io/created-by": "resume-controller-manager",
					// controlled by field: web.image.tag
					"app.kubernetes.io/version": resumesv1alpha1.LabelValue(parent.Spec.Web.Image.Tag),
				},
			},
			"spec": map[string]any{
				"selector": map[string]any{
					"matchLabels": map[string]any{
						"app.kubernetes.io/name":      "hugo",
						"app.kubernetes.io/component": "webfront",
						"app.kubernetes.io/part-of":   "resume",
						// controlled by field: profile.firstName
						// controlled by field: profile.lastName
						"app.kubernetes.io/instance": "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					},
				},
				"template": map[string]any{
					"metadata": map[string]any{
						// roll the pod when the Profile-driven config/data changes
						"annotations": map[string]any{
							"resume.jefedavis.dev/config-checksum": configChecksum,
						},
						"labels": map[string]any{
							"app.kubernetes.io/name":      "hugo",
							"app.kubernetes.io/component": "webfront",
							"app.kubernetes.io/part-of":   "resume",
							// controlled by field: profile.firstName
							// controlled by field: profile.lastName
							"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
							"app.kubernetes.io/managed-by": "resume-operator",
							"app.kubernetes.io/created-by": "resume-controller-manager",
							// controlled by field: web.image.tag
							"app.kubernetes.io/version": resumesv1alpha1.LabelValue(parent.Spec.Web.Image.Tag),
						},
					},
					"spec": map[string]any{
						"containers": []any{
							map[string]any{
								"name": "resume",
								// controlled by field: web.image.registry
								// controlled by field: web.image.name
								// controlled by field: web.image.tag
								"image": imageRef(parent.Spec.Web.Image.Registry, parent.Spec.Web.Image.Name, parent.Spec.Web.Image.Tag),
								// controlled by field: web.image.pullPolicy
								"imagePullPolicy": parent.Spec.Web.Image.PullPolicy,
								// The operator owns the full Hugo server command.
								// --noBuildLock/--renderToMemory (with HUGO_CACHEDIR and
								// HUGO_RESOURCEDIR pointed under /tmp in the image) let Hugo
								// run under a read-only root filesystem.
								"args": []any{
									"server",
									// controlled by field: baseURL
									"--baseURL=https://" + parent.Spec.BaseURL,
									"--appendPort=false",
									"--source=/site",
									"--bind=0.0.0.0",
									"--disableFastRender",
									"--noBuildLock",
									"--renderToMemory",
								},
								"volumeMounts": []any{
									map[string]any{
										"mountPath": "/site/data",
										"name":      "profile-mount",
									},
									map[string]any{
										"mountPath": "/site/data/experience/",
										"name":      "experience-mount",
									},
									map[string]any{
										"mountPath": "/site/data/certs",
										"name":      "certs-mount",
									},
									map[string]any{
										"mountPath": "/site/config.toml",
										"subPath":   "config.toml",
										"name":      "config",
									},
									// writable scratch so Hugo's cache (HUGO_CACHEDIR=/tmp)
									// works under a read-only root filesystem.
									map[string]any{
										"mountPath": "/tmp",
										"name":      "tmp",
									},
								},
							},
						},
						"volumes": []any{
							map[string]any{
								"name": "profile-mount",
								"configMap": map[string]any{
									"name": "resume-profile",
								},
							},
							map[string]any{
								"name": "experience-mount",
								"configMap": map[string]any{
									"name": "resume-experience",
								},
							},
							map[string]any{
								"name": "certs-mount",
								"configMap": map[string]any{
									"name": "resume-cert",
								},
							},
							map[string]any{
								"name": "config",
								"configMap": map[string]any{
									"name": "resume-config",
								},
							},
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

	// inject pod labels/annotations, podSecurityContext, and the resume
	// container securityContext from the Profile spec.
	if err := applyPodCustomizations(parent, resourceObj, "resume", parent.Spec.SecurityContext); err != nil {
		return nil, err
	}

	resourceObjs = append(resourceObjs, resourceObj)

	return resourceObjs, nil
}
