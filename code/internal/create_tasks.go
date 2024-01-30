// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// This function will return a queue populated with tasks
func ButlerSetup(bc *ButlerConfig, cmd *cobra.Command) (err error) {
	allPaths := getFilePaths([]string{bc.Paths.WorkspaceRoot}, bc.Paths.AllowedPaths, bc.Paths.BlockedPaths, true)

	if bc.Git.GitRepo && !bc.Task.ShouldRunAll {
		allDirtyPaths, err := getDirtyPaths(bc.Git.PublishBranch)
		if err != nil {
			return err
		}
		dirtyFolders := getUniqueFolders(allDirtyPaths)

		if err = shouldRunAll(bc, allDirtyPaths, dirtyFolders); err != nil {
			return err
		}
	} else {
		bc.Task.ShouldRunAll = true
	}

	fmt.Printf("\nallPaths: %v\n\n", allPaths)

	// Next steps:

	// 1. run preliminary commands

	// 2. create workspace array for each language

	// 3. get internal dependencies for each language

	// 4. get external dependencies for each language

	// 5. determine dirty workspaces

	// 6. create tasks for each language

	// 7. Return the populated task queue

	return
}

// Determines if Butler requires a full build.
func shouldRunAll(bc *ButlerConfig, allDirtyPaths, dirtyFolders []string) (err error) {
	// get current git branch name
	currentBranch, err := getCurrentBranch()
	if err != nil {
		return
	}

	bc.Task.ShouldRunAll = strings.EqualFold(getEnvOrDefault(envRunAll, ""), "true") || currentBranch == bc.Git.PublishBranch
	bc.Task.ShouldPublish = currentBranch == bc.Git.PublishBranch
	bc.Task.ShouldRunAll = bc.Task.ShouldRunAll || criticalPathChanged(allDirtyPaths, bc.Paths.CriticalPaths)

	return
}

// getCurrentBranch returns the whitespace trimmed result of `git branch --show-current`
// which should be the branch, or an error.
func getCurrentBranch() (branch string, err error) {
	path, err := execLookPathStub(gitCommand)
	if err != nil {
		return
	}

	cmd := &exec.Cmd{
		Path: path,
		Args: []string{path, "branch", "--show-current"},
	}

	execCmd, err := execOutputStub(cmd)
	if err != nil {
		return
	}

	branch = getEnvOrDefault(envBranch, strings.TrimSpace(string(execCmd)))
	return
}

// Takes a set of directories and calls recurseFilePath on each. recurseFilePaths will return the
// children of each directory and all of the child files will get returned by getFilePaths.
func getFilePaths(dirs, allowed, blocked []string, shouldRecurse bool) []string {
	paths := make([]string, len(dirs))

	for _, base := range dirs {
		paths = recurseFilepaths(paths, allowed, blocked, base, shouldRecurse)
	}

	return paths
}

// Reads the directory at `path` and either appends filepaths or appends the result of a further
// call to recurseFilepaths.
func recurseFilepaths(paths, allowed, blocked []string, path string, shouldRecurse bool) []string {
	fileInfos, _ := os.ReadDir(path)
	for _, fi := range fileInfos {
		path := filepath.Join(path, fi.Name())
		if !isAllowed(path, allowed, blocked) {
			continue
		}
		if !fi.IsDir() {
			paths = append(paths, path)
		} else if shouldRecurse {
			paths = recurseFilepaths(paths, allowed, blocked, path, shouldRecurse)
		}
	}
	return paths
}

func isAllowed(path string, allowed, blocked []string) bool {
	isAllowed := false
	for _, key := range allowed {
		if strings.Contains(path, key) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return false
	}

	isBlocked := false
	for _, key := range blocked {
		if strings.Contains(path, key) {
			isBlocked = true
			break
		}
	}

	return !isBlocked
}

// getDirtyPaths calls 'git diff --name-only {branch}' if branch is not blank, or
// 'git diff --name-only' without a branch.  It returns a list of file names that
// changed.
func getDirtyPaths(branch string) ([]string, error) {
	branch = strings.TrimSpace(branch)
	path, err := execLookPathStub(gitCommand)
	if err != nil {
		return nil, err
	}

	cmd := &exec.Cmd{
		Path: path,
		Args: []string{path, "diff", "--name-only"},
	}
	if branch != "" {
		cmd.Args = append(cmd.Args, branch)
	}

	output, err := execOutputStub(cmd)

	return getLines(output, []byte{'\n'}), err
}

// getLines splits on the splitOn slice, converts each grouping to a string and trims space from it.
// Returns the set of converted lines.
func getLines(input, splitOn []byte) (lines []string) {
	split := bytes.Split(input, splitOn)
	lines = make([]string, len(split))
	for _, lineBytes := range split {
		line := bytes.TrimSpace(lineBytes)
		if len(line) > 0 {
			lines = append(lines, string(line))
		}
	}
	return
}

func criticalPathChanged(dirtyPaths, criticalPaths []string) (result bool) {
	for _, path := range criticalPaths {
		for _, dirtyPath := range dirtyPaths {
			if dirtyPath == path || strings.HasPrefix(dirtyPath, path) {
				result = true
				return
			}
		}
	}
	return
}

// getUniqueFolders returns a sorted set of unique folders based on a set of paths.
func getUniqueFolders(paths []string) []string {
	folderMap := make(map[string]bool)
	for _, path := range paths {
		path = filepath.Dir(path)
		folderMap[path] = true
	}

	keys := make([]string, len(folderMap))
	for key := range folderMap {
		keys = append(keys, key)
	}

	return keys
}
