"""Page objects for E2E UI testing using pywinauto (UIA)."""

import os
import time
from pywinauto import Desktop

from config import (
    APP_PATH, APP_TITLE, LAUNCH_TIMEOUT, ACTION_TIMEOUT, ARTIFACT_DIR,
    MAIN_TABS, NETOPS_SUBTABS, SECOPS_SUBTABS, SYSOP_SUBTABS,
    LOGS_SUBTABS, DEVOPS_SUBTABS, AIOPS_SUBTABS,
)


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  Base page object
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class BasePage:
    """Base page object with common UIA helpers and artifact collection."""

    def __init__(self, app):
        self.app = app

    # -- Element resolution -------------------------------------------

    def by_id(self, auto_id: str, control_type: str = "TabItem"):
        """Resolve element by automation-id (preferred for Wails/Radix)."""
        return self.app.window(auto_id=auto_id, control_type=control_type)

    def by_name(self, name: str, control_type: str = "TabItem"):
        """Resolve element by name."""
        return self.app.window(name=name, control_type=control_type)

    def by_class(self, class_name: str):
        """Resolve element by class name."""
        return self.app.window(class_name=class_name)

    # -- Wait helpers -------------------------------------------------

    def wait_visible(self, spec, timeout: int | None = None):
        """Wait until the element is visible."""
        spec.wait("visible", timeout=timeout or ACTION_TIMEOUT)

    def wait_gone(self, spec, timeout: int | None = None):
        """Wait until the element is gone."""
        spec.wait("gone", timeout=timeout or ACTION_TIMEOUT)

    def wait_window(self, title: str, timeout: int | None = None):
        """Wait for a top-level window with *title* to appear."""
        return self.app.window(title=title).wait(
            "visible", timeout=timeout or LAUNCH_TIMEOUT,
        )

    def wait_until(self, fn, timeout: int | None = None, interval: float = 0.3):
        """Poll *fn* every *interval* s until it returns truthy."""
        deadline = time.time() + (timeout or ACTION_TIMEOUT)
        while time.time() < deadline:
            if fn():
                return
            time.sleep(interval)
        raise TimeoutError(f"wait_until timed out after {timeout or ACTION_TIMEOUT}s")

    # -- Actions ------------------------------------------------------

    def click(self, spec):
        self.wait_visible(spec)
        spec.click()

    def type_text(self, spec, text: str):
        self.wait_visible(spec)
        spec.click()
        spec.type_keys(text)

    def get_text(self, spec) -> str:
        self.wait_visible(spec)
        return spec.window_text()

    def screenshot(self, name: str):
        """Save a PNG screenshot of the main window to the artifact dir."""
        path = os.path.join(
            ARTIFACT_DIR,
            f"{name}_{int(time.time())}.png",
        )
        os.makedirs(ARTIFACT_DIR, exist_ok=True)
        self.app.window().capture_as_image().save(path)

    # -- Tracing ------------------------------------------------------

    @staticmethod
    def trace(msg: str):
        print(f"[E2E] {msg}")

    def click_traced(self, spec, name: str):
        self.trace(f"Clicking {name}")
        self.click(spec)

    def type_traced(self, spec, text: str, name: str):
        self.trace(f"Typing into {name}")
        self.type_text(spec, text)


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  Main window
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class MainWindow(BasePage):
    """Top-level window with main navigation tabs."""

    def __init__(self, app):
        super().__init__(app)
        self._tab_cache: dict[str, object] = {}

    def get_main_tab(self, tab_key: str):
        """Return the UIA wrapper for a main navigation tab."""
        auto_id = MAIN_TABS.get(tab_key, f"main-tab-{tab_key}")
        return self.by_id(auto_id)

    def click_main_tab(self, tab_key: str):
        """Click a top-level navigation tab."""
        tab = self.get_main_tab(tab_key)
        self.click_traced(tab, f"main_tab_{tab_key}")


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  Dashboard / Settings (no sub-tabs yet)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class DashboardPage(BasePage):
    """Dashboard page (no sub-tabs currently)."""


class SettingsPage(BasePage):
    """Settings page (no sub-tabs currently)."""


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  Section pages (each has sub-tabs)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class _SubtabMixin:
    """Mixin for pages that expose sub-tab navigation/verification."""

    SUBTABS: dict[str, str] = {}

    def get_subtab(self, name: str):
        """Return the UIA wrapper for a sub-tab by display name."""
        auto_id = self.SUBTABS.get(name, name)
        return self.by_id(auto_id)

    def click_subtab(self, name: str):
        """Click a sub-tab."""
        tab = self.get_subtab(name)
        self.click_traced(tab, f"{self.__class__.__name__}_subtab_{name}")

    def verify_subtabs(self):
        """Verify every expected sub-tab exists and is visible."""
        for name, auto_id in self.SUBTABS.items():
            tab = self.by_id(auto_id)
            tab.wait("visible", timeout=ACTION_TIMEOUT)
            assert tab.exists(), f"Sub-tab '{name}' ({auto_id}) does not exist"


class NetOpsPage(_SubtabMixin, BasePage):
    """NetOps page with sub-tabs."""
    SUBTABS = NETOPS_SUBTABS


class SecOpsPage(_SubtabMixin, BasePage):
    """SecOps page with sub-tabs."""
    SUBTABS = SECOPS_SUBTABS


class SysOpsPage(_SubtabMixin, BasePage):
    """SysOps page with sub-tabs."""
    SUBTABS = SYSOP_SUBTABS


class LogsPage(_SubtabMixin, BasePage):
    """Logs page with sub-tabs."""
    SUBTABS = LOGS_SUBTABS


class DevOpsPage(_SubtabMixin, BasePage):
    """DevOps page with Radix UI tabs."""
    SUBTABS = DEVOPS_SUBTABS


class AIOpsPage(_SubtabMixin, BasePage):
    """AIOps page with Radix UI tabs."""
    SUBTABS = AIOPS_SUBTABS
