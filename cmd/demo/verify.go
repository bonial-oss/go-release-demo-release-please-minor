// SPDX-FileCopyrightText: 2026 Bonial International GmbH
// SPDX-License-Identifier: Apache-2.0

package demo

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bonial-oss/go-release-demo-release-please-minor/internal/buildinfo"
	"github.com/spf13/cobra"
)

// ownerRepo is the canonical GitHub coordinate this binary was released under.
// Kept as a constant so builds from forks don't accidentally try to verify
// against upstream release assets.
const ownerRepo = "bonial-oss/go-release-demo-release-please-minor"

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify this binary's signature and checksum against its GitHub Release.",
	Long: `verify checks that the running binary matches the corresponding entry
in the checksums.txt of its release, and (if cosign is on PATH)
verifies the cosign signature over that checksums.txt against the
workflow identity that produced the release.

Requires network access to github.com and (optionally) rekor.sigstore.dev.
Requires the "cosign" CLI on PATH for Level 2 verification; without it,
falls back to SHA-256 checksum verification only.`,
	RunE: runVerify,
}

func init() {
	verifyCmd.Flags().BoolP("insecure-skip-signature", "", false,
		"skip cosign signature verification (checksum only) even if cosign is available")
	verifyCmd.Flags().Duration("timeout", 30*time.Second, "HTTP timeout for release-asset downloads")
}

func runVerify(cmd *cobra.Command, args []string) error {
	w := cmd.OutOrStdout()

	// 1. Locate own binary
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate self: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
	}

	// 2. Sanity-check embedded version
	version := buildinfo.Version
	if version == "" || version == "dev" {
		return fmt.Errorf("this binary lacks embedded release version info (built from source, not a signed release)")
	}

	// 3. Compute own SHA-256
	ownSum, err := sha256File(exe)
	if err != nil {
		return fmt.Errorf("hash self: %w", err)
	}

	// 4. Download checksums.txt (+.sig, +.pem)
	timeout, _ := cmd.Flags().GetDuration("timeout")
	skipSig, _ := cmd.Flags().GetBool("insecure-skip-signature")

	tmp, err := os.MkdirTemp("", "demo-verify-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	baseURL := fmt.Sprintf("https://github.com/%s/releases/download/%s", ownerRepo, version)
	client := &http.Client{Timeout: timeout}

	for _, name := range []string{"checksums.txt", "checksums.txt.sig", "checksums.txt.pem"} {
		if err := download(client, baseURL+"/"+name, filepath.Join(tmp, name)); err != nil {
			return fmt.Errorf("download %s: %w", name, err)
		}
	}

	// 5. Verify own SHA-256 appears in checksums.txt
	checksums, err := os.ReadFile(filepath.Join(tmp, "checksums.txt")) // #nosec G304 -- path is joined from a dir we created via os.MkdirTemp and a fixed filename, not user input
	if err != nil {
		return err
	}
	if !checksumMatches(string(checksums), ownSum) {
		return fmt.Errorf("SHA-256 %s not found in checksums.txt for %s (binary does not match this release)", ownSum, version)
	}

	// 6. Optionally verify cosign signature
	sigResult := "SKIPPED (--insecure-skip-signature set)"
	if !skipSig {
		cosignPath, err := exec.LookPath("cosign")
		if err != nil {
			sigResult = "SKIPPED (cosign CLI not found on PATH)"
		} else {
			if err := cosignVerify(cosignPath, tmp); err != nil {
				return fmt.Errorf("cosign signature verification failed: %w", err)
			}
			sigResult = "VALID (cosign keyless via Sigstore)"
		}
	}

	if _, err := fmt.Fprintf(w, "Verification passed:\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Binary:    %s\n", exe); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Version:   %s\n", version); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Commit:    %s\n", buildinfo.Commit); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Date:      %s\n", buildinfo.Date); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  SHA256:    %s\n", ownSum); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Signature: %s\n", sigResult); err != nil {
		return err
	}
	return nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path) // #nosec G304 -- callers pass either the caller's own resolved executable path or a path we constructed under a temp dir we created
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func checksumMatches(checksums, target string) bool {
	for _, line := range strings.Split(checksums, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[0] == target {
			return true
		}
	}
	return false
}

func download(client *http.Client, url, path string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}
	f, err := os.Create(path) // #nosec G304 -- path is joined from a dir this function's caller created via os.MkdirTemp and a fixed asset filename
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(f, resp.Body)
	return err
}

func cosignVerify(cosignPath, dir string) error {
	identityRegex := fmt.Sprintf("^https://github.com/%s/\\.github/workflows/release\\.yaml@refs/heads/main$", ownerRepo)
	c := exec.Command(cosignPath, // #nosec G204 -- cosignPath comes from exec.LookPath("cosign"); all other args are a fixed identity regex derived from the ownerRepo constant plus paths under a dir this binary created
		"verify-blob",
		"--certificate-identity-regexp", identityRegex,
		"--certificate-oidc-issuer", "https://token.actions.githubusercontent.com",
		"--signature", filepath.Join(dir, "checksums.txt.sig"),
		"--certificate", filepath.Join(dir, "checksums.txt.pem"),
		filepath.Join(dir, "checksums.txt"),
	)
	var stderr strings.Builder
	c.Stderr = &stderr
	c.Stdout = io.Discard
	if err := c.Run(); err != nil {
		return errors.New(strings.TrimSpace(stderr.String()))
	}
	return nil
}
