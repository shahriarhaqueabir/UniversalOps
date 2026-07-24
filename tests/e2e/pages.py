"""Page objects for E2E UI testing using pywinauto (UIA)."""

import os
import time
import pytest
from pywinauto import Desktop

from config import (
    APP_PATH, APP_TITLE, LAUNCH_TIMEOUT, ACTION_TIMEOUT, ARTIFACT_DIR,
    MAIN_TABS, MAIN_TAB_LABELS, NETOPS_SUBTABS, NETOPS_SUBTAB_LABELS,
    SECOPS_SUBTABS, SECOPS_SUBTAB_LABELS, SYSOP_SUBTABS, SYSOP_SUBTAB_LABELS,
    LOGS_SUBTABS, LOGS_SUBTAB_LABELS, DEVOPS_SUBTABS, DEVOPS_SUBTAB_LABELS,
    AIOPS_SUBTABS, AIOPS_SUBTAB_LABELS,
)


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  Base page object
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class BasePage:
    """Base page object with common UIA helpers and artifact collection."""

    def __init__(self, app):
        self.app = app

    # -- Element resolution -------------------------------------------

    def by_id(self, auto_id: str, control_type: str = "TabItem", fallback_name: str | None = None):
        """Resolve an element by automation id, falling back to its accessible name."""
        spec = self.app.window(auto_id=auto_id, control_type=control_type)
        if spec.exists(timeout=1):
            return spec
        if fallback_name:
            return self.app.window(title=fallback_name, control_type=control_type)
        return spec

    def by_name(self, name: str, control_type: str = "TabItem"):
        """Resolve element by name."""
        return self.app.window(title=name, control_type=control_type)

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


def dismiss_onboarding(app):
    """Dismiss first-run setup for isolated smoke-test profiles.

    Strategy (first-wins):
      1. Try clicking the "Skip setup and use defaults" button by its
         UIA accessible name — works when WebView2 routes clicks.
      2. Fall back to sending the Escape key — the OnboardingModal
         treats Escape as "skip all" via its keydown handler, which is
         more reliable for WebView2 content because it bypasses the UIA
         click-pattern layer entirely.
    """
    window = app.window(title=APP_TITLE)

    # Strategy 1 — UIA button click
    skip = window.child_window(
        title="Skip setup and use defaults",
        control_type="Button",
    )
    if skip.exists(timeout=4):
        try:
            skip.click_input()
            BasePage.trace("dismissed onboarding via button click")
            return
        except Exception:
            BasePage.trace("button click failed, trying Escape key fallback")

    # Strategy 2 — Escape key (more reliable for WebView2)
    try:
        window.set_focus()
        window.type_keys("{ESC}", set_foreground=False)
        BasePage.trace("dismissed onboarding via Escape key")
    except Exception as exc:
        BasePage.trace(f"Escape key fallback also failed: {exc}")


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
        return self.by_id(auto_id, fallback_name=MAIN_TAB_LABELS.get(tab_key))

    def click_main_tab(self, tab_key: str):
        """Click a top-level navigation tab."""
        if os.environ.get("E2E_ENABLE_UIA_DOM") != "1":
            pytest.skip("WebView2 DOM automation is opt-in; run browser-level E2E with E2E_ENABLE_UIA_DOM=1")
        webview = self.app.window(title="Universal-Ops - Web content", control_type="Pane")
        if webview.exists(timeout=1):
            exposed = [c for c in webview.descendants() if c.element_info.control_type in ("Button", "TabItem")]
            if not exposed:
                pytest.skip("WebView2 exposes no DOM controls through UI Automation; use browser-level E2E for interactions")
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
    SUBTAB_LABELS: dict[str, str] = {}

    def get_subtab(self, name: str):
        """Return the UIA wrapper for a sub-tab by display name."""
        auto_id = self.SUBTABS.get(name, name)
        return self.by_id(auto_id, fallback_name=self.SUBTAB_LABELS.get(name))

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
    SUBTAB_LABELS = NETOPS_SUBTAB_LABELS


class SecOpsPage(_SubtabMixin, BasePage):
    """SecOps page with sub-tabs."""
    SUBTABS = SECOPS_SUBTABS
    SUBTAB_LABELS = SECOPS_SUBTAB_LABELS


class SysOpsPage(_SubtabMixin, BasePage):
    """SysOps page with sub-tabs."""
    SUBTABS = SYSOP_SUBTABS
    SUBTAB_LABELS = SYSOP_SUBTAB_LABELS


class LogsPage(_SubtabMixin, BasePage):
    """Logs page with sub-tabs."""
    SUBTABS = LOGS_SUBTABS
    SUBTAB_LABELS = LOGS_SUBTAB_LABELS


class DevOpsPage(_SubtabMixin, BasePage):
    """DevOps page with Radix UI tabs."""
    SUBTABS = DEVOPS_SUBTABS
    SUBTAB_LABELS = DEVOPS_SUBTAB_LABELS


class AIOpsPage(_SubtabMixin, BasePage):
    """AIOps page with Radix UI tabs."""
    SUBTABS = AIOPS_SUBTABS
    SUBTAB_LABELS = AIOPS_SUBTAB_LABELS
