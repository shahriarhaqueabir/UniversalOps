"""E2E tests: tab navigation across all sections of Universal-Ops.

Tests use the conftest.py app() fixture (isolated data dirs per test).
"""

import os
import sys

import pytest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from config import (
    APP_PATH, APP_TITLE, LAUNCH_TIMEOUT, ACTION_TIMEOUT, ARTIFACT_DIR,
    MAIN_TABS, NETOPS_SUBTABS, SECOPS_SUBTABS, SYSOP_SUBTABS,
    LOGS_SUBTABS, DEVOPS_SUBTABS, AIOPS_SUBTABS,
)
from pages import (
    MainWindow, NetOpsPage, SecOpsPage, SysOpsPage,
    LogsPage, DevOpsPage, AIOpsPage, dismiss_onboarding,
)


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  Tests
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class TestMainTabs:
    """Verify all top-level tabs exist and are clickable."""

    def test_all_main_tabs_accessible(self, app):
        main = MainWindow(app)
        for tab_key in MAIN_TABS:
            main.click_main_tab(tab_key)


class TestNetOpsSubtabs:
    """Verify NetOps sub-tabs (Overview, Probes, Resolution, etc.)."""

    def test_netops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        main.click_main_tab("netops")
        netops = NetOpsPage(app)
        netops.verify_subtabs()


class TestSecOpsSubtabs:
    """Verify SecOps sub-tabs (Firewall, Users, Listening, Defender, Events)."""

    def test_secops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        main.click_main_tab("secops")
        secops = SecOpsPage(app)
        secops.verify_subtabs()


class TestSysOpsSubtabs:
    """Verify SysOps sub-tabs (Analysis, Runtime, Inventory)."""

    def test_sysops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        main.click_main_tab("sysops")
        sysops = SysOpsPage(app)
        sysops.verify_subtabs()


class TestLogsSubtabs:
    """Verify Logs sub-tabs (Overview, Live Stream)."""

    def test_logs_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        main.click_main_tab("logs")
        logs = LogsPage(app)
        logs.verify_subtabs()


class TestDevOpsSubtabs:
    """Verify DevOps sub-tabs (Terminal, PowerShell, Services, etc.)."""

    def test_devops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        main.click_main_tab("devops")
        devops = DevOpsPage(app)
        devops.verify_subtabs()


class TestAIOpsSubtabs:
    """Verify AIOps sub-tabs (Analyst Chat, Anomaly Detection, AI Insights)."""

    def test_aiops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        main.click_main_tab("aiops")
        aiops = AIOpsPage(app)
        aiops.verify_subtabs()
