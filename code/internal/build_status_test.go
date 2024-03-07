// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal_test

import (
	"testing"

	"selinc.com/butler/code/internal"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_BuildStatusEnum(t *testing.T) {
	Convey("BuildStatus Stringer Expected Output Tests", t, func() {
		So(internal.BuildStatusClean.String(), ShouldEqual, "Clean")
		So(internal.BuildStatusDirty.String(), ShouldEqual, "Dirty")
		So(internal.BuildStatusUnknown.String(), ShouldEqual, "Unknown")
		So(internal.BuildStatusFail.String(), ShouldEqual, "Fail")
		So(internal.BuildStatus(-1).String(), ShouldEqual, "Unknown")
		So(internal.BuildStatus(1000).String(), ShouldEqual, "Unknown")
	})
}

func Test_BuildStepEnum(t *testing.T) {
	Convey("BuildStep Stringer Expected Output Tests", t, func() {
		So(internal.BuildStepUnknown.String(), ShouldEqual, "Unknown")
		So(internal.BuildStepSpec.String(), ShouldEqual, "Spec")
		So(internal.BuildStepLint.String(), ShouldEqual, "Lint")
		So(internal.BuildStepTest.String(), ShouldEqual, "Test")
		So(internal.BuildStepBuild.String(), ShouldEqual, "Build")
		So(internal.BuildStepPublish.String(), ShouldEqual, "Publish")
		So(internal.BuildStep(-1).String(), ShouldEqual, "Unknown")
		So(internal.BuildStep(1000).String(), ShouldEqual, "Unknown")
	})

	Convey("BuildStep marshal JSON expected results", t, func() {
		b, err := internal.BuildStepSpec.MarshalJSON()
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, `"Spec"`)
	})

	Convey("BuildStep unmarshal JSON expected results", t, func() {
		var bs internal.BuildStep
		err := bs.UnmarshalJSON([]byte(`"Spec"`))
		So(err, ShouldBeNil)
		So(bs, ShouldEqual, internal.BuildStepSpec)
	})

	Convey("BuildStep unmarshal JSON with error", t, func() {
		var bs internal.BuildStep
		err := bs.UnmarshalJSON([]byte(`}}`))
		So(err, ShouldNotBeNil)
		So(bs, ShouldEqual, internal.BuildStepUnknown)
	})
}
