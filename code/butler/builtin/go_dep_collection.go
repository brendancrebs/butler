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
	gitName  = "git"
	goName   = "go"
	golangId = "golang"
	goMod    = "go.mod"
)

// Returns the list of standard libraries for the current go executable. The libraries will be
// returned as a list of strings. The first string will be "true" or "false" to indicate whether the
// go version has been changed compared to the main branch.
func getStdLibs() (results []string, err error) {
	cmd := exec.Command(goName, "list", "std")
	output, err := execOutputStub(cmd)
	if err != nil {
		return
	}

	s := bufio.NewScanner(bytes.NewBuffer(output))

	for s.Scan() {
		results = append(results, s.Text())
	}
	changeSet, err := getChangedModFileDeps(os.Getenv(envBranch))
	if err != nil {
		return
	}
	versionChanged := didVersionChange(changeSet)
	results = append([]string{strconv.FormatBool(versionChanged)}, results...)
	return
}

// Returns the list of go package dependencies for a given package. It returns nothing if the folder
// is not or does not contains any go files.
func getWorkspaceDeps(directory string) (results []string, err error) {
	cmd := exec.Command(goName, "list", "-test", "-f", `{{join .Deps "\n"}}`, directory)
	output, err := execOutputStub(cmd)
	if err != nil {
		return
	}
	s := bufio.NewScanner(bytes.NewBuffer(output))

	for s.Scan() {
		results = append(results, s.Text())
	}

	return
}

// Returns the list of changed dependencies listed in the mod file.
func getChangedModFileDeps(branch string) (changeSet []string, err error) {
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

	cmd := exec.Command(gitName, "diff")
	if branch != "" {
		cmd.Args = append(cmd.Args, branch)
	}
	cmd.Args = append(cmd.Args, []string{"--", filename}...)

	var b []byte
	if b, err = execOutputStub(cmd); err == nil {
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
