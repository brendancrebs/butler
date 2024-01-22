// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_loadConfig(t *testing.T) {
	Convey("Ability to parse a yaml config for all of its values", t, func() {
		temp := configPath
		configPath = "./test_data/test_helpers/.butler.base.yaml"

		config := &ButlerConfig{}
		expected := &ButlerConfig{
			Allowed:         []string{"test_repo"},
			Blocked:         []string{"ci/", "specs", ".devcontainer"},
			WorkspaceRoot:   "/workspaces/butler",
			GitRepo:         true,
			ShouldLint:      true,
			ShouldTest:      true,
			ShouldBuild:     false,
			ShouldPublish:   false,
			ShouldRunAll:    false,
			PublishBranch:   "main",
			ResultsFilePath: "./butler_results.json",
		}
		err := config.Load(configPath)
		configPath = temp

		So(err, ShouldBeNil)
		So(config, ShouldResemble, expected)
	})

	Convey("Failure to parse config file is covered", t, func() {
		temp := configPath
		configPath = "./test_data/bad_configs/invalid.butler.bad"

		config := &ButlerConfig{}
		expected := &ButlerConfig{
			ResultsFilePath: "./butler_results.json",
		}

		err := config.Load(configPath)
		configPath = temp

		So(err, ShouldNotBeNil)
		So(config, ShouldResemble, expected)
	})
}

func Test_loadButlerIgnore(t *testing.T) {
	Convey("Paths successfully parsed from .butler.ignore.", t, func() {
		testConfig := &ButlerConfig{
			WorkspaceRoot: "./test_data/test_helpers",
		}
		expected := &ButlerConfig{
			Allowed:       []string{"good_path"},
			Blocked:       []string{"bad_path"},
			WorkspaceRoot: "./test_data/test_helpers",
		}
		err := testConfig.LoadButlerIgnore()

		So(err, ShouldBeNil)
		So(testConfig, ShouldResemble, expected)
	})
	Convey("Nothing returned if no .butler.ignore file is found.", t, func() {
		testConfig := &ButlerConfig{
			WorkspaceRoot: "/",
		}
		err := testConfig.LoadButlerIgnore()

		So(err, ShouldBeNil)
		So(testConfig.Allowed, ShouldBeNil)
		So(testConfig.Blocked, ShouldBeNil)
	})
	Convey("Failure to parse .butler.ignore.", t, func() {
		testConfig := &ButlerConfig{
			WorkspaceRoot: "./test_data/bad_configs",
		}
		err := testConfig.LoadButlerIgnore()

		So(err, ShouldNotBeNil)
		So(testConfig.Allowed, ShouldEqual, []string{"good_path"})
		So(testConfig.Blocked, ShouldBeNil)
	})
}
