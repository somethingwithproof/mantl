import asyncio
import json
from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest


def test_ready_and_publish_contract(example, monkeypatch, caplog):
    async def scenario():
        example.nc = None
        with pytest.raises(example.HTTPException) as error:
            await example.ready()
        assert error.value.status_code == 503
        example.nc = SimpleNamespace(is_closed=False)
        assert (await example.ready())["nats_connected"]
        js = SimpleNamespace(publish=AsyncMock(return_value=SimpleNamespace(seq=7)))
        monkeypatch.setattr(
            example, "get_nats_connection", AsyncMock(return_value=(example.nc, js))
        )
        event = example.Event(type="order.created\r\nforged", source="fixture", data={})
        caplog.set_level("INFO")
        result = await example.publish_event(event)
        assert result.status == "published"
        payload = json.loads(js.publish.call_args.args[1])
        assert payload["type"] == event.type
        record = next(r for r in caplog.records if r.message.startswith("Event published:"))
        assert "\n" not in record.message
        assert "\r" not in record.message
        assert "\\r\\n" in record.message
        for failure, status in [(TimeoutError(), 504), (RuntimeError("broker failed"), 500)]:
            js.publish.side_effect = failure
            with pytest.raises(example.HTTPException) as error:
                await example.publish_event(event)
            assert error.value.status_code == status
            assert caplog.records[-1].exc_info is not None
        result = await example.publish_batch_events([event])
        assert result["failed"] == 1
        with pytest.raises(example.HTTPException) as error:
            await example.publish_batch_events([event] * 101)
        assert error.value.status_code == 400
        js.stream_info = AsyncMock(side_effect=RuntimeError("broker failed"))
        with pytest.raises(example.HTTPException) as error:
            await example.stream_info()
        assert error.value.status_code == 500

    asyncio.run(scenario())


def test_openapi_errors(example):
    paths = example.app.openapi()["paths"]
    assert {"500", "504"} <= paths["/api/v1/events/publish"]["post"]["responses"].keys()
    assert "503" in paths["/ready"]["get"]["responses"]
    assert "400" in paths["/api/v1/events/publish-batch"]["post"]["responses"]
    assert "500" in paths["/api/v1/stream/info"]["get"]["responses"]


def test_startup_stream_and_clean_shutdown(example, monkeypatch):
    async def scenario():
        js = SimpleNamespace(
            add_stream=AsyncMock(),
            publish=AsyncMock(return_value=SimpleNamespace(seq=1)),
            stream_info=AsyncMock(
                return_value=SimpleNamespace(
                    config=SimpleNamespace(name="EVENTS", subjects=["EVENTS.*"]),
                    state=SimpleNamespace(messages=1, bytes=20, first_seq=1, last_seq=1),
                )
            ),
        )
        nc = SimpleNamespace(is_closed=False, jetstream=lambda: js, drain=AsyncMock())
        monkeypatch.setattr(example.nats, "connect", AsyncMock(return_value=nc))
        example.nc = None
        await example.startup_event()
        js.add_stream.assert_awaited_once()
        result = await example.publish_batch_events(
            [example.Event(type="order.created", source="fixture", data={})]
        )
        assert result["published"] == 1
        assert result["failed"] == 0
        assert (await example.stream_info())["messages"] == 1
        await example.shutdown_event()
        nc.drain.assert_awaited_once()

    asyncio.run(scenario())
