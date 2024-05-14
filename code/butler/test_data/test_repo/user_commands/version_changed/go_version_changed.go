// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

const (
	gitName        = "git"
	goName         = "go"
	golangId       = "golang"
	goMod          = "go.mod"
	versionChanged = true
)

func main() {
	cmd := exec.Command(goName, "list", "std")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	s := bufio.NewScanner(bytes.NewBuffer(output))

	var results []string
	for s.Scan() {
		results = append(results, s.Text())
	}

	results = append([]string{strconv.FormatBool(versionChanged)}, results...)
	jsonDeps, _ := json.Marshal(results)
	fmt.Println(string(jsonDeps))
	return
}
