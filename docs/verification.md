<!--
SPDX-FileCopyrightText: 2026 Bonial International GmbH
SPDX-License-Identifier: Apache-2.0
-->

# Verifying a release

Every release of `go-release-demo-release-please-minor` is signed by
[Sigstore cosign](https://docs.sigstore.dev/cosign/) (keyless via OIDC)
and carries a
[SLSA L3 provenance attestation](https://slsa.dev/spec/v1.0/provenance).
This document shows how to verify a downloaded archive at four levels
of rigor.

Assumes you've downloaded these assets from a release page (e.g.
<https://github.com/bonial-oss/go-release-demo-release-please-minor/releases/tag/v0.1.0>):

```
demo_0.1.0_linux_amd64.tar.gz    # the archive
checksums.txt                    # SHA-256 of every archive
checksums.txt.sig                # cosign signature over checksums.txt
checksums.txt.pem                # signing certificate (short-lived, Fulcio-issued)
multiple.intoto.jsonl            # SLSA provenance attestation
```

All commands below use `v0.1.0` as the example version — substitute your
actual version.

## Level 1 — Checksum only (any Unix)

Confirms the archive wasn't corrupted in transit. Does NOT prove
authenticity.

```bash
sha256sum -c checksums.txt --ignore-missing
```

Expected output:

```
demo_0.1.0_linux_amd64.tar.gz: OK
```

## Level 2 — Cosign signature (requires `cosign` v2)

Confirms the archive was produced by the release workflow of THIS repo,
not by an attacker.

Install cosign: <https://docs.sigstore.dev/cosign/system_config/installation/>

```bash
cosign verify-blob \
  --certificate-identity-regexp='^https://github.com/bonial-oss/go-release-demo-release-please-minor/\.github/workflows/release\.yaml@refs/heads/main$' \
  --certificate-oidc-issuer='https://token.actions.githubusercontent.com' \
  --signature checksums.txt.sig \
  --certificate checksums.txt.pem \
  checksums.txt
```

Expected output ends with:

```
Verified OK
```

If verification succeeds, `checksums.txt` is authentic. Combined with
Level 1, that transitively verifies every archive listed in it. **The
tag-to-release binding lives in this signature** — the certificate's
subject identifies which repo/workflow/branch signed it.

## Level 3 — SLSA provenance (requires `slsa-verifier` v2.x)

Confirms the archive was built from a specific branch by a specific
workflow — proves build provenance, not just signing.

Install `slsa-verifier`: <https://github.com/slsa-framework/slsa-verifier#installation>

```bash
slsa-verifier verify-artifact \
  --provenance-path multiple.intoto.jsonl \
  --source-uri github.com/bonial-oss/go-release-demo-release-please-minor \
  --source-branch main \
  demo_0.1.0_linux_amd64.tar.gz
```

Note the flag: `--source-branch main` (not `--source-tag`). The release
workflow runs on `push: main`, so the SLSA provenance encodes
`refs/heads/main` as the source ref, not the release tag. Using
`--source-tag` would fail with `verifying tag: invalid ref: ""`. The
tag-to-release binding is provided by the cosign signature (Level 2).

Expected output ends with:

```
PASSED: SLSA verification passed
```

## Level 4 — CLI self-check

The binary itself can verify its own signature. Runs Levels 1 + 2
internally (requires `cosign` on `PATH`; if absent, skips signature
check and falls back to Level 1).

```bash
tar xzf demo_0.1.0_linux_amd64.tar.gz
./demo verify
```

Expected:

```
Verification passed:
  Binary:    /path/to/demo
  Version:   v0.1.0
  Commit:    abc123def456…
  SHA256:    <64 hex chars>
  Signature: VALID (cosign keyless via Sigstore)
```

## Reproducibility

Every release is built with `-trimpath` and pinned Go toolchain, so the
same commit produces byte-identical binaries on any x86_64 Linux runner.
A CI job (`verify-release.yaml`) independently rebuilds each tagged
commit and diffs against the released binaries — a green check on the
release page means "byte-identical rebuild succeeded across all four
platforms."

To reproduce locally:

```bash
git clone https://github.com/bonial-oss/go-release-demo-release-please-minor
cd go-release-demo-release-please-minor
git checkout v0.1.0
GORELEASER_CURRENT_TAG=v0.1.0 goreleaser build \
  --clean --single-target --id demo \
  -o rebuilt_demo
sha256sum rebuilt_demo
```

Compare the output to the corresponding entry in `checksums.txt`.
