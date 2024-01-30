// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
		{"false returned when there are no file matches", []string{"code/main", "helpers/common"}, []string{"test_path/critical_file"}, false},
		{"critical file correctly matched", []string{"test_path/critical_file1", "code/main"}, []string{"test_path/critical_file1", "test_path/critical_file2"}, true},
		{"false returned when there are no folder matches", []string{"code/", "helpers/"}, []string{"test_path/critical_folder/"}, false},
		{"critical folder correctly matched", []string{"test_path/critical_folder1/test_folder", "code/"}, []string{"test_path/critical_folder1/", "test_path/critical_folder2"}, true},
	}

	for _, test := range tests {
		Convey(test.desc, t, func() {
			So(criticalPathChanged(test.dirtyFolders, test.criticalFolders), ShouldResemble, test.expected)
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
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "git executable not found")
	})
}
