"""Verify that the pinned ML dependencies import and the health handler responds."""

import asyncio
import importlib.util
import sys
from pathlib import Path

import sklearn


def main():
    source = Path(__file__).resolve().parents[1] / "examples/ml-inference-service/src/main.py"
    spec = importlib.util.spec_from_file_location("mantl_ml_example", source)
    if spec is None or spec.loader is None:
        raise RuntimeError("cannot load the ML example")
    module = importlib.util.module_from_spec(spec)
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    health = asyncio.run(module.health())
    if health.get("status") != "healthy":
        raise RuntimeError("ML example health check failed")
    print(f"ML dependencies import and health responds; scikit-learn {sklearn.__version__}")


if __name__ == "__main__":
    main()
