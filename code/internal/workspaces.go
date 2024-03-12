// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"path/filepath"
	"strings"
	"sync"

	"selinc.com/butler/code/internal/builtin"
)

type Workspace struct {
	Location     string
	Name         string
	IsDirty      bool
	Dependencies []string
}

// Collects workspaces for a language
func getWorkspaces(lang *Language, paths []string) (workspaces []*Workspace) {
	var allDirs map[string]bool
	if lang.WorkspaceFile != "" {
		allDirs = getMatchingDirs(paths, lang.WorkspaceFile)
	} else {
		allDirs = getMatchingDirs(paths, lang.FileExtension)
	}

	workspaces = concurrentGetWorkspaces(lang.Name, lang.StdLibDeps, allDirs)
	return
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

// Returns the full set of workspaces.  Brute force multithreading, spins out the requests and lets
// the go runtime handle the workload.
func concurrentGetWorkspaces(languageId string, stdLibs []string, allDirs map[string]bool) (ws []*Workspace) {
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	for dir := range allDirs {
		wg.Add(1)
		// must proxy dir into a different variable to make it safe to use inside the closure.
		go func(thisDir string) {
			var (
				deps       = builtin.GetWorkspaceDeps(languageId, thisDir)
				prunedDeps = difference(deps, stdLibs)
				workspace  = &Workspace{Location: thisDir, Dependencies: prunedDeps}
			)
			mu.Lock()
			ws = append(ws, workspace)
			wg.Done()
			mu.Unlock()
		}(dir)
	}
	wg.Wait()
	return
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
