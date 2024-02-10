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
	"time"

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

// Task maintains state and output from various build tasks.
type Task struct {
	Name        string   `json:"name"`
	Language    string   `json:"language"`
	Path        string   `json:"path"`
	Logs        string   `json:"logs"`
	Error       string   `json:"error"`
	Arguments   []string `json:"arguments"`
	err         error
	CmdCreator  func() *exec.Cmd `json:"-"`
	Run         func() error     `json:"-"`
	Cmd         *exec.Cmd        `json:"-"`
	logsBuilder strings.Builder
	Attempts    int           `json:"attempts"`
	Retries     int           `json:"-"`
	Step        BuildStep     `json:"step"`
	Duration    time.Duration `json:"duration"`
}

// This function will return a queue populated with tasks
func GetTasks(bc *ButlerConfig, cmd *cobra.Command) (taskQueue *Queue, err error) {
	taskQueue = &Queue{tasks: make([]*Task, 0)}
	allPaths, dirtyFolders, err := butlerSetup(bc)
	if err != nil {
		return
	}

	fmt.Printf("\nallPaths: %v\n\n", allPaths)

	// Next steps:

	// 1. run preliminary commands
	if err = preliminaryCommands(bc.Languages); err != nil {
		return
	}

	// 2. create workspace array for each language
	if err = CreateWorkspaces(bc, allPaths); err != nil {
		return
	}

	// 3. get internal dependencies for each language

	// 4. get external dependencies for each language

	// 5. determine dirty workspaces
	for _, lang := range bc.Languages {
		evaluateDirtiness(lang.Workspaces, dirtyFolders)
	}

	// 6. create tasks for each language
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

	path, err := execLookPathStub(gitCommand)
	if err != nil {
		return
	}

	cmd := &exec.Cmd{
		Path: path,
		Args: []string{path, "branch", "--show-current"},
	}

	execCmd, err := execOutputStub(cmd)
	branch = strings.TrimSpace(string(execCmd))

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

func evaluateDirtiness(workspaces []*Workspace, dirtyFolders []string) {
	workspaceIsDirty := make(map[string]bool)
	mapDFs := convertToStringBoolMap(dirtyFolders)

	// hierarchical .. not sure this is the case since a go package does not
	// need to import something that is under it - e.g. don't rebuild all of
	// a library directory because of one change.
	for _, ws := range workspaces {
		for _, path := range dirtyFolders {
			// path contains the relative path info without a leading "./". The Workspace.Location
			// however can be prepended with "./", so strip it when found.
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

func (t *Task) String() string {
	const (
		maxPathLength = 60
	)
	path := t.Path
	path = trimFromLeftToLength(path, maxPathLength)

	path = fmt.Sprintf("%-*s", maxPathLength, path)

	return fmt.Sprintf("%-15s%-8s %s", t.Step, t.Language, path)
}

func trimFromLeftToLength(value string, maxLength int) string {
	if len(value) > maxLength {
		value = "..." + value[len(value)-(maxLength-3):]
	}
	return value
}
