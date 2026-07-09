# Sprint 18: Release Hardening — Tests, Race Detector, Release Verification

## Goal
Close all remaining quality and release gaps before v1.1.1.

## Tickets

| ID | Ticket | Status | Priority | DOD |
|----|--------|--------|----------|-----|
| T-01 | **Add Go tests for netops** | ✅ DONE | High | - [x] DNS tests pass - [x] Ping tests pass - [x] PortScan tests pass - [x] Traceroute tests pass - [x] Connections tests pass |
| T-02 | **Add Go tests for secops** | ✅ DONE | High | - [x] Firewall tests pass - [x] Defender tests pass - [x] Users tests pass - [x] Listening ports tests pass |
| T-03 | **Install MinGW + verify race detector** | 🔲 TODO | Med | - [ ] `choco install mingw` succeeds - [ ] `wails build -race` produces clean binary |
| T-04 | **Verify GitHub Actions release v1.1.0** | 🔲 TODO | Med | - [ ] Release page shows v1.1.0 assets - [ ] Windows .exe exists - [ ] Checksums present |
| T-05 | **Dependabot re-scan** | 🔲 TODO | Low | - [ ] Dependabot tab shows 0 open alerts (or fewer than 19) |
| T-06 | **Tag & push v1.1.1** | 🔲 TODO | Low | - [ ] All above complete - [ ] `git tag v1.1.1` pushed |
