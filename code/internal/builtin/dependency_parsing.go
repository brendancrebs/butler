// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package builtin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const languageConfigPath = "./builtin/config/languages.json"

var execOutputStub = (*exec.Cmd).Output

type Language struct {
	Id      string
	Aliases []string
}

func GetLanguageId(languageName string) (languageId string, err error) {
	languageId, err = getMethods(languageName)
	if err != nil {
		err = fmt.Errorf("error getting language id for %s: %v\n", languageName, err)
	}
	return
}

// obtains language id from config
func getMethods(name string) (languageId string, err error) {
	languageId = strings.ToLower(name)
	output, err := os.ReadFile(languageConfigPath)
	if err != nil {
		return
	}

	// Parse the languages.json data
	var jsonLanguages []Language
	if err = json.Unmarshal(output, &jsonLanguages); err != nil {
		return
	}

	// Retrieve the language based on the provided key
	for _, lang := range jsonLanguages {
		if languageId == lang.Id {
			return
		}
		for _, alias := range lang.Aliases {
			if languageId == alias {
				return lang.Id, nil
			}
		}
	}
	return "", errors.New("Language not found")
}

func GetStdLibs(languageId string) (stdLibs []string, err error) {
	switch languageId {
	case golangId:
		stdLibs, err = goGetStdLibs()
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
