# SecOps Module Rework — Design Spec

**Date**: 2026-07-14
**Status**: Approved
**Scope**: Core 8 categories (specialized 6 deferred)

## Goal

Transform SecOps from a 5-tab monolith into a sidebar-navigated security operations center answering:

> "Am I secure, and if not, what should I fix first?"

**Workflow**: Assess → Detect → Respond

## Key Decisions

- **Built-in only**: Use OS-native commands (netsh, PowerShell, ss, iptables, /etc/*, Get-WinEvent). No external tool dependencies.
- **Self-contained tabs**: Each tab owns its own `useQuery` hooks. No prop drilling from parent.
- **Core 8 first**: Overview, Identity & Access, Network Security, Endpoint Security, Log & Event Analysis, Security Hardening, Security Audit, Incident Response.
- **Deferred**: Vulnerability Management, File Integrity Monitoring, Compliance, Secrets & Certificates, Patch Management, Threat Detection.

## Sidebar Structure

```
ASSESSMENT
  ├── Overview
  ├── Identity & Access
  ├── Network Security
  └── Endpoint Security

DETECTION
  ├── Log & Event Analysis
  └── Security Hardening

RESPONSE
  ├── Security Audit
  └── Incident Response
```

## Existing Backend (11 methods, no changes needed for Phase 1)

| Method | File | Used By |
|--------|------|---------|
| `GetFirewallRules()` | firewall.go | Network Security, Overview |
| `GetFirewallProfiles()` | firewall.go | Network Security |
| `SetFirewallRuleState()` | firewall.go | Network Security |
| `GetUsers()` | users.go | Identity & Access |
| `GetGroups()` | users.go | Identity & Access |
| `GetListeningPorts()` | listening.go | Network Security, Security Audit |
| `GetDefenderStatus()` | defender.go | Endpoint Security |
| `GetScheduledTasks()` | tasks.go | Endpoint Security, Security Audit |
| `GetSecurityEvents()` | events.go | Log & Event Analysis, Overview |
| `GetSecurityScore()` | workflows.go | Overview |
| `GetRisks()` | workflows.go | Overview |
| `GetSecuritySummary()` | workflows.go | Overview |
| `GetFirewallStatus()` | secops.go (binding) | Overview, Network Security, Security Audit |

## New Backend Methods (15 total)

### `security.go` (new file)

| Method | Platform | Implementation |
|--------|----------|---------------|
| `GetPasswordPolicy()` | Win | `net accounts` parsing — max age, min length, lockout threshold, complexity |
| | Linux | `/etc/login.defs` + `pam_pwquality.conf` parsing |
| `GetFailedLogins()` | Win | `Get-WinEvent -FilterHashtable @{Id=4625}` — last 50 events |
| | Linux | `/var/log/auth.log` grep "Failed password" + `lastb` |
| `GetAccountLockouts()` | Win | `Search-ADAccount -LockedOut` (local: `net user` + SAM lookup) |
| | Linux | `pam_tally2 --user` or `/var/log/auth.log` "account locked" |

### `endpoint.go` (new file)

| Method | Platform | Implementation |
|--------|----------|---------------|
| `GetDiskEncryptionStatus()` | Win | `Get-BitLockerVolume` — mount point, protection status, encryption method |
| | Linux | `lsblk --discard` + `dmsetup status` (LUKS) + `cryptsetup status` |
| `GetSecureBootStatus()` | Win | `Confirm-SecureBootUEFI` |
| | Linux | `/sys/firmware/efi` existence check + `mokutil --sb-state` |
| `GetRunningServices()` | Win | `Get-Service` — name, status, startup type, display name |
| | Linux | `systemctl list-units --type=service --all` + `/etc/init.d/` |

### `network.go` (new file)

| Method | Platform | Implementation |
|--------|----------|---------------|
| `GetTLSCertificates()` | Win | `Get-ChildItem Cert:\LocalMachine\My` — subject, issuer, notAfter, key size |
| | Linux | `/etc/ssl/certs/` + `openssl x509 -enddate -noout` for each cert |
| `GetPublicExposure()` | Win | `Get-ListeningPorts()` filtered by `is_external == true` |
| | Linux | `ss -tulnp` filtered by non-loopback, non-127.0.0.1 |

### `hardening.go` (new file)

| Method | Platform | Implementation |
|--------|----------|---------------|
| `GetHardeningChecks()` | Win | Aggregated: firewall on, Defender on, UAC enabled, auto-update on, guest disabled, remote desktop off, SMBv1 disabled |
| | Linux | Aggregated: firewall active, SSH root no, password auth, no world-writable in /etc, UFW/status, auto-update config |
| `GetSSHConfig()` | Linux only | Parse `/etc/ssh/sshd_config` — PermitRootLogin, PasswordAuthentication, PubkeyAuthentication, X11Forwarding, Protocol, MaxAuthTries |

### `events.go` (extend existing)

| Method | Platform | Implementation |
|--------|----------|---------------|
| `GetPrivilegeEvents()` | Win | `Get-WinEvent -FilterHashtable @{Id=4672,4673,4674}` — privilege use events |
| | Linux | `journalctl` or `/var/log/auth.log` — "sudo" entries |
| `GetEventTimeline()` | Win | Merge security events (4625, 4720, 4726, 4732, 4733, 4740, 1102, 4672) into chronological timeline |
| | Linux | Merge auth.log + journalctl security events |

### `audit.go` (new file)

| Method | Platform | Implementation |
|--------|----------|---------------|
| `RunSecurityAuditChecklist()` | Both | Aggregates: firewall status, secure boot, disk encryption, exposed RDP/SSH, password policy, certificates, scheduled tasks anomaly. Returns `[]AuditCheckItem` with pass/fail/remediation. |

### `response.go` (new file)

| Method | Platform | Implementation |
|--------|----------|---------------|
| `IsolateHost()` | Win | `netsh advfirewall set allprofiles state on` + add block-all inbound rule (allow only Wails port) |
| | Linux | `iptables -A INPUT -j DROP` + `iptables -A INPUT -p tcp --dport <wails-port> -j ACCEPT` |
| `KillProcess(pid)` | Both | Win: `Stop-Process -Id $pid -Force`; Linux: `kill -9 $pid` |
| `BlockIP(ip)` | Win | `netsh advfirewall firewall add rule name="Block $ip" dir=in action=block remoteip=$ip` |
| | Linux | `iptables -A INPUT -s $ip -j DROP` |
| `DisableAccount(username)` | Win | `net user $username /active:no` |
| | Linux | `passwd -l $username` |
| `CaptureEvidence()` | Both | Collect: `netstat -ano`/`ss -tulnp`, `tasklist`/`ps aux`, recent event logs, running services → zip |
| `ExportForensicBundle()` | Both | Call `CaptureEvidence()` + export to user-selected directory |

## Binding Types (new, in `internal/app/Types.go`)

```go
type PasswordPolicy struct {
    MaxAge       int  `json:"max_age"`
    MinLength    int  `json:"min_length"`
    Complexity   bool `json:"complexity"`
    LockoutThreshold int `json:"lockout_threshold"`
    LockoutDuration  int `json:"lockout_duration"`
}

type FailedLogin struct {
    Time      string `json:"time"`
    Username  string `json:"username"`
    SourceIP  string `json:"source_ip"`
    Count     int    `json:"count"`
}

type LockedAccount struct {
    Username string `json:"username"`
    LockedSince string `json:"locked_since"`
}

type DiskEncryption struct {
    Volume    string `json:"volume"`
    Encrypted bool   `json:"encrypted"`
    Method    string `json:"method"`
    Status    string `json:"status"`
}

type SecureBoot struct {
    Enabled bool   `json:"enabled"`
    State   string `json:"state"`
}

type SystemService struct {
    Name        string `json:"name"`
    DisplayName string `json:"display_name"`
    Status      string `json:"status"`
    StartupType string `json:"startup_type"`
}

type TLSCertificate struct {
    Subject    string `json:"subject"`
    Issuer     string `json:"issuer"`
    NotAfter   string `json:"not_after"`
    KeySize    int    `json:"key_size"`
    IsExpiring bool   `json:"is_expiring"`
    DaysLeft   int    `json:"days_left"`
}

type SSHConfig struct {
    PermitRootLogin       string `json:"permit_root_login"`
    PasswordAuthentication string `json:"password_authentication"`
    PubkeyAuthentication  string `json:"pubkey_authentication"`
    X11Forwarding         string `json:"x11_forwarding"`
    MaxAuthTries          string `json:"max_auth_tries"`
}

type HardeningCheck struct {
    Category     string `json:"category"`
    Check        string `json:"check"`
    Passed       bool   `json:"passed"`
    Severity     string `json:"severity"`
    Remediation  string `json:"remediation"`
}

type AuditCheckItem struct {
    Category    string `json:"category"`
    Check       string `json:"check"`
    Passed      bool   `json:"passed"`
    Description string `json:"description"`
    Remediation string `json:"remediation"`
}

type SecurityAuditResult struct {
    Score     int              `json:"score"`
    Total     int              `json:"total"`
    Passed    int              `json:"passed"`
    Failed    int              `json:"failed"`
    Items     []AuditCheckItem `json:"items"`
    Timestamp string           `json:"timestamp"`
}
```

## Frontend Tab Design

### 1. Overview
- **Source**: `secops/OverviewTab.tsx`
- **Calls**: `GetSecurityScore`, `GetRisks`, `GetSecuritySummary`, `GetFirewallStatus`, `GetSecurityEvents` (last 5)
- **Sections**: Score donut, top risks, recommendations, recent events, firewall status
- **No new backend methods**

### 2. Identity & Access
- **Source**: `secops/IdentityTab.tsx`
- **Calls**: `GetUsers`, `GetGroups`, `GetPasswordPolicy`, `GetFailedLogins`, `GetAccountLockouts`
- **Sections**: User card grid, group badges, password policy status, failed login table, lockout list

### 3. Network Security
- **Source**: `secops/NetworkSecurityTab.tsx`
- **Calls**: `GetListeningPorts`, `GetFirewallRules`, `GetFirewallStatus`, `GetTLSCertificates`, `GetPublicExposure`
- **Sections**: External ports highlight, firewall rules table, TLS cert expiry list, exposure summary

### 4. Endpoint Security
- **Source**: `secops/EndpointTab.tsx`
- **Calls**: `GetDefenderStatus`, `GetDiskEncryptionStatus`, `GetSecureBootStatus`, `GetRunningServices`, `GetScheduledTasks`
- **Sections**: Defender status cards, encryption status, secure boot, service list, scheduled tasks

### 5. Log & Event Analysis
- **Source**: `secops/EventsTab.tsx`
- **Calls**: `GetSecurityEvents`, `GetPrivilegeEvents`, `GetEventTimeline`, `GetFailedLogins`
- **Sections**: Category filter buttons, event table, privilege escalation events, timeline view

### 6. Security Hardening
- **Source**: `secops/HardeningTab.tsx`
- **Calls**: `GetHardeningChecks`, `GetSSHConfig` (Linux only)
- **Sections**: Pass/fail check cards grouped by category, SSH config display, remediation actions

### 7. Security Audit
- **Source**: `secops/AuditTab.tsx`
- **Calls**: `RunSecurityAuditChecklist`
- **Sections**: Score display, 12-point checklist with pass/fail, remediation text per failed item, "Re-audit" button

### 8. Incident Response
- **Source**: `secops/ResponseTab.tsx`
- **Calls**: `IsolateHost`, `KillProcess`, `BlockIP`, `DisableAccount`, `CaptureEvidence`, `ExportForensicBundle`
- **Sections**: Action buttons with confirm dialogs, evidence capture, forensic bundle export

## Shared Components

`secops/components.tsx`:
- `SectionBriefing` — title + objective + checklist (reuse from NetOps pattern)
- `StatusBadge` — colored badge for pass/fail/warning/error states
- `MiniStat` — metric card with icon
- `SeverityDot` — colored dot for critical/high/medium/low

## TypeScript Types (new, in `types/index.ts`)

All new interfaces mirror Go binding types 1:1 with snake_case keys:
`PasswordPolicy`, `FailedLogin`, `LockedAccount`, `DiskEncryption`, `SecureBoot`, `SystemService`, `TLSCertificate`, `SSHConfig`, `HardeningCheck`, `AuditCheckItem`, `SecurityAuditResult`

## File Structure

```
cmd/opsforall-gui/frontend/src/pages/
├── SecOps.tsx                    (rewritten: sidebar navigator)
├── SecOps.test.tsx               (updated: new test cases)
└── secops/
    ├── components.tsx            (shared: SectionBriefing, StatusBadge, MiniStat, SeverityDot)
    ├── OverviewTab.tsx
    ├── IdentityTab.tsx
    ├── NetworkSecurityTab.tsx
    ├── EndpointTab.tsx
    ├── EventsTab.tsx
    ├── HardeningTab.tsx
    ├── AuditTab.tsx
    └── ResponseTab.tsx

internal/secops/
├── firewall.go           (existing, no changes)
├── users.go              (existing, no changes)
├── listening.go          (existing, no changes)
├── defender.go           (existing, no changes)
├── tasks.go              (existing, no changes)
├── events.go             (existing + extend: GetPrivilegeEvents, GetEventTimeline)
├── workflows.go          (existing, no changes)
├── security.go           (NEW: password policy, failed logins, lockouts)
├── endpoint.go           (NEW: disk encryption, secure boot, services)
├── network.go            (NEW: TLS certs, public exposure)
├── hardening.go          (NEW: hardening checks, SSH config)
├── audit.go              (NEW: security audit checklist)
├── response.go           (NEW: incident response actions)
└── secops_test.go        (existing + extend)

internal/app/
├── SecOps.go             (existing + extend: new bound methods)
└── Types.go              (existing + extend: new binding types)
```

## Testing Strategy

- **Frontend**: Update `SecOps.test.tsx` with mock data for new methods. Test sidebar navigation, each tab renders, key interactions.
- **Go**: Unit tests per new file. Cross-platform where possible, `t.Skip` for platform-specific.
- **Lint**: `npm run lint` + `npx tsc --noEmit` + `go vet ./...`

## Deferred (Phase 2)

| Category | Requires |
|----------|----------|
| Vulnerability Management | CVE database integration, package version checking |
| File Integrity Monitoring | Baseline hash storage, periodic comparison |
| Compliance | CIS/NIST/ISO benchmark definitions |
| Secrets & Certificates | SSH key enumeration, API key detection |
| Patch Management | OS update API integration |
| Threat Detection | Process behavior analysis, network anomaly detection |
