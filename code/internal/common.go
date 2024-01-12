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

// concatenates and removes duplicates from two slices
func concatSlices(ignore, config []string) []string {
	slice := cleanPaths(append(ignore, config...))
	seen := make(map[string]bool)
	result := []string{}

	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
