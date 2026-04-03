import asyncio
import os
import signal
from contextlib import suppress

from fastapi import FastAPI, Query
import uvicorn


app = FastAPI(title="Task7 Python Service", version="1.0.0")


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "up", "service": "python"}


@app.get("/work")
async def work(seconds: int = Query(default=10, ge=0, le=120)) -> dict[str, int | str]:
    # Simulate long-running request that should finish during graceful shutdown.
    await asyncio.sleep(seconds)
    return {"status": "done", "service": "python", "work_seconds": seconds}


@app.on_event("shutdown")
async def on_shutdown() -> None:
    # Hook to show the shutdown path was executed.
    print("python-service shutdown hook finished")


async def run() -> None:
    config = uvicorn.Config(
        app,
        host="127.0.0.1",
        port=8001,
        log_level="info",
        timeout_graceful_shutdown=15,
    )
    server = uvicorn.Server(config)
    loop = asyncio.get_running_loop()

    def handle_stop_signal() -> None:
        server.should_exit = True

    for sig in (signal.SIGINT, signal.SIGTERM):
        with suppress(NotImplementedError):
            loop.add_signal_handler(sig, handle_stop_signal)

    await server.serve()


def _must(ok: bool, message: str) -> None:
    if not ok:
        raise RuntimeError(message)


async def run_self_tests() -> None:
    health_data = await health()
    _must(health_data["status"] == "up", "health status must be up")
    _must(health_data["service"] == "python", "health service must be python")

    work_data = await work(0)
    _must(work_data["status"] == "done", "work status must be done")
    _must(work_data["service"] == "python", "work service must be python")
    _must(work_data["work_seconds"] == 0, "work seconds must be 0")

    print("Python task7 self-tests passed")


if __name__ == "__main__":
    if os.getenv("RUN_SELF_TESTS") == "1":
        asyncio.run(run_self_tests())
    else:
        asyncio.run(run())
