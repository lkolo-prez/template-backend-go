# Security Policy

We take the security of `template-backend-go` and services derived from it seriously.
Thank you for helping keep the ecosystem safe.

## Reporting a vulnerability

**Do not open a public issue, pull request or discussion for a security problem.**

Report privately through **GitHub Security Advisories**:

- <https://github.com/lkolo-prez/template-backend-go/security/advisories/new>

If you cannot use GitHub Advisories, email **office@aircloud.com.pl** and, where possible,
encrypt with the key referenced in our [`security.txt`](.well-known/security.txt).

Please include:

- a description of the issue and its impact;
- the affected version, commit SHA or image digest;
- reproduction steps or a proof-of-concept;
- any suggested remediation.

## Our commitment (SLA)

| Stage | Target |
| --- | --- |
| Acknowledgement of your report | within **2 business days** |
| Initial triage & severity (CVSS) assessment | within **5 business days** |
| Fix or mitigation for **critical/high** issues | within **30 days** |
| Fix for medium/low issues | next scheduled release |
| Coordinated disclosure & advisory publication | by mutual agreement, after a fix ships |

We follow **coordinated disclosure**. We will keep you updated on progress, credit you in the
advisory (unless you prefer to remain anonymous), and let you know when a fix is released.

## Supported versions

This is a template; downstreams own their own release trains. For the template itself, only
the latest released minor version on `main` receives security fixes.

| Version | Supported |
| --- | --- |
| latest `main` / newest release | ✅ |
| older releases | ❌ |

## Scope

In scope: the application code (`cmd/`, `internal/`), the `Dockerfile`, the CI/CD workflows
and their supply-chain configuration. Out of scope: third-party dependencies (report those
upstream; we track them via Dependabot, govulncheck, Trivy and dependency review), and
issues that require a compromised build host or privileged local access.

## Security machinery in this repo

- Static & dynamic analysis: `golangci-lint` (incl. `gosec`), `gosec` SARIF, **CodeQL**.
- Vulnerabilities: **govulncheck**, **Trivy** image scan, **dependency review**.
- Secrets: **gitleaks** (full-history scan) + a pre-commit hook.
- Supply chain: SHA-pinned actions, **harden-runner** egress control, **SBOM** (CycloneDX),
  **cosign** keyless signatures, **SLSA build provenance**, **OpenSSF Scorecard**.

A machine-readable contact is published at
[`/.well-known/security.txt`](.well-known/security.txt).
