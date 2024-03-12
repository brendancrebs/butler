// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal_test

import (
	"testing"

	"selinc.com/butler/code/internal"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_loadConfig(t *testing.T) {
	Convey("Ability to parse a yaml config for all of its values", t, func() {
		temp := internal.ConfigPath
		internal.ConfigPath = "./test_data/test_helpers/.butler.base.yaml"

		expectedLanguage := &internal.Language{
			Name:                     "golang",
			FileExtension:            ".go",
			BuiltinStdLibsMethod:     true,
			BuiltinExternalDepMethod: true,
			TaskExec: &internal.TaskCommands{
				LintCommand:    "echo go lint command",
				TestCommand:    "echo go test command",
				BuildCommand:   "echo go build command",
				PublishCommand: "echo go publish command",
				SetUpCommands:  []string{"echo go test"},
			},
			DepCommands: &internal.DependencyCommands{
				ExternalDepCommand: "go run ./user_commands/go_external_deps_method.go",
			},
		}

		config := &internal.ButlerConfig{}
		expected := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
				AllowedPaths:    []string{"test_repo"},
				BlockedPaths:    []string{"ci/", "specs", ".devcontainer", "bad_configs"},
				WorkspaceRoot:   "../..",
				ResultsFilePath: internal.ButlerResultsPath,
			},
			Git: &internal.GitConfigurations{
				PublishBranch: "main",
				GitRepo:       true,
			},
			Task: &internal.TaskConfigurations{
				Coverage:      "0",
				ShouldLint:    true,
				ShouldTest:    true,
				ShouldBuild:   false,
				ShouldPublish: false,
				ShouldRunAll:  false,
			},
			Languages: []*internal.Language{expectedLanguage},
		}

		err := config.Load(internal.ConfigPath)
		internal.ConfigPath = temp

		So(err, ShouldBeNil)
		So(config, ShouldResemble, expected)
	})

	Convey("Failure to parse config file is covered", t, func() {
		temp := internal.ConfigPath
		internal.ConfigPath = "./test_data/bad_configs/invalid.butler.bad"

		config := &internal.ButlerConfig{}
		expected := &internal.ButlerConfig{}

		err := config.Load(internal.ConfigPath)
		internal.ConfigPath = temp

		So(err, ShouldNotBeNil)
		So(config, ShouldResemble, expected)
	})
}

func Test_validateConfig(t *testing.T) {
	testLanguage := &internal.Language{
		Name: "test",
		TaskExec: &internal.TaskCommands{
			SetUpCommands: []string{"echo test"},
		},
		DepCommands: &internal.DependencyCommands{
			ExternalDepCommand: "echo get deps",
		},
	}

	Convey("Catch config that doesn't have workspace root set", t, func() {
		testConfig := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
				AllowedPaths: []string{"test_repo"},
			},
			Task: &internal.TaskConfigurations{
				Coverage: "100",
			},
			Languages: []*internal.Language{testLanguage},
		}
		err := testConfig.ValidateConfig()
		So(err.Error(), ShouldContainSubstring, "no workspace root has been set.\nPlease set a workspace root in the config")
	})

	Convey("Invalid coverage set in config.", t, func() {
		testConfig := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
				AllowedPaths:  []string{"test_repo"},
				WorkspaceRoot: "test_repo",
			},
			Task: &internal.TaskConfigurations{
				Coverage: "",
			},
			Languages: []*internal.Language{testLanguage},
		}
		type template struct {
			testValue     string
			expectedError string
		}
		testCoverageValues := []template{
			{"1000", "the test coverage threshold has been set to 1000 in the config. Please set coverage to a number between 0 and 100"},
			{"-25", "the test coverage threshold has been set to -25 in the config. Please set coverage to a number between 0 and 100"},
			{"false", `strconv.Atoi: parsing "false": invalid syntax`},
		}

		for _, test := range testCoverageValues {
			testConfig.Task.Coverage = test.testValue
			err := testConfig.ValidateConfig()
			So(err.Error(), ShouldContainSubstring, test.expectedError)
		}
	})

	Convey("git repo set to false in the absence of a publish branch", t, func() {
		testConfig := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
				AllowedPaths:  []string{"test_repo"},
				WorkspaceRoot: "test_repo",
			},
			Task: &internal.TaskConfigurations{
				Coverage: "100",
			},
			Languages: []*internal.Language{testLanguage},
			Git: &internal.GitConfigurations{
				PublishBranch: "",
			},
		}

		err := testConfig.ValidateConfig()
		So(err, ShouldBeNil)
		So(testConfig.Git.GitRepo, ShouldBeFalse)
	})

	Convey("no languages added to config", t, func() {
		testConfig := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
				AllowedPaths:  []string{"test_repo"},
				WorkspaceRoot: "test_repo",
			},
			Task: &internal.TaskConfigurations{
				Coverage: "100",
			},
			Git: &internal.GitConfigurations{
				PublishBranch: "main",
			},
		}

		err := testConfig.ValidateConfig()
		So(err.Error(), ShouldContainSubstring, "no languages have been provided in the config")
	})

	Convey("no allowed paths added to config", t, func() {
		testConfig := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
				WorkspaceRoot: "test_repo",
			},
			Task: &internal.TaskConfigurations{
				Coverage: "100",
			},
			Languages: []*internal.Language{testLanguage},
			Git: &internal.GitConfigurations{
				PublishBranch: "main",
			},
		}

		err := testConfig.ValidateConfig()
		So(err.Error(), ShouldContainSubstring, "butler has not been given permission to search for workspaces in any directories.")
	})
}

func Test_loadButlerIgnore(t *testing.T) {
	Convey("Paths successfully parsed from .butler.ignore.", t, func() {
		testConfig := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
				WorkspaceRoot: "./test_data/test_helpers",
			},
		}
		expected := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
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
		testConfig := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
				WorkspaceRoot: "/",
			},
		}
		err := testConfig.LoadButlerIgnore()

		So(err, ShouldBeNil)
		So(testConfig.Paths.AllowedPaths, ShouldBeNil)
		So(testConfig.Paths.BlockedPaths, ShouldBeNil)
	})
	Convey("Failure to parse .butler.ignore.", t, func() {
		testConfig := &internal.ButlerConfig{
			Paths: &internal.ButlerPaths{
				WorkspaceRoot: "./test_data/bad_configs",
			},
		}
		err := testConfig.LoadButlerIgnore()

		So(err, ShouldNotBeNil)
		So(testConfig.Paths.AllowedPaths, ShouldEqual, []string{"good_path"})
		So(testConfig.Paths.BlockedPaths, ShouldBeNil)
	})
}
