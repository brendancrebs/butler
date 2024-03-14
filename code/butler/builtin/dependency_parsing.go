// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package builtin

import (
	"fmt"
	"os"
	"strings"
)

const (
	envWorkspaceRoot = "WORKSPACE_ROOT"
	envBranch        = "GIT_BRANCH"
)

func GetStdLibs(languageId string) (stdLibs []string, err error) {
	languageId = strings.ToLower(languageId)
	switch languageId {
	case golangId:
		stdLibs, err = goGetStdLibs()
	default:
		err = fmt.Errorf("language id '%v' not found", languageId)
	}

	return
}

func GetExternalDependencies(languageId string) (externalDeps []string, err error) {
	languageId = strings.ToLower(languageId)
	switch languageId {
	case golangId:
		externalDeps, _, err = goGetChangedModFileDeps(os.Getenv(envBranch))
	default:
		err = fmt.Errorf("language id '%v' not found", languageId)
	}

	return
}

func GetWorkspaceDeps(languageId, directory string) (deps []string) {
	switch languageId {
	case golangId:
		deps = goGetPkgDeps(directory)
	}
	return
}
