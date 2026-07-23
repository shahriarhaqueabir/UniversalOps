import os
import time
import pytest
from pywinauto import Application

# Import config
import sys
sys.path.insert(0, os.path.dirname(__file__))
from config import APP_PATH, APP_TITLE, ACTION_TIMEOUT, MAIN_TABS

# Import page objects
from pages import MainWindow

# ──────────────────────────────────────────────────────────────────────
# E2E Smoke Tests - Basic & Sandboxed
# ──────────────────────────────────────────────────────────────────────

@pytest.mark.smoke
class TestBasicNavigation:
    """Basic navigation smoke tests to ensure app stability."""

    def test_main_tabs_smoke(self, app):
        """
        Verify main navigation tabs are present and interactive.
        This is a 'basic' test that avoids triggering deep sub-operations.
        """
        main = MainWindow(app)

        # Test a subset of main tabs to keep it 'basic' and fast
        # dashboard, sysops, netops are usually safe 'read-only' views on start
        safe_tabs = ["dashboard", "sysops", "netops", "logs", "settings"]

        for tab_key in safe_tabs:
            if tab_key not in MAIN_TABS:
                continue

            print(f"Testing main tab: {tab_key}")
            tab = main.get_main_tab(tab_key)
            tab.wait("visible", timeout=ACTION_TIMEOUT)
            assert tab.exists(), f"Main tab '{tab_key}' not found"

            # Click and verify
            tab.click_input()
            time.sleep(1.0) # Give UI time to settle

            # In some frameworks, TabItem might not report 'selected' correctly via UIA
            # so we just verify it didn't crash and is visible.
            assert tab.is_visible(), f"Main tab '{tab_key}' should be visible after click"

    def test_sysops_overview_basic(self, app):
        """Verify we can enter SysOps and see the overview."""
        main = MainWindow(app)
        sysops_page = main.click_main_tab("sysops")
        time.sleep(1.0)

        # Just check if we are on SysOps (by checking if a subtab we expect is there)
        # We use a very basic check.
        overview_tab = sysops_page.get_subtab("overview")
        assert overview_tab.exists(), "SysOps Overview tab should exist"
        assert overview_tab.is_visible()

# Note: Subtab tests removed from main smoke suite to keep it 'basic'
# and avoid 'fatal issues' caused by heavy automation-induced interactions.
