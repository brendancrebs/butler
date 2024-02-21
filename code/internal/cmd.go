// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
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
	cmd   = getCommand()
	flags = &ButlerConfig{
		Paths: &ButlerPaths{},
		Git:   &GitConfigurations{},
		Task:  &TaskConfigurations{},
	}
	execOutputStub = (*exec.Cmd).Output
	configPath     string
)

// Execute is the entrypoint into the Butler
func Execute() {
	// all errors are handled internal to this call.
	_ = cmd.Execute()
}

func parseFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&configPath, "cfg", ".butler.base.yaml", "Path to config file.")
	cmd.PersistentFlags().StringVarP(&flags.Task.Coverage, "coverage", "c", "0",
		"the percentage of code coverage that is acceptable for tests to pass.")
	cmd.PersistentFlags().BoolVarP(&flags.Task.ShouldRunAll, "all", "a", false, "Runs all tasks regardless of diff.")

	cmd.PersistentFlags().BoolVarP(&flags.Task.ShouldLint, "lint", "l", false, "Enables linting")

	cmd.PersistentFlags().BoolVarP(&flags.Task.ShouldTest, "test", "t", false, "Enables testing")

	cmd.PersistentFlags().BoolVarP(&flags.Task.ShouldBuild, "build", "b", false, "Enables building")

	cmd.PersistentFlags().BoolVarP(&flags.Task.ShouldPublish, "publish", "p", false,
		"Enables publishing.  Publishing also requires --publish-branch and --build-id.")

	cmd.PersistentFlags().BoolVarP(&flags.Git.GitRepo, "git-repository", "g", false,
		"sets whether the repository is a git repository or not.")

	cmd.PersistentFlags().StringVar(&flags.Git.PublishBranch, "publish-branch", envBranchName,
		"Branch when we will publish or diff to, based on equality to current branch name.")

	cmd.PersistentFlags().StringVar(&flags.Paths.WorkspaceRoot, "workspace-root", ".",
		"The root of the repository where Butler will start searching.")
}

func run(cmd *cobra.Command, args []string) (err error) {
	var taskQueue *Queue

	config := &ButlerConfig{}
	if err = config.Load(configPath); err != nil {
		return
	}
	config.applyFlagsToConfig(cmd, flags)

	if err = config.LoadButlerIgnore(); err != nil {
		return
	}

	fmt.Fprintln(cmd.OutOrStdout(), config)

	taskQueue, err = GetTasks(config, cmd)
	if err != nil {
		return
	}

	defer publishResults(config)

	for _, task := range taskQueue.tasks {
		fmt.Printf("\ntask: %+v", task)
	}

	return
}

// For now this just prints the config to Butler results since no tasks are being created.
func publishResults(bc *ButlerConfig) {
	resultBytes, _ := json.MarshalIndent(bc, "", "\t")
	_ = os.WriteFile(bc.Paths.ResultsFilePath, resultBytes, 0o600)
}

// // goGetModFileDiff returns the list of changed dependencies listed in the mod file.
// func goGetChangedModFileDeps(branch string) (changeSet []string, languageVersionChanged bool, err error) {
// 	err = filepath.WalkDir(workspaceRoot, func(path string, d fs.DirEntry, walkErr error) error {
// 		walkErr = ifErrNil(walkErr, func() (innerErr error) {
// 			if d.IsDir() {
// 				return nil
// 			}
// 			if filepath.Base(path) == "go.mod" && !strings.Contains(path, "tools") {
// 				var cs []string
// 				cs, innerErr = singleFileDiff(path, branch)
// 				changeSet = append(changeSet, cs...)
// 			}
// 			return
// 		})

// 		return walkErr
// 	})

// 	err = ifErrNil(err, func() error {
// 		languageVersionChanged = goDidLanguageVersionChange(changeSet)
// 		return nil
// 	})

// 	return
// }

// // goDidLanguageVersionChange checks a diff in the form of []string for a line that contains "go".
// // Preprocessing of the diff should limit it to only lines with a leading '+'.
// func goDidLanguageVersionChange(modDiff []string) (didChange bool) {
// 	for _, line := range modDiff {
// 		if strings.Split(line, " ")[0] == "go" {
// 			didChange = true
// 			break
// 		}
// 	}
// 	return
// }

// // singleFileDiff calls 'git diff {filename} {branch}'
// // if filename is empty it returns nil.
// // if branch is empty, then it is with respect to the last commit on the same branch.
// // it returns the list of the first content of each line that changed, split on spaces.
// // Used to get diff information from the go.mod file and results in changed packages.
// func singleFileDiff(filename, branch string) (changes []string, err error) {
// 	// inverted error logic
// 	filename = strings.TrimSpace(filename)
// 	branch = strings.TrimSpace(branch)
// 	var path string

// 	if path, err = exec.LookPath("git"); err == nil {
// 		cmd := &exec.Cmd{
// 			Path: path,
// 			Args: []string{path, "diff"},
// 		}
// 		if branch != "" {
// 			cmd.Args = append(cmd.Args, branch)
// 		}
// 		cmd.Args = append(cmd.Args, []string{"--", filename}...)

// 		var b []byte
// 		if b, err = cmd.Output(); err == nil {
// 			changes = pruneAdditiveChanges(getLines(b, []byte{'\n'}))
// 		}
// 	}

// 	return
// }
