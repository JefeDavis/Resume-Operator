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

package version

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var CLIVersion = "dev"

type VersionInfo struct {
	CLIVersion  string   `json:"cliVersion"`
	APIVersions []string `json:"apiVersions"`
}

type VersionFunc func(*VersionSubCommand) error

type VersionSubCommand struct {
	*cobra.Command

	// options
	Name         string
	Description  string
	SubCommandOf *cobra.Command

	VersionFunc VersionFunc
}

// NewBaseVersionSubCommand returns a subcommand that is meant to belong to a parent
// subcommand but have subcommands itself.
func NewBaseVersionSubCommand(parentCommand *cobra.Command) *VersionSubCommand {
	versionCmd := &VersionSubCommand{
		Name:         "version",
		Description:  "display the version information",
		SubCommandOf: parentCommand,
	}

	versionCmd.Setup()

	return versionCmd
}

// Setup sets up this command to be used as a command.
func (v *VersionSubCommand) Setup() {
	v.Command = &cobra.Command{
		Use:   v.Name,
		Short: v.Description,
		Long:  v.Description,
	}

	// run the version function if the function signature is set
	if v.VersionFunc != nil {
		v.RunE = v.version
	}

	// add this as a subcommand of another command if set
	if v.SubCommandOf != nil {
		v.SubCommandOf.AddCommand(v.Command)
	}
}

// version run the function to display version information about a workload.
func (v *VersionSubCommand) version(cmd *cobra.Command, args []string) error {
	return v.VersionFunc(v)
}

// GetParent is a convenience function written when the CLI code is scaffolded
// to return the parent command and avoid scaffolding code with bad imports.
func GetParent(c interface{}) *cobra.Command {
	switch subcommand := c.(type) {
	case *VersionSubCommand:
		return subcommand.Command
	case *cobra.Command:
		return subcommand
	}

	panic(fmt.Sprintf("subcommand is not proper type: %T", c))
}

// Display will parse and print the information stored on the VersionInfo object.
func (v *VersionInfo) Display() error {
	output, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to determine versionInfo, %s", err)
	}

	outputStream := os.Stdout

	if _, err := outputStream.WriteString(fmt.Sprintln(string(output))); err != nil {
		return fmt.Errorf("failed to write to stdout, %s", err)
	}

	return nil
}
