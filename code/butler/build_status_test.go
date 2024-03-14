// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler_test

import (
	"testing"

	"selinc.com/butler/code/butler"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_BuildStatusEnum(t *testing.T) {
	Convey("BuildStatus Stringer Expected Output Tests", t, func() {
		So(butler.BuildStatusClean.String(), ShouldEqual, "Clean")
		So(butler.BuildStatusDirty.String(), ShouldEqual, "Dirty")
		So(butler.BuildStatusUnknown.String(), ShouldEqual, "Unknown")
		So(butler.BuildStatusFail.String(), ShouldEqual, "Fail")
		So(butler.BuildStatus(-1).String(), ShouldEqual, "Unknown")
		So(butler.BuildStatus(1000).String(), ShouldEqual, "Unknown")
	})
}

func Test_BuildStepEnum(t *testing.T) {
	Convey("BuildStep Stringer Expected Output Tests", t, func() {
		So(butler.BuildStepUnknown.String(), ShouldEqual, "Unknown")
		So(butler.BuildStepSpec.String(), ShouldEqual, "Spec")
		So(butler.BuildStepLint.String(), ShouldEqual, "Lint")
		So(butler.BuildStepTest.String(), ShouldEqual, "Test")
		So(butler.BuildStepBuild.String(), ShouldEqual, "Build")
		So(butler.BuildStepPublish.String(), ShouldEqual, "Publish")
		So(butler.BuildStep(-1).String(), ShouldEqual, "Unknown")
		So(butler.BuildStep(1000).String(), ShouldEqual, "Unknown")
	})

	Convey("BuildStep marshal JSON expected results", t, func() {
		b, err := butler.BuildStepSpec.MarshalJSON()
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, `"Spec"`)
	})

	Convey("BuildStep unmarshal JSON expected results", t, func() {
		var bs butler.BuildStep
		err := bs.UnmarshalJSON([]byte(`"Spec"`))
		So(err, ShouldBeNil)
		So(bs, ShouldEqual, butler.BuildStepSpec)
	})

	Convey("BuildStep unmarshal JSON with error", t, func() {
		var bs butler.BuildStep
		err := bs.UnmarshalJSON([]byte(`}}`))
		So(err, ShouldNotBeNil)
		So(bs, ShouldEqual, butler.BuildStepUnknown)
	})
}
