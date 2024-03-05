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
			Name:                     "golang",
			FileExtension:            ".go",
			BuiltinStdLibsMethod:     true,
			BuiltinExternalDepMethod: true,
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
				WorkspaceRoot:   "/workspaces/butler",
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

func Test_validateConfig(t *testing.T) {
	testLanguage := &Language{
		Name: "test",
		TaskExec: &TaskCommands{
			SetUpCommands: []string{"echo test"},
		},
		DepCommands: &DependencyCommands{
			ExternalDepCommand: "echo get deps",
		},
	}

	Convey("Catch config that doesn't have workspace root set", t, func() {
		testConfig := &ButlerConfig{
			Paths: &ButlerPaths{
				AllowedPaths: []string{"test_repo"},
			},
			Task: &TaskConfigurations{
				Coverage: "100",
			},
			Languages: []*Language{testLanguage},
		}
		err := testConfig.validateConfig()
		So(err.Error(), ShouldContainSubstring, "no workspace root has been set.\nPlease set a workspace root in the config")
	})

	Convey("Invalid coverage set in config.", t, func() {
		testConfig := &ButlerConfig{
			Paths: &ButlerPaths{
				AllowedPaths:  []string{"test_repo"},
				WorkspaceRoot: "test_repo",
			},
			Task: &TaskConfigurations{
				Coverage: "",
			},
			Languages: []*Language{testLanguage},
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
			err := testConfig.validateConfig()
			So(err.Error(), ShouldContainSubstring, test.expectedError)
		}
	})

	Convey("git repo set to false in the absence of a publish branch", t, func() {
		testConfig := &ButlerConfig{
			Paths: &ButlerPaths{
				AllowedPaths:  []string{"test_repo"},
				WorkspaceRoot: "test_repo",
			},
			Task: &TaskConfigurations{
				Coverage: "100",
			},
			Languages: []*Language{testLanguage},
			Git: &GitConfigurations{
				PublishBranch: "",
			},
		}

		err := testConfig.validateConfig()
		So(err, ShouldBeNil)
		So(testConfig.Git.GitRepo, ShouldBeFalse)
	})

	Convey("no languages added to config", t, func() {
		testConfig := &ButlerConfig{
			Paths: &ButlerPaths{
				AllowedPaths:  []string{"test_repo"},
				WorkspaceRoot: "test_repo",
			},
			Task: &TaskConfigurations{
				Coverage: "100",
			},
			Git: &GitConfigurations{
				PublishBranch: "main",
			},
		}

		err := testConfig.validateConfig()
		So(err.Error(), ShouldContainSubstring, "no languages have been provided in the config")
	})

	Convey("no allowed paths added to config", t, func() {
		testConfig := &ButlerConfig{
			Paths: &ButlerPaths{
				WorkspaceRoot: "test_repo",
			},
			Task: &TaskConfigurations{
				Coverage: "100",
			},
			Languages: []*Language{testLanguage},
			Git: &GitConfigurations{
				PublishBranch: "main",
			},
		}

		err := testConfig.validateConfig()
		So(err.Error(), ShouldContainSubstring, "butler has not been given permission to search for workspaces in any directories.")
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
