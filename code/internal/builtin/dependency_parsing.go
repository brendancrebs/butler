// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package builtin

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

const languageConfigPath = "/workspaces/butler/code/internal/builtin/config/languages.json"

type Language struct {
	Id      string
	Aliases []string
}

func GetLanguageId(languageName string) (languageId string, err error) {
	languageId, err = getMethods(languageName)
	if err != nil {
		err = fmt.Errorf("Error getting language id for %s: %v\n", languageName, err)
	}

	fmt.Printf("\nID: %v\n", languageId)
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
	return "", fmt.Errorf("Language not found")
}

func GetStdLibs(languageId string) (stdLibs map[string]string, err error) {
	var m map[string]string
	method := reflect.ValueOf(m).MethodByName("goStdLibs")
	stdLibs = method.Call(nil)
	return
}
