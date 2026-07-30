// SPDX-FileCopyrightText: 2026 Bonial International GmbH
// SPDX-License-Identifier: Apache-2.0

// Package buildinfo holds build-time metadata injected via -ldflags -X.
// Values are "dev" (or empty) in local development builds and set to real
// values by GoReleaser at release time.
package buildinfo

// Version is the release version, e.g. "v0.1.0" or "dev".
var Version = "dev"

// Commit is the full git commit SHA the binary was built from, or "unknown".
var Commit = "unknown"

// Date is the commit date in RFC3339 format, or "unknown".
var Date = "unknown"

// TreeState is "clean" if built from a clean git tree, "dirty" if the tree
// had uncommitted changes at build time, or "unknown".
var TreeState = "unknown"
