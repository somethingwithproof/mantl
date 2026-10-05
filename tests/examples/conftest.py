"""Isolated example fixtures; no broker, database, cloud or telemetry connections."""

import importlib.util
import os
from contextlib import ExitStack
from pathlib import Path
from unittest.mock import patch

import pytest

APPS = {
    "publisher": "event-driven-arch/publisher/publisher.py",
    "subscriber": "event-driven-arch/subscriber/subscriber.py",
    "batch": "batch-processing-job/src/batch_processor.py",
    "products": "ecommerce-microservices/src/products-api/main.py",
    "ml": "ml-inference-service/src/main.py",
}


@pytest.fixture(scope="session")
def example():
    name = os.environ["MANTL_TEST_EXAMPLE"]
    os.environ["DATABASE_URL"] = "postgresql://localhost/fixture"
    os.environ["AWS_EC2_METADATA_DISABLED"] = "true"
    os.environ["AWS_DEFAULT_REGION"] = "us-east-1"
    path = Path(__file__).resolve().parents[2] / "examples" / APPS[name]
    spec = importlib.util.spec_from_file_location("mantl_example", path)
    module = importlib.util.module_from_spec(spec)
    with ExitStack() as stack:
        if name == "ml":
            # Disable exporter workers while preserving local tracing behavior.
            stack.enter_context(patch("opentelemetry.sdk.trace.export.BatchSpanProcessor"))
            stack.enter_context(patch("opentelemetry.sdk.trace.TracerProvider.add_span_processor"))
        spec.loader.exec_module(module)
    return module


def pytest_collection_modifyitems(items):
    name = os.environ.get("MANTL_TEST_EXAMPLE")
    for item in items:
        if item.path.stem != "test_" + str(name):
            item.add_marker(pytest.mark.skip(reason="Runs in a separate locked environment"))
