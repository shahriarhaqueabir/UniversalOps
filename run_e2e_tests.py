import os
import sys
import subprocess

# 1. CONFIGURE PYWIN32 DLL LOADING
dll_dir = r'E:\pip-packages\pywin32_system32'
if os.path.exists(dll_dir):
    os.add_dll_directory(dll_dir)

# 2. CONFIGURE PYTHON PATH
extra_paths = [
    r'E:\pip-packages',
    r'E:\pip-packages\win32',
    r'E:\pip-packages\win32\lib',
    r'E:\pip-packages\Pythonwin'
]
for p in extra_paths:
    if os.path.exists(p) and p not in sys.path:
        sys.path.insert(0, p)

# 3. CONFIGURE ENVIRONMENT FOR TESTS
os.environ["APP_PATH"] = r"E:\Projects\projectx\UniversalOps\build\bin\universal-ops.exe"
os.environ["APP_TITLE"] = "Universal-Ops Operations Platform"
os.environ["PYTHONPATH"] = ";".join(extra_paths) + ";" + os.environ.get("PYTHONPATH", "")

# 4. RUN PYTEST
import pytest

if __name__ == "__main__":
    print("--- Starting Universal-Ops E2E Tests ---")
    print(f"App Path: {os.environ['APP_PATH']}")

    # Change directory to tests/e2e so it finds pytest.ini and local imports
    os.chdir(os.path.join(os.path.dirname(__file__), "tests", "e2e"))

    # Run the navigation and pipeline smoke suite. The tab tests are the
    # desktop smoke gate even though they are not individually marker-tagged.
    sys.exit(pytest.main(["test_tabs.py", "test_pipeline.py", "-v"]))
