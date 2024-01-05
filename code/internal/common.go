// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"os"
)

// returns the value of the given environment variable if it has been set. Otherwise return the
// default.
func getEnvOrDefault(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}

// merges two maps into one.
func mergeMaps(ignoreMap, configMap map[string]bool) map[string]bool {
	result := make(map[string]bool)

	for key, value := range ignoreMap {
		result[key] = value
	}

	for key, value := range configMap {
		result[key] = value
	}

	return result
}
