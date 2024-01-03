// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package helpers

import (
	"os"
	"strings"
)

func GetEnvOrDefault(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}

func SplitCommand(cmd string) []string {
	commandParts := []string{}
	splitCmd := strings.Fields(cmd)
	commandParts = append(commandParts, splitCmd...)
	return commandParts
}

// ifErrNil allows for reversing normal logic to make test coverage easier.
func IfErrNil(err error, f func() error) error {
	if err != nil {
		return err
	}
	return f()
}

// merge two maps
func MergeMaps(ignoreMap, configMap map[string]bool) map[string]bool {
	result := make(map[string]bool)

	for key, value := range ignoreMap {
		result[key] = value
	}

	for key, value := range configMap {
		result[key] = value
	}

	return result
}
