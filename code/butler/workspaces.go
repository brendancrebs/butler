// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"path/filepath"
	"strings"
	"sync"

	"selinc.com/butler/code/butler/builtin"
)

// Represents the location where tasks will be executed for a language. It also tracks dependencies
// being used within the workspace.
type Workspace struct {
	Location     string
	IsDirty      bool
	Dependencies []string
}

// Collects workspaces for a language
func (lang *Language) getWorkspaces(paths []string, publishBranch string) {
	allDirs := make(map[string]bool)
	for _, pattern := range lang.FilePatterns {
		for k, v := range getMatchingDirs(paths, pattern) {
			allDirs[k] = v
		}
	}

	lang.concurrentGetWorkspaces(allDirs, publishBranch)
}

// getMatchingDirs returns the map of directories that contain `pattern`.
func getMatchingDirs(dirs []string, pattern string) (matches map[string]bool) {
	matches = make(map[string]bool)
	for _, dir := range dirs {
		if !strings.Contains(dir, pattern) {
			continue
		}
		matches[filepath.Dir(dir)] = true
	}
	return
}

// Returns the full set of workspaces. Brute force multithreading, spins out the requests and lets
// the go runtime handle the workload.
func (lang *Language) concurrentGetWorkspaces(allDirs map[string]bool, publishBranch string) {
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	for dir := range allDirs {
		wg.Add(1)
		// must proxy dir into a different variable to make it safe to use inside the closure.
		go func(thisDir string) {
			workspace := &Workspace{Location: thisDir}
			if lang.DepCommands.Workspace != "" && publishBranch != "" {
				deps := builtin.GetWorkspaceDeps(lang.Name, thisDir)
				prunedDeps := difference(deps, lang.StdLibDeps)
				workspace.Dependencies = prunedDeps
			}

			mu.Lock()
			lang.Workspaces = append(lang.Workspaces, workspace)
			wg.Done()
			mu.Unlock()
		}(dir)
	}
	wg.Wait()
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
