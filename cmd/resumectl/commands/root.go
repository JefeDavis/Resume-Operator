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

package commands

import (
	"github.com/spf13/cobra"

	// common imports for subcommands
	cmdgenerate "github.com/jefedavis/resume-operator/cmd/resumectl/commands/generate"
	cmdinit "github.com/jefedavis/resume-operator/cmd/resumectl/commands/init"
	cmdversion "github.com/jefedavis/resume-operator/cmd/resumectl/commands/version"

	// specific imports for workloads
	generateresumes "github.com/jefedavis/resume-operator/cmd/resumectl/commands/generate/resumes"
	initresumes "github.com/jefedavis/resume-operator/cmd/resumectl/commands/init/resumes"
	versionresumes "github.com/jefedavis/resume-operator/cmd/resumectl/commands/version/resumes"
	//+kubebuilder:scaffold:operator-builder:subcommands:imports
)

// ResumectlCommand represents the base command when called without any subcommands.
type ResumectlCommand struct {
	*cobra.Command
}

// NewResumectlCommand returns an instance of the ResumectlCommand.
func NewResumectlCommand() *ResumectlCommand {
	c := &ResumectlCommand{
		Command: &cobra.Command{
			Use:   "resumectl",
			Short: "Manage profile collection and components",
			Long:  "Manage profile collection and components",
		},
	}

	c.addSubCommands()

	return c
}

// Run represents the main entry point into the command
// This is called by main.main() to execute the root command.
func (c *ResumectlCommand) Run() {
	cobra.CheckErr(c.Execute())
}

func (c *ResumectlCommand) newInitSubCommand() {
	parentCommand := cmdinit.GetParent(cmdinit.NewBaseInitSubCommand(c.Command))
	_ = parentCommand

	// add the init subcommands
	initresumes.NewProfileSubCommand(parentCommand)
	initresumes.NewJobExperienceSubCommand(parentCommand)
	initresumes.NewCertificationSubCommand(parentCommand)
	//+kubebuilder:scaffold:operator-builder:subcommands:init
}

func (c *ResumectlCommand) newGenerateSubCommand() {
	parentCommand := cmdgenerate.GetParent(cmdgenerate.NewBaseGenerateSubCommand(c.Command))
	_ = parentCommand

	// add the generate subcommands
	generateresumes.NewProfileSubCommand(parentCommand)
	generateresumes.NewJobExperienceSubCommand(parentCommand)
	generateresumes.NewCertificationSubCommand(parentCommand)
	//+kubebuilder:scaffold:operator-builder:subcommands:generate
}

func (c *ResumectlCommand) newVersionSubCommand() {
	parentCommand := cmdversion.GetParent(cmdversion.NewBaseVersionSubCommand(c.Command))
	_ = parentCommand

	// add the version subcommands
	versionresumes.NewProfileSubCommand(parentCommand)
	versionresumes.NewJobExperienceSubCommand(parentCommand)
	versionresumes.NewCertificationSubCommand(parentCommand)
	//+kubebuilder:scaffold:operator-builder:subcommands:version
}

// addSubCommands adds any additional subCommands to the root command.
func (c *ResumectlCommand) addSubCommands() {
	c.newInitSubCommand()
	c.newGenerateSubCommand()
	c.newVersionSubCommand()
}
