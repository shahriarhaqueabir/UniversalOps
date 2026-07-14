import os
import time
import pytest
from pywinauto import Application

# Import config
import sys
sys.path.insert(0, os.path.dirname(__file__))
from config import APP_PATH, APP_TITLE, LAUNCH_TIMEOUT, ACTION_TIMEOUT, ARTIFACT_DIR, MAIN_TABS, NETOPS_SUBTABS, SECOPS_SUBTABS, SYSOP_SUBTABS, LOGS_SUBTABS, DEVOPS_SUBTABS, AIOPS_SUBTABS

# Import page objects
from pages import MainWindow, NetOpsPage, SecOpsPage, SysOpsPage, LogsPage, DevOpsPage, AIOpsPage


# ──────────────────────────────────────────────────────────────────────
# Fixtures
# ──────────────────────────────────────────────────────────────────────

@pytest.fixture(scope="function")
def app(request):
    """Launch fresh app instance per test with isolated user data dirs."""
    if not APP_PATH:
        pytest.exit("APP_PATH environment variable not set", returncode=1)
    if not APP_TITLE:
        pytest.exit("APP_TITLE environment variable not set", returncode=1)

    # Create isolated environment per test
    import tempfile
    import shutil
    tmp_dir = tempfile.mkdtemp(prefix="opsforall_e2e_")
    sandbox_env = os.environ.copy()
    sandbox_env["QT_ACCESSIBILITY"] = "1"
    sandbox_env["APPDATA"] = os.path.join(tmp_dir, "AppData", "Roaming")
    sandbox_env["LOCALAPPDATA"] = os.path.join(tmp_dir, "AppData", "Local")
    sandbox_env["TEMP"] = sandbox_env["TMP"] = os.path.join(tmp_dir, "Temp")
    for p in (sandbox_env["APPDATA"], sandbox_env["LOCALAPPDATA"], sandbox_env["TEMP"]):
        os.makedirs(p, exist_ok=True)

    # Launch via subprocess so we can pass custom env
    import subprocess
    import shlex
    proc = subprocess.Popen(
        [APP_PATH] + shlex.split(os.environ.get("APP_ARGS", "")),
        env=sandbox_env,
    )

    # Connect pywinauto by PID
    pw_app = Application(backend="uia").connect(process=proc.pid, timeout=LAUNCH_TIMEOUT)
    win = pw_app.window(title=APP_TITLE)
    win.wait("visible", timeout=LAUNCH_TIMEOUT)
    win.set_focus()

    yield win

    # Teardown
    try:
        win.close()
        proc.wait(timeout=5)
    except Exception:
        proc.kill()
    shutil.rmtree(tmp_dir, ignore_errors=True)


# ──────────────────────────────────────────────────────────────────────
# Main Window Tab Tests
# ──────────────────────────────────────────────────────────────────────

class TestMainTabs:
    """Verify all top-level tabs exist and are clickable."""

    def test_all_main_tabs_accessible(self, app):
        """All main navigation tabs should be present and clickable."""
        main = MainWindow(app)
        for tab_key, auto_id in MAIN_TABS.items():
            tab = main.get_main_tab(tab_key)
            tab.wait("visible", timeout=5)
            assert tab.exists(), f"Main tab '{tab_key}' not found"
            # Click to verify it's interactive
            tab.click_input()
            time.sleep(0.2)
            assert tab.is_selected(), f"Main tab '{tab_key}' not selected after click"


# ──────────────────────────────────────────────────────────────────────
# NetOps Sub-tab Tests
# ──────────────────────────────────────────────────────────────────────

class TestNetOpsSubtabs:
    """Verify NetOps sub-tabs (Overview, Probes, Resolution, etc.)."""

    def test_netops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        netops = main.click_main_tab("netops")
        time.sleep(0.5)  # Let React render

        for subtab_key, auto_id in NETOPS_SUBTABS.items():
            tab = netops.get_subtab(subtab_key)
            tab.wait("visible", timeout=5)
            assert tab.exists(), f"NetOps sub-tab '{subtab_key}' not found"

            # Click and verify content loads
            netops.click_subtab(subtab_key)
            time.sleep(0.3)
            assert tab.is_selected(), f"NetOps sub-tab '{subtab_key}' not selected after click"


