---
name: OpsForAll Go toolchain mismatch
description: Why `go build`/`go vet` fail in this sandbox for the OpsForAll (Hawkward) project, and how to verify Go changes anyway.
---

`go.mod` declares `go 1.26.5`, and a direct dependency (`github.com/ollama/ollama`)
requires `go >= 1.26.0`. The sandbox's package-management module list tops out at
`go-1.25`, and the module toolchain auto-download is blocked (GOPROXY points at
Replit's internal firewall proxy, which doesn't mirror that exact toolchain).

**Why it matters:** `go build ./...` and `go vet ./...` fail immediately with a
toolchain version error, even on completely unmodified code — this is not a
regression from any particular change, so don't waste time bisecting edits to
find "what broke the build."

**How to apply:** Verify Go edits with `gofmt -l`/`gofmt -e <file>` (catches
syntax errors) plus careful manual review instead of a full build. If a newer
Go module ever becomes available via `listAvailableModules({ language: "go" })`,
re-check whether it satisfies both the go.mod directive and the ollama
dependency's minimum before assuming the build is fixed.
