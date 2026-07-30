// SPDX-FileCopyrightText: 2026 Bonial International GmbH
// SPDX-License-Identifier: Apache-2.0

// Command demo is a trivial CLI used to demonstrate immutable Go releases
// via the Combo 1 toolchain (release-please + GoReleaser). It has two
// subcommands: `version` prints build metadata, `verify` checks the
// running binary's cosign signature against the corresponding GitHub
// Release. It is intentionally small — the release plumbing is the point.
package main

import (
	"fmt"
	"os"

	"github.com/bonial-oss/go-release-demo-release-please-minor/cmd/demo"
)

func main() {
	if err := demo.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
