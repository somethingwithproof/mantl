"""Basic smoke tests to ensure the test harness runs.
These keep `pytest` from exiting with code 5 (no tests collected) in minimal setups.
"""

import shutil


def test_pytest_runs() -> None:
    assert True


def test_kustomize_in_path() -> None:
    # kustomize is needed by CI validators
    assert shutil.which("kustomize"), "kustomize not found in PATH"
