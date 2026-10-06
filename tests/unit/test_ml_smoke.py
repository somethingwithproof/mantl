# SPDX-License-Identifier: Apache-2.0
"""The ML compatibility check must fail on unavailable or unhealthy services."""

import importlib.util
import sys
from types import SimpleNamespace

import pytest


@pytest.fixture
def smoke(project_root, monkeypatch):
    monkeypatch.setitem(sys.modules, "sklearn", SimpleNamespace(__version__="fixture"))
    spec = importlib.util.spec_from_file_location(
        "test_ml_smoke", project_root / "ci/smoke_ml_example.py"
    )
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


@pytest.mark.parametrize("spec", [None, SimpleNamespace(loader=None)])
def test_unavailable_module_fails_check(smoke, monkeypatch, spec):
    monkeypatch.setattr(smoke.importlib.util, "spec_from_file_location", lambda *args: spec)
    with pytest.raises(RuntimeError, match="cannot load"):
        smoke.main()


def test_unhealthy_service_fails_check(smoke, monkeypatch):
    async def health():
        return {"status": "unhealthy"}

    spec = SimpleNamespace(
        name="mantl_ml_example", loader=SimpleNamespace(exec_module=lambda module: None)
    )
    monkeypatch.setattr(smoke.importlib.util, "spec_from_file_location", lambda *args: spec)
    monkeypatch.setattr(
        smoke.importlib.util, "module_from_spec", lambda spec: SimpleNamespace(health=health)
    )
    monkeypatch.setitem(sys.modules, spec.name, None)
    with pytest.raises(RuntimeError, match="health check failed"):
        smoke.main()
