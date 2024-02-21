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

		expectedLanguage := &Language{
			Name:          "golang",
			FileExtension: ".go",
			TaskExec: &TaskCommands{
				LintCommand:    "echo go lint command",
				TestCommand:    "echo go test command",
				BuildCommand:   "echo go build command",
				PublishCommand: "echo go publish command",
				SetUpCommands:  []string{"echo go test"},
			},
			DepCommands: &DependencyCommands{
				ExternalDepCommand: "go run /workspaces/butler/user_commands/go_external_deps_method.go",
			},
		}

		config := &ButlerConfig{}
		expected := &ButlerConfig{
			Paths: &ButlerPaths{
				AllowedPaths:    []string{"test_repo"},
				BlockedPaths:    []string{"ci/", "specs", ".devcontainer", "bad_configs"},
				WorkspaceRoot:   "../..",
				ResultsFilePath: butlerResultsPath,
			},
			Git: &GitConfigurations{
				PublishBranch: "main",
				GitRepo:       true,
			},
			Task: &TaskConfigurations{
				Coverage:      "0",
				ShouldLint:    true,
				ShouldTest:    true,
				ShouldBuild:   false,
				ShouldPublish: false,
				ShouldRunAll:  false,
			},
			Languages: []*Language{expectedLanguage},
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
		expected := &ButlerConfig{}

		err := config.Load(configPath)
		configPath = temp

		So(err, ShouldNotBeNil)
		So(config, ShouldResemble, expected)
	})
}

func Test_loadButlerIgnore(t *testing.T) {
	Convey("Paths successfully parsed from .butler.ignore.", t, func() {
		testConfig := &ButlerConfig{
			Paths: &ButlerPaths{
				WorkspaceRoot: "./test_data/test_helpers",
			},
		}
		expected := &ButlerConfig{
			Paths: &ButlerPaths{
				AllowedPaths:  []string{"good_path"},
				BlockedPaths:  []string{"bad_path"},
				WorkspaceRoot: "./test_data/test_helpers",
			},
		}
		err := testConfig.LoadButlerIgnore()

		So(err, ShouldBeNil)
		So(testConfig, ShouldResemble, expected)
	})
	Convey("Nothing returned if no .butler.ignore file is found.", t, func() {
		testConfig := &ButlerConfig{
			Paths: &ButlerPaths{
				WorkspaceRoot: "/",
			},
		}
		err := testConfig.LoadButlerIgnore()

		So(err, ShouldBeNil)
		So(testConfig.Paths.AllowedPaths, ShouldBeNil)
		So(testConfig.Paths.BlockedPaths, ShouldBeNil)
	})
	Convey("Failure to parse .butler.ignore.", t, func() {
		testConfig := &ButlerConfig{
			Paths: &ButlerPaths{
				WorkspaceRoot: "./test_data/bad_configs",
			},
		}
		err := testConfig.LoadButlerIgnore()

		So(err, ShouldNotBeNil)
		So(testConfig.Paths.AllowedPaths, ShouldEqual, []string{"good_path"})
		So(testConfig.Paths.BlockedPaths, ShouldBeNil)
	})
}
