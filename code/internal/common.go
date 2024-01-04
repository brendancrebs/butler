// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"os"
)

func getEnvOrDefault(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}

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
