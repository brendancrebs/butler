// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler_test

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"selinc.com/butler/code/butler"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_createTasks(t *testing.T) {
	Convey("Coverage for createTasks", t, func() {
		inputLanguage := &butler.Language{
			Name: "golang",
			TaskExec: &butler.TaskCommands{
				LintCommand:    "echo go lint command",
				TestCommand:    "echo go test command",
				BuildCommand:   "echo go build command",
				PublishCommand: "echo go publish command",
			},
			DepCommands: &butler.DependencyCommands{
				ExternalDepCommand: "echo go external dep command",
			},
			Workspaces: []*butler.Workspace{
				{Location: "./test_data/test_repo/go_test", Name: "go_test", IsDirty: true, Dependencies: []string{}},
				{Location: "./test_data/test_repo", Name: "go_test2", IsDirty: true, Dependencies: []string{}},
				{Location: "./test_data", Name: "go_test3", IsDirty: true, Dependencies: []string{}},
			},
		}
		testQueue := &butler.Queue{Tasks: make([]*butler.Task, 0)}

		inputLanguage.CreateTasks(testQueue, butler.BuildStepTest, inputLanguage.TaskExec.LintCommand, inputLanguage.TaskExec.LintPath)
		So(inputLanguage.Workspaces, ShouldNotBeNil)
	})
}

func Test_preliminaryCommands(t *testing.T) {
	inputLanguage := &butler.Language{
		Name: "test",
		TaskExec: &butler.TaskCommands{
			LintCommand:   "echo lint command",
			SetUpCommands: []string{""},
		},
	}
	Convey("empty command passed", t, func() {
		langs := []*butler.Language{inputLanguage}

		stdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := butler.PreliminaryCommands(langs)

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
		langs := []*butler.Language{inputLanguage}
		butler.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "command"}) {
				return nil, errors.New("test command failed")
			}
			return nil, nil
		}

		err := butler.PreliminaryCommands(langs)
		So(err.Error(), ShouldContainSubstring, "error executing 'fail command':\nerror: test command failed\noutput:")
	})
}

func Test_executeUserMethods(t *testing.T) {
	testPath := "."
	testLangName := "test"
	Convey("error on empty user method", t, func() {
		_, err := butler.ExecuteUserMethods("", testPath, testLangName)
		So(err.Error(), ShouldResemble, "dependency commands not supplied for the language test.\n")
	})

	Convey("error on user command start", t, func() {
		undo := replaceStubs()
		defer undo()

		butler.ExecStartStub = func(cmd *exec.Cmd) error {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "command"}) {
				return errors.New("test command start failed")
			}
			return nil
		}
		_, err := butler.ExecuteUserMethods("fail command", testPath, testLangName)
		So(err.Error(), ShouldResemble, "error starting execution of 'fail command': test command start failed\n")
	})

	Convey("error on stdout read", t, func() {
		undo := replaceStubs()
		defer undo()
		butler.ExecStartStub = func(cmd *exec.Cmd) error {
			return nil
		}
		butler.ReadStub = func(reader io.Reader, buffer []byte) (int, error) {
			return 0, errors.New("failed to read")
		}
		_, err := butler.ExecuteUserMethods("fail command", testPath, testLangName)
		So(err.Error(), ShouldResemble, "error executing 'fail command': failed to read\n")
	})

	Convey("error on user command wait", t, func() {
		undo := replaceStubs()
		defer undo()

		butler.ExecStartStub = func(cmd *exec.Cmd) error {
			return nil
		}
		butler.ReadStub = func(reader io.Reader, buffer []byte) (int, error) {
			return 0, nil
		}
		butler.ExecWaitStub = func(cmd *exec.Cmd) error {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "command"}) {
				return errors.New("test command wait failed")
			}
			return nil
		}
		_, err := butler.ExecuteUserMethods("fail command", testPath, testLangName)
		So(err.Error(), ShouldResemble, "error executing 'fail command': test command wait failed\n")
	})
}
