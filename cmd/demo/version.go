// SPDX-FileCopyrightText: 2026 Bonial International GmbH
// SPDX-License-Identifier: Apache-2.0

package demo

import (
	"fmt"
	"runtime"

	"github.com/bonial-oss/go-release-demo-release-please-minor/internal/buildinfo"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print build metadata for this binary.",
	Long: "Prints the version, commit, commit date, git tree state at build " +
		"time, and Go runtime info. Values other than Go runtime info come " +
		"from -ldflags -X injections done by GoReleaser at release time.",
	RunE: func(cmd *cobra.Command, args []string) error {
		w := cmd.OutOrStdout()
		if _, err := fmt.Fprintf(w, "Version:    %s\n", buildinfo.Version); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Commit:     %s\n", buildinfo.Commit); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Date:       %s\n", buildinfo.Date); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "TreeState:  %s\n", buildinfo.TreeState); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Go:         %s\n", runtime.Version()); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH); err != nil {
			return err
		}
		return nil
	},
}
