from redis import Redis
import os
import json
from datetime import datetime, timezone
from config import APP_NAME
from typing import Any


RESULT_STORE_ADDRESS = os.environ["RESULT_STORE_ADDRESS"]
redis_client = Redis(host=RESULT_STORE_ADDRESS, port=6379)


def write_result(event: Any) -> None:
    key = f"result:{APP_NAME}"
    value = {
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "event": event,
    }
    redis_client.rpush(key, json.dumps(value))
