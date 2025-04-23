from parliament import Context
import requests
from opentelemetry.propagate import inject, extract
import tracing
from opentelemetry import trace, Context as OtelContext
from config import FUNCTION_NAME, HANDLERS, read_config, Route
from typing import Any, Dict, List, Tuple, Optional

import threading
from concurrent.futures import ThreadPoolExecutor

if 'tracer' not in globals():
    tracer = tracing.instrument_app(FUNCTION_NAME)
    
def get_headers(component: str, span_context: Optional[OtelContext] = None) -> Dict[str, str]:
    """
    Generates headers for the next component, including trace context.
    """
    headers = {
        "X-Forward-To": component,
        "Content-Type": "application/json"
    }
    if span_context:
        inject(headers, context=span_context)
    return headers

def handle_request(context: Context, component: str) -> Tuple[Any, OtelContext]:
    """
    Handles the request by invoking the appropriate handler and preparing
    the next component's details.
    """
    with tracer.start_as_current_span(component, context=extract(context.request.headers)) as span:
        if component in HANDLERS:
            event_out = HANDLERS[component](context)
            return  event_out, trace.set_span_in_context(span)
        return {}, trace.set_span_in_context(span)

def forward_request(route: Route, event_out: Any, span_context: Optional[OtelContext]) -> None:
    """
    Asynchronously forwards the request to the next component if there is one.
    """
    if not route.component:
        return

    headers = get_headers(route.component, span_context)

    def send_async_request():
        try:
            requests.post(url=route.url, json=event_out, headers=headers)
        except Exception as e:
            print(f"Async forwarding failed: {e}")

    # Start the async thread
    threading.Thread(target=send_async_request, daemon=True).start()

def process_local_route(
    route: Route,
    context: Context,
    routing_table: Dict[str, List[Route]],
    processing_queue: List[Route],
    send_queue: List[Route]
) -> Tuple[Context, Any, OtelContext]:
    """
    Processes a local route and updates the queues for further processing.
    """
    next_component = route.component
    if next_component in HANDLERS:
        event_out, span_context = handle_request(context, next_component)
        context = event_out

        for next_route in routing_table[next_component]:
            if next_route.url == "local":
                processing_queue.append(next_route)
            else:
                send_queue.append(next_route)
    else:
        raise ValueError(f"Component {next_component} not found in HANDLERS")
    return context, event_out, span_context

def main(context: Context) -> Tuple[str, int]:
    """
    Entry point of the function. Processes the routing table using parallel processing.
    """
    component = context.request.headers.get("X-Forward-To")
    if not component:
        return f"No component found with name {component}", 400

    processing_queue = []
    send_queue = []
    
    # Read the routing table from Redis
    routing_table = read_config()
    for next_route in routing_table[component]:
        if next_route.url == "local":
            processing_queue.append(next_route)
        else:
            send_queue.append(next_route)

    event_out, span_context = None, None

    with ThreadPoolExecutor() as executor:
        while processing_queue:
            futures = []
            for route in processing_queue:
                futures.append(executor.submit(process_local_route, route, context, routing_table, processing_queue, send_queue))
            processing_queue.clear()

            for future in futures:
                context, event_out, span_context = future.result()

        # Process the send queue in parallel
        futures = [executor.submit(forward_request, route, event_out, span_context) for route in send_queue]
        for future in futures:
            future.result()

    return "ok", 200


