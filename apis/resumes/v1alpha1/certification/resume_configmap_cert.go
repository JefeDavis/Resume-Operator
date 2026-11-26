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

// certEntry renders a single certification's YAML document for the Hugo data file.
func certEntry(cert *resumesv1alpha1.Certification) string {
	return `
---
title: ` + cert.Spec.Title + `
issuer: ` + cert.Spec.Issuer + `
earnedDate: ` + cert.Spec.EarnedDate + `
alias: ` + cert.Spec.Alias + `
validationURL: ` + cert.Spec.ValidationURL + `
imageURL: ` + cert.Spec.ImageURL + ``
}

// AggregateConfigMapResumeCert builds the resume-cert ConfigMap from all
// certifications belonging to the collection. The collection (Profile) owns the
// ConfigMap, so there is a single stable owner instead of each certification
// fighting over the owner reference. Each certification is keyed by its resource
// name, which is always set and unique within the namespace.
func AggregateConfigMapResumeCert(
	collection *resumesv1alpha1.Profile,
	certs []resumesv1alpha1.Certification,
) client.Object {
	data := map[string]any{}
	for i := range certs {
		cert := &certs[i]
		data[fmt.Sprintf("%s.yaml", cert.Name)] = certEntry(cert)
	}

	resourceObj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]any{
				"name": "resume-cert",
				"labels": map[string]any{
					"app.kubernetes.io/name":      "hugo",
					"app.kubernetes.io/component": "data",
					"app.kubernetes.io/part-of":   "resume",
					// controlled by collection field: profile.firstName
					// controlled by collection field: profile.lastName
					"app.kubernetes.io/instance":   "resume-" + collection.Spec.Profile.FirstName + "" + collection.Spec.Profile.LastName + "",
					"app.kubernetes.io/managed-by": "resume-operator",
					"app.kubernetes.io/created-by": "resume-controller-manager",
					// controlled by collection field: web.image.tag
					"app.kubernetes.io/version": resumesv1alpha1.LabelValue(collection.Spec.Web.Image.Tag),
				},
			},
			"data": data,
		},
	}

	resourceObj.SetNamespace(collection.Namespace)

	return resourceObj
}
