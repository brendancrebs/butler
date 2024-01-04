// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

// cSpell:ignore simplejs curr

package internal

import (
	"bytes"
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

func Test_RunWithErr(t *testing.T) {
	os.Setenv("BUTLER_SHOULD_RUN_ALL", "true")
	Convey("Just running the command for outer coverage of Execute", t, func() {
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/.butler.base.yaml", "--all"})
		Execute()

		// Success determined by existence of the results json file.
		_, err := os.Stat("./butler_results.json")
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, "")
	})

	Convey("config parse fails due to bad path", t, func() {
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/test_helpers/invalid.butler.bad"})
		Execute()

		// butler_results.json should still exists despite error
		_, err := os.Stat("./butler_results.json")
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "Error: stat ./test_data/test_helpers/invalid.butler.bad: no such file or directory")
	})

	Convey(".butler.ignore parse fails due to bad syntax", t, func() {
		cmd = getCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"--publish-branch", currBranch, "--cfg", "./test_data/bad_configs/.butler.base.yaml"})
		Execute()

		_, err := os.Stat("./butler_results.json")
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldContainSubstring, "Error: Butler ignore parse error: yaml: unmarshal errors")
	})

	Convey("Butler setup fails", t, func() {
	})
}
