// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_BuildStatusEnum(t *testing.T) {
	Convey("BuildStatus Stringer Expected Output Tests", t, func() {
		So(BuildStatusClean.String(), ShouldEqual, "Clean")
		So(BuildStatusDirty.String(), ShouldEqual, "Dirty")
		So(BuildStatusUnknown.String(), ShouldEqual, "Unknown")
		So(BuildStatusFail.String(), ShouldEqual, "Fail")
		So(BuildStatus(-1).String(), ShouldEqual, "Unknown")
		So(BuildStatus(1000).String(), ShouldEqual, "Unknown")
	})
}

func Test_BuildStepEnum(t *testing.T) {
	Convey("BuildStep Stringer Expected Output Tests", t, func() {
		So(BuildStepUnknown.String(), ShouldEqual, "Unknown")
		So(BuildStepSpec.String(), ShouldEqual, "Spec")
		So(BuildStepLint.String(), ShouldEqual, "Lint")
		So(BuildStepTest.String(), ShouldEqual, "Test")
		So(BuildStepBuild.String(), ShouldEqual, "Build")
		So(BuildStepPublish.String(), ShouldEqual, "Publish")
		So(BuildStep(-1).String(), ShouldEqual, "Unknown")
		So(BuildStep(1000).String(), ShouldEqual, "Unknown")
	})

	Convey("BuildStep marshal JSON expected results", t, func() {
		b, err := BuildStepSpec.MarshalJSON()
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, `"Spec"`)
	})

	Convey("BuildStep unmarshal JSON expected results", t, func() {
		var bs BuildStep
		err := bs.UnmarshalJSON([]byte(`"Spec"`))
		So(err, ShouldBeNil)
		So(bs, ShouldEqual, BuildStepSpec)
	})

	Convey("BuildStep unmarshal JSON with error", t, func() {
		var bs BuildStep
		err := bs.UnmarshalJSON([]byte(`}}`))
		So(err, ShouldNotBeNil)
		So(bs, ShouldEqual, BuildStepUnknown)
	})
}
