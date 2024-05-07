// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"encoding/json"
	"fmt"
)

// BuildStatus is the overall status values for the build.
type BuildStatus int

// BuildStatus enum values.
const (
	BuildStatusUnknown BuildStatus = iota
	BuildStatusClean               // First time pass on all tasks and no errors.
	BuildStatusDirty               // It took more than one attempt on at least one task.
	BuildStatusFail                // One or more tasks contain an error.
)

// String returns the string equivalent of the BuildStatus.
func (status BuildStatus) String() string {
	if status < BuildStatusUnknown || status > BuildStatusFail {
		status = BuildStatusUnknown
	}
	return []string{"Unknown", "Clean", "Dirty", "Fail"}[status]
}

// BuildStep governs the priority that a task is run at.
type BuildStep int

// BuildStep enum values.
const (
	BuildStepUnknown BuildStep = iota
	BuildStepLint              // language level linting: spelling and static analysis.
	BuildStepTest              // unit tests and coverage.
	BuildStepBuild             // build, package and deployment steps.
	BuildStepPublish           // push the results to Artifactory, K: drive, etc.
)

// String returns the string equivalent of the BuildStep.
func (step BuildStep) String() string {
	if step < BuildStepUnknown || step > BuildStepPublish {
		step = BuildStepUnknown
	}
	return []string{"Unknown", "Lint", "Test", "Build", "Publish"}[step]
}

// MarshalJSON marshals the enum as a quoted json string.
func (step BuildStep) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", step)), nil
}

// UnmarshalJSON un-marshals a quoted json string to the enum value.
func (step *BuildStep) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value.
	*step = toBuildStep[j]
	return nil
}

// maps the build steps to the build commands for a language.
func getBuildCommands(lang *Language) map[BuildStep]string {
	return map[BuildStep]string{
		BuildStepUnknown: "",
		BuildStepLint:    lang.TaskExec.Lint,
		BuildStepTest:    lang.TaskExec.Test,
		BuildStepBuild:   lang.TaskExec.Build,
		BuildStepPublish: lang.TaskExec.Publish,
	}
}

var buildSteps = []BuildStep{BuildStepUnknown, BuildStepLint, BuildStepTest, BuildStepBuild, BuildStepPublish}

var toBuildStep = map[string]BuildStep{
	BuildStepUnknown.String(): BuildStepUnknown,
	BuildStepLint.String():    BuildStepLint,
	BuildStepTest.String():    BuildStepTest,
	BuildStepBuild.String():   BuildStepBuild,
	BuildStepPublish.String(): BuildStepPublish,
}
