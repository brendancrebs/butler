// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler_test

import (
	"testing"

	"selinc.com/butler/code/butler"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_criticalPathChanged(t *testing.T) {
	type template struct {
		desc          string
		dirtyPaths    []string
		criticalPaths []string
		expected      bool
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
			So(butler.CriticalPathChanged(test.dirtyPaths, test.criticalPaths), ShouldResemble, test.expected)
		})
	}
}

func Test_queue(t *testing.T) {
	Convey("Queue can push and pop as expected.", t, func() {
		testQueue := &butler.Queue{}
		testQueue.Enqueue(&butler.Task{})
		testQueue.Enqueue(&butler.Task{})
		testQueue.Enqueue(&butler.Task{})
		So(testQueue.Size(), ShouldEqual, 3)

		testQueue.Dequeue()
		testQueue.Dequeue()
		So(testQueue.Size(), ShouldEqual, 1)
	})
}

func Test_evaluate_dirtiness(t *testing.T) {
	testWorkspaces := []*butler.Workspace{
		{Location: "root/workspace1", IsDirty: false, Dependencies: []string{"dep1", "dep2"}},
		{Location: "root/workspace2", IsDirty: false, Dependencies: []string{"dep3", "dep4"}},
		{Location: "root/workspace3", IsDirty: false, Dependencies: []string{"dep4"}},
		{Location: "root/workspace4", IsDirty: false, Dependencies: []string{}},
	}
	Convey("dirty workspaces marked as expected", t, func() {
		expected := []*butler.Workspace{}
		for _, ws := range testWorkspaces {
			newWs := new(butler.Workspace)
			*newWs = *ws
			expected = append(expected, newWs)
		}
		expected[0].IsDirty = true
		expected[3].IsDirty = true
		dirtyPaths := []string{"root/workspace1", "root/workspace4"}
		butler.EvaluateDirtiness(testWorkspaces, dirtyPaths)
		So(testWorkspaces, ShouldResemble, expected)
	})
}
