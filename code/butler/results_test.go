// Copyright (c) 2023 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"errors"
	"os/exec"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_GetDFOutputByField(t *testing.T) {
	dfPath, err := exec.LookPath("df")
	Convey("We can get a path for the 'df' command", t, func() {
		So(err, ShouldBeNil)
	})

	Convey("Test successful call", t, func() {
		s := getDFOutputByField(dfPath, "/", "size")
		i, err := strconv.Atoi(s)
		So(err, ShouldBeNil)
		So(i, ShouldBeGreaterThan, 0)
	})
	Convey("Getting a field from df that does not exist results in an error", t, func() {
		s := getDFOutputByField(dfPath, "/", "invalid")
		So(s, ShouldEqual, "exit status 1")
	})
	Convey("Unexpected results error when passing no path", t, func() {
		s := getDFOutputByField(dfPath, "-h", "size")
		So(s, ShouldEqual, "unexpected df output")
	})
}

func Test_assignErrorStringIfError(t *testing.T) {
	type template struct {
		desc     string
		input    string
		err      error
		expected string
	}
	tests := []*template{
		{"input with no error", "hello", nil, "hello"},
		{"input with error", "hello", errors.New("goodbye"), "goodbye"},
	}
	for _, test := range tests {
		Convey(test.desc, t, func() {
			So(assignErrorStringIfError(test.input, test.err), ShouldEqual, test.expected)
		})
	}
}

func Test_getStatusFromTasks(t *testing.T) {
	type template struct {
		desc     string
		input    []*Task
		expected BuildStatus
	}
	tests := []*template{
		{"nil input", nil, BuildStatusClean},
		{"dirty input", []*Task{{Attempts: 2}}, BuildStatusDirty},
		{"error input", []*Task{{Attempts: 2, err: errors.New("bad stuff happened")}}, BuildStatusFail},
	}
	for _, test := range tests {
		Convey(test.desc, t, func() {
			So(getStatusFromTasks(test.input), ShouldEqual, test.expected)
		})
	}
}
