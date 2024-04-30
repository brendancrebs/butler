// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package main

import example "test_data/test_repo/go_tests/go_tests2"

// exclamation concatenates two strings and adds an exclamation.
func exclamation(a, b string) string {
	return example.Concat(a, b) + "!"
}
