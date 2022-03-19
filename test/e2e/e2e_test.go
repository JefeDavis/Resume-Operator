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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8syaml "sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
	"github.com/nukleros/operator-builder-tools/pkg/resources"
	kbresource "sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
)

// E2ETestSuiteConfig represents the entire suite of tests.
type E2ETestSuiteConfig struct {
	dynamicClient    dynamic.Interface
	client           kubernetes.Clientset
	controllerConfig controllerConfig
	tests            []*E2ETest
}

type controllerConfig struct {
	Namespace string `yaml:"namespace"`
	Prefix    string `yaml:"namePrefix"`
}

// E2EComponentTestSuite represents an indvidual component test.
type E2EComponentTestSuite struct {
	suite.Suite

	suiteConfig E2ETestSuiteConfig
}

// E2ECollectionTestSuite represents an individual collection test.
type E2ECollectionTestSuite struct {
	suite.Suite

	suiteConfig E2ETestSuiteConfig
}

// E2ETest represents an individual test.
type E2ETest struct {
	suiteConfig        *E2ETestSuiteConfig
	namespace          string
	sampleManifestFile string
	unstructured       *unstructured.Unstructured
	workload           workload.Workload
	collectionTester   *E2ETest
	children           []client.Object
	getChildrenFunc    getChildren
	logSyntax          string
}

type getChildren func(*E2ETest) error
type readyChecker func() (bool, error)

const (
	controllerName          = "controller-manager"
	controllerKustomization = "../../config/default/kustomization.yaml"
	waitTimeout             = 90 * time.Second
	waitInterval            = 3 * time.Second
)

// deletableWhitelist is a representation of known kinds which may be
// deleted for our test
var deletableWhitelist = []string{
	"Deployment",
	"Secret",
	"ConfigMap",
	"DaemonSet",
	"Pod",
	"Service",
	"Ingress",
	"StorageClass",
}

//
// test entrypoint
//
func TestMain(t *testing.T) {
	// setup the test suite
	e2eTestSuite := new(E2ETestSuiteConfig)
	require.NoErrorf(t, setupSuite(e2eTestSuite), "error setting up test suite")

	// setup the tests
	collectionSuite := &E2ECollectionTestSuite{suiteConfig: *e2eTestSuite}
	componentSuite := &E2EComponentTestSuite{suiteConfig: *e2eTestSuite}

	// execute the tests
	t.Run("TestE2ESuite", func(t *testing.T) {
		// run collection test suite first
		suite.Run(t, collectionSuite)

		// run component test suite, in parallel, next
		suite.Run(t, componentSuite)
	})

	// teardown the test suites
	componentSuite.teardown()
	collectionSuite.teardown()

	// check all controller logs for errors
	if os.Getenv("DEPLOY_IN_CLUSTER") == "true" {
		require.NoErrorf(t, testControllerLogsNoErrors(e2eTestSuite, ""), "found errors in controller logs")
	}

	// perform final teardown
	require.NoErrorf(t, finalTeardown(), "error tearing down test suite")
}

//
// setup
//

// setupSuite is the common logic for both collection and component tests to run.
func setupSuite(s *E2ETestSuiteConfig) error {
	// create rest config from kubeconfig
	var err error
	var config *rest.Config
	if os.Getenv("KUBECONFIG") != "" {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	}

	if err != nil {
		return fmt.Errorf("unable to create rest config from kubeconfig; %w", err)
	}

	// create client
	restClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable create rest client from kubeconfig; %w", err)
	}
	s.client = *restClient

	// create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create dynamic client from kubeconfig; %w", err)
	}
	s.dynamicClient = dynamicClient

	// get the controller configuration from yaml
	if err := readYamlFile(controllerKustomization, &s.controllerConfig); err != nil {
		return fmt.Errorf("unable to fetch controller configuration; %w", err)
	}

	// run deploy
	return deploy(s)
}

// SetupTest is called once at the beginning of each test.  Component tests run in parallel,
// but collection tests do not.
func (s *E2EComponentTestSuite) SetupTest() {
	s.T().Parallel()
}

