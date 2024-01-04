// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package main

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func ExampleMain() {
	os.Args = []string{"", "-h"}
	main()
}

func Test_Main(t *testing.T) {
	Convey("Run error cases for coverage", t, func() {
		os.Args = []string{"", "-h"}
		main()
	})
}
