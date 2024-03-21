// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"selinc.com/butler/code/butler"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_loadConfig(t *testing.T) {
	Convey("Ability to parse a yaml config for all of its values", t, func() {
		temp := butler.ConfigPath
		butler.ConfigPath = "./test_data/test_configs/.butler.base.yaml"
		wd, _ := os.Getwd()
		expectedRoot := filepath.Join(wd, "./test_data")

		expectedLanguage := &butler.Language{
			Name:                     "golang",
			FilePatterns:             []string{".go"},
			BuiltinStdLibsMethod:     true,
			BuiltinExternalDepMethod: true,
			TaskExec: &butler.TaskCommands{
				LintCommand:    "echo go lint command",
				TestCommand:    "echo go test command",
				BuildCommand:   "echo go build command",
				PublishCommand: "echo go publish command",
				SetUpCommands:  []string{"echo go test"},
			},
		}

		config := &butler.ButlerConfig{}
		expected := &butler.ButlerConfig{
			PublishBranch: "main",
			Paths: &butler.ButlerPaths{
				AllowedPaths:    []string{"good_path", "test_repo"},
				BlockedPaths:    []string{"bad_path", "blocked_dir"},
				WorkspaceRoot:   expectedRoot,
				ResultsFilePath: "./butler_results.json",
			},
			Task: &butler.TaskConfigurations{
				Coverage:      "0",
				ShouldLint:    true,
				ShouldTest:    true,
				ShouldBuild:   false,
				ShouldPublish: false,
				ShouldRunAll:  false,
			},
			Languages: []*butler.Language{expectedLanguage},
		}

		err := config.Load(butler.ConfigPath)
		butler.ConfigPath = temp

		So(err, ShouldBeNil)
		So(config, ShouldResemble, expected)
	})

	Convey("Failure to parse config file is covered", t, func() {
		temp := butler.ConfigPath
		butler.ConfigPath = "./test_configs/invalid.butler.bad"

		config := &butler.ButlerConfig{}
		expected := &butler.ButlerConfig{}

		err := config.Load(butler.ConfigPath)
		butler.ConfigPath = temp

		So(err, ShouldNotBeNil)
		So(config, ShouldResemble, expected)
	})
}

func Test_validateConfig(t *testing.T) {
	testLanguage := &butler.Language{
		Name: "test",
		TaskExec: &butler.TaskCommands{
			SetUpCommands: []string{"echo test"},
		},
		DepCommands: &butler.DependencyCommands{
			ExternalDepCommand: "echo get deps",
		},
	}

	Convey("Catch config that doesn't have workspace root set", t, func() {
		testConfig := &butler.ButlerConfig{
			Paths: &butler.ButlerPaths{
				AllowedPaths: []string{"test_repo"},
			},
			Task: &butler.TaskConfigurations{
				Coverage: "100",
			},
			Languages: []*butler.Language{testLanguage},
		}
		err := testConfig.ValidateConfig()
		So(err.Error(), ShouldContainSubstring, "no workspace root has been set.\nPlease set a workspace root in the config")
	})

	Convey("Invalid workspace root", t, func() {
		ws, _ := os.Getwd()
		ws = filepath.Join(ws, "fail")
		testConfig := &butler.ButlerConfig{
			Paths: &butler.ButlerPaths{
				WorkspaceRoot: ws,
			},
		}
		expectedError := fmt.Sprintf("chdir %v: no such file or directory", ws)
		err := testConfig.ValidateConfig()
		So(err.Error(), ShouldContainSubstring, expectedError)
	})

	Convey("Invalid coverage set in config.", t, func() {
		testConfig := &butler.ButlerConfig{
			Paths: &butler.ButlerPaths{
				AllowedPaths:  []string{"test_repo"},
				WorkspaceRoot: ".",
			},
			Task: &butler.TaskConfigurations{
				Coverage: "",
			},
			Languages: []*butler.Language{testLanguage},
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

	Convey("no languages added to config", t, func() {
		testConfig := &butler.ButlerConfig{
			PublishBranch: "main",
			Paths: &butler.ButlerPaths{
				AllowedPaths:  []string{"test_repo"},
				WorkspaceRoot: ".",
			},
			Task: &butler.TaskConfigurations{
				Coverage: "100",
			},
		}

		err := testConfig.ValidateConfig()
		So(err.Error(), ShouldContainSubstring, "no languages have been provided in the config")
	})

	Convey("no allowed paths added to config", t, func() {
		testConfig := &butler.ButlerConfig{
			PublishBranch: "main",
			Paths: &butler.ButlerPaths{
				WorkspaceRoot: ".",
			},
			Task: &butler.TaskConfigurations{
				Coverage: "100",
			},
			Languages: []*butler.Language{testLanguage},
		}

		err := testConfig.ValidateConfig()
		So(err.Error(), ShouldContainSubstring, "butler has not been given permission to search for workspaces in any directories.")
	})
}

func Test_loadButlerIgnore(t *testing.T) {
	Convey("Paths successfully parsed from .butler.ignore.", t, func() {
		testConfig := &butler.ButlerConfig{
			Paths: &butler.ButlerPaths{
				WorkspaceRoot: ".",
			},
		}
		temp := butler.ConfigPath
		butler.ConfigPath = "./test_configs/.butler.base.yaml"
		expected := &butler.ButlerConfig{
			Paths: &butler.ButlerPaths{
				AllowedPaths:  []string{"good_path"},
				BlockedPaths:  []string{"bad_path", "blocked_dir"},
				WorkspaceRoot: ".",
			},
		}
		err := testConfig.LoadButlerIgnore()
		butler.ConfigPath = temp

		So(err, ShouldBeNil)
		So(testConfig, ShouldResemble, expected)
	})
	Convey("Nothing returned if no .butler.ignore file is found.", t, func() {
		temp := butler.ConfigPath
		butler.ConfigPath = "./test_data/test_configs/bad_ignore_configs/missing_ignore_dir/.butler.base.yaml"
		config := &butler.ButlerConfig{
			Paths: &butler.ButlerPaths{},
		}
		_ = config.Load(butler.ConfigPath)
		err := config.LoadButlerIgnore()
		butler.ConfigPath = temp

		So(err, ShouldBeNil)
		So(config.Paths.AllowedPaths, ShouldBeNil)
		So(config.Paths.BlockedPaths, ShouldBeNil)
	})
	Convey("Failure to parse .butler.ignore.", t, func() {
		temp := butler.ConfigPath
		butler.ConfigPath = "./test_configs/bad_ignore_configs/.butler.base.yaml"
		testConfig := &butler.ButlerConfig{
			Paths: &butler.ButlerPaths{
				WorkspaceRoot: ".",
			},
		}

		err := testConfig.LoadButlerIgnore()
		butler.ConfigPath = temp

		So(err, ShouldNotBeNil)
		So(testConfig.Paths.AllowedPaths, ShouldEqual, []string{"good_path"})
		So(testConfig.Paths.BlockedPaths, ShouldBeNil)
	})
}
