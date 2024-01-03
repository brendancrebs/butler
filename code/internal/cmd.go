// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"selinc.com/butler/code/helpers"
)

const (
	envBuildID         = "BUILD_ID"
	envRunAll          = "BUTLER_SHOULD_RUN_ALL"
	envPublish         = "BUTLER_SHOULD_PUBLISH"
	envBitbucketCommit = "GIT_COMMIT"
	envBranch          = "GIT_BRANCH"
)

func getCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "butler <flags>",
		Short: "butler is a build test lint runner",
		Long: `BuTLeR is a multi-threaded build, test, lint, and publish tool.  When not on ` +
			`the {publish-branch} it only runs against files diffed against the {publish-branch}.`,
		RunE: run,
	}

	parseFlags(root)

	return root
}

var (
	cmd        = getCommand()
	flags      = &ButlerConfig{}
	configPath string
)

// Execute is the entrypoint into the Butler
func Execute() {
	// Set Min/Max thread count for VITest Environment Variables to prevent timeouts from resource hogging.
	os.Setenv("VITEST_MIN_THREADS", "1")
	os.Setenv("VITEST_MAX_THREADS", "1")

	// all errors are handled internal to this call.
	_ = cmd.Execute()
}

func parseFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&configPath, "cfg", "/workspaces/butler/.butler.base.yaml", "Path to config file.")
	cmd.PersistentFlags().BoolVarP(&flags.ShouldRunAll, "all", "a", false, "Runs all tasks regardless of diff.")

	cmd.PersistentFlags().BoolVarP(&flags.ShouldLint, "lint", "l", false, "Enables linting")

	cmd.PersistentFlags().BoolVarP(&flags.ShouldTest, "test", "t", false, "Enables testing")

	cmd.PersistentFlags().BoolVarP(&flags.ShouldBuild, "build", "b", false, "Enables building")

	cmd.PersistentFlags().BoolVarP(&flags.ShouldPublish, "publish", "p", false,
		"Enables publishing.  Publishing also requires --publish-branch and --build-id.")

	cmd.PersistentFlags().StringVar(&flags.PublishBranch, "publish-branch", branchNameEnv,
		"Branch when we will publish or diff to, based on equality to current branch name.")

	cmd.PersistentFlags().StringVar(&flags.BuildID, "build-id", buildIDEnv,
		"Build-id value in the release tag.  Publish only when not equal to \"\".")

	cmd.PersistentFlags().StringVar(&flags.DockerfileSuffix, "dockerfile-suffix", ".x.Dockerfile",
		"The suffix on Dockerfiles that are intended to build and publish.")

	cmd.PersistentFlags().StringVar(&flags.TagDateFormat, "date-format", "060102",
		"Format of the date field in the release tag.")
	cmd.PersistentFlags().StringVar(&flags.ReleaseVersion, "release-version", "",
		"Enables a release build for the given version")
	cmd.PersistentFlags().StringVar(&flags.WorkspaceRoot, "workspace-root", "/workspaces/butler",
		"The root of the repository where Butler will start searching.")
}

func run(cmd *cobra.Command, args []string) (err error) {
	fmt.Printf("Hello world\n")
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %s", err)
	}
	config.applyFlagsToConfig(cmd, flags)

	ignorePaths, err := loadButlerIgnore(config)
	if err != nil {
		return
	}

	if ignorePaths != nil {
		config.Allowed = helpers.MergeMaps(ignorePaths.Allowed, config.Allowed)
		config.Blocked = helpers.MergeMaps(ignorePaths.Blocked, config.Blocked)
	}

	_ = printConfig(config, cmd)
	return
}
