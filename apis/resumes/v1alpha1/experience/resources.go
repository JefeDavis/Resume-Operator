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
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
)

// sampleJobExperience is a sample containing all fields
const sampleJobExperience = `apiVersion: resumes.jefedavis.dev/v1alpha1
kind: JobExperience
metadata:
  name: jobexperience-sample
  namespace: default
spec:
  #collection:
    #name: "profile-sample"
    #namespace: "default"
  employer: "Employer"
  location: "Location"
  startDate: "2006-01-02"
  endDate: "Present"
  positions:
    - title: "Title"
      startDate: ""
      endDate: ""
      highlights: []
`

// sampleJobExperienceRequired is a sample containing only required fields
const sampleJobExperienceRequired = `apiVersion: resumes.jefedavis.dev/v1alpha1
kind: JobExperience
metadata:
  name: jobexperience-sample
  namespace: default
spec:
  #collection:
    #name: "profile-sample"
    #namespace: "default"
  employer: "Employer"
  location: "Location"
  startDate: "2006-01-02"
  endDate: "Present"
  positions:
    - title: "Title"
`

// Sample returns the sample manifest for this custom resource.
func Sample(requiredOnly bool) string {
	if requiredOnly {
		return sampleJobExperienceRequired
	}

	return sampleJobExperience
}

// Generate returns the child resources that are associated with this workload given
// appropriate structured inputs.
func Generate(
	workloadObj resumesv1alpha1.JobExperience,
	collectionObj resumesv1alpha1.Profile,
) ([]client.Object, error) {
	resourceObjects := []client.Object{}

	for _, f := range CreateFuncs {
		resources, err := f(&workloadObj, &collectionObj)
		if err != nil {
			return nil, err
		}

		resourceObjects = append(resourceObjects, resources...)
	}

	return resourceObjects, nil
}

// GenerateForCLI returns the child resources that are associated with this workload given
// appropriate YAML manifest files.
func GenerateForCLI(workloadFile []byte, collectionFile []byte) ([]client.Object, error) {
	var workloadObj resumesv1alpha1.JobExperience
	if err := yaml.Unmarshal(workloadFile, &workloadObj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into workload, %w", err)
	}

	if err := workload.Validate(&workloadObj); err != nil {
		return nil, fmt.Errorf("error validating workload yaml, %w", err)
	}

	var collectionObj resumesv1alpha1.Profile
	if err := yaml.Unmarshal(collectionFile, &collectionObj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into collection, %w", err)
	}

	if err := workload.Validate(&collectionObj); err != nil {
		return nil, fmt.Errorf("error validating collection yaml, %w", err)
	}

	return Generate(workloadObj, collectionObj)
}

// CreateFuncs is an array of functions that are called to create the child resources for the controller
// in memory during the reconciliation loop prior to persisting the changes or updates to the Kubernetes
// database.
var CreateFuncs = []func(
	*resumesv1alpha1.JobExperience,
	*resumesv1alpha1.Profile,
) ([]client.Object, error){
	CreateConfigMapResumeExperience,
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
	*resumesv1alpha1.JobExperience,
	*resumesv1alpha1.Profile,
) ([]client.Object, error){}

func ConvertWorkload(component, collection workload.Workload) (
	*resumesv1alpha1.JobExperience,
	*resumesv1alpha1.Profile,
	error,
) {
	p, ok := component.(*resumesv1alpha1.JobExperience)
	if !ok {
		return nil, nil, resumesv1alpha1.ErrUnableToConvertJobExperience
	}

	c, ok := collection.(*resumesv1alpha1.Profile)
	if !ok {
		return nil, nil, resumesv1alpha1.ErrUnableToConvertProfile
	}

	return p, c, nil
}
