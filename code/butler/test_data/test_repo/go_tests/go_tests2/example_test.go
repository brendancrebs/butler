// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package example

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestConcat tests the Concat function.
func TestConcat(t *testing.T) {
	Convey("Expected string", t, func() {
		str := Concat("hello ", "world")
		So(str, ShouldResemble, "hello world")
	})
}
