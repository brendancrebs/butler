// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"errors"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_separateCriticalFiles(t *testing.T) {
	type template struct {
		desc                    string
		inputPaths              []string
		expectedCriticalFiles   []string
		expectedCriticalFolders []string
	}
	workspaceRoot := "/workspaces/butler/code/internal/test_data"

	tests := []template{
		{"empty slice", []string{}, []string(nil), []string(nil)},
		{"nil slice", nil, []string(nil), []string(nil)},
		{"returns lone critical file", []string{"test_helpers/.butler.base.yaml"}, []string{path.Join(workspaceRoot, "test_helpers/.butler.base.yaml")}, []string(nil)},
		{"returns correct critical files and folders", []string{"test_helpers/.butler.base.yaml", "bad_configs", "test_helpers/.butler.ignore.yaml", "test_repo/go_test"}, []string{path.Join(workspaceRoot, "test_helpers/.butler.base.yaml"), path.Join(workspaceRoot, "test_helpers/.butler.ignore.yaml")}, []string{path.Join(workspaceRoot, "bad_configs"), path.Join(workspaceRoot, "test_repo/go_test")}},
	}

	for _, test := range tests {
		Convey(test.desc, t, func() {
			criticalFiles, criticalFolders, err := separateCriticalPaths(workspaceRoot, test.inputPaths)
			So(err, ShouldBeNil)
			So(criticalFiles, ShouldResemble, test.expectedCriticalFiles)
			So(criticalFolders, ShouldResemble, test.expectedCriticalFolders)
		})
	}

	Convey("fails when incorrect critical path included", t, func() {
		criticalFiles, criticalFolders, err := separateCriticalPaths(workspaceRoot, []string{"incorrect_path/fail"})
		So(err, ShouldNotBeNil)
		So(criticalFiles, ShouldResemble, []string(nil))
		So(criticalFolders, ShouldResemble, []string(nil))
	})
}

func Test_criticalFileChanged(t *testing.T) {
	type template struct {
		desc          string
		dirtyFiles    []string
		criticalFiles []string
		expected      bool
	}
	tests := []template{
		{"empty dirtyFiles slice", []string{}, []string{"test_path/critical_file"}, false},
		{"empty criticalFiles slice", []string{"code/main", "helpers/common"}, []string{}, false},
		{"nil slice", nil, []string(nil), false},
		{"false returned when there are no matches", []string{"code/main", "helpers/common"}, []string{"test_path/critical_file"}, false},
		{"critical file correctly matched", []string{"test_path/critical_file1", "code/main"}, []string{"test_path/critical_file1", "test_path/critical_file2"}, true},
	}

	for _, test := range tests {
		Convey(test.desc, t, func() {
			So(criticalFileChanged(test.dirtyFiles, test.criticalFiles), ShouldResemble, test.expected)
		})
	}
}

func Test_criticalFolderChanged(t *testing.T) {
	type template struct {
		desc            string
		dirtyFolders    []string
		criticalFolders []string
		expected        bool
	}
	tests := []template{
		{"empty dirtyFolders slice", []string{}, []string{"test_path/critical_folder/"}, false},
		{"empty criticalFolders slice", []string{"code/", "helpers/"}, []string{}, false},
		{"nil slice", nil, []string(nil), false},
		{"false returned when there are no matches", []string{"code/", "helpers/"}, []string{"test_path/critical_folder/"}, false},
		{"critical folder correctly matched", []string{"test_path/critical_folder1/test_folder", "code/"}, []string{"test_path/critical_folder1/", "test_path/critical_folder2"}, true},
	}

	for _, test := range tests {
		Convey(test.desc, t, func() {
			So(criticalFolderChanged(test.dirtyFolders, test.criticalFolders), ShouldResemble, test.expected)
		})
	}
}

func Test_getCurrentBranchCoverage(t *testing.T) {
	Convey("fails when git executable not found", t, func() {
		undo := replaceStubs()
		defer undo()

		execLookPathStub = func(executable string) (string, error) { return "", errors.New("git executable not found") }
		branch, err := getCurrentBranch()

		So(branch, ShouldEqual, "")
		So(err.Error(), ShouldContainSubstring, "git executable not found")
	})
}
