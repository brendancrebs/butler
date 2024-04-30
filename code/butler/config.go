// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

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
	envShouldPublish, _ = strconv.ParseBool(GetEnvOrDefault(envPublish, "false"))
	envBranchName       = strings.TrimSpace(GetEnvOrDefault(envBranch, ""))
)

// ButlerPaths specifies the allowed and ignored paths within the .butler.ignore.yaml.
type ButlerPaths struct {
	AllowedPaths    []string `yaml:"allowedPaths,omitempty"`
	IgnorePaths     []string `yaml:"ignorePaths,omitempty"`
	CriticalPaths   []string `yaml:"criticalPaths,omitempty"`
	WorkspaceRoot   string   `yaml:"workspaceRoot,omitempty"`
	ResultsFilePath string   `yaml:"resultsFilePath,omitempty"`
}

type TaskConfigurations struct {
	Coverage string `yaml:"coverage,omitempty"`
	RunAll   bool   `yaml:"runAll,omitempty"`
	Lint     bool   `yaml:"lint,omitempty"`
	Test     bool   `yaml:"test,omitempty"`
	Build    bool   `yaml:"build,omitempty"`
	Publish  bool   `yaml:"publish,omitempty"`
}

// ButlerConfig specifies the Butler configuration options.
type ButlerConfig struct {
	PublishBranch string              `yaml:"publishBranch,omitempty"`
	Paths         *ButlerPaths        `yaml:"paths,omitempty"`
	Task          *TaskConfigurations `yaml:"tasks,omitempty"`
	Languages     []*Language         `yaml:"languages,omitempty"`
	Subscribers   []string            `yaml:"resultSubscribers,omitempty"`
}

// updates the config settings with values passed using the cli.
func (bc *ButlerConfig) applyFlagsToConfig(cmd *cobra.Command, flags *ButlerConfig) {
	bc.PublishBranch = useFlagIfChanged(bc.PublishBranch, flags.PublishBranch,
		cmd.Flags().Changed("publish-branch"))
	bc.Paths.WorkspaceRoot = useFlagIfChanged(bc.Paths.WorkspaceRoot, flags.Paths.WorkspaceRoot,
		cmd.Flags().Changed("workspace-root"))
	bc.Task.Coverage = useFlagIfChanged(bc.Task.Coverage, flags.Task.Coverage, cmd.Flags().Changed("coverage"))
	bc.Task.RunAll = useFlagIfChanged(bc.Task.RunAll, flags.Task.RunAll, cmd.Flags().Changed("all"))
	bc.Task.Build = useFlagIfChanged(bc.Task.Build, flags.Task.Build, cmd.Flags().Changed("build"))
	bc.Task.Lint = useFlagIfChanged(bc.Task.Lint, flags.Task.Lint, cmd.Flags().Changed("lint"))
	bc.Task.Test = useFlagIfChanged(bc.Task.Test, flags.Task.Test, cmd.Flags().Changed("test"))
	bc.Task.Publish = useFlagIfChanged(bc.Task.Publish, flags.Task.Publish,
		cmd.Flags().Changed("publish"))
}

// loads Butler config.
func (bc *ButlerConfig) Load(configPath string) (err error) {
	if _, err = os.Stat(configPath); err != nil {
		return
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(content, bc); err != nil {
		return
	}

	if err = bc.LoadButlerIgnore(); err != nil {
		return
	}

	err = bc.ValidateConfig()
	return
}

// Sets defaults for the Butler config by satisfying the go yaml package v2 unmarshaler interface.
// https://pkg.go.dev/gopkg.in/yaml.v2#Unmarshaler
func (bc *ButlerConfig) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	type defaultConfig ButlerConfig
	defaults := defaultConfig{
		Paths: &ButlerPaths{
			ResultsFilePath: butlerResultsPath,
		},
		Task: &TaskConfigurations{
			Coverage: "0",
			RunAll:   false,
			Lint:     false,
			Test:     false,
			Build:    false,
			Publish:  envShouldPublish,
		},
	}

	if err = unmarshal(&defaults); err != nil {
		return
	}

	*bc = ButlerConfig(defaults)
	return
}

// Checks that fields critical for Butler operation have been correctly supplied to the config file.
func (bc *ButlerConfig) ValidateConfig() (err error) {
	if err = bc.setWorkspace(); err != nil {
		return
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
	for _, lang := range bc.Languages {
		if err = lang.validateLanguage(); err != nil {
			return
		}
	}
	return
}

// validates that the workspace root has been set in the config. If so Butler will cd into the
// workspace root directory and set the WORKSPACE_ROOT environment variable.
func (bc *ButlerConfig) setWorkspace() (err error) {
	if bc.Paths.WorkspaceRoot == "" {
		return errors.New("no workspace root has been set.\nPlease set a workspace root in the config")
	}

	bc.Paths.WorkspaceRoot, _ = filepath.Abs(bc.Paths.WorkspaceRoot)

	if err = os.Chdir(bc.Paths.WorkspaceRoot); err != nil {
		return
	}
	os.Setenv(envWorkspaceRoot, bc.Paths.WorkspaceRoot)
	return
}

// Uses go stringer interface for printing the Butler config.
func (bc *ButlerConfig) String() string {
	configPretty, _ := yaml.Marshal(bc)
	return fmt.Sprintf(`Butler Configuration:\n
	Used config file %s\n
	<<<yaml\n
	# Configuration, including command line flags, used for the current run.\n
	%s>>>\n\n`, ConfigPath, string(configPretty))
}

func (bc *ButlerConfig) LoadButlerIgnore() (err error) {
	ignorePath := path.Join(filepath.Dir(ConfigPath), ignoreFileName)
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
	bc.Paths.IgnorePaths = concatPaths(paths.IgnorePaths, bc.Paths.IgnorePaths)
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
func GetEnvOrDefault(name, defaultValue string) string {
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
