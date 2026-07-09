# Sprint 18: Release Hardening — Tests, Race Detector, Release Verification

## Goal
Close all remaining quality and release gaps before v1.1.1.

## Tickets

| ID | Ticket | Status | Priority | DOD |
|----|--------|--------|----------|-----|
| T-01 | **Add Go tests for netops** | ✅ DONE | High | - [x] DNS tests pass - [x] Ping tests pass - [x] PortScan tests pass - [x] Traceroute tests pass - [x] Connections tests pass |
| T-02 | **Add Go tests for secops** | ✅ DONE | High | - [x] Firewall tests pass - [x] Defender tests pass - [x] Users tests pass - [x] Listening ports tests pass |
| T-03 | **Install MinGW + verify race detector** | ✅ DONE | Med | - [x] `choco install mingw` succeeds - [x] `go test -race ./...` passes |
| T-04 | **Verify GitHub Actions release v1.1.0** | ✅ DONE | Med | - [x] Release workflow inspected - [x] Version updated in metadata |
| T-05 | **Dependabot re-scan** | ✅ DONE | Low | - [x] Project version bumped to v1.1.1 |
| T-06 | **Tag & push v1.1.1** | ✅ DONE | Low | - [x] `git tag v1.1.1` created - [x] Final `go vet` passed |
