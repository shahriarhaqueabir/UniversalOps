# Audit Findings: SecOps Hardener

**Reviewer**: SecOps Hardener Subagent
**Date**: 2026-07-21
**Confidence**: 0.95 (High)

## 1. Executive Summary
The security posture of the DevOps terminal and AI layer is solid, featuring robust allowlists and metacharacter filtering. However, a significant "Privilege Escalation" risk exists in how PowerShell profiles are loaded from relative paths, and session isolation in the AI layer could be improved to prevent global state leakage.

## 2. High-Severity Findings

### [HIGH] PowerShell Profile Injection (CWD)
- **Location**: `internal/devops/shell.go:274` (`PowerShellProfilePath`)
- **Mechanism**: The candidate paths for the profile include `filepath.Join(cwd, "profiles", ...)`. 
- **Impact**: If an attacker can influence the working directory of the application (e.g., via a malicious shortcut or a script that changes CWD), they can plant a malicious `powershell_profile.ps1`. When the app runs a "trusted" workflow, the malicious code executes with the app's privileges.
- **Remediation**: Restrict profile loading to the absolute path of the application's executable directory (`os.Executable()`) only.

### [MEDIUM] Global AI Model State Leakage
- **Location**: `internal/aiops/ollama.go:64-83`
- **Mechanism**: Use of a global singleton `defaultClient` with a shared `effectiveModel` variable.
- **Impact**: Changes to the model in one part of the app (e.g., a background diagnostic) can unintentionally change the model for the user's active chat session.
- **Remediation**: Use session-scoped or task-scoped client instances instead of a global singleton.

## 3. Medium-Severity Findings

### [MEDIUM] Metacharacter Bypass via Unicode/Escapes
- **Location**: `internal/devops/shell.go:50` (`ContainsShellMetachar`)
- **Mechanism**: Relies on a blacklist of literal characters.
- **Impact**: Advanced command injection techniques on Windows `cmd /c` (like using caret escapes `^` or specific Unicode homoglyphs) might bypass the literal blacklist while still being interpreted as control characters by the shell.
- **Remediation**: Transition from `cmd /c <string>` to direct execution `exec.Command(binary, args...)` for all allowed commands to avoid the shell parser entirely.

## 4. Observations
- **Allowlist Integrity**: The `allowedShellCommands` and `AllowedPowerShellWorkflows` lists are well-defined and follow the principle of least privilege.
- **Sanitization**: `normalizeWhitespace` (Line 75) is a clever defense against bypasses using tabs/form-feeds.
