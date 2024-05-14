// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/cobra"
)

func Test_createTasks(t *testing.T) {
	Convey("Coverage for createTasks", t, func() {
		inputLanguage := &Language{
			Name: "golang",
			TaskCmd: &TaskCommands{
				Lint:    "echo go lint command",
				Test:    "echo go test command",
				Build:   "echo go build command",
				Publish: "echo go publish command",
			},
			DepCommands: &DependencyCommands{
				External: "echo go external dep command",
			},
			Workspaces: []*Workspace{
				{Location: "./test_data/test_repo/go_test", IsDirty: true, Dependencies: []string{}},
				{Location: "./test_data/test_repo", IsDirty: true, Dependencies: []string{}},
				{Location: "./test_data", IsDirty: true, Dependencies: []string{}},
			},
		}
		testQueue := &Queue{}

		inputLanguage.createTasks(testQueue, BuildStepTest, inputLanguage.TaskCmd.Lint, false)
		So(inputLanguage.Workspaces, ShouldNotBeNil)
	})
}

func Test_runCommandList(t *testing.T) {
	inputLanguage := &Language{
		Name: "test",
		TaskCmd: &TaskCommands{
			Lint:  "echo lint command",
			SetUp: []string{""},
		},
	}
	cmd := &cobra.Command{}
	Convey("empty command passed", t, func() {
		langs := []*Language{inputLanguage}

		stdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := langs[0].runCommandList(cmd, langs[0].TaskCmd.SetUp)

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = stdout

		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "empty command, skipping")
	})

	Convey("Handle command failure", t, func() {
		undo := replaceStubs()
		defer undo()

		inputLanguage.TaskCmd.SetUp = []string{"fail command"}
		langs := []*Language{inputLanguage}
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if len(cmd.Args) == 2 && cmd.Args[0] == "fail" {
				return nil, errors.New("test command failed")
			}
			return nil, nil
		}

		err := langs[0].runCommandList(cmd, langs[0].TaskCmd.SetUp)
		So(err.Error(), ShouldContainSubstring, "error executing 'fail command'\nerror: test command failed\noutput:")
	})
}

func Test_executeUserMethods(t *testing.T) {
	testLangName := "test"
	Convey("error on empty user method", t, func() {
		_, err := ExecuteUserMethods("", testLangName)
		So(err.Error(), ShouldResemble, "dependency commands not supplied for the language test")
	})

	Convey("error on user command start", t, func() {
		undo := replaceStubs()
		defer undo()

		execStartStub = func(cmd *exec.Cmd) error {
			if len(cmd.Args) == 2 && cmd.Args[0] == "fail" {
				return errors.New("test command start failed")
			}
			return nil
		}
		_, err := ExecuteUserMethods("fail command", testLangName)
		So(err.Error(), ShouldResemble, "error starting execution of 'fail command': test command start failed")
	})

	Convey("error on stdout read", t, func() {
		undo := replaceStubs()
		defer undo()
		execStartStub = func(cmd *exec.Cmd) error {
			return nil
		}
		readStub = func(reader io.Reader, buffer []byte) (int, error) {
			return 0, errors.New("failed to read")
		}
		_, err := ExecuteUserMethods("fail command", testLangName)
		So(err.Error(), ShouldResemble, "error executing 'fail command': failed to read")
	})

	Convey("error on user command wait", t, func() {
		undo := replaceStubs()
		defer undo()

		execStartStub = func(cmd *exec.Cmd) error {
			return nil
		}
		readStub = func(reader io.Reader, buffer []byte) (int, error) {
			return 0, nil
		}
		execWaitStub = func(cmd *exec.Cmd) error {
			if len(cmd.Args) == 2 && cmd.Args[0] == "fail" {
				return errors.New("test command wait failed")
			}
			return nil
		}
		_, err := ExecuteUserMethods("fail command", testLangName)
		So(err.Error(), ShouldResemble, "error executing 'fail command': test command wait failed")
	})
}

func Test_validateLanguage(t *testing.T) {
	type template struct {
		desc          string
		language      *Language
		expectedError error
	}
	testConfig := &ButlerConfig{
		Task: &TaskConfigurations{},
	}
	tests := []template{
		{"language config validation passes", &Language{Name: "test", WorkspaceFiles: []string{".test"}}, nil},
		{"fails when a name id is not provided", &Language{WorkspaceFiles: []string{".test"}}, errors.New("a language was supplied in the config without a name. Please supply a language identifier for each language in the config")},
		{"fails when no file patterns are supplied", &Language{Name: "test"}, errors.New("no file patterns supplied for 'test'. Please see the 'workspaceFiles' options in the language_options.md spec for more information")},
	}

	for _, test := range tests {
		Convey(test.desc, t, func() {
			So(test.language.validateLanguage(testConfig), ShouldResemble, test.expectedError)
		})
	}
}
