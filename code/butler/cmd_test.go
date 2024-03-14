// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

// cSpell:ignore simplejs curr

package butler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"selinc.com/butler/code/butler"

	. "github.com/smartystreets/goconvey/convey"
)

var currBranch = func() string {
	b, err := butler.GetCurrentBranch()
	if err != nil {
		t := testing.T{}
		t.Fatal(err)
	}

	return b
}()

func replaceStubs() (undo func()) {
	originalExecOutputStub := (*exec.Cmd).Output
	butler.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) { return cmd.Output() }

	originalExecStartStub := (*exec.Cmd).Start
	butler.ExecStartStub = func(cmd *exec.Cmd) error { return cmd.Start() }

	originalExecWaitStub := (*exec.Cmd).Wait
	butler.ExecWaitStub = func(cmd *exec.Cmd) error { return cmd.Wait() }

	originalReadStub := (io.Reader).Read
	butler.ReadStub = func(reader io.Reader, buffer []byte) (int, error) {
		n, err := reader.Read(buffer)
		return n, err
	}

	return func() {
		butler.ExecOutputStub = originalExecOutputStub
		butler.ExecStartStub = originalExecStartStub
		butler.ExecWaitStub = originalExecWaitStub
		butler.ReadStub = originalReadStub
		_ = os.Remove(butler.ButlerResultsPath)
	}
}

func Test_RunWithErr(t *testing.T) {
	Convey("Just running the command for outer coverage of Execute", t, func() {
		undo := replaceStubs()
		defer undo()

		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		butler.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{butler.GitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			return nil, nil
		}
		butler.Execute()

		// Success determined by existence of the results json file.
		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("Running command with env vars enabled", t, func() {
		os.Setenv(butler.EnvRunAll, "true")
		originalEnvBranch := butler.GetEnvOrDefault(butler.EnvBranch, "")
		os.Setenv(butler.EnvBranch, strings.TrimSpace(currBranch))
		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		butler.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{butler.GitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			return nil, nil
		}
		butler.Execute()

		os.Unsetenv(butler.EnvRunAll)
		os.Setenv(butler.EnvBranch, originalEnvBranch)

		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("Running command with no GIT_BRANCH env var", t, func() {
		originalEnvBranch := butler.GetEnvOrDefault(butler.EnvBranch, "")
		os.Unsetenv(butler.EnvBranch)
		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		butler.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{butler.GitCommand, "branch", "--show-current"}) {
				gitBranchReturn, _ := json.Marshal(currBranch)
				return gitBranchReturn, nil
			}
			return nil, nil
		}
		butler.Execute()

		os.Setenv(butler.EnvBranch, originalEnvBranch)

		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("Running command with git turned off", t, func() {
		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.no_git.yaml", "--all"})
		butler.Execute()

		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("config parse fails due to bad path", t, func() {
		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/invalid.butler.bad"})
		butler.Execute()

		// butler_results.json should still exist despite error
		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "Error: stat ./test_data/test_helpers/invalid.butler.bad: no such file or directory")
	})

	Convey("config parse fails from path being invalid.", t, func() {
		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/bad_configs"})
		butler.Execute()

		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "Error: read ./test_data/bad_configs: is a directory")
	})

	Convey(".butler.ignore parse fails from inability to read file", t, func() {
		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/bad_configs/ignore_dir/.butler.base.yaml"})
		butler.Execute()

		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "Error: read test_data/bad_configs/ignore_dir/.butler.ignore.yaml: is a directory")
	})

	Convey(".butler.ignore parse fails due to bad syntax", t, func() {
		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/bad_configs/.butler.base.yaml"})
		butler.Execute()

		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "cannot unmarshal !!map into []string")
	})

	Convey("Butler setup fails when git fails", t, func() {
		undo := replaceStubs()
		defer undo()

		originalEnvBranch := butler.GetEnvOrDefault(butler.EnvBranch, "")
		os.Unsetenv(butler.EnvBranch)

		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		butler.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{butler.GitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			if reflect.DeepEqual(cmd.Args, []string{butler.GitCommand, "branch", "--show-current"}) {
				return nil, errors.New("error getting git branch")
			}
			return nil, nil
		}
		butler.Execute()

		os.Setenv(butler.EnvBranch, originalEnvBranch)

		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "error getting git branch")
	})

	Convey("Butler setup fails when git diff fails", t, func() {
		undo := replaceStubs()
		defer undo()

		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		butler.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{butler.GitCommand, "diff", "--name-only", currBranch}) {
				return nil, errors.New("git diff failed")
			}
			return nil, nil
		}
		butler.Execute()

		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "git diff failed")
	})

	Convey("Butler setup fails when preliminary command fails", t, func() {
		undo := replaceStubs()
		defer undo()

		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.bad_command.yaml"})
		butler.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "command"}) {
				return nil, errors.New("test command failed")
			}
			return nil, nil
		}
		butler.Execute()
		So(stderr.String(), ShouldContainSubstring, "test command failed")
	})

	Convey("Butler failed when unsupported language supplied", t, func() {
		undo := replaceStubs()
		defer undo()

		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.invalid_lang.yaml"})

		butler.Execute()
		So(stderr.String(), ShouldContainSubstring, "Error: error getting language id for invalid: Language not found")
	})

	Convey("Butler fails when dependency parsing fails", t, func() {
		undo := replaceStubs()
		defer undo()

		butler.Cmd = butler.GetCommand()
		stderr := new(bytes.Buffer)
		butler.Cmd.SetErr(stderr)
		butler.Cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.user_command.yaml"})
		butler.ExecOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{butler.GitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			return nil, nil
		}
		butler.ExecStartStub = func(cmd *exec.Cmd) error {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "command"}) {
				return errors.New("test command start failed")
			}
			return nil
		}
		butler.Execute()

		_, err := os.Stat(butler.ButlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "Error: error starting execution of 'fail command': test command start failed\n\n")
	})
}
