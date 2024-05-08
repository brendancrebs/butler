// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package builtin

import (
	"errors"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	goTestDirectory = "../test_data/test_repo/go_test"
)

func Test_getStdLibs(t *testing.T) {
	Convey("Correct go libraries should be returned", t, func() {
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
		results, err := getStdLibs()

		So(err, ShouldBeNil)
		So(results[0], ShouldEqual, "false")
		for _, testPackage := range resultsShouldContain {
			So(results, ShouldContain, testPackage)
		}
		for _, testPackage := range resultsShouldNotContain {
			So(results, ShouldNotContain, testPackage)
		}
	})

	Convey("Returns error when 'go list std' fails", t, func() {
		undo := replaceStubs()
		defer undo()

		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{goName, "list", "std"}) {
				return nil, errors.New("go list std failed")
			}
			return cmd.Output()
		}
		results, err := getStdLibs()

		So(err.Error(), ShouldEqual, "go list std failed")
		So(len(results), ShouldEqual, 0)
	})

	Convey("Returns error when getChangedModFileDeps() fails", t, func() {
		undo := replaceStubs()
		defer undo()

		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{gitName, "diff", testBranch, "--", filepath.Join(testWorkspaceRoot, goMod)}) {
				return nil, errors.New("git diff failed")
			}
			return cmd.Output()
		}
		resultsShouldContain := []string{"archive/tar", "archive/zip", "bufio", "bytes", "compress/bzip2", "unicode", "unicode/utf16", "unicode/utf8", "unsafe", "vendor/golang.org/x/text/unicode/norm"}
		results, err := getStdLibs()

		So(err.Error(), ShouldEqual, "git diff failed")
		for _, testPackage := range resultsShouldContain {
			So(results, ShouldContain, testPackage)
		}
	})
}

func Test_getWorkspaceDeps(t *testing.T) {
	Convey("Returns a set of deps for a go workspace as expected", t, func() {
		resultsShouldContain := []string{"github.com/smartystreets/goconvey/convey", "testing"}
		results, err := getWorkspaceDeps(goTestDirectory)

		So(err, ShouldBeNil)
		for _, testPackage := range resultsShouldContain {
			So(results, ShouldContain, testPackage)
		}
	})

	Convey("Returns err when go command fails", t, func() {
		execOutputStub = func(cmd *exec.Cmd) ([]byte, error) {
			if reflect.DeepEqual(cmd.Args, []string{goName, "list", "-test", "-f", `{{join .Deps "\n"}}`, goTestDirectory}) {
				return nil, errors.New("go list failed")
			}
			return cmd.Output()
		}
		results, err := getWorkspaceDeps(goTestDirectory)

		So(err.Error(), ShouldEqual, "go list failed")
		So(len(results), ShouldEqual, 0)
	})
}

func Test_pruneAdditiveChanges(t *testing.T) {
	tests := []struct {
		desc            string
		input, expected []string
	}{
		{"changeset with no added lines", []string{"hello", "goodbye", "no more"}, []string{"", "", ""}},
		{"changeset with removed lines", []string{"hello", "-goodbye", "-no more"}, []string{"", "", ""}},
		{"changeset with additive changes", []string{"hello", "+goodbye", "+no more"}, []string{"", "", "", "goodbye", "no"}},
		{"changeset with multiple pluses", []string{"hello", "++goodbye", "+no more"}, []string{"", "", "", "no"}},
	}
	for _, test := range tests {
		Convey(test.desc, t, func() {
			So(pruneAdditiveChanges(test.input), ShouldResemble, test.expected)
		})
	}
}

func Test_convertLinesToStrings(t *testing.T) {
	tests := []struct {
		desc      string
		input     []byte
		splitOn   []byte
		wantLines []string
	}{
		{
			desc:      "empty input results in no lines",
			input:     []byte{},
			splitOn:   []byte("doesn't matter"),
			wantLines: []string{},
		},
		{
			desc:      "splits line correctly 1",
			input:     []byte("This\nis\na\nstrange\nmessage!"),
			splitOn:   []byte{'\n'},
			wantLines: []string{"This", "a", "is", "message!", "strange"},
		},
		{
			desc:      "Splits line correctly 2",
			input:     []byte("abc123"),
			splitOn:   []byte{'c'},
			wantLines: []string{"123", "ab"},
		},
	}
	for _, test := range tests {
		Convey(test.desc, t, func() {
			lines := convertLinesToStrings(test.input, test.splitOn)
			So(len(lines), ShouldEqual, len(test.wantLines))
			for i, line := range lines {
				So(line, ShouldResemble, test.wantLines[i])
			}
		})
	}
}

func Test_didVersionChange(t *testing.T) {
	tests := []struct {
		desc     string
		input    []string
		expected bool
	}{
		{"empty slice", []string{}, false},
		{"nil slice", nil, false},
		{"contains go version change", []string{"go 1.19.0"}, true},
	}
	for _, test := range tests {
		Convey(test.desc, t, func() {
			So(didVersionChange(test.input), ShouldResemble, test.expected)
		})
	}
}
