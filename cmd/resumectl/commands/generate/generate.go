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

package generate

import (
	"fmt"

	"github.com/spf13/cobra"
)

type GenerateFunc func(*GenerateSubCommand) error

type GenerateSubCommand struct {
	*cobra.Command

	// flags
	WorkloadManifest   string
	CollectionManifest string
	APIVersion         string

	// options
	Name                  string
	Description           string
	CollectionKind        string
	UseCollectionManifest bool
	WorkloadKind          string
	UseWorkloadManifest   bool
	SubCommandOf          *cobra.Command

	// execution
	GenerateFunc GenerateFunc
}

// NewBaseGenerateSubCommand returns a subcommand that is meant to belong to a parent
// subcommand but have subcommands itself.
func NewBaseGenerateSubCommand(parentCommand *cobra.Command) *GenerateSubCommand {
	generateCmd := &GenerateSubCommand{
		Name:                  "generate",
		Description:           "generate child resource manifests from a workload's custom resource",
		UseCollectionManifest: false,
		UseWorkloadManifest:   false,
		SubCommandOf:          parentCommand,
	}

	generateCmd.Setup()

	return generateCmd
}

// Setup sets up this command to be used as a command.
func (g *GenerateSubCommand) Setup() {
	g.Command = &cobra.Command{
		Use:   g.Name,
		Short: g.Description,
		Long:  g.Description,
	}

	// run the generate function if the function signature is set
	if g.GenerateFunc != nil {
		g.RunE = g.generate
	}

	// add workload-manifest flag if this subcommand requests it
	if g.UseWorkloadManifest {
		g.Flags().StringVarP(
			&g.WorkloadManifest,
			"workload-manifest",
			"w",
			"",
			fmt.Sprintf("filepath to the %s workload manifest used to generate child resources", g.WorkloadKind),
		)

		if err := g.MarkFlagRequired("workload-manifest"); err != nil {
			panic(err)
		}
	}

	// add collection-manifest flag if this subcommand requests it
	if g.UseCollectionManifest {
		g.Command.Flags().StringVarP(
			&g.CollectionManifest,
			"collection-manifest",
			"c",
			"",
			fmt.Sprintf("filepath to the %s collection manifest used to generate child resources", g.CollectionKind),
		)

		if err := g.MarkFlagRequired("collection-manifest"); err != nil {
			panic(err)
		}
	}

	// add this as a subcommand of another command if set
	if g.SubCommandOf != nil {
		g.SubCommandOf.AddCommand(g.Command)
	}
}

// GetParent is a convenience function written when the CLI code is scaffolded
// to return the parent command and avoid scaffolding code with bad imports.
func GetParent(c interface{}) *cobra.Command {
	switch subcommand := c.(type) {
	case *GenerateSubCommand:
		return subcommand.Command
	case *cobra.Command:
		return subcommand
	}

	panic(fmt.Sprintf("subcommand is not proper type: %T", c))
}

// generate creates child resource manifests from a workload's custom resource.
func (g *GenerateSubCommand) generate(cmd *cobra.Command, args []string) error {
	return g.GenerateFunc(g)
}
