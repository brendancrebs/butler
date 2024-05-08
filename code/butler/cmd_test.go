// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

// cSpell:ignore simplejs curr

package butler

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

	. "github.com/smartystreets/goconvey/convey"
)

var currBranch = func() string {
	b, err := getCurrentBranch()
	if err != nil {
		t := testing.T{}
		t.Fatal(err)
	}

	return b
}()

func replaceStubs() (undo func()) {
	originalExecOutputStub := (*exec.Cmd).Output
	execOutputStub = func(cmd *exec.Cmd) ([]byte, error) { return cmd.Output() }

	originalExecStartStub := (*exec.Cmd).Start
	execStartStub = func(cmd *exec.Cmd) error { return cmd.Start() }

	originalExecWaitStub := (*exec.Cmd).Wait
	execWaitStub = func(cmd *exec.Cmd) error { return cmd.Wait() }

	originalReadStub := (io.Reader).Read
	readStub = func(reader io.Reader, buffer []byte) (int, error) {
		n, err := reader.Read(buffer)
		return n, err
	}

	originalWd, _ := os.Getwd()

	return func() {
		execOutputStub = originalExecOutputStub
		execStartStub = originalExecStartStub
		execWaitStub = originalExecWaitStub
		readStub = originalReadStub
		_ = os.Chdir(originalWd)
		_ = os.Remove(butlerResultsPath)
	}
}

func Test_RunWithErr(t *testing.T) {
	Convey("Just running the command for outer coverage of Execute", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/.butler.base.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			return nil, nil
		}
		Execute()

		// Success determined by existence of the results json file.
		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("Running command with env vars enabled", t, func() {
		undo := replaceStubs()
		defer undo()

		os.Setenv(envRunAll, "true")
		originalEnvBranch := GetEnvOrDefault(envBranch, "")
		os.Setenv(envBranch, strings.TrimSpace(currBranch))
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--config", "./test_data/test_configs/.butler.base.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			return nil, nil
		}
		Execute()

		os.Unsetenv(envRunAll)
		os.Setenv(envBranch, originalEnvBranch)

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("Running command with no GIT_BRANCH env var", t, func() {
		undo := replaceStubs()
		defer undo()

		originalEnvBranch := GetEnvOrDefault(envBranch, "")
		os.Unsetenv(envBranch)
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--config", "./test_data/test_configs/.butler.base.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitCommand, "branch", "--show-current"}) {
				gitBranchReturn, _ := json.Marshal(currBranch)
				return gitBranchReturn, nil
			}
			return nil, nil
		}
		Execute()

		os.Setenv(envBranch, originalEnvBranch)

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("Running command with git turned off", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--config", "./test_data/test_configs/no_git.yaml", "--all"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("config parse fails due to bad path", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/.butler.invalid.bad"})
		Execute()

		// butler_results.json should still exist despite error
		_, err := os.Stat(butlerResultsPath)
		So(err.Error(), ShouldContainSubstring, "stat ./butler_results.json: no such file or directory")
		So(stderr.String(), ShouldContainSubstring, "Error: stat ./test_data/test_configs/.butler.invalid.bad: no such file or directory")
	})

	Convey("config parse fails from path being invalid.", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/path_configs"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err.Error(), ShouldContainSubstring, "stat ./butler_results.json: no such file or directory")
		So(stderr.String(), ShouldContainSubstring, "Error: read ./test_data/test_configs/path_configs: is a directory")
	})

	Convey("path config parse fails from inability to read file", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/path_configs/ignore_dir/.butler.base.yaml"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err.Error(), ShouldContainSubstring, "stat ./butler_results.json: no such file or directory")
		So(stderr.String(), ShouldContainSubstring, "Error: read test_data/test_configs/path_configs/ignore_dir/.butler.paths.yaml: is a directory")
	})

	Convey("path config parse fails due to bad syntax", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/path_configs/.butler.base.yaml"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err.Error(), ShouldContainSubstring, "stat ./butler_results.json: no such file or directory")
		So(stderr.String(), ShouldContainSubstring, "cannot unmarshal !!map into []string")
	})

	Convey("Butler setup fails when git fails", t, func() {
		undo := replaceStubs()
		defer undo()

		originalEnvBranch := GetEnvOrDefault(envBranch, "")
		os.Unsetenv(envBranch)

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/.butler.base.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			if reflect.DeepEqual(cmd.Args, []string{gitCommand, "branch", "--show-current"}) {
				return nil, errors.New("error getting git branch")
			}
			return nil, nil
		}
		Execute()

		os.Setenv(envBranch, originalEnvBranch)

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "error getting git branch")
	})

	Convey("Butler setup fails when git diff fails", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/.butler.base.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitCommand, "diff", "--name-only", currBranch}) {
				return nil, errors.New("git diff failed")
			}
			return nil, nil
		}
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "git diff failed")
	})

	Convey("Butler setup fails when preliminary command fails", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/bad_setup_command.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "setup"}) {
				return nil, errors.New("setup command failed")
			}
			return nil, nil
		}
		Execute()
		So(stderr.String(), ShouldContainSubstring, "setup command failed")
	})

	Convey("Butler cleanup fails when teardown command fails", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/bad_cleanup_command.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "cleanup"}) {
				return nil, errors.New("cleanup command failed")
			}
			return nil, nil
		}
		Execute()
		So(stderr.String(), ShouldContainSubstring, "cleanup command failed")
	})

	Convey("Butler failed when unsupported language supplied", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--config", "./test_data/test_configs/unknown_lang.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			return nil, nil
		}

		Execute()
		So(stderr.String(), ShouldContainSubstring, "Error: language id 'invalid' not found")
	})

	Convey("Butler fails when dependency parsing fails", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--config", "./test_data/test_configs/user_command_fail.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			return nil, nil
		}
		execStartStub = func(cmd *exec.Cmd) error {
			if reflect.DeepEqual(cmd.Args, []string{"fail", "command"}) {
				return errors.New("test command start failed")
			}
			return nil
		}
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "Error: error starting execution of 'fail command': test command start failed\n")
	})

	Convey("Butler successfully executes user commands", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/user_command.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			return nil, nil
		}
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("Butler failed when unnamed language supplied", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/invalid_language_1.yaml"})

		Execute()
		So(stderr.String(), ShouldContainSubstring, "Error: a language supplied in the config without a name. Please supply a language identifier for each language in the config")
	})

	Convey("Butler failed when no file pattern supplied", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/invalid_language_2.yaml"})

		Execute()
		So(stderr.String(), ShouldContainSubstring, "Error: no file patterns supplied for 'invalid'. Please see the 'FilePatterns' options in the config spec for more information")
	})

	Convey("Butler collects workspace dependencies when enabled", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--config", "./test_data/test_configs/workspace_deps.yaml"})
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitCommand, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"../test_data/test_repo/go_test/example.go"})
				return gitDiffReturn, nil
			}
			return nil, nil
		}
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "")
	})

	Convey("Butler runs without allowed paths", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--config", "./test_data/test_configs/no_allowed_paths.yaml"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "")
	})
}
