// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package helpers

import "os"

func GetEnvOrDefault(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
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
