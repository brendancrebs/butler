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

type Queue struct {
	tasks []*Task
}

func (q *Queue) Enqueue(task *Task) {
	q.tasks = append(q.tasks, task)
}

func (q *Queue) Dequeue() *Task {
	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task
}

func (q *Queue) Size() int {
	return len(q.tasks)
}

// This function will return a queue populated with tasks
func GetTasks(bc *ButlerConfig, cmd *cobra.Command) (taskQueue *Queue, err error) {
	taskQueue = &Queue{tasks: make([]*Task, 0)}
	allPaths, dirtyFolders, err := butlerSetup(bc)
	if err != nil {
		return
	}

	fmt.Printf("\nallPaths: %v\n\n", allPaths)

	if err = preliminaryCommands(bc.Languages); err != nil {
		return
	}

	for _, lang := range bc.Languages {
		if err = lang.getExternalDeps(bc); err != nil {
			return
		}

		lang.Workspaces, err = getWorkspaces(lang, allPaths)
		if err != nil {
			return nil, fmt.Errorf("Error getting workspaces for %s:\n%v\n", lang.Name, err)
		}
	}

	for _, lang := range bc.Languages {
		evaluateDirtiness(lang.Workspaces, dirtyFolders)
	}

	if err = PopulateTaskQueue(bc, taskQueue, cmd); err != nil {
		return
	}

	return
}

func butlerSetup(bc *ButlerConfig) (allPaths []string, dirtyFolders []string, err error) {
	allPaths = getFilePaths([]string{bc.Paths.WorkspaceRoot}, bc.Paths.AllowedPaths, bc.Paths.BlockedPaths, true)

	if bc.Git.GitRepo && !bc.Task.ShouldRunAll {
		allDirtyPaths, err := getDirtyPaths(bc.Git.PublishBranch)
		if err != nil {
			return nil, nil, err
		}
		dirtyFolders = getUniqueFolders(allDirtyPaths)

		if err = shouldRunAll(bc, allDirtyPaths, dirtyFolders); err != nil {
			return nil, nil, err
		}
	} else {
		bc.Task.ShouldRunAll = true
	}
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
	if branch = getEnvOrDefault(envBranch, ""); branch != "" {
		return
	}

	cmd := exec.Command(gitCommand, "branch", "--show-current")

	output, err := execOutputStub(cmd)
	branch = strings.TrimSpace(string(output))

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

	args := []string{gitCommand, "diff", "--name-only"}
	if branch != "" {
		args = append(args, branch)
	}

	cmd := exec.Command(args[0], args[1:]...)

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

func evaluateDirtiness(workspaces []*Workspace, dirtyFolders []string) {
	workspaceIsDirty := make(map[string]bool)
	mapDFs := convertToStringBoolMap(dirtyFolders)

	for _, ws := range workspaces {
		for _, path := range dirtyFolders {
			if !strings.Contains(path, strings.TrimPrefix(ws.Location, "./")) {
				continue
			}
			ws.IsDirty = true
			workspaceIsDirty[ws.Location] = true
			break
		}
	}

	madeAdditionalWorkspacesDirty := true
	for madeAdditionalWorkspacesDirty {
		madeAdditionalWorkspacesDirty = false
		for _, ws := range workspaces {
			for _, dep := range ws.WorkspaceDependencies {
				initialState := ws.IsDirty
				ws.IsDirty = ws.IsDirty || mapDFs[dep] || workspaceIsDirty[dep]
				madeAdditionalWorkspacesDirty = madeAdditionalWorkspacesDirty || (initialState != ws.IsDirty)
			}
		}
	}
}

// convertToStringBoolMap converts a []string to a map[string]bool.
func convertToStringBoolMap(values []string) map[string]bool {
	valueMap := make(map[string]bool)
	for _, value := range values {
		valueMap[value] = true
	}
	return valueMap
}
