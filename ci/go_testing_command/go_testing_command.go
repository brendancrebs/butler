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

// This method takes the original go testing command and checks it for coverage. Failing if less
// than 100%
func main() {
	cmd := exec.Command("/usr/local/go/bin/go", "test", "-tags=testonly", "-count=1", "-cover", "-v")

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
	re := regexp.MustCompile(`coverage: (\d+(\.\d+)?)% of statements`)
	matches := re.FindStringSubmatch(output)

	if len(matches) > 1 {
		coverage := matches[1]
		if !strings.HasPrefix(coverage, "100") {
			return false
		}
	}

	return true
}
