"""E2E test configuration for the current Universal-Ops navigation model."""

import os

APP_PATH = os.environ.get(
    'APP_PATH',
    os.path.join(
        os.path.dirname(os.path.abspath(__file__)),
        '..', '..', 'build', 'bin', 'universal-ops.exe',
    ),
)
APP_TITLE = os.environ.get('APP_TITLE', 'Universal-Ops Operations Platform')
LAUNCH_TIMEOUT = int(os.environ.get('LAUNCH_TIMEOUT', '45'))
ACTION_TIMEOUT = int(os.environ.get('ACTION_TIMEOUT', '15'))
ARTIFACT_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'artifacts')

MAIN_TABS = {
    'dashboard': 'main-tab-dashboard', 'sysops': 'main-tab-sysops',
    'workflows': 'main-tab-workflows', 'netops': 'main-tab-netops',
    'secops': 'main-tab-secops', 'devops': 'main-tab-devops',
    'aiops': 'main-tab-aiops', 'reports': 'main-tab-reports',
    'alerts': 'main-tab-alerts', 'logs': 'main-tab-logs',
    'settings': 'main-tab-settings',
}

MAIN_TAB_LABELS = {
    'dashboard': 'Dashboard', 'sysops': 'System Ops', 'workflows': 'Workflow Library',
    'netops': 'Network Ops', 'secops': 'Security Ops', 'devops': 'DevOps',
    'aiops': 'AI Ops', 'reports': 'Reports', 'alerts': 'Alerts',
    'logs': 'Logs', 'settings': 'Settings',
}

NETOPS_SUBTABS = {
    key: f'netops-tab-{key}' for key in (
        'overview', 'connections', 'interfaces', 'arp', 'ping', 'dns',
        'traceroute', 'portscan', 'bandwidth', 'dns-advanced', 'multi-ping',
        'health', 'vpn', 'discovery', 'actions',
    )
}
NETOPS_SUBTAB_LABELS = {
    'overview': 'Overview', 'connections': 'Connections', 'interfaces': 'Interfaces',
    'arp': 'ARP Table', 'ping': 'Ping', 'dns': 'DNS', 'traceroute': 'Traceroute',
    'portscan': 'Port Scan', 'bandwidth': 'Bandwidth', 'dns-advanced': 'DNS Advanced',
    'multi-ping': 'Multi-Ping', 'health': 'Health Check', 'vpn': 'VPN',
    'discovery': 'Discovery', 'actions': 'Actions',
}

SECOPS_SUBTABS = {
    key: f'secops-tab-{key}' for key in (
        'overview', 'identity', 'perimeter', 'endpoint', 'events', 'hardening',
        'audit', 'response',
    )
}
SECOPS_SUBTAB_LABELS = {
    'overview': 'Overview', 'identity': 'Identity & Access', 'perimeter': 'Perimeter Security',
    'endpoint': 'Endpoint Security', 'events': 'Log & Events', 'hardening': 'Security Hardening',
    'audit': 'Security Audit', 'response': 'Incident Response',
}

SYSOP_SUBTABS = {
    key: f'sysops-tab-{key}' for key in (
        'system-info', 'hardware', 'cpu', 'memory', 'disk', 'packages', 'processes',
        'services', 'scheduler', 'logs', 'users', 'diagnostics', 'actions',
    )
}
SYSOP_SUBTAB_LABELS = {
    'system-info': 'System Info', 'hardware': 'Hardware', 'cpu': 'CPU', 'memory': 'Memory',
    'disk': 'Disk', 'packages': 'Installed Apps', 'processes': 'Processes',
    'services': 'Services', 'scheduler': 'Scheduler', 'logs': 'Logs', 'users': 'Users',
    'diagnostics': 'Diagnostics', 'actions': 'Actions',
}

LOGS_SUBTABS = {key: f'logs-tab-{key}' for key in ('overview', 'live', 'audit')}
LOGS_SUBTAB_LABELS = {'overview': 'Overview', 'live': 'Live Stream', 'audit': 'Audit'}

DEVOPS_SUBTABS = {
    key: f'devops-tab-{key}' for key in (
        'overview', 'powershell', 'bash', 'docker', 'kubernetes', 'diagnostics',
        'services', 'servers', 'environment',
    )
}
DEVOPS_SUBTAB_LABELS = {
    'overview': 'Overview', 'powershell': 'PS', 'bash': 'Bash', 'docker': 'Docker',
    'kubernetes': 'K8s', 'diagnostics': 'Health', 'services': 'Services',
    'servers': 'Servers', 'environment': 'Env',
}

AIOPS_SUBTABS = {key: f'aiops-tab-{key}' for key in ('ai-chat', 'anomalies', 'insights')}
AIOPS_SUBTAB_LABELS = {'ai-chat': 'Analyst Chat', 'anomalies': 'Anomaly Detection', 'insights': 'AI Insights'}
