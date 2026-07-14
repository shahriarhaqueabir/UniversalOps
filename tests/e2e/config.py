import os

# ──────────────────────────────────────────────────────────────────────
# Test Configuration
# ──────────────────────────────────────────────────────────────────────

# Set via environment or edit directly
APP_PATH = os.environ.get(
    "APP_PATH",
    r"E:\Projects\projectx\AllOpsFull\build\bin\OpsForAll.exe"
)

APP_TITLE = os.environ.get("APP_TITLE", "OpsForAll")

LAUNCH_TIMEOUT = int(os.environ.get("LAUNCH_TIMEOUT", "15"))
ACTION_TIMEOUT = int(os.environ.get("ACTION_TIMEOUT", "10"))
ARTIFACT_DIR = os.path.join(os.path.dirname(__file__), "artifacts")

# ──────────────────────────────────────────────────────────────────────
# UIA Automation IDs for tabs (must match frontend data-automation-id)
# ──────────────────────────────────────────────────────────────────────

# Main navigation tabs
MAIN_TABS = {
    "dashboard": "main-tab-dashboard",
    "sysops": "main-tab-sysops",
    "netops": "main-tab-netops",
    "secops": "main-tab-secops",
    "devops": "main-tab-devops",
    "aiops": "main-tab-aiops",
    "logs": "main-tab-logs",
    "settings": "main-tab-settings",
}

# NetOps sub-tabs
NETOPS_SUBTABS = {
    "overview": "netops-tab-overview",
    "ping": "netops-tab-ping",
    "dns": "netops-tab-dns",
    "connections": "netops-tab-connections",
    "interfaces": "netops-tab-interfaces",
    "traceroute": "netops-tab-traceroute",
    "portscan": "netops-tab-portscan",
    "bandwidth": "netops-tab-bandwidth",
}

# SecOps sub-tabs
SECOPS_SUBTABS = {
    "firewall": "secops-tab-firewall",
    "users": "secops-tab-users",
    "listening": "secops-tab-listening",
    "defender": "secops-tab-defender",
    "events": "secops-tab-events",
}

# SysOps sub-tabs
SYSOP_SUBTABS = {
    "overview": "sysops-tab-overview",
    "processes": "sysops-tab-processes",
    "system-info": "sysops-tab-system-info",
}

# Logs sub-tabs
LOGS_SUBTABS = {
    "overview": "logs-tab-overview",
    "live": "logs-tab-live",
}

# DevOps sub-tabs
DEVOPS_SUBTABS = {
    "overview": "devops-tab-overview",
    "terminal": "devops-tab-terminal",
    "powershell-pro": "devops-tab-powershell-pro",
    "services": "devops-tab-services",
    "containers": "devops-tab-containers",
    "git": "devops-tab-git",
    "servers": "devops-tab-servers",
    "environment": "devops-tab-environment",
    "file-browser": "devops-tab-file-browser",
    "toolbox": "devops-tab-toolbox",
}

# AIOps sub-tabs
AIOPS_SUBTABS = {
    "ai-chat": "aiops-tab-ai-chat",
    "anomalies": "aiops-tab-anomalies",
    "insights": "aiops-tab-insights",
}