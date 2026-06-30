# GitHub Copilot instructions

Repository-wide guidance for GitHub Copilot in `template-backend-go` — a production Go HTTP
microservice template (`chi`, `log/slog` JSON, OpenTelemetry, distroless/static/non-root).
Go **1.23+**, module `github.com/aircloud/template-backend-go`.

## General

- Prefer the Go standard library; the only HTTP dependency is `go-chi/chi/v5`. Do not add
  heavy frameworks or new dependencies without a clear need.
- Match the existing style: `gofumpt`-formatted, `goimports`-grouped, comments are full
  sentences ending with a period.
- Build, test and lint via the Makefile: `make build`, `make test`, `make cover`, `make lint`,
  `make security`, `make ci`. Coverage must stay at or above 60%.

## Writing code

- **Errors**: always check them; wrap with `fmt.Errorf("...: %w", err)`; compare with
  `errors.Is` / `errors.As`. Never discard an error silently.
- **Logging**: use `log/slog` with the JSON handler only. Attribute keys are **snake_case**.
  Inside handlers use context-aware methods (`InfoContext`, `LogAttrs`). No `fmt.Print`/`log`.
- **HTTP handlers**: cap request bodies with `http.MaxBytesReader`, decode with
  `DisallowUnknownFields`, validate input with `go-playground/validator`, and respond via the
  `writeJSON` / `writeError` / `writeValidationError` helpers in `internal/handlers/respond.go`.
- **Context**: never make an outbound request without a context; honor request cancellation.
- **Concurrency**: guard shared state with a mutex (see the widgets store); run tests with `-race`.
- Do **not** add `init()` functions or new global mutable state.

## Tests

- Table-driven tests with `testify`; use `t.Parallel()` where safe. New behavior needs tests.

## Security & workflows

- Never hardcode secrets; use environment variables and `.env.example`. gitleaks runs in CI.
- Keep the container distroless/static/non-root — do not add a shell or installer steps.
- When editing `.github/workflows/*`: keep every action pinned to a full 40-hex SHA with a
  `# vX.Y.Z` comment, keep top-level `permissions: contents: read`, keep the `concurrency`
  group, and keep `step-security/harden-runner` as the first step of each job. Auth is OIDC —
  never introduce long-lived secrets.

## Commits

Use Conventional Commits (`feat:`, `fix:`, `docs:`, `ci:`, `build:`, `refactor:`, `chore:`).
