# template-backend-go

[![CI](https://github.com/aircloud/template-backend-go/actions/workflows/ci.yml/badge.svg)](https://github.com/aircloud/template-backend-go/actions/workflows/ci.yml)
[![Container](https://github.com/aircloud/template-backend-go/actions/workflows/container.yml/badge.svg)](https://github.com/aircloud/template-backend-go/actions/workflows/container.yml)
[![CodeQL](https://github.com/aircloud/template-backend-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/aircloud/template-backend-go/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/aircloud/template-backend-go/badge)](https://scorecard.dev/viewer/?uri=github.com/aircloud/template-backend-go)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](./LICENSE)

Opinionated, security-first **Go HTTP microservice** template. It ships a production-shaped
service built on the standard library plus [chi](https://github.com/go-chi/chi): structured
`log/slog` JSON logging, OpenTelemetry tracing, 12-factor config, graceful shutdown,
deny-by-default security headers, and a multi-stage build into a **distroless, static,
non-root** image — all wired to **GitHub Actions with OIDC** so there are **no long-lived
registry or cloud secrets**.

> This is a **template repository**. The org `aircloud`, registry `ghcr.io/aircloud` and the
> module path `github.com/aircloud/template-backend-go` are **examples** — search-and-replace
> them for your org. Target toolchain year: **2026** (Go 1.23+).

## Why this exists

- **Keyless by design**: the container workflow pushes to GHCR and signs/attests with
  GitHub OIDC (cosign keyless + SLSA build provenance). No PATs, no registry passwords.
- **Secure by default**: distroless `static-debian12:nonroot` runtime (uid 65532), no shell,
  read-only-friendly, request-body limits, strict response-security headers, request timeouts.
- **Supply-chain hardened**: every action pinned to a 40-hex SHA, `harden-runner` egress
  control on each job, SBOM (CycloneDX via Syft), Trivy image scan, govulncheck + gosec,
  CodeQL, gitleaks, dependency review and OpenSSF Scorecard.

## Repository layout

```
cmd/server/          Process entrypoint: config, logging, OTel, server, graceful shutdown
internal/
  config/            12-factor env config (caarlos0/env) + validation
  httpserver/        chi router, middleware (request-id, recover, security headers, logging)
  handlers/          Health probes + the v1 widgets CRUD API (in-memory store)
  observability/     OpenTelemetry tracer setup (OTLP/HTTP, no-op when disabled)
docs/                Architecture (Mermaid) + branch-protection ruleset
.github/workflows/   CI, container build/sign/attest, CodeQL, scorecard, gitleaks, release
Dockerfile           Multi-stage -> distroless/static nonroot
Makefile             Developer task runner (`make help`)
```

## Prerequisites

| Tool | Version |
| --- | --- |
| Go | >= 1.23 |
| Docker (with Buildx) | >= 24 (for `make docker`) |
| golangci-lint | v1.61.0 (`make setup` installs it) |
| pre-commit | >= 3.7 (optional, recommended) |

The `.devcontainer` provides the full toolchain; VS Code will offer to reopen in it.

## Quick start

```bash
# 1. Install developer tooling (lint, vuln, gosec, air, pre-commit hooks)
make setup

# 2. Run the service locally (development logging, debug level)
make run
#   -> http server listening on 0.0.0.0:8080

# 3. Exercise it
curl -s localhost:8080/healthz | jq
curl -s -X POST localhost:8080/api/v1/widgets \
  -H 'content-type: application/json' \
  -d '{"name":"gizmo","color":"#00ff88","quantity":3}' | jq

# 4. Run the full local pipeline before pushing
make ci          # tidy, format-check, vet, lint, test, cover, security
```

## Make targets

Run `make help` for the authoritative list. The most-used targets:

| Target | What it does |
| --- | --- |
| `make setup` | Install golangci-lint, govulncheck, gosec, air; wire pre-commit |
| `make run` | Run the server locally (`ENVIRONMENT=development LOG_LEVEL=debug`) |
| `make run-watch` | Run with live reload via [air](https://github.com/air-verse/air) |
| `make build` | Build the static `./bin/server` binary (`CGO_ENABLED=0`, trimmed, stamped) |
| `make test` | `go test -race` across all packages |
| `make cover` | Tests with coverage, enforcing the `COVER_MIN` (60%) gate |
| `make lint` | `golangci-lint run` (strict config in `.golangci.yml`) |
| `make format` / `make format-check` | Apply / verify gofmt + goimports formatting |
| `make vet` | `go vet ./...` |
| `make security` | `govulncheck` + `gosec` |
| `make sbom` | Generate `sbom.cdx.json` (CycloneDX via Syft) |
| `make docker` | Build the container image (`ghcr.io/aircloud/template-backend-go`) |
| `make ci` | Full local pipeline (tidy → format-check → vet → lint → test → cover → security) |
| `make clean` | Remove build/test artifacts and caches |

## HTTP API

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/healthz` | Liveness probe — `200` while the process is up |
| `GET` | `/readyz` | Readiness probe — `503` while draining, else `200` |
| `GET` | `/api/v1/widgets` | List widgets (newest first) |
| `POST` | `/api/v1/widgets` | Create a widget (`201`, `Location` header) |
| `GET` | `/api/v1/widgets/{id}` | Fetch one widget |
| `PUT` | `/api/v1/widgets/{id}` | Replace a widget |
| `DELETE` | `/api/v1/widgets/{id}` | Delete a widget (`204`) |

Errors use a canonical envelope — `{"error":"...","details":{...}}` — and validation
failures return `422` with field-level `details`. The widget store is **in-memory** and
exists only to demonstrate request decoding, validation and structured responses; replace
it with a real datastore in a derived service.

## Configuration

All configuration is via environment variables (12-factor); every value has a safe default
so the service boots with none set. See [`internal/config/config.go`](internal/config/config.go)
and [`.env.example`](.env.example).

| Variable | Default | Notes |
| --- | --- | --- |
| `HOST` / `PORT` | `0.0.0.0` / `8080` | Listen address |
| `SERVICE_NAME` | `template-backend-go` | Logged + used as the OTel service name |
| `ENVIRONMENT` | `development` | `development` \| `staging` \| `production` |
| `LOG_LEVEL` | `info` | `debug` \| `info` \| `warn` \| `error` |
| `READ_HEADER_TIMEOUT` … `SHUTDOWN_TIMEOUT` | see config | HTTP server timeouts |
| `MAX_HEADER_BYTES` | `1048576` | Request header cap |
| `OTEL_ENABLED` | `false` | Enable OTLP/HTTP trace export |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4318` | Collector endpoint |
| `OTEL_TRACES_SAMPLER_ARG` | `1.0` | Trace sample ratio (0–1) |

## CI/CD flow

| Trigger | Workflow | What runs |
| --- | --- | --- |
| Pull request / push | `ci.yml` | gofmt, `go mod tidy` check, golangci-lint, race tests + coverage gate, govulncheck, gosec (SARIF), hadolint, image build (no push) |
| Push to `main` / `v*.*.*` tag | `container.yml` | Buildx build & push to GHCR, SBOM (Syft/CycloneDX), Trivy scan, **cosign keyless signature**, SBOM attestation, **SLSA build provenance** — all via OIDC |
| PR / push / schedule | `codeql.yml` | CodeQL `security-extended` + `security-and-quality` |
| PR | `dependency-review.yml` | Blocks high-severity / disallowed-license dependency changes |
| Push / PR / schedule | `gitleaks.yml` | Secret scanning across full history |
| Push / merge to `main` | `release.yml` | release-please (SemVer from Conventional Commits) + GoReleaser binaries |
| Schedule / branch-protection | `scorecard.yml` | OpenSSF Scorecard, SARIF to code scanning |

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the runtime + supply-chain diagrams and
[docs/BRANCH-PROTECTION.md](docs/BRANCH-PROTECTION.md) for the required GitHub Ruleset.

## Security posture

- **Runtime**: distroless static non-root image, no shell/package manager, explicit
  `USER 65532`, binary-based `HEALTHCHECK` (`/server -healthcheck`), conservative response
  security headers, 1 MiB request-body cap, per-request timeouts, panic recovery.
- **Build**: `CGO_ENABLED=0` static binary, `-trimpath`, stripped, reproducible-friendly.
- **Supply chain**: SHA-pinned actions, harden-runner egress allow-lists, SBOM + cosign
  keyless signing + build provenance, Trivy/govulncheck/gosec/CodeQL/gitleaks/Scorecard.
- **Reporting**: see [SECURITY.md](SECURITY.md). Do **not** open a public issue for a vuln.

## Contributing & governance

- Contribution workflow + Conventional Commits: [CONTRIBUTING.md](CONTRIBUTING.md).
- Community expectations: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
- AI-assistant guidance lives in [CLAUDE.md](CLAUDE.md), [AGENTS.md](AGENTS.md),
  [.github/copilot-instructions.md](.github/copilot-instructions.md) and
  [.cursor/rules/00-project.mdc](.cursor/rules/00-project.mdc).

## License

[Apache-2.0](./LICENSE) — Copyright 2026 AirCloud.
