# Sprint 13: Release Hardening & UI Audit

## Objective
Complete the comprehensive UI audit, fix remaining broken/wired buttons, clean up root directory, archive old plans, and verify the build so v1.0.0 release is solid.

| ID | Ticket | Status | Priority | DOD |
|----|--------|--------|----------|-----|
| T-01 | **Root Cleanup** — Remove stale `.exe`/`.log` files, update `.gitignore` | 🔲 TODO | High | - [ ] Root exe/log files removed - [ ] .gitignore updated - [ ] No tracked binaries |
| T-02 | **Archive Old Plans** — Move outdated sprint files to `docs/archive/` | 🔲 TODO | Med | - [ ] 12 old plan files moved - [ ] `plans/` has only current plan |
| T-03 | **Dashboard Button Handlers** — Add onClick for QUICK DIAGNOSTIC and GENERATE BRIEFING, clean up void states | 🔲 TODO | High | - [ ] Quick Diagnostic calls Dash.GetQuickDiagnostic - [ ] Generate Briefing navigates to AIOps - [ ] No void-suppressed unused state |
| T-04 | **NetOps Traceroute Tab** — Implement missing traceroute tab content | 🔲 TODO | High | - [ ] Traceroute tab renders with input + results - [ ] Backend Traceroute method verified |
| T-05 | **SecOps Full Audit** — Verify all 5 tabs work with proper states | 🔲 TODO | High | - [ ] Firewall toggle works - [ ] Users list returns real data - [ ] Listening ports works - [ ] Defender status works - [ ] Events loads |
| T-06 | **DevOps Full Audit** — Verify all 4 tabs work | 🔲 TODO | High | - [ ] Terminal runs + blocks dangerous commands - [ ] PowerShell workflow enforcement - [ ] Services list/control - [ ] FileBrowser works |
| T-07 | **AIOps/NetworkDesign/Logs/Settings** — Verify remaining pages | 🔲 TODO | Med | - [ ] AIOps Ollama check on load - [ ] NetworkDesign save/load - [ ] Logs refresh/filter - [ ] Settings wired |
| T-08 | **Build Verification** | 🔲 TODO | High | - [ ] `wails build` passes - [ ] `npm run build` passes |
| T-09 | **Test Verification** | 🔲 TODO | High | - [ ] `go test ./...` all pass - [ ] Frontend `npm test` all pass |
| T-10 | **Document No-Programming-Skills Flow** | 🔲 TODO | Low | - [ ] README enhanced with release screenshots - [ ] Clear download + run steps |
| T-11 | **Memory Update** — Record session state | 🔲 TODO | Low | - [ ] `.memory/` updated with sprint results |
