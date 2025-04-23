from redis import Redis
import os
import json
import dataclasses
from typing import Dict, List

REDIS_URL = os.environ["NODE_IP"]
redis_client = Redis(host=REDIS_URL, port=6379)

FUNCTION_NAME = ""
HANDLERS = {
    # REGISTER COMPONENT HANDLERS HERE
}

@dataclasses.dataclass
class Route:
    """
    Represents a route in the routing table.
    """
    component: str
    url: str

def read_config() -> Dict[str, List[Route]]:
    """
    Reads configuration keys from Redis and populates HANDLERS dictionary.
    """
    config = redis_client.get(FUNCTION_NAME)
    if not config:
        raise ValueError(f"Configuration for {FUNCTION_NAME} not found in Redis")
    
    routing_table = json.loads(config)
    return routing_table
        