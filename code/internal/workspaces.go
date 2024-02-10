// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"bufio"
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Workspace struct {
	Location              string
	Name                  string
	IsDirty               bool
	WorkspaceDependencies []string
}

const (
	goPath = "/usr/local/go/bin/go"
)

// Collects workspaces for a language
func getWorkspaces(baseConfig *ButlerConfig, lang *Language, paths []string) (workspaces []*Workspace, err error) {
	var allDirs map[string]bool
	if lang.WorkspaceFile != "" {
		allDirs = getMatchingDirs(paths, lang.WorkspaceFile)
	} else {
		allDirs = getMatchingDirs(paths, lang.FileExtension)
	}
	stdLibs, err := goGetStdLibs(goPath)
	if err != nil {
		return
	}
	workspaces = concurrentGetWorkspaces(baseConfig, allDirs, goPath, stdLibs)
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

// goGetStdLibs returns the list of go std libs for the current go executable.
func goGetStdLibs(goPath string) ([]string, error) {
	cmd := &exec.Cmd{
		Path: goPath,
		Args: []string{goPath, "list", "std"},
	}

	b, err := cmd.Output()
	s := bufio.NewScanner(bytes.NewBuffer(b))
	var results []string
	for s.Scan() {
		results = append(results, s.Text())
	}

	return results, err
}

// Returns the full set of workspaces.  Brute force multithreading, spins out the requests and lets
// the go runtime handle the workload.
func concurrentGetWorkspaces(baseConfig *ButlerConfig, allDirs map[string]bool, goPath string, stdLibs []string) (ws []*Workspace) {
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	for dir := range allDirs {
		wg.Add(1)
		// must proxy dir into a different variable to make it safe to use inside the closure.
		go func(thisDir string) {
			var (
				deps       = goGetPkgDeps(goPath, thisDir)
				prunedDeps = difference(deps, stdLibs)
				workspace  = &Workspace{Location: thisDir, WorkspaceDependencies: prunedDeps}
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

// goGetPkgDeps returns the list of go package dependencies for a given package.
// It returns nothing if the folder is not or does not contains any go files.
func goGetPkgDeps(goPath, directory string) []string {
	cmd := &exec.Cmd{
		Path: goPath,
		// use go templating to make each entry its own line.
		// See: https://dave.cheney.net/2014/09/14/go-list-your-swiss-army-knife
		Args: []string{goPath, "list", "-test", "-f", `{{join .Deps "\n"}}`, directory},
	}

	b, _ := cmd.Output()
	s := bufio.NewScanner(bytes.NewBuffer(b))

	var results []string
	for s.Scan() {
		results = append(results, s.Text())
	}

	return results
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
