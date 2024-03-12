// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal_test

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"selinc.com/butler/code/internal"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_createTasks(t *testing.T) {
	Convey("Coverage for createTasks", t, func() {
		inputLanguage := &internal.Language{
			Name: "golang",
			TaskExec: &internal.TaskCommands{
				LintCommand:    "echo go lint command",
				TestCommand:    "echo go test command",
				BuildCommand:   "echo go build command",
				PublishCommand: "echo go publish command",
			},
			DepCommands: &internal.DependencyCommands{
				ExternalDepCommand: "echo go external dep command",
			},
			Workspaces: []*internal.Workspace{
				{Location: "./test_data/test_repo/go_test", Name: "go_test", IsDirty: true, Dependencies: []string{}},
				{Location: "./test_data/test_repo", Name: "go_test2", IsDirty: true, Dependencies: []string{}},
				{Location: "./test_data", Name: "go_test3", IsDirty: true, Dependencies: []string{}},
			},
		}
		testQueue := &internal.Queue{Tasks: make([]*internal.Task, 0)}

		inputLanguage.CreateTasks(testQueue, internal.BuildStepTest, inputLanguage.TaskExec.LintCommand, inputLanguage.TaskExec.LintPath)
		So(inputLanguage.Workspaces, ShouldNotBeNil)
	})
}

func Test_preliminaryCommands(t *testing.T) {
	inputLanguage := &internal.Language{
		Name: "test",
		TaskExec: &internal.TaskCommands{
			LintCommand:   "echo lint command",
			SetUpCommands: []string{""},
		},
	}
	Convey("empty command passed", t, func() {
		langs := []*internal.Language{inputLanguage}

		stdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := internal.PreliminaryCommands(langs)

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = stdout

		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "empty command, skipping")
	})

	Convey("Handle command failure", t, func() {
		undo := replaceStubs()
		defer undo()

		inputLanguage.TaskExec.SetUpCommands = []string{"fail command"}
		langs := []*internal.Language{inputLanguage}
		internal.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "command"}) {
				return nil, errors.New("test command failed")
			}
			return nil, nil
		}

		err := internal.PreliminaryCommands(langs)
		So(err.Error(), ShouldContainSubstring, "error executing 'fail command':\nerror: test command failed\noutput:")
	})
}

func Test_executeUserMethods(t *testing.T) {
	testPath := "."
	testLangName := "test"
	Convey("error on empty user method", t, func() {
		_, err := internal.ExecuteUserMethods("", testPath, testLangName)
		So(err.Error(), ShouldResemble, "dependency commands not supplied for the language test.\n")
	})

	Convey("error on user command start", t, func() {
		undo := replaceStubs()
		defer undo()

		internal.ExecStartStub = func(cmd *exec.Cmd) error {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "command"}) {
				return errors.New("test command start failed")
			}
			return nil
		}
		_, err := internal.ExecuteUserMethods("fail command", testPath, testLangName)
		So(err.Error(), ShouldResemble, "error starting execution of 'fail command': test command start failed\n")
	})

	Convey("error on stdout read", t, func() {
		undo := replaceStubs()
		defer undo()
		internal.ExecStartStub = func(cmd *exec.Cmd) error {
			return nil
		}
		internal.ReadStub = func(reader io.Reader, buffer []byte) (int, error) {
			return 0, errors.New("failed to read")
		}
		_, err := internal.ExecuteUserMethods("fail command", testPath, testLangName)
		So(err.Error(), ShouldResemble, "error executing 'fail command': failed to read\n")
	})

	Convey("error on user command wait", t, func() {
		undo := replaceStubs()
		defer undo()

		internal.ExecStartStub = func(cmd *exec.Cmd) error {
			return nil
		}
		internal.ReadStub = func(reader io.Reader, buffer []byte) (int, error) {
			return 0, nil
		}
		internal.ExecWaitStub = func(cmd *exec.Cmd) error {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "command"}) {
				return errors.New("test command wait failed")
			}
			return nil
		}
		_, err := internal.ExecuteUserMethods("fail command", testPath, testLangName)
		So(err.Error(), ShouldResemble, "error executing 'fail command': test command wait failed\n")
	})
}
