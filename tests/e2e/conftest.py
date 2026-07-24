"""Pytest fixtures and hooks for E2E testing."""

import os
import sys
import subprocess
import shutil
import tempfile
import time

# ── pywin32 DLL path (Windows) — MUST be set before pywinauto imports ──

if sys.platform == "nt":
    # Dynamically locate pywin32 DLL directory (workstation-agnostic)
    try:
        import importlib.util
        for mod_name in ("win32api", "win32file"):
            spec = importlib.util.find_spec(mod_name)
            if spec and spec.origin:
                pywin32_site = os.path.dirname(os.path.dirname(spec.origin))
                dll_dir = os.path.join(pywin32_site, "pywin32_system32")
                if os.path.exists(dll_dir):
                    os.add_dll_directory(dll_dir)
                    break
    except Exception:
        pass  # pywinauto will surface its own ImportError if DLLs are missing

import pytest
from pywinauto import Application

from config import (
    APP_PATH, APP_TITLE, LAUNCH_TIMEOUT, ACTION_TIMEOUT, ARTIFACT_DIR,
)
from pages import dismiss_onboarding


# ── Fixtures ────────────────────────────────────────────────────────

@pytest.fixture(scope="function")
def app():
    """Launch a fresh app instance per test with isolated user data dirs."""
    if not APP_PATH:
        pytest.exit("APP_PATH environment variable not set", returncode=1)
    if not APP_TITLE:
        pytest.exit("APP_TITLE environment variable not set", returncode=1)

    # Isolated data dir
    tmpdir = tempfile.mkdtemp(prefix="opsforall_e2e_")
    portable_data = os.path.join(tmpdir, "data")
    os.makedirs(portable_data, exist_ok=True)
    open(os.path.join(portable_data, ".onboarded"), "w", encoding="utf-8").close()
    user_data = os.path.join(tmpdir, "AppData", "Roaming")
    os.makedirs(user_data, exist_ok=True)

    env = os.environ.copy()
    env["QT_ACCESSIBILITY"] = "1"
    for var in ["APPDATA", "LOCALAPPDATA", "TEMP", "TMP"]:
        env[var] = os.path.join(tmpdir, "AppData", "Local") if var == "LOCALAPPDATA" else tmpdir

    proc = subprocess.Popen(
        [APP_PATH],
        cwd=tmpdir,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    try:
        # Connect via PID (retry loop up to LAUNCH_TIMEOUT)
        deadline = time.time() + LAUNCH_TIMEOUT
        while time.time() < deadline:
            try:
                app = Application(backend="uia").connect(process=proc.pid)
                break
            except Exception:
                time.sleep(1)
        else:
            raise TimeoutError(f"Could not connect to app (PID {proc.pid})")

        app.window(title=APP_TITLE).wait("visible", timeout=LAUNCH_TIMEOUT)
        app.window(title=APP_TITLE).set_focus()
        dismiss_onboarding(app)

        yield app

    finally:
        # Teardown: close the app window
        try:
            app.window(title=APP_TITLE).close()
        except Exception:
            pass

        # Kill leftover process
        try:
            proc.kill()
        except Exception:
            pass

        # Clean up temp dir
        try:
            shutil.rmtree(tmpdir, ignore_errors=True)
        except Exception:
            pass


# ── Hooks ───────────────────────────────────────────────────────────

@pytest.hookimpl(tryfirst=True, hookwrapper=True)
def pytest_runtest_makereport(item, call):
    """Capture a screenshot on test failure."""
    outcome = yield
    rep = outcome.get_result()

    if rep.when == "call" and rep.failed:
        app_fixture = item.funcargs.get("app")
        if app_fixture is not None:
            os.makedirs(ARTIFACT_DIR, exist_ok=True)
            try:
                win = app_fixture.window(title=APP_TITLE)
                img = win.capture_as_image()
                path = os.path.join(ARTIFACT_DIR, f"FAIL_{item.name}.png")
                img.save(path)
            except Exception:
                pass
