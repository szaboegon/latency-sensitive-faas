from parliament import Context #type: ignore
import requests
from opentelemetry.propagate import inject, extract
import tracing
from opentelemetry import trace
from opentelemetry.context import Context as OtelContext
from config import FUNCTION_NAME, HANDLERS, read_config, Route
from typing import Any, Dict, Deque, Tuple, Optional, TypedDict
import threading
from collections import deque
from event import Event, extract_event, create_event

if "tracer" not in globals():
    tracer = tracing.instrument_app(FUNCTION_NAME)

    
class RouteToProcess(TypedDict):
    """
    Represents a route to be processed.
    """
    route: Route
    input: Event
    

def get_headers(
    component: str, span_context: Optional[OtelContext] = None
) -> Dict[str, str]:
    """
    Generates headers for the next component, including trace context.
    """
    headers = {"X-Forward-To": component, "Content-Type": "application/json"}
    if span_context:
        inject(headers, context=span_context)
    return headers


def handle_request(context: Context, component: str) -> Tuple[Any, OtelContext]:
    """
    Handles the request by invoking the appropriate handler and preparing
    the next component's details.
    """
    with tracer.start_as_current_span(
        component, context=extract(context.request.headers)
    ) as span:
        if component in HANDLERS:
            handler = HANDLERS[component]
            event_out = handler(context)
            return event_out, trace.set_span_in_context(span)
        return {}, trace.set_span_in_context(span)


def forward_request(
    route: Route, event_out: Any, span_context: Optional[OtelContext]
) -> None:
    """
    Asynchronously forwards the request to the next component if there is one.
    """
    if not route["component"]:
        return

    headers = get_headers(route["component"], span_context)

    def send_async_request():
        try:
            requests.post(url=route["url"], json=event_out, headers=headers)
        except Exception as e:
            print(f"Async forwarding failed: {e}")

    # Start the async thread
    threading.Thread(target=send_async_request, daemon=True).start()


def main(context: Context) -> Tuple[str, int]:
    """
    Entry point of the function. Processes the routing table using parallel processing.
    """
    component = context.request.headers.get("X-Forward-To")
    if not component:
        return f"No component found with name {component}", 400

    processing_queue: Deque[RouteToProcess] = deque(
        [RouteToProcess(route=Route(component=component, url="local"), input=extract_event(context))] 
    )
    output, span_context = None, None

    # Read the routing table from Redis
    routing_table = read_config()

    try:
        while processing_queue:
            current = processing_queue.popleft()
            component = current["route"]["component"]

            output, span_context = handle_request(current["input"], component)

            for next_route in routing_table.get(component, []):
                if next_route["url"] == "local":
                    event_in = create_event(output)
                    route_to_process = RouteToProcess(route=next_route, input=event_in)
                    processing_queue.append(route_to_process)
                else:
                    forward_request(next_route, output, span_context)
    except KeyError as e:
        return f"Invalid routing table entry: {e} is missing", 400

    return "ok", 200
