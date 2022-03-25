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
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
)

// CreateConfigMapResumeExperience creates the resume-experience ConfigMap resource.
func CreateConfigMapResumeExperience(
	parent *resumesv1alpha1.JobExperience,
	collection *resumesv1alpha1.Profile,
) ([]client.Object, error) {
	fileName := strings.ReplaceAll(parent.Spec.Employer, " ", "-")
	fileName = strings.ReplaceAll(fileName, ".", "")
	fileName = strings.ReplaceAll(fileName, ",", "")

	var experienceBuffer bytes.Buffer

	experience := template.New("Experience")
	experience, _ = experience.Parse(experienceTemplate)
	if err := experience.Execute(&experienceBuffer, *parent); err != nil {
		return nil, fmt.Errorf("unable to scaffold experience yaml for ConfigMap, %w", err)
	}

	resourceObjs := []client.Object{}
	resourceObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "resume-experience",
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
				// controlled by field: employer
				// controlled by field: location
				// controlled by field: startDate
				// controlled by field: endDate
				// controlled by field: position.title
				// controlled by field: position.startDate
				// controlled by field: position.endDate
				// controlled by field: position.highlights
				fmt.Sprintf("%s.yaml", fileName): experienceBuffer.String(),
			},
		},
	}

	resourceObj.SetNamespace(parent.Namespace)

	resourceObjs = append(resourceObjs, resourceObj)

	return resourceObjs, nil
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