// setup is called upon entering a test.  This is separate from the above
// method as it populates specific metadata about an individual test that is
// not otherwise available during the SetupTest method.
func (tester *E2ETest) setup() error {
	// get the sample manifest from yaml
	yamlFile, err := readYamlManifest(tester.sampleManifestFile, tester.unstructured)
	if err != nil {
		return fmt.Errorf("unable to fetch sample manifest; %w", err)
	}

	// get the proper object from the manifest object
	if err := k8syaml.Unmarshal(yamlFile, tester.workload); err != nil {
		return fmt.Errorf("unable to unmarshal yaml to api object; %w", err)
	}

	// ensure the namespace for the underlying manifest matches the tester namespace
	tester.unstructured.SetNamespace(tester.namespace)
	tester.workload.SetNamespace(tester.namespace)

	// get the proper collection object from the manifest object
	if tester.collectionTester != nil {
		collection := &unstructured.Unstructured{}
		collectionYaml, err := readYamlManifest(tester.collectionTester.sampleManifestFile, collection)
		if err != nil {
			return fmt.Errorf("unable to fetch sample collection manifest; %w", err)
		}

		if err := k8syaml.Unmarshal(collectionYaml, tester.collectionTester.workload); err != nil {
			return fmt.Errorf("unable to unmarshal collection yaml to api object; %w", err)
		}

		// ensure the namespace for the underlying manifest matches the collection tester namespace
		tester.collectionTester.unstructured.SetNamespace(tester.collectionTester.namespace)
		tester.collectionTester.workload.SetNamespace(tester.collectionTester.namespace)
	}

	// get and store the non-mutated child objects
	if err := tester.getChildrenFunc(tester); err != nil {
		return fmt.Errorf("unable to unmarshal yaml to api object; %w", err)
	}

	// create a namespace for each test case
	// NOTE: cluster-scoped resources will not have a namespace and therefore will
	// not receive an individual namespace for their test case
	if tester.namespace != "" {
		if err := createNamespaceForTest(tester); err != nil {
			return fmt.Errorf("failed to create namespace for test; %w", err)
		}
	}

	return nil
}

//
// deploy
//
// DEPLOY="true" will run all tasks to deploy into the cluster to include:
//   - docker build
//   - docker push
//   - crd install
//   - controller deployment
//
// DEPLOY_IN_CLUSTER="true" ensures that the controller is running before proceeding.
// if this option is not used, a separate process such as the 'make run' target
// should be handling the controller functions for the test.
//
func deploy(s *E2ETestSuiteConfig) error {
	// install crds
	if os.Getenv("DEPLOY") == "true" {
		installCommand := exec.Command("make", "-C", "../..", "install")
		_, err := installCommand.Output()
		if err != nil {
			return fmt.Errorf("failed to run 'make install' target; %w", err)
		}
	}

	if os.Getenv("DEPLOY_IN_CLUSTER") == "true" {
		if os.Getenv("DEPLOY") == "true" {
			// build image
			buildCommand := exec.Command("make", "-C", "../..", "docker-build")
			_, err := buildCommand.Output()
			if err != nil {
				return fmt.Errorf("failed to run 'make docker-build' target; %w", err)
			}

			// push image
			pushCommand := exec.Command("make", "-C", "../..", "docker-push")
			_, err = pushCommand.Output()
			if err != nil {
				return fmt.Errorf("failed to run 'make docker-push' target; %w", err)
			}

			// deploy controller
			deployCommand := exec.Command("make", "-C", "../..", "deploy")
			_, err = deployCommand.Output()
			if err != nil {
				return fmt.Errorf("failed to run 'make deploy' target; %w", err)
			}
		}

		// wait for controller to be ready
		if err := waitForController(s); err != nil {
			return fmt.Errorf("failed to wait for controller for test; %w", err)
		}
	}

	return nil
}

//
// teardown
//
// make undeploy will teardown the operator and all of its associated custom
// resources
//

// finalTeardown is the last teardown operation that happens in the E2E testing.
func finalTeardown() error {
	// run teardown
	if os.Getenv("TEARDOWN") == "true" {
		var undeployCommand *exec.Cmd

		if os.Getenv("DEPLOY_IN_CLUSTER") == "true" {
			undeployCommand = exec.Command("make", "-C", "../..", "undeploy")
		} else {
			undeployCommand = exec.Command("make", "-C", "../..", "uninstall")
		}

		_, err := undeployCommand.Output()
		if err != nil {
			return fmt.Errorf("failed to run 'make undeploy/uninstall' target with error; %w", err)
		}
	}

	return nil
}

