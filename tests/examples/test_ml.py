import asyncio
from types import SimpleNamespace
from unittest.mock import Mock

import pytest


def test_model_errors_and_predictions(example, caplog):
    async def scenario():
        example.models.clear()
        with pytest.raises(example.HTTPException) as error:
            await example.ready()
        assert error.value.status_code == 503
        calls = [
            ("predict", example.PredictionRequest(features={"feature1": 1})),
            ("predict_batch", example.BatchPredictionRequest(instances=[{"feature1": 1}])),
        ]
        for name, request in calls:
            handler = getattr(example, name)
            with pytest.raises(example.HTTPException) as error:
                await handler(request)
            assert error.value.status_code == 404
        example.models["v1"] = example.DemoModel("fixture", "v1")
        assert (await example.ready())["models"] == 1
        assert (await example.predict(calls[0][1])).prediction == pytest.approx(0.55)
        assert (await example.predict_batch(calls[1][1])).batch_size == 1
        example.models["v1"] = SimpleNamespace(
            name="fixture",
            version="v1",
            predict=Mock(side_effect=RuntimeError("model failed")),
            predict_batch=Mock(side_effect=RuntimeError("model failed")),
        )
        for name, request in calls:
            handler = getattr(example, name)
            with pytest.raises(example.HTTPException) as error:
                await handler(request)
            assert error.value.status_code == 500
            assert caplog.records[-1].exc_info is not None

    asyncio.run(scenario())


def test_openapi_errors(example):
    paths = example.app.openapi()["paths"]
    assert "503" in paths["/ready"]["get"]["responses"]
    for path in ("/api/v1/predict", "/api/v1/predict/batch"):
        assert {"404", "500"} <= paths[path]["post"]["responses"].keys()


def test_model_metadata_and_input_validation(example):
    async def scenario():
        example.models.clear()
        await example.startup_event()
        metadata = await example.list_models()
        assert {m.version for m in metadata} == {"v1", "v2"}
        assert (await example.health())["models_loaded"] == 2
        for cls, kwargs in [
            (example.PredictionRequest, {"features": {}}),
            (example.BatchPredictionRequest, {"instances": []}),
            (example.BatchPredictionRequest, {"instances": [{"feature1": 1}] * 1001}),
        ]:
            with pytest.raises(ValueError):
                cls(**kwargs)

    asyncio.run(scenario())
