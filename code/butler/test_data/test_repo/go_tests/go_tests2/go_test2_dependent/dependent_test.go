// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExclamation(t *testing.T) {
	Convey("Expected Sum", t, func() {
		str := exclamation("hello ", "world")
		So(str, ShouldEqual, "hello world!")
	})
}
