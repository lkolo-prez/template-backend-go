# CLAUDE.md

Guidance for Claude Code (and other AI assistants) working in this repository.
This is a production Go HTTP microservice template. Treat every change as if it ships.

## What this is

A 12-factor Go service: `chi` router, `log/slog` JSON logging, OpenTelemetry tracing,
graceful shutdown, distroless/static/non-root container. Module path:
`github.com/aircloud/template-backend-go`. Go **1.23+**.

## Build, test and lint commands

Use the Makefile — do not invent commands. Run `make help` for the full list.

```bash
make setup         # install golangci-lint v1.61.0, govulncheck, gosec, air; wire pre-commit
make run           # run locally (ENVIRONMENT=development LOG_LEVEL=debug)
make run-watch     # live reload via air
make build         # static ./bin/server (CGO_ENABLED=0, -trimpath, ldflags-stamped)
make test          # go test -race -count=1 ./...
make cover         # tests + coverage; enforces COVER_MIN=60
make lint          # golangci-lint run ./... (strict; config in .golangci.yml)
make format        # go fmt + golangci-lint --fix
make format-check  # fail if not gofmt-clean
make vet           # go vet ./...
make security      # govulncheck ./... + gosec
make sbom          # CycloneDX SBOM -> sbom.cdx.json (syft)
make docker        # build ghcr.io/aircloud/template-backend-go image
make ci            # tidy + format-check + vet + lint + test + cover + security
```

**Before proposing a change as done, run `make ci`** (or at minimum `make format-check vet
lint test`). Add or update tests for any behavior change; coverage must stay >= 60%.

## Project layout

- `cmd/server/main.go` — entrypoint: config load, slog JSON logger, OTel setup, `http.Server`
  with timeouts, signal-based graceful shutdown, and a `-healthcheck` mode used by the
  Docker `HEALTHCHECK`.
- `internal/config/` — env-driven `Config` (`caarlos0/env` + `go-playground/validator`).
  All settings have safe defaults; see the struct tags for env var names.
- `internal/httpserver/` — `New(Options)` builds the chi router and the middleware chain:
  `RequestID` -> `Recover` -> `SecurityHeaders` -> `Logging` -> `Timeout` (-> OTel if enabled).
- `internal/handlers/` — health probes (`/healthz`, `/readyz`) and the v1 widgets CRUD API
  with an in-memory store, request-body limits and validation. Responses go through the
  helpers in `respond.go` (`writeJSON` / `writeError` / `writeValidationError`).
- `internal/observability/` — `SetupTracing` (OTLP/HTTP exporter; no-op when disabled).

## Conventions (follow the existing code)

- **Errors**: wrap with `%w`; check every error (`errcheck`/`errorlint` enforce this).
- **Logging**: `log/slog` only, JSON handler, **snake_case attribute keys**, prefer the
  context-aware methods (`InfoContext`, `LogAttrs`) inside handlers. `sloglint` is strict.
- **HTTP**: never start a request without a context; cap request bodies (`MaxBytesReader`)
  and set timeouts. Keep `internal/httpserver` transport-only; logic lives in `handlers`.
- **No `init()` functions, no global mutable state** beyond what already exists.
- **Style**: `gofumpt` + `goimports`. Comments are full sentences ending with a period
  (`godot`). Exported identifiers are documented.
- **Tests**: table-driven, use `testify`, `t.Parallel()` where safe, race detector on.

## Security expectations

- Keep the container distroless/static/non-root; do not add a shell or package manager.
- Never commit secrets or a real `.env` — use `.env.example`. gitleaks runs in CI and pre-commit.
- In `.github/workflows`, **every action is pinned to a 40-hex SHA** with a `# vX.Y.Z`
  comment; jobs start from `permissions: contents: read`, set a `concurrency` group, and use
  `step-security/harden-runner` as the first step. Preserve all of this when editing workflows.
- Registry/cloud auth is **GitHub OIDC** — never introduce long-lived secrets or PATs.

## Commits & PRs

- **Conventional Commits**; commits must be **signed (verified)**.
- Update `README.md` / `docs/` when public API or config changes.
- See `CONTRIBUTING.md` for the full workflow and `docs/BRANCH-PROTECTION.md` for the ruleset.
