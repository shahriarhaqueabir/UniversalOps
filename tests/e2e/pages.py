import os
import time
import json
from pywinauto import Desktop
from config import (
    APP_PATH, APP_TITLE, LAUNCH_TIMEOUT, ACTION_TIMEOUT,
    ARTIFACT_DIR, MAIN_TABS, NETOPS_SUBTABS, SECOPS_SUBTABS,
    SYSOP_SUBTABS, LOGS_SUBTABS, DEVOPS_SUBTABS, AIOPS_SUBTABS
)


class BasePage:
    """Base page object with common UIA helpers and artifact collection."""

    def __init__(self, window):
        self.window = window

    # ─── Locator Helpers ────────────────────────────────────────────

    def by_id(self, auto_id, **kwargs):
        """Find by AutomationId (most stable)."""
        return self.window.child_window(auto_id=auto_id, **kwargs)

    def by_name(self, name, **kwargs):
        """Find by visible text / accessible name."""
        return self.window.child_window(title=name, **kwargs)

    def by_class(self, cls, index=0, **kwargs):
        """Find by control class + index (fragile, last resort)."""
        return self.window.child_window(class_name=cls, found_index=index, **kwargs)

    # ─── Waits ──────────────────────────────────────────────────────

    def wait_visible(self, spec, timeout=ACTION_TIMEOUT):
        spec.wait("visible", timeout=timeout)
        return spec

    def wait_gone(self, spec, timeout=ACTION_TIMEOUT):
        spec.wait_not("visible", timeout=timeout)
        return spec

    def wait_window(self, title, timeout=LAUNCH_TIMEOUT):
        """Wait for a new top-level window (dialogs, popups)."""
        dlg = Desktop(backend="uia").window(title=title)
        dlg.wait("visible", timeout=timeout)
        return dlg

    def wait_until(self, fn, timeout=ACTION_TIMEOUT, interval=0.3):
        """Poll arbitrary condition."""
        deadline = time.time() + timeout
        while time.time() < deadline:
            try:
                if fn():
                    return True
            except Exception:
                pass
            time.sleep(interval)
        raise TimeoutError(f"Condition not met within {timeout}s")

    # ─── Actions ────────────────────────────────────────────────────

    def click(self, spec):
        self.wait_visible(spec)
        spec.click_input()

    def type_text(self, spec, text):
        self.wait_visible(spec)
        ctrl = spec.wrapper_object()
        try:
            ctrl.set_edit_text(text)
        except Exception:
            # Fallback for controls without ValuePattern (e.g., Qt 5.x)
            import pywinauto.keyboard as kb
            ctrl.click_input()
            kb.send_keys("^a")
            kb.send_keys(text, with_spaces=True)

    def get_text(self, spec):
        ctrl = spec.wrapper_object()
        for attr in ("window_text", "get_value"):
            try:
                v = getattr(ctrl, attr)()
                if v:
                    return v
            except Exception:
                pass
        return ""

    # ─── Artifacts ──────────────────────────────────────────────────

    def screenshot(self, name):
        os.makedirs(ARTIFACT_DIR, exist_ok=True)
        path = os.path.join(ARTIFACT_DIR, f"{name}.png")
        self.window.capture_as_image().save(path)
        return path

    def trace(self, action, spec=None, text=None):
        """Per-step trace for flaky test diagnosis (opt-in via E2E_TRACE=1)."""
        if os.environ.get("E2E_TRACE") != "1":
            return

        os.makedirs(ARTIFACT_DIR, exist_ok=True)
        step = getattr(self, "_step", 0) + 1
        self._step = step
        idx = f"{step:03d}"

        try:
            self.window.capture_as_image().save(
                os.path.join(ARTIFACT_DIR, f"step_{idx}_{action}.png"))
        except Exception:
            pass

        rec = {
            "ts": time.time(),
            "step": step,
            "action": action,
            "locator": getattr(spec, "criteria", None) if spec else None,
            "text": text if os.environ.get("E2E_TRACE_INCLUDE_TEXT") == "1" else (
                "<redacted>" if text else None),
        }
        with open(os.path.join(ARTIFACT_DIR, "trace.jsonl"), "a") as f:
            f.write(json.dumps(rec) + "\n")

    def click_traced(self, spec, action_name):
        self.trace(f"{action_name}_before", spec)
        self.click(spec)
        self.trace(f"{action_name}_after", spec)

    def type_traced(self, spec, text, action_name):
        self.trace(f"{action_name}_before", spec, text)
        self.type_text(spec, text)
        self.trace(f"{action_name}_after", spec)


class MainWindow(BasePage):
    """Top-level window with main navigation tabs."""

    def __init__(self, window):
        super().__init__(window)
        self._tab_cache = {}

    def get_main_tab(self, tab_key):
        """Get main navigation tab by key (e.g., 'netops')."""
        auto_id = MAIN_TABS[tab_key]
        if auto_id not in self._tab_cache:
            self._tab_cache[auto_id] = self.by_id(auto_id, control_type="TabItem")
        return self._tab_cache[auto_id]

    def click_main_tab(self, tab_key):
        """Click a main navigation tab and return the appropriate page object."""
        tab = self.get_main_tab(tab_key)
        self.click_traced(tab, f"main_tab_{tab_key}")
        time.sleep(0.3)  # Let React render

        # Return the appropriate page object
        if tab_key == "netops":
            return NetOpsPage(self.window)
        elif tab_key == "secops":
            return SecOpsPage(self.window)
        elif tab_key == "sysops":
            return SysOpsPage(self.window)
        elif tab_key == "logs":
            return LogsPage(self.window)
        elif tab_key == "devops":
            return DevOpsPage(self.window)
        elif tab_key == "aiops":
            return AIOpsPage(self.window)
        elif tab_key == "dashboard":
            return DashboardPage(self.window)
        elif tab_key == "settings":
            return SettingsPage(self.window)
        return BasePage(self.window)


