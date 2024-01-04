// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"selinc.com/butler/code/helpers"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	defaultRegistry = "hop-docker-dev.artifactory.metro.ad.selinc.com"
	tagDateFormat   = "060102"
	BaseConfigName  = ".butler.base.yaml"
	ignoreName      = ".butler.ignore.yaml"
)

var (
	shouldPublishEnv, _ = strconv.ParseBool(helpers.GetEnvOrDefault(envPublish, "false"))
	buildIDEnv          = strings.TrimSpace(helpers.GetEnvOrDefault(envBuildID, ""))
	branchNameEnv       = strings.TrimSpace(helpers.GetEnvOrDefault(envBranch, ""))
)

// ButlerPaths specifies the allowed and blocked paths within the .butler.ignore.yaml.
type ButlerPaths struct {
	Allowed map[string]bool `yaml:"allowed-paths,omitempty"`
	Blocked map[string]bool `yaml:"blocked-paths,omitempty"`
}

// ButlerConfig specifies the Butler configuration options.
type ButlerConfig struct {
	Allowed         map[string]bool `yaml:"allowed-paths,omitempty"`
	Blocked         map[string]bool `yaml:"blocked-paths,omitempty"`
	Registry        string          `yaml:"Registry,omitempty"`
	PublishBranch   string          `yaml:"publish-branch,omitempty"`
	BuildID         string          `yaml:"build-id,omitempty"`
	TagDateFormat   string          `yaml:"tag-date-format,omitempty"`
	ResultsFilePath string          `yaml:"results-file-path,omitempty"`
	ReleaseVersion  string          `yaml:"release-version,omitempty"`
	WorkspaceRoot   string          `yaml:"workspace-root,omitempty"`
	CriticalPaths   []string        `yaml:"critical-paths,omitempty"`
	ShouldRunAll    bool            `yaml:"should-run-all,omitempty"`
	ShouldLint      bool            `yaml:"should-lint,omitempty"`
	ShouldTest      bool            `yaml:"should-test,omitempty"`
	ShouldBuild     bool            `yaml:"should-build,omitempty"`
	ShouldPublish   bool            `yaml:"should-publish,omitempty"`
}

func (bc *ButlerConfig) applyFlagsToConfig(cmd *cobra.Command, flags *ButlerConfig) {
	bc.PublishBranch = useFlagIfChangedString(bc.PublishBranch, flags.PublishBranch,
		cmd.Flags().Changed("publish-branch"))
	bc.BuildID = useFlagIfChangedString(bc.BuildID, flags.BuildID,
		cmd.Flags().Changed("build-id"))
	bc.TagDateFormat = useFlagIfChangedString(bc.TagDateFormat, flags.TagDateFormat,
		cmd.Flags().Changed("date-format"))
	bc.ReleaseVersion = useFlagIfChangedString(bc.ReleaseVersion, flags.ReleaseVersion,
		cmd.Flags().Changed("release-version"))
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
func loadConfig() (config *ButlerConfig, err error) {
	config = &ButlerConfig{
		// load config with env vars.
		Registry:        defaultRegistry,
		BuildID:         buildIDEnv,
		ShouldPublish:   shouldPublishEnv,
		PublishBranch:   branchNameEnv,
		TagDateFormat:   tagDateFormat,
		ResultsFilePath: "./butler_results.json",
	}

	if _, err := os.Stat(configPath); err != nil {
		return nil, err
	}
	content, _ := os.ReadFile(configPath)
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, fmt.Errorf("Configuration parse error: %w", err)
	}

	config.Allowed = appendSep(config.Allowed)
	config.Blocked = appendSep(config.Blocked)

	return
}

// loads butler ignore file if it exists
func loadButlerIgnore(bc *ButlerConfig) (paths *ButlerPaths, err error) {
	ignorePath := path.Join(bc.WorkspaceRoot, ignoreName)
	if _, err := os.Stat(ignorePath); err != nil {
		return nil, nil
	}
	content, _ := os.ReadFile(ignorePath)
	err = yaml.Unmarshal(content, &paths)
	if err != nil {
		return nil, fmt.Errorf("Butler ignore parse error: %w", err)
	}

	paths.Allowed = appendSep(paths.Allowed)
	paths.Blocked = appendSep(paths.Blocked)
	return
}

func printConfig(config *ButlerConfig, cmd *cobra.Command) (err error) {
	configPretty, err := yaml.Marshal(config)

	fmt.Fprintf(cmd.OutOrStdout(), "Butler Configuration:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Used config file %s\n", configPath)
	fmt.Fprintf(cmd.OutOrStdout(), "<<<yaml\n")
	fmt.Fprintf(cmd.OutOrStdout(), "# Configuration, including command line flags, used for the current run.\n")
	fmt.Fprintf(cmd.OutOrStdout(), "%s", string(configPretty))
	fmt.Fprintf(cmd.OutOrStdout(), ">>>\n\n")

	return
}

// useFlagIfChangedBool returns b if c or a.
//
//nolint:unparam // okay always false.
func useFlagIfChangedBool(a, b, c bool) bool {
	if c {
		return b
	}
	return a
}

// useFlagIfChangedString returns b if c or a.
//
//nolint:unparam // okay always "".
func useFlagIfChangedString(a, b string, c bool) string {
	if c {
		return b
	}
	return a
}

func appendSep(in map[string]bool) map[string]bool {
	out := make(map[string]bool)

	for key := range in {
		if strings.HasSuffix(key, "/") {
			out[key] = true
			continue
		}

		out[key+"/"] = true
	}

	return out
}
