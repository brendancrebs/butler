// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"bytes"
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

// Returns a queue populated with tasks
func getTasks(bc *ButlerConfig, cmd *cobra.Command) (taskQueue *Queue, err error) {
	taskQueue = &Queue{tasks: make([]*Task, 0)}
	allPaths, dirtyFolders, err := butlerSetup(bc)
	if err != nil {
		return
	}

	for _, lang := range bc.Languages {
		if err = lang.runCommandList(cmd, lang.TaskCmd.SetUp); err != nil {
			return
		}
		if !bc.Task.RunAll {
			if err = lang.getDependencies(bc); err != nil {
				return
			}
			dirtyFolders = append(dirtyFolders, lang.ExternalDeps...)
		}

		lang.getWorkspaces(bc, allPaths)
	}

	if !bc.Task.RunAll {
		for _, lang := range bc.Languages {
			EvaluateDirtiness(lang.Workspaces, dirtyFolders)
		}
	}

	populateTaskQueue(bc, taskQueue, cmd)

	return
}

// First gathers all non-blocked file paths in a repo. If a repo uses git, files with a git diff
// will be collected for determine dirty workspaces later. This will also be used to determine if a
// critical file has been changed.
func butlerSetup(bc *ButlerConfig) (allPaths []string, dirtyFolders []string, err error) {
	allPaths = getFilePaths([]string{bc.Paths.WorkspaceRoot}, bc.Paths.AllowedPaths, bc.Paths.IgnorePaths, true)
	if bc.PublishBranch != "" && !bc.Task.RunAll {
		allDirtyPaths, err := getDirtyPaths(bc.PublishBranch)
		if err != nil {
			return nil, nil, err
		}
		dirtyFolders = getUniqueFolders(allDirtyPaths)

		if err = shouldRunAll(bc, allDirtyPaths); err != nil {
			return nil, nil, err
		}
	} else {
		bc.Task.RunAll = true
	}
	return
}

// Checks several conditions to see wether Butler requires a full build. A list of conditions can be
// found in the butler_config.md file in the specs folder. In the spec it will be in the general
// options under the "runAll" tag.
func shouldRunAll(bc *ButlerConfig, allDirtyPaths []string) (err error) {
	// get current git branch name
	currentBranch, err := getCurrentBranch()
	if err != nil {
		return
	}

	bc.Task.RunAll = strings.EqualFold(GetEnvOrDefault(envRunAll, ""), "true") || currentBranch == bc.PublishBranch
	bc.Task.Publish = strings.EqualFold(GetEnvOrDefault(envPublish, ""), "true") || currentBranch == bc.PublishBranch
	bc.Task.RunAll = bc.Task.RunAll || CriticalPathChanged(allDirtyPaths, bc.Paths.CriticalPaths)

	return
}

// Returns the whitespace trimmed result of `git branch --show-current` which should be the branch,
// or an error.
func getCurrentBranch() (branch string, err error) {
	if branch = GetEnvOrDefault(envBranch, ""); branch != "" {
		return
	}

	cmd := exec.Command(gitCommand, "branch", "--show-current")

	output, err := execOutputStub(cmd)
	branch = strings.TrimSpace(string(output))

	return
}

// Takes a set of directories and calls recurseFilePath on each. recurseFilePaths() will return the
// children of each directory and all of the child files will get returned by getFilePaths.
func getFilePaths(dirs, allowed, ignored []string, shouldRecurse bool) []string {
	paths := make([]string, len(dirs))

	for _, base := range dirs {
		paths = recurseFilepaths(paths, allowed, ignored, base, shouldRecurse)
	}

	return paths
}

// Reads the directory at `path` and either appends filepaths or appends the result of a further
// call to recurseFilepaths().
func recurseFilepaths(paths, allowed, ignored []string, path string, shouldRecurse bool) []string {
	fileInfos, _ := os.ReadDir(path)
	for _, fi := range fileInfos {
		path := filepath.Join(path, fi.Name())
		if !isAllowed(path, allowed, ignored) {
			continue
		}
		if !fi.IsDir() {
			paths = append(paths, path)
		} else if shouldRecurse {
			paths = recurseFilepaths(paths, allowed, ignored, path, shouldRecurse)
		}
	}
	return paths
}

// Checks a path to see if it is on a paths marked as 'allowed' in the config and checks that it
// also hasn't been blocked.
func isAllowed(path string, allowed, ignored []string) bool {
	isAllowed := false
	if len(allowed) == 0 {
		isAllowed = true
	} else {
		for _, key := range allowed {
			if strings.Contains(path, key) {
				isAllowed = true
				break
			}
		}
	}
	if !isAllowed {
		return false
	}

	isIgnored := false
	for _, key := range ignored {
		if strings.Contains(path, key) {
			isIgnored = true
			break
		}
	}

	return !isIgnored
}

// Calls 'git diff --name-only {branch}' if branch is not blank, or 'git diff --name-only' without
// a branch.  It returns a list of file names that changed.
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

// Splits on the splitOn slice, converts each grouping to a string and trims space from it. Returns
// the set of converted lines.
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

// Checks if any of the critical paths listed in the config are dirty paths or the parent to dirty
// paths in the case of directories.
func CriticalPathChanged(dirtyPaths, criticalPaths []string) (result bool) {
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

// Returns a sorted set of unique folders based on a set of paths.
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

// Takes a language's workspaces and the set of dirty folders and determines which workspaces
// require tasks.
func EvaluateDirtiness(workspaces []*Workspace, dirtyFolders []string) {
	dirtyWorkspaces := make(map[string]bool)
	mapDFs := convertToStringBoolMap(dirtyFolders)

	for _, ws := range workspaces {
		for _, path := range dirtyFolders {
			if !strings.Contains(path, strings.TrimPrefix(ws.Location, "./")) {
				continue
			}
			ws.IsDirty = true
			dirtyWorkspaces[ws.Location] = true
			break
		}
	}

	madeWsDirty := true
	for madeWsDirty {
		madeWsDirty = false
		for _, ws := range workspaces {
			for _, dep := range ws.Dependencies {
				initialState := ws.IsDirty
				ws.IsDirty = ws.IsDirty || mapDFs[dep] || dirtyWorkspaces[dep]
				madeWsDirty = madeWsDirty || (initialState != ws.IsDirty)
			}
		}
	}
}

// Converts a []string to a map[string]bool.
func convertToStringBoolMap(values []string) map[string]bool {
	valueMap := make(map[string]bool)
	for _, value := range values {
		valueMap[value] = true
	}
	return valueMap
}
