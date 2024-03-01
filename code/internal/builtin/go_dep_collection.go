// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package builtin

import (
	"bufio"
	"bytes"
	"os/exec"
)

const (
	goExec   = "go"
	golangId = "golang"
)

// goGetStdLibs returns the list of go std libs for the current go executable.
func goGetStdLibs() ([]string, error) {
	goPath, _ := exec.LookPath(goExec)
	cmd := exec.Command(goPath, "list", "std")

	output, err := execOutputStub(cmd)
	s := bufio.NewScanner(bytes.NewBuffer(output))
	var results []string
	for s.Scan() {
		results = append(results, s.Text())
	}

	return results, err
}

// goGetPkgDeps returns the list of go package dependencies for a given package.
// It returns nothing if the folder is not or does not contains any go files.
func goGetPkgDeps(directory string) []string {
	goPath, _ := exec.LookPath(goExec)
	cmd := exec.Command(goPath, "list", "-test", "-f", `{{join .Deps "\n"}}`, directory)

	output, _ := execOutputStub(cmd)
	s := bufio.NewScanner(bytes.NewBuffer(output))

	var results []string
	for s.Scan() {
		results = append(results, s.Text())
	}

	return results
}
