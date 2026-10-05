import asyncio
from types import SimpleNamespace
from unittest.mock import AsyncMock, Mock

import pytest


def test_shutdown_interrupts_backoff(example):
    async def scenario():
        example.shutdown_requested = asyncio.Event()
        task = asyncio.create_task(example.wait_for_shutdown(60))
        await asyncio.sleep(0)
        example.signal_handler(15, None)
        await asyncio.wait_for(task, 0.1)
        assert example.shutdown_requested.is_set()
        example.shutdown_requested = asyncio.Event()
        await example.wait_for_shutdown(0.001)

    asyncio.run(scenario())


def test_record_and_checkpoint_errors(example, caplog):
    async def scenario():
        processor = object.__new__(example.BatchProcessor)
        processor.s3_client = SimpleNamespace(
            put_object=Mock(side_effect=RuntimeError("storage failed"))
        )
        processor.checkpoint_key = "fixture"
        processor.save_checkpoint(1)
        assert caplog.records[-1].exc_info is not None
        processor.db_pool = SimpleNamespace(
            acquire=Mock(side_effect=RuntimeError("database failed"))
        )
        assert not await processor.process_record({"id": 1, "data": "fixture"})
        assert caplog.records[-1].message == "Failed to mark record as failed"
        assert caplog.records[-1].exc_info is not None

    asyncio.run(scenario())


def test_job_error_closes_pool_and_shutdown_exit(example, monkeypatch, caplog):
    async def scenario():
        processor = object.__new__(example.BatchProcessor)
        processor.initialize = AsyncMock()
        processor.cleanup = AsyncMock()
        processor.load_checkpoint = Mock(return_value=None)
        processor.fetch_batch = AsyncMock(side_effect=RuntimeError("fetch failed"))
        example.shutdown_requested = asyncio.Event()
        with pytest.raises(RuntimeError, match="fetch failed"):
            await processor.run()
        processor.cleanup.assert_awaited_once()
        assert caplog.records[-1].exc_info is not None
        example.shutdown_requested.set()
        monkeypatch.setattr(
            example, "push_to_gateway", Mock(side_effect=RuntimeError("metrics failed"))
        )
        with pytest.raises(SystemExit) as error:
            await processor.run()
        assert error.value.code == 130
        assert processor.cleanup.await_count == 2
        assert caplog.records[-1].exc_info is not None

    asyncio.run(scenario())


def test_checkpoint_does_not_advance_past_interrupted_batch(example, monkeypatch):
    async def scenario():
        example.shutdown_requested = asyncio.Event()
        processor = object.__new__(example.BatchProcessor)
        processor.initialize = AsyncMock()
        processor.cleanup = AsyncMock()
        processor.load_checkpoint = Mock(return_value=10)
        processor.save_checkpoint = Mock()
        processor.fetch_batch = AsyncMock(return_value=[{"id": 11}, {"id": 12}])

        async def interrupted(record):
            example.shutdown_requested.set()
            return True

        processor.process_record = interrupted
        monkeypatch.setattr(example, "push_to_gateway", Mock())
        with pytest.raises(SystemExit) as error:
            await processor.run()
        assert error.value.code == 130
        processor.save_checkpoint.assert_not_called()
        processor.cleanup.assert_awaited_once()

    asyncio.run(scenario())


@pytest.mark.parametrize("endpoint", ["http://collector:4317", "collector:4317", "https://"])
def test_rejects_insecure_transport(example, monkeypatch, endpoint):
    monkeypatch.setenv("PUSHGATEWAY_URL", endpoint)
    validate = example.pushgateway_endpoint
    with pytest.raises(RuntimeError, match="HTTPS"):
        validate()


def test_verified_transport_endpoint(example, monkeypatch):
    monkeypatch.setenv("PUSHGATEWAY_URL", "https://collector.example:4317")
    assert example.pushgateway_endpoint() == "https://collector.example:4317"
