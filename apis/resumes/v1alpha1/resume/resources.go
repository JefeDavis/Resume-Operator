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
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
)

// sampleProfile is a sample containing all fields
const sampleProfile = `apiVersion: resumes.jefedavis.dev/v1alpha1
kind: Profile
metadata:
  name: profile-sample
  namespace: default
spec:
  profile:
    firstName: "John"
    lastName: "Doe"
    phoneNumber: ""
    email: ""
    linkedinURL: ""
    githubURL: ""
    location: "South Carolina"
    overview: ""
    coreCompetencies: ""
    projects: ""
    skills: ""
  web:
    image:
      tag: "latest"
      registry: ""
      name: "jefedavis/resume"
      pullPolicy: "IfNotPresent"
  baseURL: "example.com"
  pageTitle: "John Doe - CV"
  pageCount: "1"
  pdf:
    image:
      registry: ""
      name: "jefedavis/resume"
      tag: "latest"
      pullPolicy: "IfNotPresent"
  certIssuer: "letsencrypt-staging"
  ingressClass: "nginx"
`

// sampleProfileRequired is a sample containing only required fields
const sampleProfileRequired = `apiVersion: resumes.jefedavis.dev/v1alpha1
kind: Profile
metadata:
  name: profile-sample
  namespace: default
spec:
`

// Sample returns the sample manifest for this custom resource.
func Sample(requiredOnly bool) string {
	if requiredOnly {
		return sampleProfileRequired
	}

	return sampleProfile
}

// Generate returns the child resources that are associated with this workload given
// appropriate structured inputs.
func Generate(collectionObj resumesv1alpha1.Profile) ([]client.Object, error) {
	resourceObjects := []client.Object{}

	for _, f := range CreateFuncs {
		resources, err := f(&collectionObj)
		if err != nil {
			return nil, err
		}

		resourceObjects = append(resourceObjects, resources...)
	}

	return resourceObjects, nil
}

// GenerateForCLI returns the child resources that are associated with this workload given
// appropriate YAML manifest files.
func GenerateForCLI(collectionFile []byte) ([]client.Object, error) {
	var collectionObj resumesv1alpha1.Profile
	if err := yaml.Unmarshal(collectionFile, &collectionObj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into collection, %w", err)
	}

	if err := workload.Validate(&collectionObj); err != nil {
		return nil, fmt.Errorf("error validating collection yaml, %w", err)
	}

	return Generate(collectionObj)
}

// CreateFuncs is an array of functions that are called to create the child resources for the controller
// in memory during the reconciliation loop prior to persisting the changes or updates to the Kubernetes
// database.
var CreateFuncs = []func(
	*resumesv1alpha1.Profile,
) ([]client.Object, error){
	CreateConfigMapResumeConfig,
	CreateConfigMapResumeProfile,
	CreateDeploymentResume,
	CreateDeploymentPdfConverter,
	CreateServicePdfConverterSvc,
	CreateServiceResumeSvc,
	CreateIngressResume,
}

// InitFuncs is an array of functions that are called prior to starting the controller manager.  This is
// necessary in instances which the controller needs to "own" objects which depend on resources to
// pre-exist in the cluster. A common use case for this is the need to own a custom resource.
// If the controller needs to own a custom resource type, the CRD that defines it must
// first exist. In this case, the InitFunc will create the CRD so that the controller
// can own custom resources of that type.  Without the InitFunc the controller will
// crash loop because when it tries to own a non-existent resource type during manager
// setup, it will fail.
var InitFuncs = []func(
	*resumesv1alpha1.Profile,
) ([]client.Object, error){}

func ConvertWorkload(component workload.Workload) (*resumesv1alpha1.Profile, error) {
	p, ok := component.(*resumesv1alpha1.Profile)
	if !ok {
		return nil, resumesv1alpha1.ErrUnableToConvertProfile
	}

	return p, nil
}