// teardownSuite is called once at the very end of all tests.
func teardownSuite(s *E2ETestSuiteConfig) error {
	for _, e2eTest := range s.tests {
		// delete the custom resources for the tests
		if err := deleteCustomResource(e2eTest); err != nil {
			return fmt.Errorf("failed to delete custom resource: %+v; %w", e2eTest, err)
		}

		// delete the namespaces for the tests
		if e2eTest.namespace != "" {
			if err := deleteNamespaceForTest(e2eTest); err != nil {
				return fmt.Errorf("failed to delete namespace during teardown: %s; %w", e2eTest.namespace, err)
			}
		}
	}

	return nil
}

// TearDownSuite runs the logic to teardown a collection test suite.
func (s *E2ECollectionTestSuite) teardown() {
	if len(s.suiteConfig.tests) > 0 {
		require.NoErrorf(s.T(), teardownSuite(&s.suiteConfig), "unable to teardown collection test suite")
	}
}

// TearDownSuite runs the logic to teardown a component test suite.
func (s *E2EComponentTestSuite) teardown() {
	if len(s.suiteConfig.tests) > 0 {
		require.NoErrorf(s.T(), teardownSuite(&s.suiteConfig), "unable to teardown component test suite")
	}
}

//
// helpers
//
func readYamlManifest(path string, destination *unstructured.Unstructured) ([]byte, error) {
	// read the yaml file
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s; %w", path, err)
	}

	// decode yaml into unstructured.Unstructured
	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err = dec.Decode(yamlFile, nil, destination)
	if err != nil {
		return nil, fmt.Errorf("error decoding sample manifest %s; %w\n\nwith data: %s", path, err, yamlFile)
	}

	return yamlFile, nil
}

func readYamlFile(path string, destination interface{}) error {
	// read the yaml file
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read file %s; %w", path, err)
	}

	// store config in memory
	if err = yaml.Unmarshal(yamlFile, destination); err != nil {
		return fmt.Errorf("unable to unmarshal yaml file %s; %w", path, err)
	}

	return nil
}

func newNamespaceStub(namespaceName string) *v1.Namespace {
	return &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: resources.NamespaceVersion,
			Kind:       resources.NamespaceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
}

func namespaceExists(tester *E2ETest) (bool, error) {
	_, err := tester.suiteConfig.client.CoreV1().Namespaces().Get(
		context.TODO(),
		tester.namespace,
		metav1.GetOptions{},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func getPlural(kind string) string {
	pluralMap := map[string]string{
		"resourcequota": "resourcequotas",
	}
	plural := kbresource.RegularPlural(kind)

	if pluralMap[plural] != "" {
		return pluralMap[plural]
	}

	return plural
}

func getUpdatableChild(tester *E2ETest, name, namespace, kind string) client.Object {
	for _, child := range tester.children {
		if child.GetObjectKind().GroupVersionKind().Kind == kind {
			if child.GetName() == name && child.GetNamespace() == namespace {
				return child
			}
		}
	}

	return nil
}

func getDeletableChild(tester *E2ETest) client.Object {
	for _, whitelistKind := range deletableWhitelist {
		for _, child := range tester.children {
			if child.GetObjectKind().GroupVersionKind().Kind == whitelistKind {
				return child
			}
		}
	}

	return nil
}

func getResourceGVR(resource client.Object) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    resource.GetObjectKind().GroupVersionKind().Group,
		Version:  resource.GetObjectKind().GroupVersionKind().Version,
		Resource: getPlural(strings.ToLower(resource.GetObjectKind().GroupVersionKind().Kind)),
	}
}

func getClientForResource(tester *E2ETest, resource client.Object) dynamic.ResourceInterface {
	if tester.namespace != "" {
		return tester.suiteConfig.dynamicClient.Resource(getResourceGVR(resource)).
			Namespace(tester.namespace)
	}

	return tester.suiteConfig.dynamicClient.Resource(getResourceGVR(resource)).
		Namespace(resource.GetNamespace())
}

func getControllerDeployment(s *E2ETestSuiteConfig) (*appsv1.Deployment, error) {
	return s.client.
		AppsV1().Deployments(s.controllerConfig.Namespace).
		Get(context.TODO(), (s.controllerConfig.Prefix + controllerName), metav1.GetOptions{})
}

func createCustomResource(tester *E2ETest) error {
	_, err := getClientForResource(tester, tester.unstructured).
		Create(context.TODO(), tester.unstructured, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating custom resource: %+v; %w", tester.unstructured, err)
	}

	return waitForCustomResource(tester)
}

func createNamespaceForTest(tester *E2ETest) error {
	namespaceExists, err := namespaceExists(tester)
	if namespaceExists || err != nil {
		return err
	}

	_, err = tester.suiteConfig.client.
		CoreV1().Namespaces().
		Create(
			context.TODO(),
			newNamespaceStub(tester.namespace),
			metav1.CreateOptions{},
		)

	return err
}

