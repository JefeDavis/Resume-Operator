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

package init

import (
	"fmt"

	"github.com/spf13/cobra"
)

type InitFunc func(*InitSubCommand) error

type InitSubCommand struct {
	*cobra.Command

	// flags
	APIVersion   string
	RequiredOnly bool

	// options
	Name         string
	Description  string
	SubCommandOf *cobra.Command

	InitFunc InitFunc
}

// NewBaseInitSubCommand returns a subcommand that is meant to belong to a parent
// subcommand but have subcommands itself.
func NewBaseInitSubCommand(parentCommand *cobra.Command) *InitSubCommand {
	initCmd := &InitSubCommand{
		Name:         "init",
		Description:  "write a sample custom resource manifest for a workload to standard out",
		SubCommandOf: parentCommand,
	}

	initCmd.Setup()

	return initCmd
}

// Setup sets up this command to be used as a command.
func (i *InitSubCommand) Setup() {
	i.Command = &cobra.Command{
		Use:   i.Name,
		Short: i.Description,
		Long:  i.Description,
	}

	// run the initialize function if the function signature is set
	if i.InitFunc != nil {
		i.RunE = i.initialize
	}

	// always add the api-version flag
	i.Flags().StringVarP(
		&i.APIVersion,
		"api-version",
		"",
		"",
		"api version of the workload to generate a workload manifest for",
	)

	// always add the required-only flag
	i.Flags().BoolVarP(
		&i.RequiredOnly,
		"required-only",
		"r",
		false,
		"only print required fields in the manifest output",
	)

	// add this as a subcommand of another command if set
	if i.SubCommandOf != nil {
		i.SubCommandOf.AddCommand(i.Command)
	}
}

// GetParent is a convenience function written when the CLI code is scaffolded
// to return the parent command and avoid scaffolding code with bad imports.
func GetParent(c interface{}) *cobra.Command {
	switch subcommand := c.(type) {
	case *InitSubCommand:
		return subcommand.Command
	case *cobra.Command:
		return subcommand
	}

	panic(fmt.Sprintf("subcommand is not proper type: %T", c))
}

// initialize creates sample workload manifests for a workload's custom resource.
func (i *InitSubCommand) initialize(cmd *cobra.Command, args []string) error {
	return i.InitFunc(i)
}
