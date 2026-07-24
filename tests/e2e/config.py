"""E2E test configuration — auto_id strings and timeouts."""

import os

# ── Application paths ──

APP_PATH = os.environ.get(
    'APP_PATH',
    'E:\\Projects\\projectx\\UniversalOps\\build\\bin\\universal-ops.exe',
)
APP_TITLE = os.environ.get('APP_TITLE', 'Universal-Ops Operations Platform')

# ── Timeouts (seconds) ──

LAUNCH_TIMEOUT = int(os.environ.get('LAUNCH_TIMEOUT', '45'))
ACTION_TIMEOUT = int(os.environ.get('ACTION_TIMEOUT', '15'))

# ── Artifact directory ──

ARTIFACT_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'artifacts')

# ── Main tab auto-id map ──

MAIN_TABS: dict[str, str] = {
    'dashboard': 'main-tab-dashboard',
    'sysops':    'main-tab-sysops',
    'netops':    'main-tab-netops',
    'secops':    'main-tab-secops',
    'devops':    'main-tab-devops',
    'aiops':     'main-tab-aiops',
    'logs':      'main-tab-logs',
    'settings':  'main-tab-settings',
}

# ── Sub-tab auto-id maps ──

NETOPS_SUBTABS: dict[str, str] = {
    'overview':     'netops-tab-overview',
    'ping':         'netops-tab-ping',
    'dns':          'netops-tab-dns',
    'connections':  'netops-tab-connections',
    'interfaces':   'netops-tab-interfaces',
    'traceroute':   'netops-tab-traceroute',
    'portscan':     'netops-tab-portscan',
    'bandwidth':    'netops-tab-bandwidth',
}

SECOPS_SUBTABS: dict[str, str] = {
    'firewall':   'secops-tab-firewall',
    'users':      'secops-tab-users',
    'listening':  'secops-tab-listening',
    'defender':   'secops-tab-defender',
    'events':     'secops-tab-events',
}

SYSOP_SUBTABS: dict[str, str] = {
    'overview':     'sysops-tab-overview',
    'processes':    'sysops-tab-processes',
    'system-info':  'sysops-tab-system-info',
}

LOGS_SUBTABS: dict[str, str] = {
    'overview':  'logs-tab-overview',
    'live':      'logs-tab-live',
}

DEVOPS_SUBTABS: dict[str, str] = {
    'overview':         'devops-tab-overview',
    'terminal':         'devops-tab-terminal',
    'powershell-pro':   'devops-tab-powershell-pro',
    'services':         'devops-tab-services',
    'containers':       'devops-tab-containers',
    'git':              'devops-tab-git',
    'servers':          'devops-tab-servers',
    'environment':      'devops-tab-environment',
    'file-browser':     'devops-tab-file-browser',
    'toolbox':          'devops-tab-toolbox',
}

AIOPS_SUBTABS: dict[str, str] = {
    'ai-chat':     'aiops-tab-ai-chat',
    'anomalies':   'aiops-tab-anomalies',
    'insights':    'aiops-tab-insights',
}
