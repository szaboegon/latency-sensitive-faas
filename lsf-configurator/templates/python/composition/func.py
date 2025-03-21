from parliament import Context
import os
import requests
from opentelemetry.propagate import inject, extract
import tracing
from opentelemetry import trace

LOADBALANCER_URL = f'http://{os.environ["NODE_IP"]}:8080'
FUNCTION_NAME = ""
HANDLERS = {
    # REGISTER COMPONENT HANDLERS HERE
}

if 'tracer' not in globals():
    tracer = tracing.instrument_app(FUNCTION_NAME)
    
def get_headers(component: str, span_context: Context | None = None):
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

def handle_request(context: Context, next_component: str):
    """
    Handles the request by invoking the appropriate handler and preparing
    the next component's details.
    """
    with tracer.start_as_current_span(next_component, context=extract(context.request.headers)) as span:
        if next_component in HANDLERS:
            next_component, event_out = HANDLERS[next_component](context)
            return next_component, event_out, trace.set_span_in_context(span)
        return None, {}, trace.set_span_in_context(span)

def forward_request(next_component: str, event_out: any, span_context: Context | None):
    """
    Forwards the request to the load balancer if there is a next component.
    """
    if not next_component:
        return "ok", 200

    headers = get_headers(next_component, span_context)
    response = requests.post(LOADBALANCER_URL, json=event_out, headers=headers)
    return response.text, 200

def main(context: Context):
    """
    Entry point of the function. Manages parent context and orchestrates the pipeline.
    """
    next_component = context.request.headers.get("X-Forward-To")
    if not next_component:
        return "", {}, None
    
    next_component, event_out, span_context = handle_request(context, next_component)
    while next_component and next_component in HANDLERS:
        handle_request(context, next_component)

    return forward_request(next_component, event_out, span_context)

    
