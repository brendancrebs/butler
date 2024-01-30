// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

// cSpell:ignore simplejs curr

package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"reflect"
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

	originalExecLookPathStub := exec.LookPath
	execLookPathStub = func(executable string) (string, error) { return exec.LookPath(executable) }

	return func() {
		execOutputStub = originalExecOutputStub
		execLookPathStub = originalExecLookPathStub
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
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		gitPath, _ := execLookPathStub(gitCommand)
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitPath, "diff", "--name-only", currBranch}) {
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

	Convey("Running command with BUTLER_SHOULD_RUN_ALL env var enabled", t, func() {
		os.Setenv("BUTLER_SHOULD_RUN_ALL", "true")
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("Running command with git turned off", t, func() {
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.no_git.yaml", "--all"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("config parse fails due to bad path", t, func() {
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/invalid.butler.bad"})
		Execute()

		// butler_results.json should still exist despite error
		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "Error: stat ./test_data/test_helpers/invalid.butler.bad: no such file or directory")
	})

	Convey("config parse fails from inability to read file", t, func() {
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/bad_configs/no_read_permissions/.butler.locked.yaml"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "Error: open ./test_data/bad_configs/no_read_permissions/.butler.locked.yaml: permission denied")
	})

	Convey(".butler.ignore parse fails from inability to read file", t, func() {
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/bad_configs/no_read_permissions/.butler.base.yaml"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "Error: open test_data/bad_configs/no_read_permissions/.butler.ignore.yaml: permission denied")
	})

	Convey(".butler.ignore parse fails due to bad syntax", t, func() {
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/bad_configs/.butler.base.yaml"})
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "cannot unmarshal !!map into []string")
	})

	Convey("Butler setup fails when git fails", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		gitPath, _ := execLookPathStub(gitCommand)
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitPath, "diff", "--name-only", currBranch}) {
				gitDiffReturn, _ := json.Marshal([]string{"testPath1", "testPath2"})
				return gitDiffReturn, nil
			}
			if reflect.DeepEqual(cmd.Args, []string{gitPath, "branch", "--show-current"}) {
				return nil, errors.New("error getting git branch")
			}
			return nil, nil
		}
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "error getting git branch")
	})

	Convey("Butler setup fails when git not installed", t, func() {
		undo := replaceStubs()
		defer undo()

		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		execLookPathStub = func(executable string) (string, error) { return "", errors.New("git executable not found") }
		Execute()

		_, err := os.Stat(butlerResultsPath)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "git executable not found")
	})
}