func getResource(tester *E2ETest, resource client.Object) (client.Object, error) {
	clusterObject, err := getClientForResource(tester, resource).
		Get(context.TODO(), resource.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get resource from cluster: %v; %w", clusterObject, err)
	}

	return clusterObject, nil
}

func getControllerLogs(s *E2ETestSuiteConfig) (string, error) {
	deployment, err := getControllerDeployment(s)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve controller deployment; %w", err)
	}

	podListOpts := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(deployment.Spec.Template.Labels).String(),
	}

	controllerPods, err := s.client.CoreV1().Pods(s.controllerConfig.Namespace).List(context.TODO(), podListOpts)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve controller pods; %w", err)
	}

	buf := new(bytes.Buffer)

	for _, pod := range controllerPods.Items {
		for _, container := range pod.Spec.Containers {
			podLogOpts := v1.PodLogOptions{Container: container.Name}
			req := s.client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)

			podLogs, err := req.Stream(context.TODO())
			if err != nil {
				return "", fmt.Errorf("error opening log stream for pod %s/%s; %w", pod.Namespace, pod.Name, err)
			}

			defer podLogs.Close()

			_, err = io.Copy(buf, podLogs)
			if err != nil {
				return "", fmt.Errorf("error storing logs to string buffer; %w", err)
			}
		}
	}

	return buf.String(), nil
}

func updateResource(tester *E2ETest, resource client.Object) error {
	unstructuredResource, err := resources.ToUnstructured(resource)
	if err != nil {
		return err
	}

	_, err = getClientForResource(tester, resource).
		Update(context.TODO(), unstructuredResource, metav1.UpdateOptions{})

	return err
}

func deleteResource(tester *E2ETest, resource client.Object) error {
	return getClientForResource(tester, resource).
		Delete(context.TODO(), resource.GetName(), metav1.DeleteOptions{})
}

