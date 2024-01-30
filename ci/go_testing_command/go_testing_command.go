// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	testCoverage = "100"
	goExec       = "go"
)

// This method takes the original go testing command and checks it for coverage. Failing if less
// than 100%
func main() {
	goPath, _ := exec.LookPath(goExec)
	cmd := exec.Command(goPath, "test", "-tags=testonly", "-count=1", "-cover", "-v")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	logs := out.String()

	if err != nil {
		fmt.Println(logs)
		os.Exit(1)
	}

	// Check for coverage
	if !checkCoverage(logs) {
		fmt.Println(logs)
		os.Exit(1)
	}
}

// parses the test output and returns false if coverage is less than 100%
func checkCoverage(output string) bool {
	// parsing coverage statement from 'go test -cover' output.
	re := regexp.MustCompile(`coverage: (\d+(\.\d+)?)% of statements`)
	matches := re.FindStringSubmatch(output)

	if len(matches) > 1 {
		// We check index 1 of matches since that will contain the first parsed out instance of the
		// sub-expression defined above in 'coverage: ... of statements'. index 0 will contain the
		// entire expression.
		coverage := matches[1]
		if !strings.HasPrefix(coverage, testCoverage) {
			return false
		}
	}

	return true
}
