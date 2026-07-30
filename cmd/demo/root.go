// SPDX-FileCopyrightText: 2026 Bonial International GmbH
// SPDX-License-Identifier: Apache-2.0

// Package demo defines the cobra command tree for the `demo` CLI.
package demo

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "demo",
	Short:         "A trivial CLI demonstrating immutable Go releases (Combo 1).",
	Long:          "demo is a tiny reference binary that ships as a signed, attested, reproducible release cut via release-please + GoReleaser. Its `version` subcommand prints build metadata; its `verify` subcommand checks its own cosign signature.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command. Called from main().
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(verifyCmd)
}
