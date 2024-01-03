// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

// cSpell:ignore simplejs curr

package internal

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var currBranch = func() string {
	b, err := getCurrentBranch()
	if err != nil {
		t := testing.T{}
		t.Fatal(err)
	}

	return b
}()

func Test_RunWithError(t *testing.T) {
	Convey("Just running the command for outer coverage of Execute", t, func() {
		cmd = getCommand()
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.yaml"})
		Execute()

		// Success determined by existence of the results json file.
		_, err := os.Stat("./butler_results.json")
		So(err, ShouldBeNil)
	})
}
