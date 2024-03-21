// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

type GoMod struct {
	Module  Module
	Go      string
	Require []Dependency
}

type Module struct {
	Path string
}

type Dependency struct {
	Path     string
	Version  string
	Indirect bool
}

var goModPath = "./go.mod"

func main() {
	goModPath, _ = filepath.Abs(goModPath)
	dependencies, err := getDependencies(goModPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var dependencyStrings []string
	for _, dep := range dependencies {
		dependencyStrings = append(dependencyStrings, fmt.Sprintf("%s@%s\n", dep.Path, dep.Version))
	}
	jsonDeps, _ := json.Marshal(dependencyStrings)
	fmt.Println(string(jsonDeps))
}

func getDependencies(modFilePath string) ([]Dependency, error) {
	cmd := exec.Command("go", "mod", "edit", "-json", modFilePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var goMod GoMod
	err = json.Unmarshal(output, &goMod)
	if err != nil {
		return nil, err
	}

	return goMod.Require, nil
}
