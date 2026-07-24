"""
UIA tree inspector for Universal-Ops WebView2 onboarding.

Usage:
    python tests/e2e/inspect_uia.py

Dumps the full UIA automation tree so you can see exactly what
control types, names, and automation-ids the WebView2 surface
exposes.  Useful for debugging selector mismatches in page objects.
"""

import os
import sys
import time
import subprocess
import tempfile

# ── ensure pywinauto / pywin32 paths (workstation-agnostic) ────────
import importlib.util
for _mod_name in ("win32api", "win32file"):
    _spec = importlib.util.find_spec(_mod_name)
    if _spec and _spec.origin:
        _pywin32_site = os.path.dirname(os.path.dirname(_spec.origin))
        _dll_dir = os.path.join(_pywin32_site, "pywin32_system32")
        if os.path.exists(_dll_dir):
            os.add_dll_directory(_dll_dir)
            break

from pywinauto import Application, Desktop


APP_PATH = os.environ.get(
    "APP_PATH",
    os.path.join(
        os.path.dirname(os.path.abspath(__file__)),
        "..", "..", "build", "bin", "universal-ops.exe",
    ),
)
APP_TITLE = os.environ.get("APP_TITLE", "Universal-Ops Operations Platform")


# ── helpers ─────────────────────────────────────────────────────────

def dump_tree(element, indent: int = 0, max_depth: int = 8, max_siblings: int = 80):
    """Recursively print the UIA tree for *element*."""
    if indent > max_depth:
        return
    try:
        info = element.element_info
        ctrl_type = info.control_type or "(no-ctrl)"
        name = info.name or "(unnamed)"
        auto_id = info.automation_id or ""
        class_name = info.class_name or ""
        rect = info.rectangle
        rect_str = f"  [{rect.left},{rect.top},{rect.right},{rect.bottom}]" if rect else ""
        suffix = f"  aid={auto_id}" if auto_id else ""
        suffix += f"  class={class_name}" if class_name else ""
        print(f"{'  ' * indent}{ctrl_type}: {name!r}{suffix}{rect_str}")
    except Exception as exc:
        print(f"{'  ' * indent}  (error: {exc})")
        return

    try:
        children = element.descendants()
    except Exception:
        children = []

    shown = 0
    for child in children:
        if shown >= max_siblings:
            print(f"{'  ' * (indent+1)}... ({len(children) - shown} more)")
            break
        try:
            ci = child.element_info
            if ci.control_type == "Pane" and ci.class_name in ("Windows.UI.Core.CoreWindow",):
                # skip internal Windows.UI.Core.CoreWindow panes
                continue
        except Exception:
            pass
        dump_tree(child, indent + 1, max_depth, max_siblings)
        shown += 1


def main():
    # Launch with isolated data dir -- no .onboarded so the modal shows
    tmpdir = tempfile.mkdtemp(prefix="opsforall_inspect_")
    portable_data = os.path.join(tmpdir, "data")
    os.makedirs(portable_data, exist_ok=True)

    env = os.environ.copy()
    env["QT_ACCESSIBILITY"] = "1"
    for var in ["APPDATA", "LOCALAPPDATA", "TEMP", "TMP"]:
        env[var] = (
            os.path.join(tmpdir, "AppData", "Local")
            if var == "LOCALAPPDATA"
            else tmpdir
        )

    print(f"Launching {APP_PATH} with isolated data dir: {tmpdir}")
    print(f"  NOTE: .onboarded marker is NOT created so the onboarding modal appears\n")

    proc = subprocess.Popen(
        [APP_PATH],
        cwd=tmpdir,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    try:
        # Connect
        deadline = time.time() + 45
        app = None
        while time.time() < deadline:
            try:
                app = Application(backend="uia").connect(process=proc.pid)
                break
            except Exception:
                time.sleep(1)
        if app is None:
            raise TimeoutError("Could not connect to app")

        win = app.window(title=APP_TITLE)
        win.wait("visible", timeout=45)
        win.set_focus()
        time.sleep(3)  # let WebView2 render & accessibility settle

        # ── Dump: main window child tree (shallow) ──
        print("=" * 70)
        print("TOP-LEVEL WINDOW CHILDREN")
        print("=" * 70)
        dump_tree(win, indent=0, max_depth=2, max_siblings=40)

        # ── Deep search for WebView2 / Chromium content ──
        print("\n" + "=" * 70)
        print("SEARCHING FOR WEBVIEW2 / CHROMIUM CONTENT")
        print("=" * 70)
        webview_candidates = win.descendants(control_type="Pane")
        for pane in webview_candidates:
            try:
                ci = pane.element_info
                ci_lower = ci.class_name.lower()
                if "webview" in ci_lower or "chromium" in ci_lower or "webview2" in ci_lower:
                    print(f"\n--- WebView2 Pane: class={ci.class_name}, name={ci.name!r} ---")
                    dump_tree(pane, indent=1, max_depth=6, max_siblings=60)
            except Exception:
                continue

        # ── Full deep tree (first major subtree) ──
        print("\n" + "=" * 70)
        print("FULL DEEP TREE (depth-limited to 5, 60 siblings)")
        print("=" * 70)
        for top in win.descendants():
            try:
                ci = top.element_info
                if ci.control_type == "Pane" and ci.class_name in ("Windows.UI.Core.CoreWindow",):
                    continue
            except Exception:
                pass
            dump_tree(top, indent=1, max_depth=5, max_siblings=60)
            break  # just first major subtree

        # ── Search by aria-label / automation-id patterns ──
        print("\n" + "=" * 70)
        print("SEARCH: All buttons with aria-label patterns")
        print("=" * 70)
        for btn in win.descendants(control_type="Button"):
            try:
                name = btn.element_info.name or ""
                aid = btn.element_info.automation_id or ""
                if "skip" in name.lower() or "onboard" in name.lower() or "setup" in name.lower():
                    rect = btn.element_info.rectangle
                    print(f"  Button: name={name!r}  aid={aid}  rect=[{rect.left},{rect.top},{rect.right},{rect.bottom}]")
            except Exception:
                continue

        print("\n" + "=" * 70)
        print("SEARCH: All elements with 'skip' in name or automation-id")
        print("=" * 70)
        for elem in win.descendants():
            try:
                ci = elem.element_info
                name = (ci.name or "").lower()
                aid = (ci.automation_id or "").lower()
                if "skip" in name or "skip" in aid:
                    rect = ci.rectangle
                    print(f"  {ci.control_type}: name={ci.name!r}  aid={ci.automation_id}  rect=[{rect.left},{rect.top},{rect.right},{rect.bottom}]")
            except Exception:
                continue

        # ── Quick check ──
        print("\n" + "=" * 70)
        print("QUICK CHECK: Does the target button exist?")
        print("=" * 70)
        target = win.child_window(
            title="Skip setup and use defaults",
            control_type="Button",
        )
        print(f"  child_window(title='Skip setup and use defaults', control_type='Button'): exists={target.exists(timeout=2)}")

    finally:
        proc.kill()
        proc.wait()
        try:
            import shutil
            shutil.rmtree(tmpdir, ignore_errors=True)
        except Exception:
            pass
    print("\nDone.")


if __name__ == "__main__":
    main()
