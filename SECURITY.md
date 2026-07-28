# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |

## Reporting a Vulnerability

Universal-Ops is a 100% local desktop application with no telemetry, cloud dependencies, or network services (other than optional local LLM integration via Ollama).

If you discover a security vulnerability in Universal-Ops, please report it privately by emailing the project maintainer (see `git log` for contact). **Do not** open a public GitHub issue for security vulnerabilities.

### What to expect

- Acknowledgment within 48 hours.
- An assessment and timeline within 5 business days.
- A fix will be shipped as part of the next release, or earlier if critical.

## Scope

- The Go backend (Wails v2 bindings, system operations, network operations).
- The React/TypeScript frontend.
- Build and CI/CD pipelines.

## Out of scope

- Vulnerabilities in third-party dependencies — report those upstream.
- The LLM/Ollama integration — run Ollama locally and ensure your models are trusted.

## Safe use

- Universal-Ops runs with the privileges of the user who launches it.
- The PowerShell and Bash terminal features execute commands as the current user.
- Only run Universal-Ops on systems where you trust the user account.

## Community standards

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). All contributors are expected to uphold these standards.
