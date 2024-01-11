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
		defer func() {
			configPath = temp
		}()

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

		So(err, ShouldBeNil)
		So(config, ShouldResemble, expected)
	})

	Convey("Failure to parse config file is covered", t, func() {
		temp := configPath
		configPath = "./test_data/bad_configs/invalid.butler.bad"
		defer func() {
			configPath = temp
		}()

		config := &ButlerConfig{}
		expected := &ButlerConfig{
			ResultsFilePath: "./butler_results.json",
		}

		err := config.Load(configPath)
		So(err, ShouldNotBeNil)
		So(config, ShouldResemble, expected)
	})
}

func Test_loadButlerIgnore(t *testing.T) {
	Convey("Paths successfully parsed from Butler ignore.", t, func() {
		testConfig := &ButlerConfig{
			WorkspaceRoot: "./test_data/test_helpers",
		}
		expected := &ButlerPaths{
			Allowed: []string{"good_path"},
			Blocked: []string{"bad_path"},
		}
		paths, err := loadButlerIgnore(testConfig)

		So(err, ShouldBeNil)
		So(paths, ShouldResemble, expected)
	})
	Convey("Nothing returned in no Butler ignore if found.", t, func() {
		testConfig := &ButlerConfig{
			WorkspaceRoot: "/",
		}
		paths, err := loadButlerIgnore(testConfig)

		So(err, ShouldBeNil)
		So(paths, ShouldBeNil)
	})
	Convey("Failure to parse Butler ignore.", t, func() {
		testConfig := &ButlerConfig{
			WorkspaceRoot: "./test_data/bad_configs",
		}
		paths, err := loadButlerIgnore(testConfig)

		So(err, ShouldNotBeNil)
		So(paths, ShouldBeNil)
	})
}
