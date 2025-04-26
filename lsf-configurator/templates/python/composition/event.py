from typing import Optional, NamedTuple
from parliament import Context  # type: ignore

class Event(NamedTuple):
    """
    Represents an event to pass to components.
    """
    json: Optional[str]
    data: Optional[bytes]
      
def extract_event(context: Context) -> Event:
    """
    Extracts the event from the context.
    """
    json_data = context.request.json if context.request.json else None
    byte_data = context.request.data if context.request.data else None

    return Event(json=json_data, data=byte_data)

def create_event(data: bytes | dict) -> Event:
    """
    Creates an event from the context.
    """
    return Event(
        json=data if isinstance(data, dict) else None,
        data=data if isinstance(data, bytes) else None,
    )
                    