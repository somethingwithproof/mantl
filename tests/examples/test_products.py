import asyncio
from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest


def test_failure_and_status_contracts(example, caplog):
    async def scenario():
        example.db_pool = None
        with pytest.raises(example.HTTPException) as error:
            await example.ready()
        assert error.value.status_code == 503
        calls = [
            (
                "list_products",
                (),
                dict(category=None, min_price=None, max_price=None, limit=10, offset=0),
            ),
            ("get_product", (1,), {}),
            (
                "create_product",
                (
                    example.ProductCreate(
                        name="fixture", price="1.00", category="fixture", stock=1, sku="fixture"
                    ),
                ),
                {},
            ),
            ("update_product", (1, example.ProductUpdate(name="fixture")), {}),
            ("delete_product", (1,), {}),
            ("list_categories", (), {}),
        ]
        db = SimpleNamespace(
            fetch=AsyncMock(side_effect=RuntimeError("database failed")),
            fetchrow=AsyncMock(side_effect=RuntimeError("database failed")),
            execute=AsyncMock(side_effect=RuntimeError("database failed")),
        )
        for name, args, kwargs in calls:
            handler = getattr(example, name)
            with pytest.raises(example.HTTPException) as error:
                await handler(*args, db=db, **kwargs)
            assert error.value.status_code == 500
            assert caplog.records[-1].exc_info is not None
        db.fetchrow.side_effect = None
        db.fetchrow.return_value = None
        for name, args in [
            ("get_product", (1,)),
            ("update_product", (1, example.ProductUpdate(name="fixture"))),
        ]:
            handler = getattr(example, name)
            with pytest.raises(example.HTTPException) as error:
                await handler(*args, db=db)
            assert error.value.status_code == 404
        db.execute.side_effect = None
        db.execute.return_value = "DELETE 0"
        with pytest.raises(example.HTTPException) as error:
            await example.delete_product(1, db=db)
        assert error.value.status_code == 404
        empty_update = example.ProductUpdate()
        with pytest.raises(example.HTTPException) as error:
            await example.update_product(1, empty_update, db=db)
        assert error.value.status_code == 400
        db.fetchrow.side_effect = example.asyncpg.UniqueViolationError("duplicate")
        with pytest.raises(example.HTTPException) as error:
            await example.create_product(calls[2][1][0], db=db)
        assert error.value.status_code == 409

    asyncio.run(scenario())


def test_safe_log_and_readiness_failure(example, caplog):
    async def scenario():
        product = example.ProductCreate(
            name="fixture", price="1.00", category="fixture", stock=1, sku="sku\r\nforged"
        )
        db = SimpleNamespace(fetchrow=AsyncMock(return_value={"id": 1}))
        caplog.set_level("INFO")
        assert await example.create_product(product, db=db) == {"id": 1}
        record = next(r for r in caplog.records if r.message.startswith("Created product:"))
        assert "\n" not in record.message
        assert "skuforged" in record.message

        class UnavailablePool:
            def acquire(self):
                raise RuntimeError("database unavailable")

        example.db_pool = UnavailablePool()
        with pytest.raises(example.HTTPException) as error:
            await example.ready()
        assert error.value.status_code == 503

    asyncio.run(scenario())


def test_openapi_errors(example):
    paths = example.app.openapi()["paths"]
    assert "503" in paths["/ready"]["get"]["responses"]
    expected = {"get": {"500"}, "post": {"409", "500"}}
    for method, codes in expected.items():
        assert codes <= paths["/api/v1/products"][method]["responses"].keys()
    for method in ("get", "put", "delete"):
        assert {"404", "500"} <= paths["/api/v1/products/{product_id}"][method]["responses"].keys()
    assert "400" in paths["/api/v1/products/{product_id}"]["put"]["responses"]


def test_parameterized_filters_and_successful_updates(example):
    async def scenario():
        db = SimpleNamespace(
            fetch=AsyncMock(return_value=[{"id": 1}]),
            fetchrow=AsyncMock(return_value={"id": 1}),
            execute=AsyncMock(return_value="DELETE 1"),
        )
        result = await example.list_products(
            category="fixture", min_price=1.0, max_price=2.0, limit=10, offset=3, db=db
        )
        assert result == [{"id": 1}]
        query, *args = db.fetch.call_args.args
        assert "category = $1" in query
        assert "price >= $2" in query
        assert "price <= $3" in query
        assert args == ["fixture", 1.0, 2.0, 10, 3]
        assert await example.get_product(1, db=db) == {"id": 1}
        assert await example.update_product(
            1, example.ProductUpdate(name="changed", description=None), db=db
        ) == {"id": 1}
        assert await example.delete_product(1, db=db) is None
        db.fetch.return_value = [{"category": "fixture"}]
        assert await example.list_categories(db=db) == ["fixture"]

        class Connection:
            async def __aenter__(self):
                return SimpleNamespace(fetchval=AsyncMock(return_value=1))

            async def __aexit__(self, *args):
                return None

        example.db_pool = SimpleNamespace(acquire=Connection)
        assert (await example.ready())["status"] == "ready"

    asyncio.run(scenario())
