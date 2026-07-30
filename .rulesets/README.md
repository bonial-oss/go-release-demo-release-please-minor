<!--
SPDX-FileCopyrightText: 2026 Bonial International GmbH
SPDX-License-Identifier: Apache-2.0
-->

# Repository Rulesets

This directory documents the branch and tag rulesets applied to this repo
via `gh api`. GitHub does not natively read rulesets from repo files — the
API/UI is the source of truth. These JSON files exist so a reader can see
the intended state without visiting the web UI.

## Apply

```bash
gh api -X POST repos/bonial-oss/go-release-demo-release-please-minor/rulesets \
  --input .rulesets/main.json

gh api -X POST repos/bonial-oss/go-release-demo-release-please-minor/rulesets \
  --input .rulesets/tags.json
```

Applied ruleset IDs are recorded in the plan's execution ledger
(`.superpowers/sdd/…/progress.md` in the meta repo).

## `main.json` — branch ruleset

Enforces PR-based changes to `main`:

- Requires a pull request (0 approvals for this PoC — solo-dev setup;
  dismiss stale reviews on push; resolve review threads before merge).
- Required status checks: `test-and-lint`, `commitlint`.
- Blocks force-push (`non_fast_forward`) and branch deletion.
- No bypass actors.

`enforcement: "active"` — real enforcement.

### PoC deviations from plan-intended shape

- **`required_approving_review_count: 0`** (plan called for 1). In this
  solo-dev PoC there's no second reviewer; requiring an approval would
  deadlock all self-merges. In production, raise back to 1 (or higher)
  and add reviewers to CODEOWNERS.
- **`required_signatures` rule dropped** (plan included it). The PoC
  author's git config doesn't have `commit.gpgsign=true` and no
  automation is set up to sign commits from CI. In production, either
  enable `commit.gpgsign=true` globally (`gpg.format=ssh` with an SSH
  signing key) or gate on GitHub's "vigilant mode" — then re-add
  `{ "type": "required_signatures" }` to the ruleset.

Both deviations are documented as follow-ups; the ruleset shape in this
file matches what's actually applied to the repo.

## `tags.json` — tag ruleset (evaluate-mode, provisional)

**Current state: `enforcement: "evaluate"` (log-only, does not block).**

Intent: restrict creation/updates/deletion of `v[0-9]+.[0-9]+.[0-9]+`
tags so only the release workflow can create them. The plan calls for
`enforcement: "active"` with `bypass_actors: [{actor_id: 15368,
actor_type: "Integration", bypass_mode: "always"}]` — the well-known
GitHub Actions integration ID — so `secrets.GITHUB_TOKEN` pushes in the
release workflow bypass the ruleset while human users cannot.

Applying that shape against `bonial-oss/go-release-demo-release-please-minor`
fails with:

```
Validation Failed:
"Actor GitHub Actions integration must be part of the ruleset source
 or owner organization"
```

The `bonial-oss` org does not have the built-in GitHub Actions
integration on its ruleset bypass list. It is not something an admin can
freely enable on the org side — GitHub treats it as a distinct app grant.

### Follow-up (proper fix for production)

Create a dedicated GitHub App (working name: `bonial-release`) with
minimal scopes:

- Repository permissions: `contents: write`
- Repository access: this repo only

Install the app at the org level. Then in `tags.json`:

- Replace `actor_id: 15368` with the new app's numeric App ID.
- Replace `actor_type: "Integration"` with `actor_type: "Integration"`
  (unchanged — it's still an Integration bypass, just a different one).
- Set `enforcement: "active"`.

And in `.github/workflows/release.yaml`, replace `secrets.GITHUB_TOKEN`
with a token minted from the new app (via
`actions/create-github-app-token` or similar), so the release workflow
pushes tags under the app's identity — which the ruleset then correctly
allows to bypass.

Once the App is in place, **`verify-release.yaml` will also auto-trigger**
on `release: types: [released]`. Today it doesn't, because `promote`
publishes the release under `secrets.GITHUB_TOKEN` and GitHub's
[workflow chain-prevention rule](https://docs.github.com/en/actions/using-workflows/triggering-a-workflow#triggering-a-workflow-from-a-workflow)
suppresses the `released` event in that case. The App's identity is
distinct from `github-actions[bot]`, so the event fires normally. Until
then, operators dispatch `verify-release.yaml` manually after each
release (see the README's "Verifying a published release" section).

Until that App exists, this ruleset stays in `evaluate` mode: violations
are recorded in the ruleset's rule-suite log but not blocked. The
`immutable-releases` repo setting (Task 17) is the load-bearing
protection for release-asset immutability regardless; the tag ruleset
adds defense-in-depth on the tag layer.
