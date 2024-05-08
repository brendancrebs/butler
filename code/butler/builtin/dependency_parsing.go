// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	envWorkspaceRoot = "WORKSPACE_ROOT"
	envBranch        = "GIT_BRANCH"
)

var execOutputStub = (*exec.Cmd).Output

func GetStdLibs(languageId string) (stdLibs []string, err error) {
	languageId = strings.ToLower(languageId)
	switch languageId {
	case golangId:
		stdLibs, err = getStdLibs()
	default:
		err = fmt.Errorf("language id '%v' not found", languageId)
	}

	return
}

func GetExternalDependencies(languageId string) (externalDeps []string, err error) {
	languageId = strings.ToLower(languageId)
	switch languageId {
	case golangId:
		externalDeps, err = getChangedModFileDeps(os.Getenv(envBranch))
	default:
		err = fmt.Errorf("language id '%v' not found", languageId)
	}

	return
}

func GetWorkspaceDeps(languageId, directory string) (deps []string) {
	switch languageId {
	case golangId:
		deps, _ = getWorkspaceDeps(directory)
	}

	return
}
