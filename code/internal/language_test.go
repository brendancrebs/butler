// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_createTasks(t *testing.T) {
	Convey("Coverage for createTasks", t, func() {
		inputLanguage := &Language{
			Name: "golang",
			TaskExec: &TaskCommands{
				LintCommand:    "echo go lint command",
				TestCommand:    "echo go test command",
				BuildCommand:   "echo go build command",
				PublishCommand: "echo go publish command",
			},
			DepCommands: &DependencyCommands{
				ExternalDepCommand: "echo go external dep command",
			},
			Workspaces: []*Workspace{
				{Location: "./test_data/test_repo/go_test", Name: "go_test", IsDirty: true, WorkspaceDependencies: []string{}},
				{Location: "./test_data/test_repo", Name: "go_test2", IsDirty: true, WorkspaceDependencies: []string{}},
				{Location: "./test_data", Name: "go_test3", IsDirty: true, WorkspaceDependencies: []string{}},
			},
		}
		testQueue := &Queue{tasks: make([]*Task, 0)}

		err := createTasks(inputLanguage, testQueue, BuildStepTest, inputLanguage.TaskExec.LintCommand, inputLanguage.TaskExec.LintPath)
		So(err, ShouldBeNil)
		So(inputLanguage.Workspaces, ShouldNotBeNil)
	})
}
