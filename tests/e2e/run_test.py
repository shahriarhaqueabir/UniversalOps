"""
Standalone test runner — kept for convenience.

No path bootstrapping needed: pywin32 DLL directory is resolved
dynamically in conftest.py.
"""
import sys
import pytest

if __name__ == "__main__":
    args = sys.argv[1:] if len(sys.argv) > 1 else []
    sys.exit(pytest.main(args))