class DashboardPage(BasePage):
    """Dashboard page (no sub-tabs currently)."""
    pass


class SettingsPage(BasePage):
    """Settings page (no sub-tabs currently)."""
    pass


class NetOpsPage(BasePage):
    """NetOps page with sub-tabs."""

    def get_subtab(self, subtab_key):
        auto_id = NETOPS_SUBTABS[subtab_key]
        return self.by_id(auto_id, control_type="TabItem")

    def click_subtab(self, subtab_key):
        tab = self.get_subtab(subtab_key)
        self.click_traced(tab, f"netops_subtab_{subtab_key}")
        time.sleep(0.2)
        return self

    def verify_subtabs(self):
        results = {}
        for key, auto_id in NETOPS_SUBTABS.items():
            try:
                tab = self.by_id(auto_id, control_type="TabItem")
                tab.wait("visible", timeout=3)
                results[key] = {"exists": True, "selected": tab.is_selected()}
            except Exception as e:
                results[key] = {"exists": False, "error": str(e)}
        return results


class SecOpsPage(BasePage):
    """SecOps page with sub-tabs."""

    def get_subtab(self, subtab_key):
        auto_id = SECOPS_SUBTABS[subtab_key]
        return self.by_id(auto_id, control_type="TabItem")

    def click_subtab(self, subtab_key):
        tab = self.get_subtab(subtab_key)
        self.click_traced(tab, f"secops_subtab_{subtab_key}")
        time.sleep(0.2)
        return self

    def verify_subtabs(self):
        results = {}
        for key, auto_id in SECOPS_SUBTABS.items():
            try:
                tab = self.by_id(auto_id, control_type="TabItem")
                tab.wait("visible", timeout=3)
                results[key] = {"exists": True, "selected": tab.is_selected()}
            except Exception as e:
                results[key] = {"exists": False, "error": str(e)}
        return results


class SysOpsPage(BasePage):
    """SysOps page with sub-tabs."""

    def get_subtab(self, subtab_key):
        auto_id = SYSOP_SUBTABS[subtab_key]
        return self.by_id(auto_id, control_type="TabItem")

    def click_subtab(self, subtab_key):
        tab = self.get_subtab(subtab_key)
        self.click_traced(tab, f"sysops_subtab_{subtab_key}")
        time.sleep(0.2)
        return self

    def verify_subtabs(self):
        results = {}
        for key, auto_id in SYSOP_SUBTABS.items():
            try:
                tab = self.by_id(auto_id, control_type="TabItem")
                tab.wait("visible", timeout=3)
                results[key] = {"exists": True, "selected": tab.is_selected()}
            except Exception as e:
                results[key] = {"exists": False, "error": str(e)}
        return results


class LogsPage(BasePage):
    """Logs page with sub-tabs."""

    def get_subtab(self, subtab_key):
        auto_id = LOGS_SUBTABS[subtab_key]
        return self.by_id(auto_id, control_type="TabItem")

    def click_subtab(self, subtab_key):
        tab = self.get_subtab(subtab_key)
        self.click_traced(tab, f"logs_subtab_{subtab_key}")
        time.sleep(0.2)
        return self

    def verify_subtabs(self):
        results = {}
        for key, auto_id in LOGS_SUBTABS.items():
            try:
                tab = self.by_id(auto_id, control_type="TabItem")
                tab.wait("visible", timeout=3)
                results[key] = {"exists": True, "selected": tab.is_selected()}
            except Exception as e:
                results[key] = {"exists": False, "error": str(e)}
        return results


class DevOpsPage(BasePage):
    """DevOps page with Radix UI tabs."""

    def get_subtab(self, subtab_key):
        auto_id = DEVOPS_SUBTABS[subtab_key]
        return self.by_id(auto_id, control_type="TabItem")

    def click_subtab(self, subtab_key):
        tab = self.get_subtab(subtab_key)
        self.click_traced(tab, f"devops_subtab_{subtab_key}")
        time.sleep(0.2)
        return self

    def verify_subtabs(self):
        results = {}
        for key, auto_id in DEVOPS_SUBTABS.items():
            try:
                tab = self.by_id(auto_id, control_type="TabItem")
                tab.wait("visible", timeout=3)
                results[key] = {"exists": True, "selected": tab.is_selected()}
            except Exception as e:
                results[key] = {"exists": False, "error": str(e)}
        return results


class AIOpsPage(BasePage):
    """AIOps page with Radix UI tabs."""

    def get_subtab(self, subtab_key):
        auto_id = AIOPS_SUBTABS[subtab_key]
        return self.by_id(auto_id, control_type="TabItem")

    def click_subtab(self, subtab_key):
        tab = self.get_subtab(subtab_key)
        self.click_traced(tab, f"aiops_subtab_{subtab_key}")
        time.sleep(0.2)
        return self

    def verify_subtabs(self):
        results = {}
        for key, auto_id in AIOPS_SUBTABS.items():
            try:
                tab = self.by_id(auto_id, control_type="TabItem")
                tab.wait("visible", timeout=3)
                results[key] = {"exists": True, "selected": tab.is_selected()}
            except Exception as e:
                results[key] = {"exists": False, "error": str(e)}
        return results