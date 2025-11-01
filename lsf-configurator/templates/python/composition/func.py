from parliament import Context  # type: ignore
import requests
from opentelemetry.propagate import inject, extract
import tracing
from opentelemetry import trace
from opentelemetry.context import Context as OtelContext
from config import APP_NAME, FUNCTION_NAME, read_config
from route import Route
from typing import Any, Dict, Deque, List, Tuple, Optional, TypedDict, Union
import threading
from collections import deque
from event import Event, extract_event, create_event
from opentelemetry.trace.status import Status, StatusCode  # Add this import
from logger import setup_logging
from results import write_result
import faulthandler
import sys

faulthandler.enable(file=sys.stderr, all_threads=True)


logger = setup_logging(__name__)
tracer = tracing.instrument_app(app_name=APP_NAME, service_name=FUNCTION_NAME)


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


# Handler can return event out as a dictionary, or bytes, or even as a list of dicts/bytes
# A list signals multiple outgoing invocations
HandlerReturnType = Union[Dict[str, Any], bytes, List[Dict[str, Any]], List[bytes]]


def _handler_worker(component: str, event: Event) -> Any:
    """
    Worker function executed in a separate process.
    Dynamically imports func.HANDLERS to avoid stale global state.
    """
    from config import HANDLERS

    try:
        if component not in HANDLERS:
            raise ValueError(f"No handler registered for component '{component}'")

        handler = HANDLERS[component]
        return handler(event)
    except Exception as e:
        import traceback

        traceback.print_exc()
        raise e


def handle_request(
    event: Event, component: str, parent_context: Optional[OtelContext]
) -> Tuple[HandlerReturnType, Optional[OtelContext]]:
    """
    Handles the request by invoking the appropriate handler in a worker process.
    The handler is executed in an isolated process to prevent global tracer races.
    """
    with tracer.start_as_current_span(component, context=parent_context) as span:
        try:
            logger.info(f"Starting handler for component '{component}'.")

            # Submit the job to a worker process
            event_out = _handler_worker(component, event)

            logger.info(f"Component '{component}' processed successfully.")
            span.set_status(Status(StatusCode.OK))
            return event_out, trace.set_span_in_context(span)

        except Exception as e:
            span.set_status(Status(StatusCode.ERROR, str(e)))
            span.record_exception(e)
            logger.error(f"Error processing component '{component}': {e}")
            raise e


forward_threads: List[threading.Thread] = []


def forward_request(
    route: Route, event_out: Any, span_context: Optional[OtelContext]
) -> None:
    """
    Asynchronously forwards the request to the next component if there is one.
    """
    if not route["component"]:
        logger.warning("Attempted to forward request with empty component.")
        return

    headers = get_headers(route["component"], span_context)

    def send_async_request() -> None:
        with tracer.start_span("forward_request", context=span_context) as span:
            try:
                requests.post(url=route["url"], json=event_out, headers=headers)
                logger.info(
                    f"Request forwarded to component '{route['component']}' at URL '{route['url']}'."
                )
                span.set_status(Status(StatusCode.OK))
            except Exception as e:
                span.set_status(Status(StatusCode.ERROR, str(e)))
                span.record_exception(e)
                logger.error(
                    f"Error forwarding request to '{route['component']}' at '{route['url']}': {e}"
                )
                raise e

    t = threading.Thread(target=send_async_request)
    t.start()
    forward_threads.append(t)


def main(context: Context) -> Tuple[str, int]:
    """
    Entry point of the function. Processes the routing table using parallel processing.
    """
    global forward_threads
    forward_threads = []

    logger.info("Headers received: " + str(context.request.headers))

    component = context.request.headers.get("X-Forward-To")
    if not component:
        logger.error(f"No component found with name {component}")
        return f"No component found with name {component}", 400

    processing_queue: Deque[RouteToProcess] = deque(
        [
            RouteToProcess(
                route=Route(component=component, url="local"),
                input=extract_event(context),
            )
        ]
    )
    span_context: Optional[OtelContext] = None
    output, span_context = None, extract(context.request.headers)

    # Read routing table from redis
    with tracer.start_as_current_span("read_config", context=span_context) as span:
        try:
            routing_table = read_config()
            span_context = trace.set_span_in_context(span)
            logger.info("Routing table read from Redis successfully.")
        except Exception as e:
            span.set_status(Status(StatusCode.ERROR, str(e)))
            span.record_exception(e)
            logger.error(f"Error reading routing table from Redis: {e}")
            return "Error: Routing table could not be read from Redis", 500

    try:
        while processing_queue:
            current = processing_queue.popleft()
            component = current["route"]["component"]

            try:
                output, span_context = handle_request(
                    current["input"], component, span_context
                )
            except Exception:
                logger.error(f"Error in component '{component}', aborting workflow.")
                return f"Error processing component '{component}'", 500

            outputs = output if isinstance(output, list) else [output]
            for o in outputs:
                next_routes = routing_table.get(component, [])
                if not next_routes:
                    # No next components, write result to Redis
                    with tracer.start_as_current_span(
                        "write_result", context=span_context
                    ) as span:
                        try:
                            write_result(o)
                            logger.info(
                                f"Result for component '{component}' written to Redis."
                            )
                            span.set_status(Status(StatusCode.OK))
                        except Exception as e:
                            span.set_status(Status(StatusCode.ERROR, str(e)))
                            span.record_exception(e)
                            logger.error(
                                f"Error writing result for component '{component}': {e}"
                            )
                            return (
                                f"Error writing result for component '{component}'",
                                500,
                            )
                for next_route in next_routes:
                    if next_route["url"] == "local":
                        event_in = create_event(o)
                        route_to_process = RouteToProcess(
                            route=next_route, input=event_in
                        )
                        processing_queue.append(route_to_process)
                    else:
                        forward_request(next_route, o, span_context)
    except KeyError as e:
        logger.error(f"Invalid routing table entry: {e} is missing")
        return f"Invalid routing table entry: {e} is missing", 400

    for t in forward_threads:
        t.join()

    logger.info("All components processed successfully.")
    return "ok", 200
