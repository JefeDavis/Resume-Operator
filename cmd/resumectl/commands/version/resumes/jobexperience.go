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
	"github.com/spf13/cobra"

	cmdversion "github.com/jefedavis/resume-operator/cmd/resumectl/commands/version"

	"github.com/jefedavis/resume-operator/apis/resumes"
)

// NewJobExperienceSubCommand creates a new command and adds it to its
// parent command.
func NewJobExperienceSubCommand(parentCommand *cobra.Command) {
	versionCmd := &cmdversion.VersionSubCommand{
		Name:         "job-experience",
		Description:  "Manage resume job experience component",
		VersionFunc:  VersionJobExperience,
		SubCommandOf: parentCommand,
	}

	versionCmd.Setup()
}

func VersionJobExperience(v *cmdversion.VersionSubCommand) error {
	apiVersions := make([]string, len(resumes.JobExperienceGroupVersions()))

	for i, groupVersion := range resumes.JobExperienceGroupVersions() {
		apiVersions[i] = groupVersion.Version
	}

	versionInfo := cmdversion.VersionInfo{
		CLIVersion:  cmdversion.CLIVersion,
		APIVersions: apiVersions,
	}

	return versionInfo.Display()
}
