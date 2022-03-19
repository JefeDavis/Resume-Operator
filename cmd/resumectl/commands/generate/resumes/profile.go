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

package resumes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// common imports for subcommands
	cmdgenerate "github.com/jefedavis/resume-operator/cmd/resumectl/commands/generate"
	// specific imports for workloads
	v1alpha1profile "github.com/jefedavis/resume-operator/apis/resumes/v1alpha1/resume"
	//+kubebuilder:scaffold:operator-builder:imports
)

// NewProfileSubCommand creates a new command and adds it to its
// parent command.
func NewProfileSubCommand(parentCommand *cobra.Command) {
	generateCmd := &cmdgenerate.GenerateSubCommand{
		Name:                  "collection",
		Description:           "Manage profile workload",
		SubCommandOf:          parentCommand,
		GenerateFunc:          GenerateProfile,
		UseCollectionManifest: true,
		CollectionKind:        "Profile"}

	generateCmd.Setup()
}

// GenerateProfile runs the logic to generate child resources for a
// Profile workload.
func GenerateProfile(g *cmdgenerate.GenerateSubCommand) error {
	var apiVersion string

	collectionFilename, _ := filepath.Abs(g.CollectionManifest)
	collectionFile, err := os.ReadFile(collectionFilename)
	if err != nil {
		return fmt.Errorf("failed to open collection file %s, %w", collectionFile, err)
	}

	var collection map[string]interface{}

	if err := yaml.Unmarshal(collectionFile, &collection); err != nil {
		return fmt.Errorf("failed to unmarshal yaml into collection, %w", err)
	}

	collectionGroupVersion := strings.Split(collection["apiVersion"].(string), "/")
	collectionAPIVersion := collectionGroupVersion[len(collectionGroupVersion)-1]

	apiVersion = collectionAPIVersion

	// generate a map of all versions to generate functions for each api version created
	type generateFunc func([]byte) ([]client.Object, error)
	generateFuncMap := map[string]generateFunc{
		"v1alpha1": v1alpha1profile.GenerateForCLI,
		//+kubebuilder:scaffold:operator-builder:versionmap
	}

	generate := generateFuncMap[apiVersion]
	resourceObjects, err := generate(collectionFile)
	if err != nil {
		return fmt.Errorf("unable to retrieve resources; %w", err)
	}

	e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

	outputStream := os.Stdout

	for _, o := range resourceObjects {
		if _, err := outputStream.WriteString("---\n"); err != nil {
			return fmt.Errorf("failed to write output, %w", err)
		}

		if err := e.Encode(o, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output, %w", err)
		}
	}

	return nil
}
