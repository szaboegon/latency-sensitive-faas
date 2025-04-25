from redis import Redis
import os
import json
from typing import Dict, List, TypedDict, cast, Any, Callable
from parliament import Context #type: ignore

REDIS_URL = os.environ["NODE_IP"]
redis_client = Redis(host=REDIS_URL, port=6379)

FUNCTION_NAME = ""
HANDLERS: Dict[str, Callable[[Context], Any]] = {
    # REGISTER COMPONENT HANDLERS HERE
}

class Route(TypedDict):
    """
    Represents a route in the routing table.
    """
    component: str
    url: str
    
RoutingTable = Dict[str, List[Route]]

def read_config() -> RoutingTable:
    """
    Reads configuration keys from Redis and populates HANDLERS dictionary.
    """
    config: Any = redis_client.get(FUNCTION_NAME)
    if not config:
        raise ValueError(f"Configuration for {FUNCTION_NAME} not found in Redis")
    
    configJson = json.loads(config)
    return cast(RoutingTable, configJson)
        