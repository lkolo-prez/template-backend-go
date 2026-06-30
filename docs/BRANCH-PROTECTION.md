# Branch protection (GitHub Rulesets)

`main` is protected by a GitHub **Ruleset** (the modern replacement for classic branch
protection). Configure it under **Settings -> Rules -> Rulesets -> New branch ruleset**, or
apply it as code with the GitHub API / Terraform. These settings are required for the repo to
score well on OpenSSF Scorecard and to keep the supply chain trustworthy.

## Target

- **Ruleset name**: `protect-main`
- **Enforcement status**: Active
- **Target branches**: include the default branch (`main`).
- **Bypass list**: empty (no roles or apps bypass — not even admins).

## Required rules

| Rule | Setting |
| --- | --- |
| Restrict deletions | ✅ on |
| Restrict force pushes (non-fast-forward) | ✅ on |
| Require linear history | ✅ on |
| Require signed commits | ✅ on |
| Require a pull request before merging | ✅ on |
| — Required approvals | **2** |
| — Dismiss stale approvals on new commits | ✅ on |
| — Require review from Code Owners | ✅ on (see [`.github/CODEOWNERS`](../.github/CODEOWNERS)) |
| — Require approval of the most recent push | ✅ on |
| — Require conversation resolution before merging | ✅ on |
| Require status checks to pass | ✅ on |
| — Require branches to be up to date before merging | ✅ on |
| Block force pushes | ✅ on |

## Required status checks

Add these checks (the job names from the workflows) as **required**:

- `Lint & format` — from `ci.yml`
- `Test & coverage` — from `ci.yml`
- `Vulnerability & SAST` — from `ci.yml`
- `Build image (no push)` — from `ci.yml`
- `Analyze (go)` — from `codeql.yml`
- `Review dependency changes` — from `dependency-review.yml`
- `gitleaks` — from `gitleaks.yml`

> Status-check names must match the job `name:` exactly. If you rename a job, update the
> ruleset. New required checks only take effect after they have reported at least once.

## Signed commits

Verified commits are required. Contributors configure signing once (SSH signing is simplest):

```bash
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_ed25519.pub
git config --global commit.gpgsign true
# Add the key under GitHub -> Settings -> SSH and GPG keys -> "Signing Key".
```

## Tag protection

Protect release tags so only the release automation can create/move them:

- **Ruleset name**: `protect-tags`
- **Target**: tags matching `v*`
- Restrict creation/updates/deletions to the release workflow / maintainers.

## Repository-level settings (complementary)

- **Actions -> General**: allow only actions pinned by SHA; set default `GITHUB_TOKEN`
  permissions to **read-only**; require approval for first-time contributors' workflow runs.
- **Code security**: enable Dependabot alerts + security updates, secret scanning with push
  protection, and CodeQL default/advanced setup (this repo uses the advanced `codeql.yml`).
- **General -> Pull Requests**: allow squash merges only; auto-delete head branches.

## As code (optional)

The ruleset can be exported as JSON via
`gh api repos/aircloud/template-backend-go/rulesets` and version-controlled, or managed with
the `github_repository_ruleset` Terraform resource, so branch protection itself is reviewed
through the same PR process.
