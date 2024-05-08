// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package builtin

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testBranch        = "butler_unit_test_main"
	testWorkspaceRoot = "../test_data"
)

func replaceStubs() (undo func()) {
	originalExecOutputStub := (*exec.Cmd).Output
	execOutputStub = func(cmd *exec.Cmd) ([]byte, error) { return cmd.Output() }

	originalEnvBranch := os.Getenv(envBranch)
	originalWorkspaceRoot := os.Getenv(envWorkspaceRoot)
	os.Setenv(envBranch, testBranch)
	os.Setenv(envWorkspaceRoot, testWorkspaceRoot)

	return func() {
		execOutputStub = originalExecOutputStub
		os.Setenv(envBranch, originalEnvBranch)
		os.Setenv(envWorkspaceRoot, originalWorkspaceRoot)
	}
}

func Test_GetStdLibs(t *testing.T) {
	Convey("Returns a list of go standard libraries without error", t, func() {
		undo := replaceStubs()
		defer undo()

		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitName, "diff", testBranch, "--", filepath.Join(testWorkspaceRoot, goMod)}) {
				return nil, nil
			}
			return cmd.Output()
		}

		resultsShouldContain := []string{"archive/tar", "archive/zip", "bufio", "bytes", "compress/bzip2", "unicode", "unicode/utf16", "unicode/utf8", "unsafe", "vendor/golang.org/x/text/unicode/norm"}
		resultsShouldNotContain := []string{"selinc.com/butler", "selinc.com/builtin", "github.com/smartystreets/goconvey/convey"}
		results, err := GetStdLibs(golangId)

		So(err, ShouldBeNil)
		So(results[0], ShouldEqual, "false")
		for _, testPackage := range resultsShouldContain {
			So(results, ShouldContain, testPackage)
		}
		for _, testPackage := range resultsShouldNotContain {
			So(results, ShouldNotContain, testPackage)
		}
	})

	Convey("Returns error when an unknown language ID is passed in", t, func() {
		undo := replaceStubs()
		defer undo()

		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitName, "diff", testBranch, "--", filepath.Join(testWorkspaceRoot, goMod)}) {
				return nil, nil
			}
			return cmd.Output()
		}

		results, err := GetStdLibs("fail")

		So(err.Error(), ShouldEqual, "language id 'fail' not found")
		So(len(results), ShouldEqual, 0)
	})
}

func Test_GetExternalDependencies(t *testing.T) {
	Convey("Correct go libraries should be returned", t, func() {
		undo := replaceStubs()
		defer undo()

		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitName, "diff", testBranch, "--", filepath.Join(testWorkspaceRoot, goMod)}) {
				return nil, nil
			}
			return cmd.Output()
		}

		results, err := GetExternalDependencies(golangId)
		So(err, ShouldBeNil)
		So(len(results), ShouldEqual, 0)
	})

	Convey("Returns error when an unknown language ID is passed in", t, func() {
		undo := replaceStubs()
		defer undo()

		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitName, "diff", testBranch, "--", filepath.Join(testWorkspaceRoot, goMod)}) {
				return nil, nil
			}
			return cmd.Output()
		}

		results, err := GetExternalDependencies("fail")

		So(err.Error(), ShouldEqual, "language id 'fail' not found")
		So(len(results), ShouldEqual, 0)
	})
}

func Test_GetWorkspaceDeps(t *testing.T) {
	Convey("Returns go workspace dependencies as expected", t, func() {
		undo := replaceStubs()
		defer undo()

		resultsShouldContain := []string{"github.com/smartystreets/goconvey/convey", "testing"}
		results := GetWorkspaceDeps(golangId, goTestDirectory)

		for _, testPackage := range resultsShouldContain {
			So(results, ShouldContain, testPackage)
		}
	})
}
