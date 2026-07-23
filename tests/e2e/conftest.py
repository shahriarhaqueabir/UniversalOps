import os
import sys
import pytest
import time
import subprocess
import shlex
import tempfile
import shutil
from pywinauto import Application

# Import config
sys.path.insert(0, os.path.dirname(__file__))
from config import APP_PATH, APP_TITLE, LAUNCH_TIMEOUT, ARTIFACT_DIR

@pytest.fixture(scope="session")
def sandbox_dir():
    """Create a persistent sandbox directory for the entire session."""
    tmp_dir = tempfile.mkdtemp(prefix="opsforall_e2e_session_")
    yield tmp_dir
    # Cleanup at the very end
    shutil.rmtree(tmp_dir, ignore_errors=True)

@pytest.fixture(scope="session")
def app(sandbox_dir):
    """Launch app instance once for the entire session with isolated user data dirs."""
    if not APP_PATH or not os.path.exists(APP_PATH):
        # Don't fail hard here, maybe it's being built or path is different
        # but for E2E we usually expect it.
        pytest.exit(f"APP_PATH not found: {APP_PATH}", returncode=1)

    # 1. Setup Sandbox Environment
    sandbox_env = os.environ.copy()
    sandbox_env["APPDATA"] = os.path.join(sandbox_dir, "AppData", "Roaming")
    sandbox_env["LOCALAPPDATA"] = os.path.join(sandbox_dir, "AppData", "Local")
    sandbox_env["TEMP"] = sandbox_env["TMP"] = os.path.join(sandbox_dir, "Temp")

    # Deep Sandbox: Redirect UserProfile to prevent local config leakage/locks
    user_profile = os.path.join(sandbox_dir, "UserProfile")
    os.makedirs(user_profile, exist_ok=True)
    sandbox_env["USERPROFILE"] = user_profile
    sandbox_env["HOMEDRIVE"] = os.path.splitdrive(user_profile)[0]
    sandbox_env["HOMEPATH"] = os.path.splitdrive(user_profile)[1]

    # Critical for some UI frameworks to be discoverable by UIA
    sandbox_env["QT_ACCESSIBILITY"] = "1"

    # Wails/WebView2 specific isolation
    sandbox_env["WEBVIEW2_USER_DATA_FOLDER"] = os.path.join(sandbox_dir, "WebView2Data")

    for p in (sandbox_env["APPDATA"], sandbox_env["LOCALAPPDATA"], sandbox_env["TEMP"]):
        os.makedirs(p, exist_ok=True)

    # 2. Launch App
    print(f"\n[E2E] Launching: {APP_PATH}")
    proc = subprocess.Popen(
        [APP_PATH],
        env=sandbox_env,
        creationflags=subprocess.CREATE_NO_WINDOW if os.name == 'nt' else 0
    )

    # 3. Connect pywinauto
    time.sleep(3) # Give more time for heavy initial load

    try:
        pw_app = Application(backend="uia").connect(process=proc.pid, timeout=LAUNCH_TIMEOUT)
        win = pw_app.window(title=APP_TITLE)
        win.wait("visible", timeout=LAUNCH_TIMEOUT)
        win.set_focus()

        yield win
    except Exception as e:
        print(f"[E2E] Setup Error: {e}")
        # Take screenshot of desktop if app failed to connect
        try:
            from PIL import ImageGrab
            os.makedirs(ARTIFACT_DIR, exist_ok=True)
            ImageGrab.grab().save(os.path.join(ARTIFACT_DIR, "DESKTOP_FAILURE.png"))
        except:
            pass
        raise e
    finally:
        # 4. Teardown
        print("[E2E] Tearing down app...")
        try:
            if 'win' in locals() and win.exists():
                win.close()

            proc.terminate()
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                proc.kill()
        except Exception as e:
            print(f"[E2E] Teardown warning: {e}")
            proc.kill()

@pytest.hookimpl(tryfirst=True, hookwrapper=True)
def pytest_runtest_makereport(item, call):
    outcome = yield
    rep = outcome.get_result()
    setattr(item, f"rep_{rep.when}", rep)

    if rep.when == "call" and rep.failed:
        # Try to get 'app' fixture from any scope
        win = item.funcargs.get("app")
        if win:
            os.makedirs(ARTIFACT_DIR, exist_ok=True)
            try:
                # Sanitize filename
                safe_name = "".join([c if c.isalnum() else "_" for c in item.name])
                win.capture_as_image().save(
                    os.path.join(ARTIFACT_DIR, f"FAIL_{safe_name}.png")
                )
            except:
                pass
