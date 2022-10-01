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
	"bytes"
	"fmt"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
)

// CreateConfigMapResumeProfile creates the resume-profile ConfigMap resource.
func CreateConfigMapResumeProfile(
	parent *resumesv1alpha1.Profile,
) ([]client.Object, error) {
	var profileBuffer bytes.Buffer

	profile := template.New("Profile")
	profile, _ = profile.Parse(profileTemplate)
	if err := profile.Execute(&profileBuffer, *parent); err != nil {
		return nil, fmt.Errorf("unable to scaffold profile.yaml for ConfigMap, %w", err)
	}

	resourceObjs := []client.Object{}
	resourceObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "resume-profile",
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":      "hugo",
					"app.kubernetes.io/component": "data",
					"app.kubernetes.io/part-of":   "resume",
					// controlled by field: profile.firstName
					// controlled by field: profile.lastName
					"app.kubernetes.io/instance":   "resume-" + parent.Spec.Profile.FirstName + "" + parent.Spec.Profile.LastName + "",
					"app.kubernetes.io/managed-by": "resume-operator",
					"app.kubernetes.io/created-by": "resume-controller-manager",
					// controlled by field: web.image.tag
					"app.kubernetes.io/version": parent.Spec.Web.Image.Tag,
				},
			},
			"data": map[string]interface{}{
				// controlled by field: profile.firstName
				// controlled by field: profile.lastName
				// controlled by field: profile.phoneNumber
				// controlled by field: profile.email
				// controlled by field: profile.linkedinURL
				// controlled by field: profile.githubURL
				// controlled by field: profile.location
				// controlled by field: profile.overview
				// controlled by field: profile.coreCompetencies
				// controlled by field: profile.projects
				// controlled by field: profile.skills
				"profile.yaml": profileBuffer.String(),
			},
		},
	}

	resourceObj.SetNamespace(parent.Namespace)

	resourceObjs = append(resourceObjs, resourceObj)

	return resourceObjs, nil
}

const profileTemplate = `
---
basicInfo:
  firstName: {{ .Spec.Profile.FirstName }}
  lastName: {{ .Spec.Profile.LastName }}
  photo: img/avatar.jpg
  contacts:
    {{- if .Spec.Profile.PhoneNumber }}
    - icon: fa-solid fa-phone
      info: {{ .Spec.Profile.PhoneNumber }}
    {{- end }}
    {{- if .Spec.Profile.Email }}
    - icon: fa-solid fa-envelope
      info: {{ .Spec.Profile.Email }}
    {{- end }}
    {{- if .Spec.Profile.LinkedinURL }}
    - icon: fa-brands fa-linkedin
      info: {{ .Spec.Profile.LinkedinURL }}
    {{- end }}
    {{- if .Spec.Profile.GithubURL }}
    - icon: fa-brands fa-github
      info: {{ .Spec.Profile.GithubURL }}
    {{- end }}
    - icon: fa-solid fa-map-marker-alt
      info: {{ .Spec.Profile.Location }}
overview: {{ .Spec.Profile.Overview }}
coreCompetencies: 
  {{- range .Spec.Profile.CoreCompetencies }}
  - {{ . }}
	{{- end }}
projects:
  {{- range .Spec.Profile.Projects }}
  - {{ . }}
	{{- end }}
skills: 
  {{- range .Spec.Profile.Skills }}
  - family: {{ .Family }}
    items:
		  {{- range .Items }}
      - {{ . }}
			{{- end }}
	{{- end }}
`
