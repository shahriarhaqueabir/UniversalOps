import os
import pytest
from pywinauto import Application


# Import config
import sys
sys.path.insert(0, os.path.dirname(__file__))
from config import APP_PATH, APP_TITLE, LAUNCH_TIMEOUT, ACTION_TIMEOUT, ARTIFACT_DIR


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


@pytest.hookimpl(tryfirst=True, hookwrapper=True)
def pytest_runtest_makereport(item, call):
    """Capture screenshot on failure."""
    outcome = yield
    rep = outcome.get_result()
    setattr(item, f"rep_{rep.when}", rep)

    if rep.when == "call" and rep.failed:
        # Try to get the app fixture
        win = item.funcargs.get("app")
        if win:
            os.makedirs(ARTIFACT_DIR, exist_ok=True)
            try:
                win.capture_as_image().save(
                    os.path.join(ARTIFACT_DIR, f"FAIL_{item.name}.png")
                )
            except Exception:
                pass