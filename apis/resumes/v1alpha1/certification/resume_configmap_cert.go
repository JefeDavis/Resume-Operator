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

package certification

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
)

// CreateConfigMapResumeCert creates the resume-cert ConfigMap resource.
func CreateConfigMapResumeCert(
	parent *resumesv1alpha1.Certification,
	collection *resumesv1alpha1.Profile,
) ([]client.Object, error) {
	resourceObjs := []client.Object{}
	resourceObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "resume-cert",
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":      "hugo",
					"app.kubernetes.io/component": "data",
					"app.kubernetes.io/part-of":   "resume",
					// controlled by collection field: profile.firstName
					// controlled by collection field: profile.lastName
					"app.kubernetes.io/instance":   "resume-" + collection.Spec.Profile.FirstName + "" + collection.Spec.Profile.LastName + "",
					"app.kubernetes.io/managed-by": "resume-operator",
					"app.kubernetes.io/created-by": "resume-controller-manager",
					// controlled by collection field: web.image.tag
					"app.kubernetes.io/version": collection.Spec.Web.Image.Tag,
				},
			},
			"data": map[string]interface{}{
				// controlled by field: title
				// controlled by field: issuer
				// controlled by field: earnedDate
				// controlled by field: alias
				// controlled by field: validationURL
				// controlled by field: imageURL
				fmt.Sprintf("%s.yaml", parent.Spec.Alias): `
---
title: ` + parent.Spec.Title + `
issuer: ` + parent.Spec.Issuer + `
earnedDate: ` + parent.Spec.EarnedDate + `
alias: ` + parent.Spec.Alias + `
validationURL: ` + parent.Spec.ValidationURL + `
imageURL: ` + parent.Spec.ImageURL + ``,
			},
		},
	}

	resourceObj.SetNamespace(parent.Namespace)

	resourceObjs = append(resourceObjs, resourceObj)

	return resourceObjs, nil
}
