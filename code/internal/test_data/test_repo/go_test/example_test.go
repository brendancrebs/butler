// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestSum tests the Sum function.
func TestSum(t *testing.T) {
	Convey("Expected Sum", t, func() {
		total := Sum(5, 5)
		So(total, ShouldEqual, 10)
	})
}
