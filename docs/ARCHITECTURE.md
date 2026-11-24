# Architecture

`template-backend-go` is a single, self-contained HTTP service that follows a strict
transport / logic separation and the 12-factor methodology. This document describes the
runtime request path, the package boundaries, and the supply-chain pipeline.

## Runtime request flow

```mermaid
flowchart LR
  client([Client]) -->|HTTP| srv["http.Server (cmd/server)"]
  srv --> router["chi.Router (internal/httpserver)"]

  subgraph mw["Middleware chain (outer -> inner)"]
    direction TB
    rid[RequestID] --> rec[Recover] --> sec[SecurityHeaders] --> logm[Logging] --> to[Timeout 30s] --> otel{OTel enabled?}
  end

  router --> mw
  otel -->|yes| otelh[otelhttp handler] --> routes
  otel -->|no| routes

  subgraph routes["Routes (internal/handlers)"]
    direction TB
    health["/healthz, /readyz"]
    widgets["/api/v1/widgets CRUD"]
  end

  widgets --> store[("in-memory store\n(sync.RWMutex)")]
  health --> resp["respond.go helpers"]
  widgets --> resp
  resp -->|JSON| client

  logm -. "log/slog JSON" .-> stdout[(stdout)]
  otelh -. spans .-> collector[(OTLP/HTTP collector)]
```

### Component responsibilities

| Package | Responsibility |
| --- | --- |
| `cmd/server` | Wire-up: load config, build the slog JSON logger, set up OpenTelemetry, construct the `http.Server` with timeouts, run it, and drain on `SIGINT`/`SIGTERM`. Also implements `-healthcheck` (used by the Docker `HEALTHCHECK`) and `-version`. |
| `internal/config` | Parse + validate environment configuration (`caarlos0/env` + `go-playground/validator`). Safe defaults so the service boots with no env set. |
| `internal/httpserver` | Build the chi router and the middleware chain. Transport-only — no business logic. |
| `internal/handlers` | Health probes and the v1 widgets CRUD API (in-memory store) plus the canonical JSON response helpers. |
| `internal/observability` | Configure the OpenTelemetry tracer (OTLP/HTTP exporter; W3C propagation always installed; no-op when disabled). |

### Lifecycle

```mermaid
sequenceDiagram
  participant OS
  participant main as cmd/server
  participant srv as http.Server
  participant tp as OTel TracerProvider

  main->>main: config.Load() + validate
  main->>tp: SetupTracing(ctx)
  main->>srv: ListenAndServe (goroutine)
  Note over main: select { serverErr | ctx.Done() }
  OS-->>main: SIGTERM / SIGINT
  main->>srv: Shutdown(ctx, ShutdownTimeout)
  main->>tp: ForceFlush + Shutdown
  main-->>OS: exit 0
```

The readiness probe (`/readyz`) returns `503` while draining so an orchestrator stops routing
traffic before connections are closed.

## Configuration

Configuration is entirely environment-driven (see `internal/config/config.go` and
`.env.example`). Networking, service identity, log level, HTTP timeouts, request size caps and
OpenTelemetry settings are all tunable; every key has a safe default.

## Container

A multi-stage `Dockerfile` builds a fully static binary (`CGO_ENABLED=0`, `-trimpath`,
stripped, version-stamped via `-ldflags`) in `golang:1.27-bookworm`, then copies it into
`gcr.io/distroless/static-debian12:nonroot`. The final image has no shell or package manager,
runs as uid `65532`, exposes `8080`, and uses the binary's own `-healthcheck` mode for the
`HEALTHCHECK`. Pin the base images by digest in real deployments.

## Supply-chain pipeline

```mermaid
flowchart TD
  pr[Pull request] --> ci["ci.yml\n(gofmt, tidy, golangci-lint, race tests + coverage,\ngovulncheck, gosec SARIF, hadolint, image build no-push)"]
  pr --> codeql[codeql.yml]
  pr --> depr[dependency-review.yml]
  pr --> leaks[gitleaks.yml]

  ci --> review{2 approvals + CODEOWNERS\nsigned commits + linear history}
  review --> main[(main)]

  main --> container["container.yml (OIDC)\nBuildx build + push to GHCR"]
  container --> sbom[SBOM CycloneDX - Syft]
  container --> trivy[Trivy image scan - SARIF]
  container --> sign[cosign keyless signature]
  container --> attest[SBOM attestation + SLSA build provenance]

  main --> release["release.yml\nrelease-please -> SemVer tag"]
  release --> gor[GoReleaser signed binaries]

  cron[Schedules] --> scorecard[scorecard.yml]
  cron --> codeql
  cron --> leaks
```

Every workflow pins its actions to full 40-hex commit SHAs (with `# vX.Y.Z` comments), starts
from `permissions: contents: read`, sets a `concurrency` group, and runs
`step-security/harden-runner` first. Registry and signing auth is **GitHub OIDC** — there are
no long-lived secrets.

See [BRANCH-PROTECTION.md](BRANCH-PROTECTION.md) for the required GitHub Ruleset.
