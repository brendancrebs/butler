// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package builtin

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	goName   = "go"
	golangId = "golang"
	goMod    = "go.mod"
)

// goGetStdLibs returns the list of go std libs for the current go executable.
func goGetStdLibs() ([]string, error) {
	goPath, _ := exec.LookPath(goName)
	cmd := exec.Command(goPath, "list", "std")

	output, err := cmd.Output()
	s := bufio.NewScanner(bytes.NewBuffer(output))
	var results []string
	for s.Scan() {
		results = append(results, s.Text())
	}
	changeSet, _ := goGetChangedModFileDeps(os.Getenv(envBranch))
	versionChanged, err := didVersionChange(changeSet), err
	results = append([]string{strconv.FormatBool(versionChanged)}, results...)
	return results, err
}

// goGetPkgDeps returns the list of go package dependencies for a given package.
// It returns nothing if the folder is not or does not contains any go files.
func goGetPkgDeps(directory string) (results []string) {
	goPath, _ := exec.LookPath(goName)
	cmd := exec.Command(goPath, "list", "-test", "-f", `{{join .Deps "\n"}}`, directory)

	output, _ := cmd.Output()
	s := bufio.NewScanner(bytes.NewBuffer(output))

	for s.Scan() {
		results = append(results, s.Text())
	}

	return
}

// goGetModFileDiff returns the list of changed dependencies listed in the mod file.
func goGetChangedModFileDeps(branch string) (changeSet []string, err error) {
	err = filepath.WalkDir(os.Getenv(envWorkspaceRoot), func(path string, d fs.DirEntry, walkErr error) error {
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == goMod {
			var cs []string
			cs, walkErr = singleFileDiff(path, branch)
			changeSet = append(changeSet, cs...)
		}
		return walkErr
	})
	return
}

func singleFileDiff(filename, branch string) (changes []string, err error) {
	filename = strings.TrimSpace(filename)
	branch = strings.TrimSpace(branch)
	var path string

	path, err = exec.LookPath("git")
	if err != nil {
		return
	}
	cmd := &exec.Cmd{
		Path: path,
		Args: []string{path, "diff"},
	}
	if branch != "" {
		cmd.Args = append(cmd.Args, branch)
	}
	cmd.Args = append(cmd.Args, []string{"--", filename}...)

	var b []byte
	if b, err = cmd.Output(); err == nil {
		changes = pruneAdditiveChanges(convertLinesToStrings(b, []byte{'\n'}))
	}

	return
}

func pruneAdditiveChanges(lines []string) (changes []string) {
	changes = make([]string, len(lines))
	for _, line := range lines {
		if len(line) > 2 && line[0] == '+' && line[1] != '+' {
			line = strings.Split(strings.TrimSpace(line[1:]), " ")[0]
			changes = append(changes, line)
		}
	}
	return
}

func convertLinesToStrings(input, splitOn []byte) (lines []string) {
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

func didVersionChange(modDiff []string) (didChange bool) {
	for _, line := range modDiff {
		if strings.Split(line, " ")[0] == goName {
			didChange = true
			break
		}
	}
	return
}
