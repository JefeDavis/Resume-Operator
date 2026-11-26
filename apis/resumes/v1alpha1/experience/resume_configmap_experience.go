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

package experience

import (
	"bytes"
	"fmt"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
)

// experienceEntry renders a single job experience's YAML document for the Hugo
// data file.
func experienceEntry(exp *resumesv1alpha1.JobExperience) (string, error) {
	var buf bytes.Buffer

	tmpl, err := template.New("Experience").Parse(experienceTemplate)
	if err != nil {
		return "", fmt.Errorf("unable to parse experience template, %w", err)
	}

	if err := tmpl.Execute(&buf, *exp); err != nil {
		return "", fmt.Errorf("unable to scaffold experience yaml for ConfigMap, %w", err)
	}

	return buf.String(), nil
}

// AggregateConfigMapResumeExperience builds the resume-experience ConfigMap from
// all job experiences belonging to the collection. The collection (Profile) owns
// the ConfigMap, so there is a single stable owner instead of each experience
// fighting over the owner reference. Each experience is keyed by its resource
// name, which is always set and unique within the namespace.
func AggregateConfigMapResumeExperience(
	collection *resumesv1alpha1.Profile,
	experiences []resumesv1alpha1.JobExperience,
) (client.Object, error) {
	data := map[string]any{}

	for i := range experiences {
		exp := &experiences[i]

		entry, err := experienceEntry(exp)
		if err != nil {
			return nil, err
		}

		data[fmt.Sprintf("%s.yaml", exp.Name)] = entry
	}

	resourceObj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]any{
				"name": "resume-experience",
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

	return resourceObj, nil
}

const experienceTemplate = `
---
employer: {{ .Spec.Employer }}
location: {{ .Spec.Location }}
startDate: {{ .Spec.StartDate }}
endDate: {{ .Spec.EndDate }}
positions:
{{- range .Spec.Positions }}
  - title: {{ .Title }}
    startDate: {{ .StartDate }}
    endDate: {{ .EndDate }}
    highlights:
		{{- range .Highlights }}
      - {{ . }}
		{{- end }}
{{- end }}
`
