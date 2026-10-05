import asyncio
from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest


def test_shutdown_wakes_backoff(example):
    async def scenario():
        example.shutdown_requested = asyncio.Event()
        task = asyncio.create_task(example.wait_for_shutdown(60))
        await asyncio.sleep(0)
        example.signal_handler(15, None)
        await asyncio.wait_for(task, timeout=0.1)
        assert example.shutdown_requested.is_set()
        example.shutdown_requested = asyncio.Event()
        await example.wait_for_shutdown(0.001)
        assert not example.shutdown_requested.is_set()

    asyncio.run(scenario())


def test_failed_messages_retry_and_invalid_json_is_discarded(example, caplog):
    async def scenario():
        handler = example.EventHandler()
        handler.handle_order_created = AsyncMock(side_effect=RuntimeError("processing failed"))
        assert not await handler.handle_event({"type": "order.created"})
        assert caplog.records[-1].exc_info is not None
        msg = SimpleNamespace(data=b'{"type":"order.created"}', ack=AsyncMock(), nak=AsyncMock())
        await example.message_callback(msg, handler)
        msg.nak.assert_awaited_once_with(delay=5)
        msg.ack.assert_not_awaited()
        msg.data = b"not json"
        await example.message_callback(msg, handler)
        msg.ack.assert_awaited_once()
        msg.data = b"[]"
        await example.message_callback(msg, handler)
        assert msg.nak.call_args.kwargs == {"delay": 10}

    asyncio.run(scenario())


def test_subscriber_drains_after_failures(example, monkeypatch):
    async def scenario():
        example.shutdown_requested = asyncio.Event()
        psub = SimpleNamespace(fetch=AsyncMock(side_effect=[RuntimeError("fetch failed")]))
        js = SimpleNamespace(pull_subscribe=AsyncMock(return_value=psub))
        nc = SimpleNamespace(jetstream=lambda: js, drain=AsyncMock())
        monkeypatch.setattr(example.nats, "connect", AsyncMock(return_value=nc))
        monkeypatch.setattr(example, "start_http_server", lambda port: None)

        async def stop(timeout):
            example.shutdown_requested.set()

        monkeypatch.setattr(example, "wait_for_shutdown", stop)
        await example.main()
        nc.drain.assert_awaited_once()
        example.shutdown_requested = asyncio.Event()
        js.pull_subscribe.side_effect = RuntimeError("subscription failed")
        with pytest.raises(SystemExit) as error:
            await example.main()
        assert error.value.code == 1
        assert nc.drain.await_count == 2

    asyncio.run(scenario())


def test_empty_fetch_backoff_and_successful_ack(example, monkeypatch):
    async def scenario():
        handler = example.EventHandler()
        handler.handle_order_created = AsyncMock()
        assert await handler.handle_event({"type": "order.created"})
        message = SimpleNamespace(
            data=b'{"type":"order.created"}', ack=AsyncMock(), nak=AsyncMock()
        )
        await example.message_callback(message, handler)
        message.ack.assert_awaited_once()
        message.nak.assert_not_awaited()
        example.shutdown_requested = asyncio.Event()
        psub = SimpleNamespace(fetch=AsyncMock(return_value=[]))
        nc = SimpleNamespace(
            jetstream=lambda: SimpleNamespace(pull_subscribe=AsyncMock(return_value=psub)),
            drain=AsyncMock(),
        )
        monkeypatch.setattr(example.nats, "connect", AsyncMock(return_value=nc))
        monkeypatch.setattr(example, "start_http_server", lambda port: None)

        async def stop(timeout):
            example.shutdown_requested.set()

        monkeypatch.setattr(example, "wait_for_shutdown", stop)
        await example.main()
        assert psub.fetch.await_count == 1
        nc.drain.assert_awaited_once()

    asyncio.run(scenario())
