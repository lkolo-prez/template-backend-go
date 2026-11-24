# AGENTS.md

Instructions for AI coding agents (the [agents.md](https://agents.md) open standard, used by
Codex and compatible tools). Claude Code reads `CLAUDE.md`; the two are kept consistent.

## Project

Production Go HTTP microservice template — `chi`, `log/slog` JSON, OpenTelemetry, graceful
shutdown, distroless/static/non-root image. Module: `github.com/aircloud/template-backend-go`.
Go **1.26+**. Everything in English. Target year 2026.

## Setup commands

```bash
make setup     # install golangci-lint v2.15.0, govulncheck, gosec, air; wire pre-commit
```

## Dev / build / test commands

Drive everything through the Makefile (`make help` lists all targets):

```bash
make run           # run locally (development logging)
make run-watch     # live reload (air)
make build         # static ./bin/server binary
make test          # go test -race -count=1 ./...
make cover         # tests + 60% coverage gate
make lint          # golangci-lint run ./...
make format        # gofmt + golangci-lint --fix
make vet           # go vet ./...
make security      # govulncheck + gosec
make docker        # build the container image
make ci            # tidy + format-check + vet + lint + test + cover + security
```

## Validation before you finish

1. Run `make ci` (or at minimum `make format-check vet lint test`).
2. Add/adjust tests for any behavior change; keep coverage >= 60%.
3. Ensure `golangci-lint`, `govulncheck` and `gosec` are clean.

## Code style

- Format with `gofumpt` + `goimports` (`make format`). The strict linter set lives in
  `.golangci.yml` — do not weaken it; justify any `//nolint` inline.
- Errors: wrap with `%w`, check every return, use `errors.Is/As`.
- Logging: `log/slog` JSON only, **snake_case keys**, context-aware methods in handlers.
- No `init()` functions; no new global mutable state. Comments are full sentences with a period.
- Keep `internal/httpserver` transport-only; business logic belongs in `internal/handlers`.

## Project structure

- `cmd/server/` — entrypoint (config, logger, OTel, server, graceful shutdown, `-healthcheck`).
- `internal/config/` — env config + validation (safe defaults).
- `internal/httpserver/` — chi router + middleware chain.
- `internal/handlers/` — health probes + v1 widgets CRUD (in-memory) + response helpers.
- `internal/observability/` — OpenTelemetry tracer setup.
- `.github/workflows/` — CI, container (build/sign/attest), CodeQL, scorecard, gitleaks, release.

## Security & CI rules (must preserve)

- Never commit secrets or a real `.env`; use `.env.example`. gitleaks runs in CI + pre-commit.
- Keep the runtime image distroless/static/non-root — no shell, no package manager.
- In workflows: pin every action to a 40-hex SHA with a `# vX.Y.Z` comment, start from
  `permissions: contents: read`, set a `concurrency` group, and put `harden-runner` first.
- Auth to GHCR/cloud is **GitHub OIDC** only — never add long-lived secrets or PATs.

## Commits & PRs

- **Conventional Commits**, **signed (verified)** commits. See `CONTRIBUTING.md`.
- Update `README.md` / `docs/` when the public API or config changes.
