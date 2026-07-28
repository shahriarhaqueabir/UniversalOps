# UniversalOps — Production Runbook

> Operational procedures for building, deploying, and troubleshooting UniversalOps in production.

---

## Table of Contents
1. [Build Pipeline](#1-build-pipeline)
2. [Release Process](#2-release-process)
3. [Installation Verification](#3-installation-verification)
4. [Logging & Diagnostics](#4-logging--diagnostics)
5. [Database Maintenance](#5-database-maintenance)
6. [Performance Tuning](#6-performance-tuning)
7. [Troubleshooting](#7-troubleshooting)

---

## 1. Build Pipeline

### Local Build
```powershell
# Development build with dev tools
wails dev

# Production build (optimized, no dev tools)
wails build

# Build for specific platform
wails build -platform windows/amd64
wails build -platform linux/amd64
wails build -platform darwin/amd64
```

### CI Build (GitHub Actions)
The CI pipeline runs on every push to `main` and `develop`:
1. **Test** (3 OS matrix): `go test ./internal/...` + `npx vitest run`
2. **Lint**: `golangci-lint` + ESLint + TypeScript check
3. **Build**: `wails build` on Ubuntu (cross-compile for Windows/Linux/macOS)

### Prerequisites for Build Machine
- Go 1.26+
- Node.js 22+
- Wails CLI v2
- GCC (MinGW on Windows, build-essential on Linux, Xcode CLT on macOS)
- WebKit GTK dev headers (Linux only)

---

## 2. Release Process

### Creating a Release
1. Ensure all tests pass on `main`:
   ```bash
   go test ./internal/... -count=1 -timeout 120s
   cd cmd/opsforall-gui/frontend && npx vitest run
   ```

2. Update version in `wails.json` and `README.md` badge.

3. Tag the release:
   ```bash
   git tag -a v1.4.0 -m "v1.4.0: [brief description]"
   git push origin v1.4.0
   ```

4. The `release.yml` workflow automatically:
   - Builds the binary for Windows (amd64)
   - Creates a GitHub Release with the binary attached
   - Generates release notes from commit history

### Release Artifacts
| Artifact | Description |
|----------|-------------|
| `universal-ops-v1.x.x-windows-amd64.exe` | Portable executable |
| `universal-ops-v1.x.x-windows-amd64-installer.exe` | NSIS installer (future) |

### Post-Release
- Verify the release page shows the new version
- Test the PowerShell one-liner install from a clean machine
- Update `install.ps1` checksum if applicable

---

## 3. Installation Verification

### Quick Health Check
```powershell
# After installation, verify the app starts
.\universal-ops-*-windows-amd64.exe --version

# Check that the database initializes
Test-Path "$env:LOCALAPPDATA\UniversalOps\universalops.db"
```

### Logs Location
```
Windows: %LOCALAPPDATA%\UniversalOps\logs\
Linux:   ~/.local/share/UniversalOps/logs/
macOS:   ~/Library/Logs/UniversalOps/
```

### Data Directory
```
Windows: %LOCALAPPDATA%\UniversalOps\
Linux:   ~/.local/share/UniversalOps/
macOS:   ~/Library/Application Support/UniversalOps/
```

---

## 4. Logging & Diagnostics

### Log Levels
- `INFO`: Normal operations (startup, shutdown, periodic metrics)
- `WARN`: Recoverable issues (collector timeout, API fallback)
- `ERROR`: Failures requiring attention (database corruption, collector crash)

### Viewing Logs
```powershell
# Tail the latest log file
Get-Content "$env:LOCALAPPDATA\UniversalOps\logs\universalops.log" -Tail 50 -Wait

# Search for errors
Select-String -Path "$env:LOCALAPPDATA\UniversalOps\logs\*.log" -Pattern "ERROR"
```

### Diagnostic Commands
```bash
# Check database integrity
sqlite3 universalops.db "PRAGMA integrity_check;"

# Check WAL status
sqlite3 universalops.db "PRAGMA wal_checkpoint;"

# Count alerts
sqlite3 universalops.db "SELECT level, COUNT(*) FROM alerts GROUP BY level;"
```

---

## 5. Database Maintenance

### Backup
```powershell
# Stop UniversalOps first, then:
Copy-Item "$env:LOCALAPPDATA\UniversalOps\universalops.db" ".\backup-$(Get-Date -Format 'yyyyMMdd').db"
```

### Vacuum (reclaim space)
```powershell
# Stop UniversalOps first, then:
sqlite3 "$env:LOCALAPPDATA\UniversalOps\universalops.db" "VACUUM;"
```

### WAL Checkpoint
The application runs `PRAGMA wal_checkpoint(TRUNCATE)` periodically. To force manually:
```powershell
sqlite3 "$env:LOCALAPPDATA\UniversalOps\universalops.db" "PRAGMA wal_checkpoint(TRUNCATE);"
```

### Reset (nuclear option)
```powershell
# Deletes all data — use only as last resort
Remove-Item "$env:LOCALAPPDATA\UniversalOps\universalops.db"
Remove-Item "$env:LOCALAPPDATA\UniversalOps\universalops.db-wal"
Remove-Item "$env:LOCALAPPDATA\UniversalOps\universalops.db-shm"
```

---

## 6. Performance Tuning

### Collection Interval
Default: 3 seconds. Adjust via environment variable:
```powershell
$env:UNIVERSALOPS_COLLECT_INTERVAL = "5"  # seconds
```

### SQLite WAL Mode
WAL mode is enabled by default. If you experience high disk I/O:
```powershell
# Increase checkpoint interval (reduces write frequency)
$env:UNIVERSALOPS_WAL_CHECKPOINT_INTERVAL = "300"  # seconds
```

### Ollama Model Selection
For slower machines, use a smaller model:
```powershell
$env:UNIVERSALOPS_AI_MODEL = "llama3.2:3b"  # instead of default universalops
```

---

## 7. Troubleshooting

### App won't start
| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| "WebView2 not found" | Missing WebView2 runtime | Install from [Microsoft](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) |
| "DLL load failed" | Missing VC++ redistributable | Install [VC++ Redist](https://aka.ms/vs/17/release/vc_redist.x64.exe) |
| "Database locked" | Previous crash left WAL in bad state | Delete `universalops.db-wal` and `universalops.db-shm` |
| Blank white window | GPU driver issue | Update graphics drivers, or disable hardware acceleration |

### AIOps not working
| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| "Connection refused" | Ollama not running | Start Ollama: `ollama serve` |
| "Model not found" | `universalops` model not created | Run: `ollama create universalops -f Modelfile` |
| Slow responses | Large model on CPU | Use a smaller model: `ollama pull llama3.2:3b` |

### High CPU / Memory
- Reduce collection interval (see Performance Tuning)
- Check for runaway processes in Processes tab
- Disable unused collectors in Settings
- Restart the application to clear accumulated state

### Data loss after update
- Check `%LOCALAPPDATA%\UniversalOps\` for backup files
- The database schema migrates automatically — if migration fails, the app logs the error and falls back
- Report migration failures as GitHub issues with the log file attached