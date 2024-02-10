// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package builtin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var languageConfigPath = getRootDir()

type LanguageMethod struct {
	Name                     string
	WorkspaceMethod          string
	ExternalDependencyMethod string
	LintMethod               string
	TestMethod               string
	BuildMethod              string
	PublishMethod            string
	Aliases                  []string
}

// wrapper for built in default methods of acquiring third party repo dependencies.
func ExternalDependencyParsing(languageName string, workspaceRoot string) (changedDeps []string, err error) {
	methods, err := getMethods(languageName)
	if err != nil {
		err = fmt.Errorf("Error getting language methods for %s: %v\n", languageName, err)
	}
	fmt.Printf("\nMethods: %+v\n", methods)
	return
}

// obtains language default language methods for a given language
func getMethods(name string) (language LanguageMethod, err error) {
	data, err := os.ReadFile(languageConfigPath)
	if err != nil {
		return
	}

	// Parse the languages.json data
	var jsonData map[string]LanguageMethod
	if err = json.Unmarshal(data, &jsonData); err != nil {
		return
	}

	// Retrieve the object based on the provided key
	if language, ok := jsonData[name]; ok {
		return language, nil
	}
	return
}

func getRootDir() string {
	execDir, _ := os.Executable()
	absolutePath := filepath.Join(execDir, "../config/languages.json")
	return absolutePath
}
