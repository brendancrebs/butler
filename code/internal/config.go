// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"errors"
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
	configFileName    = ".butler.base.yaml"
	ignoreFileName    = ".butler.ignore.yaml"
	gitCommand        = "git"
	butlerResultsPath = "./butler_results.json"
	defaultCoverage   = "0"
)

var (
	envShouldPublish, _ = strconv.ParseBool(getEnvOrDefault(envPublish, "false"))
	envBranchName       = strings.TrimSpace(getEnvOrDefault(envBranch, ""))
)

// ButlerPaths specifies the allowed and blocked paths within the .butler.ignore.yaml.
type ButlerPaths struct {
	AllowedPaths    []string `yaml:"allowed-paths,omitempty"`
	BlockedPaths    []string `yaml:"blocked-paths,omitempty"`
	CriticalPaths   []string `yaml:"critical-paths,omitempty"`
	WorkspaceRoot   string   `yaml:"workspace-root,omitempty"`
	ResultsFilePath string   `yaml:"results-file-path,omitempty"`
}

type GitConfigurations struct {
	PublishBranch string `yaml:"publish-branch,omitempty"`
	GitRepo       bool   `yaml:"git-repository,omitempty"`
}

type TaskConfigurations struct {
	Coverage      string `yaml:"coverage,omitempty"`
	ShouldRunAll  bool   `yaml:"should-run-all,omitempty"`
	ShouldLint    bool   `yaml:"should-lint,omitempty"`
	ShouldTest    bool   `yaml:"should-test,omitempty"`
	ShouldBuild   bool   `yaml:"should-build,omitempty"`
	ShouldPublish bool   `yaml:"should-publish,omitempty"`
}

// ButlerConfig specifies the Butler configuration options.
type ButlerConfig struct {
	Paths     *ButlerPaths        `yaml:"paths,omitempty"`
	Git       *GitConfigurations  `yaml:"git,omitempty"`
	Task      *TaskConfigurations `yaml:"tasks,omitempty"`
	Languages []*Language         `yaml:"languages,omitempty"`
}

func (bc *ButlerConfig) applyFlagsToConfig(cmd *cobra.Command, flags *ButlerConfig) {
	bc.Git.PublishBranch = useFlagIfChanged(bc.Git.PublishBranch, flags.Git.PublishBranch,
		cmd.Flags().Changed("publish-branch"))
	bc.Paths.WorkspaceRoot = useFlagIfChanged(bc.Paths.WorkspaceRoot, flags.Paths.WorkspaceRoot,
		cmd.Flags().Changed("workspace-root"))
	bc.Task.Coverage = useFlagIfChanged(bc.Task.Coverage, flags.Task.Coverage, cmd.Flags().Changed("coverage"))
	bc.Task.ShouldRunAll = useFlagIfChanged(bc.Task.ShouldRunAll, flags.Task.ShouldRunAll, cmd.Flags().Changed("all"))
	bc.Task.ShouldBuild = useFlagIfChanged(bc.Task.ShouldBuild, flags.Task.ShouldBuild, cmd.Flags().Changed("build"))
	bc.Task.ShouldLint = useFlagIfChanged(bc.Task.ShouldLint, flags.Task.ShouldLint, cmd.Flags().Changed("lint"))
	bc.Task.ShouldTest = useFlagIfChanged(bc.Task.ShouldTest, flags.Task.ShouldTest, cmd.Flags().Changed("test"))
	bc.Task.ShouldPublish = useFlagIfChanged(bc.Task.ShouldPublish, flags.Task.ShouldPublish,
		cmd.Flags().Changed("publish"))
}

// loads base Butler config
func (bc *ButlerConfig) Load(configPath string) (err error) {
	if _, err = os.Stat(configPath); err != nil {
		return
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(content, bc)
	if err != nil {
		return
	}

	err = bc.validateConfig()
	return
}

// Sets defaults for the Butler config by satisfying the go yaml package v2 unmarshaler interface.
// https://pkg.go.dev/gopkg.in/yaml.v2#Unmarshaler
func (bc *ButlerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type defaultConfig ButlerConfig
	defaults := defaultConfig{
		Paths: &ButlerPaths{
			ResultsFilePath: butlerResultsPath,
		},
		Git: &GitConfigurations{
			PublishBranch: envBranchName,
			GitRepo:       false,
		},
		Task: &TaskConfigurations{
			Coverage:      "0",
			ShouldRunAll:  false,
			ShouldLint:    false,
			ShouldTest:    false,
			ShouldBuild:   false,
			ShouldPublish: envShouldPublish,
		},
	}

	if err := unmarshal(&defaults); err != nil {
		return err
	}

	*bc = ButlerConfig(defaults)
	return nil
}

func (bc *ButlerConfig) validateConfig() error {
	if bc.Paths.WorkspaceRoot == "" {
		return errors.New("no workspace root has been set.\nPlease set a workspace root in the config")
	}
	if bc.Task.Coverage != "" {
		cover, err := strconv.Atoi(bc.Task.Coverage)
		if err != nil {
			return err
		}
		if cover < 0 || cover > 100 {
			return fmt.Errorf(`the test coverage threshold has been set to %d in the config. Please set coverage to a number between 0 and 100`, cover)
		}
	}
	if bc.Languages == nil {
		return errors.New(`no languages have been provided in the config`)
	}
	if bc.Paths.AllowedPaths == nil {
		return errors.New(`butler has not been given permission to search for workspaces in any directories.\n Please enter a list of directories in the 'allowed-paths' config field`)
	}
	if bc.Git.PublishBranch == "" {
		bc.Git.GitRepo = false
	}
	return nil
}

func (bc *ButlerConfig) String() string {
	configPretty, _ := yaml.Marshal(bc)
	return fmt.Sprintf(`Butler Configuration:\n
	Used config file %s\n
	<<<yaml\n
	# Configuration, including command line flags, used for the current run.\n
	%s>>>\n\n`, configPath, string(configPretty))
}

func (bc *ButlerConfig) LoadButlerIgnore() (err error) {
	ignorePath := path.Join(bc.Paths.WorkspaceRoot, ignoreFileName)
	if _, err := os.Stat(ignorePath); err != nil {
		return nil
	}

	content, err := os.ReadFile(ignorePath)
	if err != nil {
		return
	}

	paths := &ButlerPaths{}
	err = yaml.Unmarshal(content, &paths)

	bc.Paths.AllowedPaths = concatPaths(paths.AllowedPaths, bc.Paths.AllowedPaths)
	bc.Paths.BlockedPaths = concatPaths(paths.BlockedPaths, bc.Paths.BlockedPaths)
	bc.Paths.CriticalPaths = concatPaths(paths.CriticalPaths, bc.Paths.CriticalPaths)

	return
}

// concatenates two slices containing file paths. Each path in the resulting slice will be cleaned
// with filepath.Clean() and duplicates will be removed.
func concatPaths(sliceA, sliceB []string) (result []string) {
	sliceA = append(sliceA, sliceB...)
	result = sliceA[:0]
	seen := map[string]bool{}
	for _, value := range sliceA {
		cleanedValue := filepath.Clean(value)
		if seen[cleanedValue] {
			continue
		}
		seen[cleanedValue] = true
		result = append(result, cleanedValue)
	}

	return
}

// returns the value of the given environment variable if it has been set. Otherwise return the
// default.
func getEnvOrDefault(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}

// returns b if c is true, otherwise a.
func useFlagIfChanged[T any](a, b T, c bool) T {
	if c {
		return b
	}
	return a
}
