<!--
SPDX-FileCopyrightText: 2026 Bonial International GmbH
SPDX-License-Identifier: Apache-2.0
-->

# go-release-demo-release-please-minor

**Combo 1** reference implementation of immutable Go releases, using
**release-please + GoReleaser**. Part of the
[go-release-demo](https://github.com/bonial-oss/go-release-demo)
evaluation project comparing three release toolchains.

## What this demonstrates

Cutting a versioned Go release where:

- The git tag cannot be moved, deleted, or re-created (tag ruleset).
- The release assets cannot be edited after publication (immutable-releases setting).
- The artifacts are signed by a verifiable workflow identity via **cosign keyless** (Sigstore, OIDC).
- The build has **SLSA L3 provenance attestations** proving how it was built.
- The build is **reproducible**: the same commit produces byte-identical binaries, testable via a rebuild-and-diff CI job.
- The release is **atomic**: the tag only exists if the full build, sign, and publish succeeded — no partial state ever reaches a consumer.
- Version proposals surface in a **Release PR** — release-please keeps a running proposal open on `main` that shows exactly what will be released and lets humans review it before merging.

## Toolchain

| Concern | Tool |
|---|---|
| Next version + Release PR + changelog | [`release-please`](https://github.com/googleapis/release-please) via [`googleapis/release-please-action@v4`](https://github.com/googleapis/release-please-action) |
| Build, sign, SBOM, release | [`goreleaser`](https://goreleaser.com) |
| Keyless signing | [`cosign`](https://docs.sigstore.dev/cosign/) |
| Build provenance | [`slsa-github-generator`](https://github.com/slsa-framework/slsa-github-generator) |
| SBOM | [`syft`](https://github.com/anchore/syft) (invoked by GoReleaser) |

## Using the released binary

Download for your platform from the [latest release](https://github.com/bonial-oss/go-release-demo-release-please-minor/releases/latest):

```bash
# Example: linux/amd64
curl -fsSL -o demo.tar.gz \
  https://github.com/bonial-oss/go-release-demo-release-please-minor/releases/download/v0.1.0/demo_0.1.0_linux_amd64.tar.gz
tar xzf demo.tar.gz
./demo version
./demo verify
```

## Verifying the release

See [`docs/verification.md`](docs/verification.md) for cut-and-paste commands
that check the signature, provenance, and reproducibility of any downloaded
archive.

The bundled `demo verify` subcommand does the same check in-binary — the
binary you downloaded proves to you that it's the binary the release
workflow produced.

## How releases are cut (Release-PR flow)

1. Merge normal PRs to `main` containing `feat:` / `fix:` / etc. commits.
2. On every push to `main`, the `Release` workflow runs release-please.
   release-please reads the commit history since the last tag and either:
   - **opens or updates** an open Release PR titled `chore(main): release X.Y.Z`
     showing the proposed next version and generated `docs/CHANGELOG.md` delta; or
   - **detects a just-merged Release PR** and emits `release_created=true`,
     at which point downstream jobs (`goreleaser`, `slsa`, `promote`)
     build and publish the release.
3. Three sequential jobs after `release_created=true`: **goreleaser** (build,
   sign, SBOM, create draft release with all assets) → **slsa** (attach SLSA
   L3 provenance to the draft) → **promote** (`gh release edit --draft=false`
   — the only write to a published release).
4. If any step fails, no tag reaches consumers (the release stays a draft or
   doesn't exist at all). Delete the draft (if any) and re-dispatch the
   workflow.

A manual `workflow_dispatch` path is supported for dry-run previews and
forced re-runs — see [`.github/workflows/release.yaml`](.github/workflows/release.yaml).

## Verifying a published release (manual dispatch)

The [`verify-release.yaml`](.github/workflows/verify-release.yaml) workflow
runs rebuild-and-diff (reproducibility) and cosign + SLSA signature
verification against every released asset. It is defined to trigger on
`release: types: [released]` but **does not auto-trigger today** — the
`promote` job publishes the release under `secrets.GITHUB_TOKEN`, and
GitHub's [chain-prevention rule](https://docs.github.com/en/actions/using-workflows/triggering-a-workflow#triggering-a-workflow-from-a-workflow)
suppresses the `released` event in that case.

After each release, dispatch verify-release manually:

```bash
gh workflow run verify-release.yaml \
  --repo bonial-oss/go-release-demo-release-please-minor \
  --ref main \
  --field tag=v0.1.0   # substitute the actual release tag
```

**Planned fix:** switch `promote` to a dedicated `bonial-release` GitHub
App token (see `.rulesets/README.md`). The App identity is distinct from
`github-actions[bot]`, so the `released` event fires normally under it.

## Development

```bash
make install-tools   # install reuse into .venv-reuse
make lint            # reuse-lint + md-lint + go-lint
make test            # go test ./... -race -cover
make build           # go build ./...
```

Individual commits must be [Conventional Commits](https://www.conventionalcommits.org/); this is enforced by CI (`.github/workflows/commitlint.yaml`). The `main` branch protects against direct pushes; merge via PR.

## License

Apache 2.0. See [LICENSE](LICENSE) and [REUSE.toml](REUSE.toml).