# ──────────────────────────────────────────────────────────────────────
# SecOps Sub-tab Tests
# ──────────────────────────────────────────────────────────────────────

class TestSecOpsSubtabs:
    """Verify SecOps sub-tabs (Firewall, Users, Listening, Defender, Events)."""

    def test_secops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        secops = main.click_main_tab("secops")
        time.sleep(0.5)

        for subtab_key, auto_id in SECOPS_SUBTABS.items():
            tab = secops.get_subtab(subtab_key)
            tab.wait("visible", timeout=5)
            assert tab.exists(), f"SecOps sub-tab '{subtab_key}' not found"

            secops.click_subtab(subtab_key)
            time.sleep(0.3)
            assert tab.is_selected(), f"SecOps sub-tab '{subtab_key}' not selected after click"


# ──────────────────────────────────────────────────────────────────────
# SysOps Sub-tab Tests
# ──────────────────────────────────────────────────────────────────────

class TestSysOpsSubtabs:
    """Verify SysOps sub-tabs (Analysis, Runtime, Inventory)."""

    def test_sysops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        sysops = main.click_main_tab("sysops")
        time.sleep(0.5)

        for subtab_key, auto_id in SYSOP_SUBTABS.items():
            tab = sysops.get_subtab(subtab_key)
            tab.wait("visible", timeout=5)
            assert tab.exists(), f"SysOps sub-tab '{subtab_key}' not found"

            sysops.click_subtab(subtab_key)
            time.sleep(0.3)
            assert tab.is_selected(), f"SysOps sub-tab '{subtab_key}' not selected after click"


# ──────────────────────────────────────────────────────────────────────
# Logs Sub-tab Tests
# ──────────────────────────────────────────────────────────────────────

class TestLogsSubtabs:
    """Verify Logs sub-tabs (Overview, Live Stream)."""

    def test_logs_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        logs = main.click_main_tab("logs")
        time.sleep(0.5)

        for subtab_key, auto_id in LOGS_SUBTABS.items():
            tab = logs.get_subtab(subtab_key)
            tab.wait("visible", timeout=5)
            assert tab.exists(), f"Logs sub-tab '{subtab_key}' not found"

            logs.click_subtab(subtab_key)
            time.sleep(0.3)
            assert tab.is_selected(), f"Logs sub-tab '{subtab_key}' not selected after click"


# ──────────────────────────────────────────────────────────────────────
# DevOps Sub-tab Tests (Radix UI)
# ──────────────────────────────────────────────────────────────────────

class TestDevOpsSubtabs:
    """Verify DevOps sub-tabs (Terminal, PowerShell, Services, etc.)."""

    def test_devops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        devops = main.click_main_tab("devops")
        time.sleep(0.5)

        for subtab_key, auto_id in DEVOPS_SUBTABS.items():
            tab = devops.get_subtab(subtab_key)
            tab.wait("visible", timeout=5)
            assert tab.exists(), f"DevOps sub-tab '{subtab_key}' not found"

            devops.click_subtab(subtab_key)
            time.sleep(0.3)
            assert tab.is_selected(), f"DevOps sub-tab '{subtab_key}' not selected after click"


# ──────────────────────────────────────────────────────────────────────
# AIOps Sub-tab Tests (Radix UI)
# ──────────────────────────────────────────────────────────────────────

class TestAIOpsSubtabs:
    """Verify AIOps sub-tabs (Analyst Chat, Anomaly Detection, AI Insights)."""

    def test_aiops_subtabs_exist_and_clickable(self, app):
        main = MainWindow(app)
        aiops = main.click_main_tab("aiops")
        time.sleep(0.5)

        for subtab_key, auto_id in AIOPS_SUBTABS.items():
            tab = aiops.get_subtab(subtab_key)
            tab.wait("visible", timeout=5)
            assert tab.exists(), f"AIOps sub-tab '{subtab_key}' not found"

            aiops.click_subtab(subtab_key)
            time.sleep(0.3)
            assert tab.is_selected(), f"AIOps sub-tab '{subtab_key}' not selected after click"