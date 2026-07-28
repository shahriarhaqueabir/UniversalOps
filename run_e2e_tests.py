import os
import sys
import subprocess
import pytest

# 1. CONFIGURE PYWIN32 DLL LOADING
# Dynamically locate pywin32 DLL directory (workstation-agnostic)
if sys.platform == "nt":
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
        pass

# 2. CONFIGURE ENVIRONMENT FOR TESTS
# Use environment variables if set (CI), otherwise fallback to local project structure
app_path = os.environ.get("APP_PATH")
if not app_path:
    # Default to build/bin relative to this script
    app_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "build", "bin", "universal-ops.exe")

os.environ["APP_PATH"] = app_path
os.environ["APP_TITLE"] = os.environ.get("APP_TITLE", "Universal-Ops Operations Platform")

if __name__ == "__main__":
    print("--- Starting Universal-Ops E2E Tests ---")
    print(f"App Path: {os.environ['APP_PATH']}")

    if not os.path.exists(os.environ['APP_PATH']):
        print(f"ERROR: Executable not found at {os.environ['APP_PATH']}")
        sys.exit(1)

    # Change directory to tests/e2e so it finds pytest.ini and local imports
    os.chdir(os.path.join(os.path.dirname(__file__), "tests", "e2e"))

    # Run the navigation and pipeline smoke suite.
    sys.exit(pytest.main(["test_tabs.py", "test_pipeline.py", "-v"]))
