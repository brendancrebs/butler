// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func getCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "butler <flags>",
		Short: "butler is a build test lint runner",
		Long: `BuTLeR is a multi-threaded build, test, lint, and publish tool.  When not on ` +
			`the {publish-branch} it only runs against files diffed against the {publish-branch}.`,
		RunE: run,
	}

	return root
}

var cmd = getCommand()

// Execute is the entrypoint into the Butler
func Execute() {
	// Set Min/Max thread count for VITest Environment Variables to prevent timeouts from resource hogging.
	os.Setenv("VITEST_MIN_THREADS", "1")
	os.Setenv("VITEST_MAX_THREADS", "1")

	// all errors are handled internal to this call.
	_ = cmd.Execute()
}

func run(cmd *cobra.Command, args []string) (err error) {
	fmt.Printf("\nHello world\n")
	return
}
