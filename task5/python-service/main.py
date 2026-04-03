from typing import Any

import httpx
from fastapi import FastAPI, HTTPException
from fastapi.testclient import TestClient
from pydantic import BaseModel, Field
from unittest.mock import AsyncMock, MagicMock, patch


GO_SERVICE_URL = "http://127.0.0.1:8080/process"

app = FastAPI(title="Task5 FastAPI client", version="1.0.0")


class Geo(BaseModel):
    lat: float
    lon: float


class Address(BaseModel):
    city: str
    street: str
    geo: Geo


class Contact(BaseModel):
    type: str
    value: str


class Item(BaseModel):
    sku: str
    name: str
    qty: int = Field(gt=0)
    price: float = Field(gt=0)
    tags: list[str] = []
    metadata: Any | None = None


class Payment(BaseModel):
    method: str
    currency: str
    paid: bool = False
    extra: dict[str, Any] = {}


class OrderRequest(BaseModel):
    request_id: str
    user_id: int = Field(gt=0)
    name: str
    address: Address
    contacts: list[Contact] = Field(min_length=1)
    items: list[Item] = Field(min_length=1)
    payment: Payment
    flags: dict[str, bool] = {}
    metadata: dict[str, str] = {}


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "up"}


@app.post("/forward")
async def forward_to_go(payload: OrderRequest) -> dict[str, Any]:
    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            response = await client.post(GO_SERVICE_URL, json=payload.model_dump())
    except httpx.RequestError as exc:
        raise HTTPException(status_code=502, detail=f"Go service unavailable: {exc}") from exc

    if response.status_code >= 400:
        raise HTTPException(
            status_code=response.status_code,
            detail={"go_error": response.json()},
        )

    return {
        "status": "ok",
        "processed_by": "python-service",
        "go_response": response.json(),
    }


def _must(ok: bool, message: str) -> None:
    if not ok:
        raise RuntimeError(message)


def run_self_tests() -> None:
    client = TestClient(app)

    health = client.get("/health")
    _must(health.status_code == 200, "health status must be 200")
    _must(health.json() == {"status": "up"}, "health body mismatch")

    payload = {
        "request_id": "REQ-SELFTEST",
        "user_id": 42,
        "name": "Ivan Petrov",
        "address": {
            "city": "Moscow",
            "street": "Tverskaya 1",
            "geo": {"lat": 55.7558, "lon": 37.6176},
        },
        "contacts": [{"type": "email", "value": "ivan@example.com"}],
        "items": [{"sku": "BK-001", "name": "Book", "qty": 2, "price": 500.0}],
        "payment": {"method": "card", "currency": "RUB", "paid": True, "extra": {}},
        "flags": {"express": True},
        "metadata": {"source": "lab10"},
    }

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.json = MagicMock(
        return_value={
            "status": "ok",
            "processed_by": "go-service",
            "request_id": "REQ-SELFTEST",
            "total_qty": 2,
            "total_price": 1000.0,
        }
    )
    instance = MagicMock()
    instance.__aenter__ = AsyncMock(return_value=instance)
    instance.__aexit__ = AsyncMock(return_value=False)
    instance.post = AsyncMock(return_value=mock_response)

    with patch("httpx.AsyncClient", return_value=instance):
        ok_response = client.post("/forward", json=payload)

    _must(ok_response.status_code == 200, "forward success status must be 200")
    ok_json = ok_response.json()
    _must(ok_json["status"] == "ok", "forward success body status mismatch")
    _must(ok_json["processed_by"] == "python-service", "processed_by mismatch")

    err_response_obj = MagicMock()
    err_response_obj.status_code = 400
    err_response_obj.json = MagicMock(return_value={"status": "error", "error": "bad payload"})
    instance_err = MagicMock()
    instance_err.__aenter__ = AsyncMock(return_value=instance_err)
    instance_err.__aexit__ = AsyncMock(return_value=False)
    instance_err.post = AsyncMock(return_value=err_response_obj)
    with patch("httpx.AsyncClient", return_value=instance_err):
        bad_response = client.post("/forward", json=payload)

    _must(bad_response.status_code == 400, "forward error must propagate 400")

    invalid = client.post("/forward", json={"not_a_order": True})
    _must(invalid.status_code == 422, "invalid request must return 422")

    print("Python self-tests passed")


if __name__ == "__main__":
    run_self_tests()
