//go:build e2e_test
// +build e2e_test

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

package e2e_test

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	resumesv1alpha1 "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1"
	"github.com/jefedavis/resume-operator/apis/resumes/v1alpha1/experience"
)

//
// resumesv1alpha1JobExperience tests
//
func resumesv1alpha1JobExperienceChildrenFuncs(tester *E2ETest) error {
	// TODO: need to run r.GetResources(request) on the reconciler to get the mutated resources
	if len(experience.CreateFuncs) == 0 {
		return nil
	}

	workload, collection, err := experience.ConvertWorkload(tester.workload, tester.collectionTester.workload)
	if err != nil {
		return fmt.Errorf("error in workload conversion; %w", err)
	}

	resourceObjects, err := experience.Generate(*workload, *collection)
	if err != nil {
		return fmt.Errorf("unable to create objects in memory; %w", err)
	}

	tester.children = resourceObjects

	return nil
}

func resumesv1alpha1JobExperienceNewHarness(namespace string) *E2ETest {
	return &E2ETest{
		namespace:          namespace,
		unstructured:       &unstructured.Unstructured{},
		workload:           &resumesv1alpha1.JobExperience{},
		sampleManifestFile: "../../config/samples/resumes_v1alpha1_jobexperience.yaml",
		getChildrenFunc:    resumesv1alpha1JobExperienceChildrenFuncs,
		logSyntax:          "controllers.resumes.JobExperience",
		collectionTester:   resumesv1alpha1ProfileNewHarness("test-resumes-v1alpha1-profile"),
	}
}

func (tester *E2ETest) resumesv1alpha1JobExperienceTest(testSuite *E2EComponentTestSuite) {
	testSuite.suiteConfig.tests = append(testSuite.suiteConfig.tests, tester)
	tester.suiteConfig = &testSuite.suiteConfig
	require.NoErrorf(testSuite.T(), tester.setup(), "failed to setup test")

	// create the custom resource
	require.NoErrorf(testSuite.T(), testCreateCustomResource(tester), "failed to create custom resource")

	// test the deletion of a child object
	require.NoErrorf(testSuite.T(), testDeleteChildResource(tester), "failed to reconcile deletion of a child resource")

	// test the update of a child object
	// TODO: need immutable fields so that we can predict which managed fields we can modify to test reconciliation
	// see https://github.com/vmware-tanzu-labs/operator-builder/issues/67

	// test the update of a parent object
	// TODO: need immutable fields so that we can predict which managed fields we can modify to test reconciliation
	// see https://github.com/vmware-tanzu-labs/operator-builder/issues/67

	// test that controller logs do not contain errors
	if os.Getenv("DEPLOY_IN_CLUSTER") == "true" {
		require.NoErrorf(testSuite.T(), testControllerLogsNoErrors(tester.suiteConfig, tester.logSyntax), "found errors in controller logs")
	}
}

func (testSuite *E2EComponentTestSuite) Test_resumesv1alpha1JobExperience() {
	tester := resumesv1alpha1JobExperienceNewHarness("test-resumes-v1alpha1-jobexperience")
	tester.resumesv1alpha1JobExperienceTest(testSuite)
}

func (testSuite *E2EComponentTestSuite) Test_resumesv1alpha1JobExperienceMulti() {
	tester := resumesv1alpha1JobExperienceNewHarness("test-resumes-v1alpha1-jobexperience-2")
	tester.resumesv1alpha1JobExperienceTest(testSuite)
}
