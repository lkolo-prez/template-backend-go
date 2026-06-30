<!--
Title MUST follow Conventional Commits, e.g.:
  feat(handlers): add cursor pagination to GET /widgets
  fix(httpserver): close response body on otel handler error
-->

## Summary

<!-- What does this change and why? Link the issue: Closes #NNN -->

## Type of change

- [ ] feat — new capability
- [ ] fix — bug fix
- [ ] refactor — no behavior change
- [ ] docs
- [ ] ci / build
- [ ] chore

## Affected areas

- [ ] cmd/server
- [ ] internal/config
- [ ] internal/httpserver
- [ ] internal/handlers
- [ ] internal/observability
- [ ] Dockerfile / image
- [ ] CI/CD

## Checklist

- [ ] `make ci` passes locally (or pre-commit + CI cover it).
- [ ] New behavior is covered by tests; coverage stays at or above the 60% gate (`make cover`).
- [ ] `golangci-lint`, `govulncheck` and `gosec` are green; any `//nolint` is justified inline.
- [ ] Public API or config changes are reflected in `README.md` / `docs/`.
- [ ] No secrets, credentials or real `.env` files are committed.
- [ ] Commits are signed (verified) and follow Conventional Commits.
