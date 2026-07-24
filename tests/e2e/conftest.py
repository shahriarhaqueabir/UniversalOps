"""Pytest fixtures and hooks for E2E testing."""

import os
import sys
import subprocess
import shlex
import shutil
import tempfile
import time

import pytest
from pywinauto import Application

from config import (
    APP_PATH, APP_TITLE, LAUNCH_TIMEOUT, ACTION_TIMEOUT, ARTIFACT_DIR,
)


# ── pywin32 DLL path (Windows CI) ───────────────────────────────────

if sys.platform == "nt":
    extra_paths = [
        "E:\\pip-packages",
        "E:\\pip-packages\\win32",
        "E:\\pip-packages\\win32\\lib",
        "E:\\pip-packages\\Pythonwin",
    ]
    for p in extra_paths:
        if os.path.exists(p):
            sys.path.insert(0, p)

    dll_dir = "E:\\pip-packages\\pywin32_system32"
    if os.path.exists(dll_dir):
        os.add_dll_directory(dll_dir)


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
    user_data = os.path.join(tmpdir, "AppData", "Roaming")
    os.makedirs(user_data, exist_ok=True)

    env = os.environ.copy()
    env["QT_ACCESSIBILITY"] = "1"
    for var in ["APPDATA", "LOCALAPPDATA", "TEMP", "TMP"]:
        env[var] = os.path.join(tmpdir, "AppData", "Local") if var == "LOCALAPPDATA" else tmpdir

    proc = subprocess.Popen(
        shlex.split(APP_PATH),
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
