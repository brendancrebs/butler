// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func ButlerSetup(bc *ButlerConfig, cmd *cobra.Command) (err error) {
	allPaths := getFilePaths([]string{bc.WorkspaceRoot}, true, bc.Allowed, bc.Blocked)
	allDirtyPaths, err := getDirtyPaths(bc.PublishBranch)
	if err != nil {
		return err
	}

	if err = shouldBuildAll(bc, allDirtyPaths); err != nil {
		return err
	}

	fmt.Printf("allPaths: %v\n", allPaths)

	return
}

// Determines if Butler requires a full build.
func shouldBuildAll(bc *ButlerConfig, allDirtyPaths []string) (err error) {
	// get current git branch name
	currentBranch, err := getCurrentBranch()
	if err != nil {
		return
	}

	rebuildAll := strings.EqualFold(getEnvOrDefault(envRunAll, ""), "true") || currentBranch == bc.PublishBranch
	bc.ShouldPublish = currentBranch == bc.PublishBranch

	criticalFiles, criticalFolders, err := separateCriticalFiles(bc.WorkspaceRoot, bc.CriticalPaths)
	rebuildAll = rebuildAll || criticalFileChanged(allDirtyPaths, criticalFiles)
	dirtyFolders := getUniqueFolders(allDirtyPaths)
	rebuildAll = rebuildAll || criticalFolderChanged(dirtyFolders, criticalFolders)

	bc.ShouldRunAll = bc.ShouldRunAll || rebuildAll

	return
}

// getCurrentBranch is returns the whitespace trimmed result of `git branch --show-current`
// which should be the branch, or an error.
func getCurrentBranch() (branch string, err error) {
	path, err := execLookPathStub("git")
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

// getFilePaths takes a set of directories and returns all of the filepaths from them.
// If `shouldRecurse`, then it calls recurseFilepaths for each folder.
func getFilePaths(dirs []string, shouldRecurse bool, allowed, blocked map[string]bool) []string {
	paths := make([]string, 0)

	for _, base := range dirs {
		paths = recurseFilepaths(base, paths, shouldRecurse, allowed, blocked)
	}

	sort.Strings(paths)
	return paths
}

// recurseFilepaths reads the directory at `path` and either appends filepaths or appends the result
// of a further call to recurseFilepaths.
func recurseFilepaths(path string, paths []string, shouldRecurse bool, allowed, blocked map[string]bool) []string {
	fileInfos, _ := os.ReadDir(path)
	for _, fi := range fileInfos {
		path := filepath.Join(path, fi.Name())
		if !allowedAndNotBlocked(path, allowed, blocked) {
			continue
		}
		if !fi.IsDir() {
			paths = append(paths, path)
		} else if shouldRecurse {
			paths = recurseFilepaths(path, paths, shouldRecurse, allowed, blocked)
		}
	}
	return paths
}

func allowedAndNotBlocked(path string, allowed, blocked map[string]bool) bool {
	isAllowed := false
	for key := range allowed {
		if strings.Contains(path, key) {
			isAllowed = true
			break
		}
	}

	isBlocked := false
	for key := range blocked {
		if strings.Contains(path, key) {
			isBlocked = true
			break
		}
	}

	return isAllowed && !isBlocked
}

// getDirtyPaths calls 'git diff --name-only {branch}' if branch is not blank, or
// 'git diff --name-only' without a branch.  It returns a list of file names that
// changed.
func getDirtyPaths(branch string) ([]string, error) {
	branch = strings.TrimSpace(branch)
	path, err := execLookPathStub("git")
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

	output, err := cmd.Output()

	return getLines(output, []byte{'\n'}), err
}

// getLines splits on the splitOn slice, converts each grouping to a string and trims space from it.
// Returns the set of converted liens.
func getLines(input, splitOn []byte) (lines []string) {
	lines = make([]string, 0)
	split := bytes.Split(input, splitOn)
	for _, lineBytes := range split {
		line := bytes.TrimSpace(lineBytes)
		if len(line) > 0 {
			lines = append(lines, string(line))
		}
	}
	sort.Strings(lines)
	return
}

func separateCriticalFiles(workspaceRoot string, criticalPaths []string) (criticalFiles []string, criticalFolders []string, err error) {
	var fi fs.FileInfo
	for _, path := range criticalPaths {
		path := filepath.Join(workspaceRoot, path)
		fi, err = os.Stat(path)
		if err != nil {
			break
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			criticalFolders = append(criticalFolders, path)
		case mode.IsRegular():
			criticalFiles = append(criticalFiles, path)
		}
	}
	return
}

func criticalFileChanged(dirtyPaths []string, criticalFiles []string) (result bool) {
	for _, file := range criticalFiles {
		for _, dirtyFile := range dirtyPaths {
			if dirtyFile == file {
				result = true
				return
			}
		}
	}
	return
}

func criticalFolderChanged(dirtyFolders []string, criticalFolders []string) (result bool) {
	for _, folder := range criticalFolders {
		for _, dirtyFolder := range dirtyFolders {
			if strings.HasPrefix(dirtyFolder, folder) {
				result = true
				break
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

	return getSortedKeys(folderMap)
}

// getSortedKeys returns the keys sorted in ascending order.
func getSortedKeys(mapStringBool map[string]bool) []string {
	keys := make([]string, 0)
	for key := range mapStringBool {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