func deleteCustomResource(tester *E2ETest) error {
	crClient := getClientForResource(tester, tester.unstructured)

	_, err := crClient.Get(context.TODO(), tester.unstructured.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if err := crClient.Delete(context.TODO(), tester.unstructured.GetName(), metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("error deleting custom resource: %+v; %w", tester.unstructured, err)
	}

	return waitForMissingResources(tester)
}

func deleteNamespaceForTest(tester *E2ETest) error {
	err := tester.suiteConfig.client.
		CoreV1().Namespaces().
		Delete(context.TODO(), tester.namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	namespaceIsMissing := func() (bool, error) {
		namespaceExists, err := namespaceExists(tester)
		if err != nil {
			return false, err
		}

		return !namespaceExists, nil
	}

	return waitFor(namespaceIsMissing)
}

func waitForMissingResources(tester *E2ETest) error {
	// wait for the resources to be missing
	childResourcesAreMissing := func() (bool, error) {
		for _, child := range tester.children {
			_, err := getClientForResource(tester, child).
				Get(context.TODO(), child.GetName(), metav1.GetOptions{})

			// we expect an IsNotFound error
			if err == nil {
				return false, nil
			}

			if errors.IsNotFound(err) {
				continue
			}
			return false, err
		}

		return true, nil
	}

	return waitFor(childResourcesAreMissing)
}

func waitForEqualResources(tester *E2ETest, resource client.Object) error {
	// wait for the resources to be equal
	childResourceIsEqual := func() (bool, error) {
		childResourceClusterObject, err := getClientForResource(tester, resource).
			Get(context.TODO(), resource.GetName(), metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("unable to get child resource from cluster: %+v; %w", resource, err)
		}

		// return equality statue of resource
		return resources.AreEqual(resource, childResourceClusterObject)
	}

	return waitFor(childResourceIsEqual)
}

func waitForChildResources(tester *E2ETest) error {
	// wait for the resources to be ready
	childResourcesAreReady := func() (bool, error) {
		childResourceClusterObjects := make([]client.Object, len(tester.children))
		for i, child := range tester.children {
			childResourceClusterObject, err := getClientForResource(tester, child).
				Get(context.TODO(), child.GetName(), metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("unable to get child resource from cluster: %+v; %w", child, err)
			}

			childResourceClusterObjects[i] = childResourceClusterObject
		}

		// get the ready status of the resources
		return resources.AreReady(childResourceClusterObjects...)
	}

	return waitFor(childResourcesAreReady)
}

func waitForCustomResource(tester *E2ETest) error {
	customResourceIsReady := func() (bool, error) {
		customResource, err := getClientForResource(tester, tester.unstructured).
			Get(context.TODO(), tester.unstructured.GetName(), metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("unable to get custom resource from cluster: %+v; %w", customResource, err)
		}

		// get the created status of the resource
		if customResource.Object["status"] == nil {
			return false, nil
		}

		createStatus := customResource.Object["status"].(map[string]interface{})["created"]
		if createStatus != nil {
			created, ok := createStatus.(bool)
			if !ok {
				return false, fmt.Errorf("unable to determine custom resource status")
			}

			return created, nil
		}

		return false, nil
	}

	return waitFor(customResourceIsReady)
}

func waitForController(s *E2ETestSuiteConfig) error {
	deploymentIsReady := func() (bool, error) {
		deployment, err := s.client.
			AppsV1().Deployments(s.controllerConfig.Namespace).
			Get(context.TODO(), (s.controllerConfig.Prefix + controllerName), metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		return resources.IsReady(deployment)
	}

	return waitFor(deploymentIsReady)
}

func waitFor(isReady readyChecker) error {
	timeout, interval := time.After(waitTimeout), time.Tick(waitInterval)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for resource")
		case <-interval:
			ready, err := isReady()
			if err != nil {
				return fmt.Errorf("error waiting for resource to be ready, %w", err)
			}

			if ready {
				return nil
			}
		}
	}
}

//
// tests
//
func testCreateCustomResource(tester *E2ETest) error {
	_, err := getClientForResource(tester, tester.unstructured).
		Create(context.TODO(), tester.unstructured, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating custom resource: %+v; %w", tester.unstructured, err)
	}

	// ensure the status ready field gets set
	if err = waitForCustomResource(tester); err != nil {
		return fmt.Errorf("failed waiting for custom resource ready status: %v; %w", tester.unstructured, err)
	}

	// double-check that the child resources are ready
	if err = waitForChildResources(tester); err != nil {
		return fmt.Errorf("child resources are not in a ready state: %v; %w", tester.unstructured, err)
	}

	return nil
}

func testDeleteChildResource(tester *E2ETest) error {
	childToDelete := getDeletableChild(tester)
	if childToDelete != nil {
		// delete the child resource
		if err := deleteResource(tester, childToDelete); err != nil {
			return fmt.Errorf("failed deleting child resource;: %+v; %w", childToDelete, err)
		}

		// wait for the child resource to return
		if err := waitForChildResources(tester); err != nil {
			return fmt.Errorf(
				"failed waiting for reconciliation after child deletion for resource: %+v; %w",
				childToDelete,
				err,
			)
		}
	}

	return nil
}

func testUpdateParentResource(tester *E2ETest, desiredStateChild client.Object) error {
	if desiredStateChild != nil {
		// update the parent resource
		if err := updateResource(tester, tester.workload); err != nil {
			return fmt.Errorf("failed updating parent resource;: %+v; %w", tester.workload, err)
		}

		// wait for the child resource to be equal
		if err := waitForEqualResources(tester, desiredStateChild); err != nil {
			return fmt.Errorf(
				"failed waiting for reconciliation after child update for resource: %+v; %w",
				desiredStateChild,
				err,
			)
		}
	}

	return nil
}

func testUpdateChildResource(tester *E2ETest, childToUpdate, desiredStateChild client.Object) error {
	if childToUpdate != nil {
		// update the child resource
		if err := updateResource(tester, childToUpdate); err != nil {
			return fmt.Errorf("failed updating child resource: %+v; %w", childToUpdate, err)
		}

		// wait for the child resource to be equal
		if err := waitForEqualResources(tester, desiredStateChild); err != nil {
			return fmt.Errorf(
				"failed waiting for reconciliation after child update for resource: %+v; %w",
				childToUpdate,
				err,
			)
		}
	}

	return nil
}

func testControllerLogsNoErrors(s *E2ETestSuiteConfig, searchSyntax string) error {
	logs, err := getControllerLogs(s)
	if err != nil {
		return fmt.Errorf("failed fetching controller logs; %w", err)
	}

	errors := []string{}

	for _, logLine := range strings.Split(logs, "\n") {
		if strings.Contains(logLine, "ERROR") && strings.Contains(logLine, searchSyntax) {
			errors = append(errors, logLine)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("found errors in controller: +%v", errors)
	}

	return nil
}
