// SPDX-FileCopyrightText: 2026 Bonial International GmbH
// SPDX-License-Identifier: Apache-2.0

package demo

import "testing"

func TestChecksumMatches(t *testing.T) {
	sample := `
aaaaaaaa  demo_0.1.0_linux_amd64.tar.gz
bbbbbbbb  demo_0.1.0_linux_arm64.tar.gz
cccccccc  demo_0.1.0_darwin_amd64.tar.gz
`
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{"present linux/amd64", "aaaaaaaa", true},
		{"present linux/arm64", "bbbbbbbb", true},
		{"absent", "deadbeef", false},
		{"empty target", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checksumMatches(sample, tc.target)
			if got != tc.want {
				t.Errorf("checksumMatches(%q) = %v, want %v", tc.target, got, tc.want)
			}
		})
	}
}
