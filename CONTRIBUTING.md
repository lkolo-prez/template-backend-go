# Contributing

Thanks for contributing to `template-backend-go`. This repo follows trunk-based development
with short-lived branches and automated quality gates.

## Ground rules

- Be respectful — see [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
- Never commit secrets, credentials or a real `.env` (use `.env.example`).
- All commits must be **signed (verified)** and follow **Conventional Commits**.

## Development setup

```bash
make setup          # installs golangci-lint, govulncheck, gosec, air; wires pre-commit
pre-commit run -a   # run all hooks against the whole tree
```

Required tools: Go >= 1.26, Docker (with Buildx) for image builds, golangci-lint v2.15.0,
and (optionally) pre-commit. The `.devcontainer` provides all of them.

## Everyday commands

```bash
make run            # run locally with development logging
make run-watch      # live reload via air
make test           # go test -race ./...
make cover          # tests + coverage gate (COVER_MIN=60)
make lint           # golangci-lint (strict)
make security       # govulncheck + gosec
make ci             # full local pipeline (run this before pushing)
```

## Branch & commit conventions

- Branch from `main`: `feat/<area>-<slug>`, `fix/<area>-<slug>`, `docs/<slug>`.
- **Conventional Commits** drive release-please and the changelog:

  ```
  feat(handlers): add cursor pagination to GET /widgets
  fix(httpserver): close response body on otel handler error
  docs(architecture): add supply-chain trust diagram
  ci(actions): pin scorecard action to v2.4.0
  ```

- A breaking change uses `!` and a footer:

  ```
  feat(api)!: move widgets under /api/v2

  BREAKING CHANGE: clients must update the base path from /api/v1 to /api/v2.
  ```

## Signed commits

Verified commits are **required** by the ruleset. Configure once:

```bash
# SSH signing (simplest)
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_ed25519.pub
git config --global commit.gpgsign true
# then add the key under GitHub -> Settings -> SSH and GPG keys -> "Signing Key"
```

## Pull request checklist

1. `make ci` passes locally (or rely on pre-commit + CI).
2. New behavior is covered by tests; coverage stays at or above the 60% gate (`make cover`).
3. `golangci-lint`, `govulncheck` and `gosec` are green; any `//nolint` is justified inline.
4. Public API or config changes are reflected in `README.md` and, if relevant,
   `docs/ARCHITECTURE.md`.
5. No secrets, credentials or real `.env` files are committed.
6. Commits are **signed (verified)** and follow **Conventional Commits**.
7. At least **2 approvals** including the relevant CODEOWNERS.

## How a change ships

```
PR  --> ci (gofmt, tidy, lint, race tests + coverage, govulncheck, gosec, hadolint, image build)
    --> codeql / dependency-review / gitleaks
    --> review (2 approvals, CODEOWNERS, signed commits, linear history)
    --> merge to main
         |-> container (build & push to GHCR, SBOM, Trivy, cosign sign, provenance — OIDC)
         |-> release-please opens/updates the release PR; merging it tags a SemVer release,
             then GoReleaser publishes signed binaries.
```

## Reporting bugs / requesting features

Use the issue templates. For **security issues**, follow [SECURITY.md](SECURITY.md) — do not
open a public issue.
