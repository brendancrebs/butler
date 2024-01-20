// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	BaseConfigName    = ".butler.base.yaml"
	ignoreName        = ".butler.ignore.yaml"
	gitCommand        = "git"
	butlerResultsPath = "./butler_results.json"
)

var (
	shouldPublishEnv, _ = strconv.ParseBool(getEnvOrDefault(envPublish, "false"))
	branchNameEnv       = strings.TrimSpace(getEnvOrDefault(envBranch, ""))
)

// ButlerPaths specifies the allowed and blocked paths within the .butler.ignore.yaml.
type ButlerPaths struct {
	Allowed []string `yaml:"allowed-paths,omitempty"`
	Blocked []string `yaml:"blocked-paths,omitempty"`
}

// ButlerConfig specifies the Butler configuration options.
type ButlerConfig struct {
	Allowed         []string `yaml:"allowed-paths,omitempty"`
	Blocked         []string `yaml:"blocked-paths,omitempty"`
	CriticalPaths   []string `yaml:"critical-paths,omitempty"`
	PublishBranch   string   `yaml:"publish-branch,omitempty"`
	ResultsFilePath string   `yaml:"results-file-path,omitempty"`
	WorkspaceRoot   string   `yaml:"workspace-root,omitempty"`
	GitRepo         bool     `yaml:"git-repository,omitempty"`
	ShouldRunAll    bool     `yaml:"should-run-all,omitempty"`
	ShouldLint      bool     `yaml:"should-lint,omitempty"`
	ShouldTest      bool     `yaml:"should-test,omitempty"`
	ShouldBuild     bool     `yaml:"should-build,omitempty"`
	ShouldPublish   bool     `yaml:"should-publish,omitempty"`
}

func (bc *ButlerConfig) applyFlagsToConfig(cmd *cobra.Command, flags *ButlerConfig) {
	bc.PublishBranch = useFlagIfChangedString(bc.PublishBranch, flags.PublishBranch,
		cmd.Flags().Changed("publish-branch"))
	bc.WorkspaceRoot = useFlagIfChangedString(bc.WorkspaceRoot, flags.WorkspaceRoot,
		cmd.Flags().Changed("workspace-root"))
	bc.ShouldRunAll = useFlagIfChangedBool(bc.ShouldRunAll, flags.ShouldRunAll, cmd.Flags().Changed("all"))
	bc.ShouldBuild = useFlagIfChangedBool(bc.ShouldBuild, flags.ShouldBuild, cmd.Flags().Changed("build"))
	bc.ShouldLint = useFlagIfChangedBool(bc.ShouldLint, flags.ShouldLint, cmd.Flags().Changed("lint"))
	bc.ShouldTest = useFlagIfChangedBool(bc.ShouldTest, flags.ShouldTest, cmd.Flags().Changed("test"))
	bc.ShouldPublish = useFlagIfChangedBool(bc.ShouldPublish, flags.ShouldPublish,
		cmd.Flags().Changed("publish"))
}

// loads base Butler config
func (bc *ButlerConfig) Load(configPath string) (err error) {
	bc.ShouldPublish = shouldPublishEnv
	bc.PublishBranch = branchNameEnv
	bc.ResultsFilePath = butlerResultsPath

	if _, err = os.Stat(configPath); err != nil {
		return
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(content, bc)
	return
}

func (bc *ButlerConfig) String() string {
	configPretty, _ := yaml.Marshal(bc)
	return fmt.Sprintf(`Butler Configuration:\n
	Used config file %s\n
	<<<yaml\n
	# Configuration, including command line flags, used for the current run.\n
	%s>>>\n\n`, configPath, string(configPretty))
}

func loadButlerIgnore(bc *ButlerConfig) (paths *ButlerPaths, err error) {
	ignorePath := path.Join(bc.WorkspaceRoot, ignoreName)
	if _, err := os.Stat(ignorePath); err != nil {
		return nil, nil
	}

	content, err := os.ReadFile(ignorePath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(content, &paths)
	return
}

// useFlagIfChangedBool returns b if c is true, otherwise a.
func useFlagIfChangedBool(a, b, c bool) bool {
	if c {
		return b
	}
	return a
}

// useFlagIfChangedString returns b if c is true, otherwise a.
func useFlagIfChangedString(a, b string, c bool) string {
	if c {
		return b
	}
	return a
}

// Uses filepath.Clean() to update the allowed/blocked filepath strings to a consistent format.
func cleanPaths(paths []string) (cleanedPaths []string) {
	for _, path := range paths {
		cleanedPath := filepath.Clean(path)
		cleanedPaths = append(cleanedPaths, cleanedPath)
	}

	return
}
